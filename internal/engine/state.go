package engine

import (
	"time"

	"github.com/realjv3/momo/internal/market"
	"github.com/realjv3/momo/internal/session"
)

const maxRecent = 8

type Meta struct {
	Float     int64
	PrevClose float64
}

type backfillState int

const (
	bfNone backfillState = iota
	bfPending
	bfDone
)

type state struct {
	meta         Meta
	date         string
	sess         session.Session
	ema          float64
	n            int
	sessVol      float64
	open         float64
	regularClose float64
	recent       []market.Bar
	lastSignal   time.Time
	bf           backfillState
	pending      []market.Bar
	retryAt      time.Time
}

// reset clears all derived state but keeps meta and backfill status.
func (s *state) reset(date string) {
	*s = state{meta: s.meta, bf: s.bf, retryAt: s.retryAt, date: date}
}

// apply folds a bar into session state.
func (s *state) apply(b market.Bar, period int) {
	t := b.StartTime()
	sess := session.At(t)

	if d := session.Date(t); d != s.date {
		s.reset(d)
	}

	if sess != s.sess {
		s.sess = sess
		s.ema, s.n, s.sessVol, s.recent = 0, 0, 0, nil
	}

	if sess == session.Closed {
		return
	}

	if s.n == 0 {
		s.ema = b.C
	} else {
		s.ema += 2 / float64(period+1) * (b.C - s.ema)
	}

	s.n++
	s.sessVol += b.V

	if b.OP > 0 {
		s.open = b.OP
	}

	if sess == session.Regular {
		s.regularClose = b.C
	}

	s.recent = append(s.recent, b)
	if len(s.recent) > maxRecent {
		s.recent = s.recent[len(s.recent)-maxRecent:]
	}
}

// baseline is the prior close: yesterday's for pre/regular, today's regular close for post.
func (s *state) baseline() float64 {
	if s.sess == session.Post && s.regularClose > 0 {
		return s.regularClose
	}

	return s.meta.PrevClose
}

func (s *state) changePct(price float64) float64 {
	base := s.baseline()
	if base <= 0 {
		return 0
	}

	return (price - base) / base * 100
}

// check returns a signal when bar b shows upward momentum under filters f.
func (s *state) check(b market.Bar, f Filters) (Signal, bool) {
	chg := s.changePct(b.C)

	if s.baseline() <= 0 || s.n < f.EMAPeriod || len(s.recent) < f.RisingBars+1 {
		return Signal{}, false
	}

	if !f.candidate(s.sess, b.C, chg, s.meta.Float) || b.C <= s.ema {
		return Signal{}, false
	}

	if s.sessVol < f.MinSessionVolume || b.V < f.MinBarVolume {
		return Signal{}, false
	}

	if f.MinChangeFromOpenPct > 0 && (s.open <= 0 || (b.C-s.open)/s.open*100 < f.MinChangeFromOpenPct) {
		return Signal{}, false
	}

	if !s.rising(f.RisingBars) {
		return Signal{}, false
	}

	prev := s.recent[len(s.recent)-2]
	if prev.C <= 0 || (b.C-prev.C)/prev.C*100 < f.MinBarChangePct {
		return Signal{}, false
	}

	if b.V < prev.V*f.VolRiseFactor {
		return Signal{}, false
	}

	if !s.lastSignal.IsZero() && b.EndTime().Sub(s.lastSignal) < time.Duration(f.CooldownMinutes)*time.Minute {
		return Signal{}, false
	}

	s.lastSignal = b.EndTime()

	return Signal{
		Time:      b.EndTime(),
		Symbol:    b.Sym,
		Session:   string(s.sess),
		ChangePct: chg,
		Price:     b.C,
		Float:     s.meta.Float,
		Volume:    s.sessVol,
	}, true
}

// rising reports whether the last n bars are consecutive minutes with each close above the prior close.
func (s *state) rising(n int) bool {
	for i := 0; i < n; i++ {
		cur, prev := s.recent[len(s.recent)-1-i], s.recent[len(s.recent)-2-i]
		if cur.C <= prev.C || cur.Start-prev.Start != 60_000 {
			return false
		}
	}

	return true
}

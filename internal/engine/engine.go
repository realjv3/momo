package engine

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/realjv3/momo/internal/market"
	"github.com/realjv3/momo/internal/session"
)

const retryDelay = 2 * time.Minute

// Signal is one row in the scanner table.
type Signal struct {
	ID        int64     `json:"id"`
	Time      time.Time `json:"time"`
	Symbol    string    `json:"symbol"`
	Session   string    `json:"session"`
	ChangePct float64   `json:"changePct"`
	Price     float64   `json:"price"`
	Float     int64     `json:"float"`
	Volume    float64   `json:"volume"`
}

// Backfiller returns minute bars for sym with start in [from, to].
type Backfiller func(ctx context.Context, sym string, from, to time.Time) ([]market.Bar, error)

type backfillResult struct {
	sym  string
	bars []market.Bar
	err  error
}

type Engine struct {
	filters  atomic.Pointer[Filters]
	syms     map[string]*state
	in       chan market.Bar
	done     chan backfillResult
	out      chan<- Signal
	backfill Backfiller
}

// New builds an engine over the symbols in meta; a nil backfill disables lazy history loading.
func New(meta map[string]Meta, f Filters, backfill Backfiller, out chan<- Signal) *Engine {
	e := &Engine{
		syms:     make(map[string]*state, len(meta)),
		in:       make(chan market.Bar, 8192),
		done:     make(chan backfillResult, 64),
		out:      out,
		backfill: backfill,
	}

	for sym, m := range meta {
		st := &state{meta: m}
		if backfill == nil {
			st.bf = bfDone
		}

		e.syms[sym] = st
	}

	e.SetFilters(f)

	return e
}

func (e *Engine) Input() chan<- market.Bar { return e.in }

func (e *Engine) Tracked() int { return len(e.syms) }

func (e *Engine) Filters() Filters { return *e.filters.Load() }

func (e *Engine) SetFilters(f Filters) {
	f = f.Normalize()
	e.filters.Store(&f)
}

func (e *Engine) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case b := <-e.in:
			e.Handle(ctx, b)
		case r := <-e.done:
			e.onBackfill(ctx, r)
		}
	}
}

// Handle processes one bar; it is not safe for concurrent use.
func (e *Engine) Handle(ctx context.Context, b market.Bar) {
	st := e.syms[b.Sym]
	if st == nil {
		return
	}

	if st.bf == bfPending {
		st.pending = append(st.pending, b)

		return
	}

	f := e.Filters()
	st.apply(b, f.EMAPeriod)

	if st.bf == bfNone && !b.EndTime().Before(st.retryAt) && f.candidate(st.sess, b.C, st.changePct(b.C), st.meta.Float) {
		st.bf = bfPending
		st.pending = []market.Bar{b}

		go e.fetch(ctx, b)

		return
	}

	e.evaluate(ctx, st, b, f)
}

func (e *Engine) fetch(ctx context.Context, b market.Bar) {
	bars, err := e.backfill(ctx, b.Sym, session.PreOpen(b.StartTime()), b.StartTime().Add(-time.Millisecond))

	select {
	case e.done <- backfillResult{sym: b.Sym, bars: bars, err: err}:
	case <-ctx.Done():
	}
}

func (e *Engine) onBackfill(ctx context.Context, r backfillResult) {
	st := e.syms[r.sym]
	f := e.Filters()

	if r.err != nil {
		log.Printf("backfill %s: %v", r.sym, r.err)
	}

	pending := st.pending
	first := pending[0].Start

	st.reset("")

	for _, b := range r.bars {
		if b.Start < first {
			st.apply(b, f.EMAPeriod)
		}
	}

	st.bf = bfDone
	if r.err != nil {
		st.bf = bfNone
		st.retryAt = pending[len(pending)-1].EndTime().Add(retryDelay)
	}

	for _, b := range pending {
		st.apply(b, f.EMAPeriod)
		e.evaluate(ctx, st, b, f)
	}
}

func (e *Engine) evaluate(ctx context.Context, st *state, b market.Bar, f Filters) {
	sig, ok := st.check(b, f)
	if !ok {
		return
	}

	select {
	case e.out <- sig:
	case <-ctx.Done():
	}
}

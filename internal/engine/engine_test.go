package engine

import (
	"context"
	"testing"
	"time"

	"github.com/realjv3/momo/internal/market"
	"github.com/realjv3/momo/internal/session"
)

// mon is a Monday in regular hours (10:00 ET).
var mon = time.Date(2026, 11, 2, 10, 0, 0, 0, session.ET)

func bar(i int, c, v float64) market.Bar {
	start := mon.Add(time.Duration(i) * time.Minute)

	return market.Bar{Sym: "X", C: c, V: v, Start: start.UnixMilli(), End: start.Add(time.Minute).UnixMilli()}
}

func run(t *testing.T, meta Meta, f Filters, bars []market.Bar) []Signal {
	t.Helper()

	out := make(chan Signal, 64)
	e := New(map[string]Meta{"X": meta}, f, nil, out)

	for _, b := range bars {
		e.Handle(context.Background(), b)
	}

	close(out)

	var sigs []Signal
	for s := range out {
		sigs = append(sigs, s)
	}

	return sigs
}

// ramp is 12 flat bars followed by a rising price with rising volume.
func ramp() []market.Bar {
	var bars []market.Bar
	for i := 0; i < 12; i++ {
		bars = append(bars, bar(i, 2.00, 10_000))
	}

	return append(bars, bar(12, 2.10, 20_000), bar(13, 2.25, 40_000), bar(14, 2.45, 80_000))
}

func TestSignalOnUpwardMomentum(t *testing.T) {
	sigs := run(t, Meta{Float: 5_000_000, PrevClose: 1.50}, DefaultFilters(), ramp())

	if len(sigs) == 0 {
		t.Fatal("expected a signal on rising price and volume above the EMA")
	}

	if got := sigs[0].ChangePct; got < 30 {
		t.Fatalf("change vs prior close: got %.1f", got)
	}
}

func TestFiltersBlock(t *testing.T) {
	cases := map[string]func(*Filters, *Meta){
		"price above max":  func(f *Filters, _ *Meta) { f.MaxPrice = 2 },
		"float above max":  func(_ *Filters, m *Meta) { m.Float = 50_000_000 },
		"unknown float":    func(_ *Filters, m *Meta) { m.Float = 0 },
		"change too small": func(f *Filters, _ *Meta) { f.MinChangePct = 90 },
		"session disabled": func(f *Filters, _ *Meta) { f.Regular = false },
		"volume too low":   func(f *Filters, _ *Meta) { f.MinSessionVolume = 10_000_000 },
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			f, m := DefaultFilters(), Meta{Float: 5_000_000, PrevClose: 1.50}
			mutate(&f, &m)

			if sigs := run(t, m, f, ramp()); len(sigs) != 0 {
				t.Fatalf("expected no signal, got %d", len(sigs))
			}
		})
	}
}

func TestCooldownSuppressesRepeats(t *testing.T) {
	bars := ramp()
	bars = append(bars, bar(15, 2.70, 160_000), bar(16, 3.00, 320_000))

	f := DefaultFilters()
	f.CooldownMinutes = 10

	if sigs := run(t, Meta{Float: 5_000_000, PrevClose: 1.50}, f, bars); len(sigs) != 1 {
		t.Fatalf("expected 1 signal inside cooldown, got %d", len(sigs))
	}
}

func TestPostMarketBaselineIsRegularClose(t *testing.T) {
	s := &state{meta: Meta{PrevClose: 1.00}}
	s.apply(bar(0, 2.00, 1000), 9)

	post := time.Date(2026, 11, 2, 16, 30, 0, 0, session.ET)
	s.apply(market.Bar{Sym: "X", C: 2.20, V: 1000, Start: post.UnixMilli(), End: post.Add(time.Minute).UnixMilli()}, 9)

	if got := s.changePct(2.20); got < 9.9 || got > 10.1 {
		t.Fatalf("post-market change should be vs regular close 2.00, got %.2f", got)
	}
}

func TestMinuteGapBlocksSignal(t *testing.T) {
	var bars []market.Bar
	for i := 0; i < 12; i++ {
		bars = append(bars, bar(i, 2.00, 10_000))
	}

	bars = append(bars, bar(12, 2.10, 20_000), bar(13, 2.25, 40_000), bar(20, 2.45, 80_000))

	if sigs := run(t, Meta{Float: 5_000_000, PrevClose: 1.50}, DefaultFilters(), bars); len(sigs) != 1 {
		t.Fatalf("expected only the pre-gap signal, got %d", len(sigs))
	}
}

func TestFailedBackfillRetries(t *testing.T) {
	calls := 0
	out := make(chan Signal, 8)
	backfill := func(context.Context, string, time.Time, time.Time) ([]market.Bar, error) {
		calls++
		if calls == 1 {
			return nil, context.DeadlineExceeded
		}

		return nil, nil
	}

	e := New(map[string]Meta{"X": {Float: 5_000_000, PrevClose: 1.50}}, DefaultFilters(), backfill, out)
	ctx := context.Background()

	e.Handle(ctx, bar(0, 2.00, 10_000))
	e.onBackfill(ctx, <-e.done)

	if st := e.syms["X"]; st.bf != bfNone {
		t.Fatalf("failed backfill should stay retryable, got %v", st.bf)
	}

	e.Handle(ctx, bar(1, 2.01, 10_000))

	if st := e.syms["X"]; st.bf != bfNone || len(st.pending) != 0 {
		t.Fatal("retry must wait for the delay")
	}

	e.Handle(ctx, bar(3, 2.02, 10_000))
	e.onBackfill(ctx, <-e.done)

	if st := e.syms["X"]; st.bf != bfDone || calls != 2 {
		t.Fatalf("expected retry to succeed, bf=%v calls=%d", st.bf, calls)
	}
}

func TestBackfillReplaysBufferedBars(t *testing.T) {
	var history []market.Bar
	for i := 0; i < 12; i++ {
		history = append(history, bar(i, 2.00, 10_000))
	}

	out := make(chan Signal, 8)
	backfill := func(context.Context, string, time.Time, time.Time) ([]market.Bar, error) { return history, nil }
	e := New(map[string]Meta{"X": {Float: 5_000_000, PrevClose: 1.50}}, DefaultFilters(), backfill, out)
	ctx := context.Background()

	e.Handle(ctx, bar(12, 2.10, 20_000))
	e.Handle(ctx, bar(13, 2.25, 40_000))
	e.Handle(ctx, bar(14, 2.45, 80_000))
	e.onBackfill(ctx, <-e.done)

	if len(out) == 0 {
		t.Fatal("expected a signal from history plus the bars buffered during backfill")
	}
}

package universe

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/realjv3/momo/internal/massive"
	"github.com/realjv3/momo/internal/store"
)

// Refresh reloads NASDAQ tickers, free float and previous closes into the store.
func Refresh(ctx context.Context, c *massive.Client, st *store.Store) (int, error) {
	tickers, err := c.NasdaqTickers(ctx)
	if err != nil {
		return 0, err
	}

	log.Printf("universe: %d NASDAQ tickers", len(tickers))

	floats, err := c.Floats(ctx, nil)
	if err != nil {
		return 0, err
	}

	byTicker := make(map[string]massive.Float, len(floats))
	for _, f := range floats {
		byTicker[f.Ticker] = f
	}

	closes, closeDate, err := c.PrevCloses(ctx, time.Now())
	if err != nil {
		return 0, err
	}

	syms := make([]store.Symbol, 0, len(tickers))

	for _, t := range tickers {
		pc := closes[t.Ticker]
		if pc <= 0 {
			continue
		}

		f := byTicker[t.Ticker]
		syms = append(syms, store.Symbol{
			Symbol:        t.Ticker,
			FreeFloat:     f.FreeFloat,
			PrevClose:     pc,
			PrevCloseDate: closeDate,
		})
	}

	prev, err := st.SymbolCount(ctx)
	if err != nil {
		return 0, err
	}

	if len(syms) == 0 || len(syms) < prev/2 {
		return 0, fmt.Errorf("refusing to replace %d symbols with %d", prev, len(syms))
	}

	if err := st.ReplaceSymbols(ctx, syms); err != nil {
		return 0, err
	}

	log.Printf("universe: stored %d symbols (prev close %s)", len(syms), closeDate)

	return len(syms), nil
}

package main

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/realjv3/momo/internal/config"
	"github.com/realjv3/momo/internal/engine"
	"github.com/realjv3/momo/internal/market"
	"github.com/realjv3/momo/internal/massive"
	"github.com/realjv3/momo/internal/session"
)

// replay runs historical minute bars for the given symbols through the engine and prints signals.
func replay(ctx context.Context, cfg config.Config, args []string) error {
	fs := flag.NewFlagSet("replay", flag.ContinueOnError)
	date := fs.String("date", "", "trading day, YYYY-MM-DD")
	syms := fs.String("symbols", "", "comma separated symbols")

	if err := fs.Parse(args); err != nil {
		return err
	}

	day, err := time.ParseInLocation("2006-01-02", *date, session.ET)
	if err != nil {
		return fmt.Errorf("-date: %w", err)
	}

	tickers := strings.Split(strings.ToUpper(*syms), ",")
	if *syms == "" {
		return fmt.Errorf("-symbols is required")
	}

	rest := massive.New(cfg.APIKey, cfg.RESTBase, cfg.RESTPerMin)

	closes, closeDate, err := rest.PrevCloses(ctx, day)
	if err != nil {
		return err
	}

	floats, err := rest.Floats(ctx, tickers)
	if err != nil {
		return err
	}

	meta := map[string]engine.Meta{}
	for _, f := range floats {
		meta[f.Ticker] = engine.Meta{Float: f.FreeFloat, PrevClose: closes[f.Ticker]}
	}

	fmt.Printf("replay %s, baseline close %s, %d symbols\n", *date, closeDate, len(meta))

	var bars []market.Bar

	for sym := range meta {
		b, err := rest.MinuteBars(ctx, sym, session.PreOpen(day), session.PreOpen(day).Add(16*time.Hour))
		if err != nil {
			return err
		}

		bars = append(bars, b...)
	}

	sort.Slice(bars, func(i, j int) bool { return bars[i].Start < bars[j].Start })

	signals := make(chan engine.Signal, len(bars)+1)
	eng := engine.New(meta, engine.DefaultFilters(), nil, signals)

	for _, b := range bars {
		eng.Handle(ctx, b)
	}

	close(signals)

	n := 0

	for s := range signals {
		n++

		fmt.Printf("%s  %-6s %-7s %+7.2f%%  $%-8.2f float %-10d vol %.0f\n",
			s.Time.In(session.ET).Format("15:04:05"), s.Symbol, s.Session, s.ChangePct, s.Price, s.Float, s.Volume)
	}

	fmt.Printf("%d signals from %d bars\n", n, len(bars))

	return nil
}

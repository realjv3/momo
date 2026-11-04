package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "time/tzdata"

	"github.com/realjv3/momo/internal/config"
	"github.com/realjv3/momo/internal/engine"
	"github.com/realjv3/momo/internal/market"
	"github.com/realjv3/momo/internal/massive"
	"github.com/realjv3/momo/internal/server"
	"github.com/realjv3/momo/internal/session"
	"github.com/realjv3/momo/internal/store"
	"github.com/realjv3/momo/internal/universe"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cmd := "serve"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	switch cmd {
	case "serve":
		err = serve(ctx, cfg)
	case "refresh":
		err = refresh(ctx, cfg)
	case "replay":
		err = replay(ctx, cfg, os.Args[2:])
	default:
		err = fmt.Errorf("unknown command %q (serve | refresh | replay)", cmd)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func refresh(ctx context.Context, cfg config.Config) error {
	st, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer st.Close()

	_, err = universe.Refresh(ctx, massive.New(cfg.APIKey, cfg.RESTBase, cfg.RESTPerMin), st)

	return err
}

func serve(ctx context.Context, cfg config.Config) error {
	st, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer st.Close()

	rest := massive.New(cfg.APIKey, cfg.RESTBase, cfg.RESTPerMin)

	meta, closeDate, err := st.SymbolMeta(ctx)
	if err != nil {
		return err
	}

	if len(meta) == 0 || closeDate < session.LastClosedDay(time.Now()).Format("2006-01-02") {
		if _, err := universe.Refresh(ctx, rest, st); err != nil {
			if len(meta) == 0 {
				return fmt.Errorf("refresh universe: %w", err)
			}

			log.Printf("refresh universe failed, using stale data: %v", err)
		}

		if meta, _, err = st.SymbolMeta(ctx); err != nil {
			return err
		}
	}

	filters, err := st.Filters(ctx)
	if err != nil {
		return err
	}

	signals := make(chan engine.Signal, 256)
	eng := engine.New(meta, filters, func(ctx context.Context, sym string, from, to time.Time) ([]market.Bar, error) {
		return rest.MinuteBars(ctx, sym, from, to)
	}, signals)
	srv := server.New(st, eng, "web/dist", cfg.Dev)

	stream := &massive.Stream{URL: cfg.WSURL, Key: cfg.APIKey, OnStatus: srv.SetFeedStatus}

	go eng.Run(ctx)
	go stream.Run(ctx, eng.Input())
	go srv.RunStatusTicker(ctx)
	go srv.RunSignals(ctx, signals)

	httpSrv := &http.Server{Addr: cfg.Addr, Handler: srv.Handler()}

	go func() {
		<-ctx.Done()
		httpSrv.Shutdown(context.Background())
	}()

	log.Printf("serving on %s, tracking %d symbols", cfg.Addr, len(meta))

	if err := httpSrv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	return nil
}

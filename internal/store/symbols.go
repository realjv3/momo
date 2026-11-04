package store

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/realjv3/momo/internal/engine"
)

type Symbol struct {
	Symbol        string
	FreeFloat     int64
	PrevClose     float64
	PrevCloseDate string
}

// ReplaceSymbols swaps the whole universe in one transaction.
func (s *Store) ReplaceSymbols(ctx context.Context, syms []Symbol) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `truncate symbols`); err != nil {
		return err
	}

	batch := &pgx.Batch{}
	for _, y := range syms {
		batch.Queue(`insert into symbols (symbol, free_float, prev_close, prev_close_date) values ($1,$2,$3,$4)`,
			y.Symbol, y.FreeFloat, y.PrevClose, y.PrevCloseDate)
	}

	res := tx.SendBatch(ctx, batch)
	for range syms {
		if _, err := res.Exec(); err != nil {
			res.Close()

			return err
		}
	}

	if err := res.Close(); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// SymbolMeta loads the engine's per-symbol inputs and the universe's prev-close date (YYYY-MM-DD).
func (s *Store) SymbolMeta(ctx context.Context) (map[string]engine.Meta, string, error) {
	rows, err := s.pool.Query(ctx, `select symbol, free_float, prev_close, coalesce(to_char(prev_close_date, 'YYYY-MM-DD'), '') from symbols`)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	meta := map[string]engine.Meta{}

	var newest string

	for rows.Next() {
		var (
			sym  string
			m    engine.Meta
			date string
		)

		if err := rows.Scan(&sym, &m.Float, &m.PrevClose, &date); err != nil {
			return nil, "", err
		}

		meta[sym] = m
		newest = max(newest, date)
	}

	return meta, newest, rows.Err()
}

func (s *Store) SymbolCount(ctx context.Context) (int, error) {
	var n int

	err := s.pool.QueryRow(ctx, `select count(*) from symbols`).Scan(&n)

	return n, err
}

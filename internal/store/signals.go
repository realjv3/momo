package store

import (
	"context"

	"github.com/realjv3/momo/internal/engine"
)

// InsertSignal persists sig and sets its ID.
func (s *Store) InsertSignal(ctx context.Context, sig *engine.Signal) error {
	return s.pool.QueryRow(ctx,
		`insert into signals (ts, symbol, session, change_pct, price, free_float, volume)
		 values ($1,$2,$3,$4,$5,$6,$7) returning id`,
		sig.Time.UTC(), sig.Symbol, sig.Session, sig.ChangePct, sig.Price, sig.Float, sig.Volume,
	).Scan(&sig.ID)
}

// RecentSignals returns the newest signals first.
func (s *Store) RecentSignals(ctx context.Context, limit int) ([]engine.Signal, error) {
	rows, err := s.pool.Query(ctx,
		`select id, ts, symbol, session, change_pct, price, free_float, volume
		 from signals order by ts desc, id desc limit $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sigs := []engine.Signal{}

	for rows.Next() {
		var g engine.Signal
		if err := rows.Scan(&g.ID, &g.Time, &g.Symbol, &g.Session, &g.ChangePct, &g.Price, &g.Float, &g.Volume); err != nil {
			return nil, err
		}

		g.Time = g.Time.UTC()
		sigs = append(sigs, g)
	}

	return sigs, rows.Err()
}

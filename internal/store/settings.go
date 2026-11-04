package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/realjv3/momo/internal/engine"
)

const filtersKey = "filters"

// Filters loads saved filters, falling back to defaults.
func (s *Store) Filters(ctx context.Context) (engine.Filters, error) {
	var raw []byte

	err := s.pool.QueryRow(ctx, `select value from settings where key=$1`, filtersKey).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return engine.DefaultFilters(), nil
	}

	if err != nil {
		return engine.Filters{}, err
	}

	f := engine.DefaultFilters()

	return f, json.Unmarshal(raw, &f)
}

func (s *Store) ResetFilters(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `delete from settings where key=$1`, filtersKey)

	return err
}

func (s *Store) SaveFilters(ctx context.Context, f engine.Filters) error {
	raw, err := json.Marshal(f)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx,
		`insert into settings (key, value) values ($1,$2) on conflict (key) do update set value=excluded.value`,
		filtersKey, raw)

	return err
}

package engine

import "github.com/realjv3/momo/internal/session"

// Filters are the live-adjustable scan criteria; zero on a Max field disables it.
type Filters struct {
	MinPrice             float64 `json:"minPrice"`
	MaxPrice             float64 `json:"maxPrice"`
	MinChangePct         float64 `json:"minChangePct"`
	MaxChangePct         float64 `json:"maxChangePct"`
	MinChangeFromOpenPct float64 `json:"minChangeFromOpenPct"`
	MinSessionVolume     float64 `json:"minSessionVolume"`
	MinBarVolume         float64 `json:"minBarVolume"`
	PriorBars            int     `json:"priorBars"`
	MinPriorBarVolume    float64 `json:"minPriorBarVolume"`
	MinFloat             float64 `json:"minFloat"`
	MaxFloat             float64 `json:"maxFloat"`
	MoveWindowMinutes    int     `json:"moveWindowMinutes"`
	MinMovePct           float64 `json:"minMovePct"`
	VolWindowMinutes     int     `json:"volWindowMinutes"`
	VolRiseFactor        float64 `json:"volRiseFactor"`
	EMAPeriod            int     `json:"emaPeriod"`
	CooldownMinutes      int     `json:"cooldownMinutes"`
	PreMarket            bool    `json:"preMarket"`
	Regular              bool    `json:"regular"`
	PostMarket           bool    `json:"postMarket"`
}

func DefaultFilters() Filters {
	return Filters{
		MinPrice:          1,
		MaxPrice:          20,
		MinChangePct:      5,
		MinSessionVolume:  350_000,
		MinBarVolume:      5_000,
		PriorBars:         2,
		MinPriorBarVolume: 2_000,
		MaxFloat:          300_000_000,
		MoveWindowMinutes: 5,
		MinMovePct:        3,
		VolWindowMinutes:  3,
		VolRiseFactor:     2,
		EMAPeriod:         9,
		CooldownMinutes:   3,
		PreMarket:         true,
		Regular:           true,
		PostMarket:        true,
	}
}

// Normalize clamps values that would break the engine.
func (f Filters) Normalize() Filters {
	f.EMAPeriod = max(f.EMAPeriod, 2)
	f.PriorBars = min(max(f.PriorBars, 0), maxRecent-1)
	f.MoveWindowMinutes = min(max(f.MoveWindowMinutes, 1), maxRecent)
	f.VolWindowMinutes = min(max(f.VolWindowMinutes, 1), maxRecent)
	f.CooldownMinutes = max(f.CooldownMinutes, 0)

	return f
}

func (f Filters) sessionOn(s session.Session) bool {
	switch s {
	case session.Pre:
		return f.PreMarket
	case session.Regular:
		return f.Regular
	case session.Post:
		return f.PostMarket
	}

	return false
}

// candidate applies the cheap per-symbol filters that gate backfill and signals.
func (f Filters) candidate(s session.Session, price, changePct float64, float int64) bool {
	switch {
	case !f.sessionOn(s):
		return false
	case price < f.MinPrice || (f.MaxPrice > 0 && price > f.MaxPrice):
		return false
	case changePct < f.MinChangePct || (f.MaxChangePct > 0 && changePct > f.MaxChangePct):
		return false
	case f.MinFloat > 0 && float < int64(f.MinFloat):
		return false
	case f.MaxFloat > 0 && (float <= 0 || float > int64(f.MaxFloat)):
		return false
	}

	return true
}

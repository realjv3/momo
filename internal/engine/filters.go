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
	MinFloat             float64 `json:"minFloat"`
	MaxFloat             float64 `json:"maxFloat"`
	MinBarChangePct      float64 `json:"minBarChangePct"`
	VolRiseFactor        float64 `json:"volRiseFactor"`
	RisingBars           int     `json:"risingBars"`
	EMAPeriod            int     `json:"emaPeriod"`
	CooldownMinutes      int     `json:"cooldownMinutes"`
	PreMarket            bool    `json:"preMarket"`
	Regular              bool    `json:"regular"`
	PostMarket           bool    `json:"postMarket"`
}

func DefaultFilters() Filters {
	return Filters{
		MinPrice:         1,
		MaxPrice:         20,
		MinChangePct:     5,
		MinSessionVolume: 50_000,
		MinBarVolume:     5_000,
		MaxFloat:         20_000_000,
		MinBarChangePct:  0.25,
		VolRiseFactor:    1.2,
		RisingBars:       2,
		EMAPeriod:        9,
		CooldownMinutes:  3,
		PreMarket:        true,
		Regular:          true,
		PostMarket:       true,
	}
}

// Normalize clamps values that would break the engine.
func (f Filters) Normalize() Filters {
	f.EMAPeriod = max(f.EMAPeriod, 2)
	f.RisingBars = min(max(f.RisingBars, 1), maxRecent-2)
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

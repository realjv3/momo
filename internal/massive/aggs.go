package massive

import (
	"context"
	"fmt"
	"time"

	"github.com/realjv3/momo/internal/market"
	"github.com/realjv3/momo/internal/session"
)

type dailyBar struct {
	T  string  `json:"T"`
	C  float64 `json:"c"`
	Ts int64   `json:"t"` // exact-match sink so "t" cannot overwrite "T" via case-insensitive decoding
}

type minuteBar struct {
	C float64 `json:"c"`
	V float64 `json:"v"`
	T int64   `json:"t"`
}

// MinuteBars returns extended-hours minute bars with start in [from, to].
func (c *Client) MinuteBars(ctx context.Context, sym string, from, to time.Time) ([]market.Bar, error) {
	u := fmt.Sprintf("%s/v2/aggs/ticker/%s/range/1/minute/%d/%d?adjusted=false&sort=asc&limit=50000", c.base, sym, from.UnixMilli(), to.UnixMilli())

	raw, err := getPaged[minuteBar](ctx, c, u)
	if err != nil {
		return nil, err
	}

	bars := make([]market.Bar, len(raw))
	for i, r := range raw {
		bars[i] = market.Bar{Sym: sym, C: r.C, V: r.V, Start: r.T, End: r.T + 60_000}
	}

	return bars, nil
}

// PrevCloses returns each symbol's close on the most recent closed trading day as of now, with that date.
func (c *Client) PrevCloses(ctx context.Context, now time.Time) (map[string]float64, string, error) {
	d := session.LastClosedDay(now)

	for i := 0; i < 10; i++ {
		if wd := d.Weekday(); wd == time.Saturday || wd == time.Sunday {
			d = d.AddDate(0, 0, -1)

			continue
		}

		date := d.Format("2006-01-02")

		bars, err := getPaged[dailyBar](ctx, c, c.base+"/v2/aggs/grouped/locale/us/market/stocks/"+date+"?adjusted=false")
		if err != nil {
			return nil, "", err
		}

		if len(bars) > 0 {
			closes := make(map[string]float64, len(bars))
			for _, b := range bars {
				closes[b.T] = b.C
			}

			return closes, date, nil
		}

		d = d.AddDate(0, 0, -1)
	}

	return nil, "", fmt.Errorf("no closed trading day found before %s", now.Format("2006-01-02"))
}

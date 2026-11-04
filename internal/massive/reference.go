package massive

import (
	"context"
	"net/url"
	"strings"
)

type Ticker struct {
	Ticker string `json:"ticker"`
}

// NasdaqTickers returns active NASDAQ-listed common stocks.
func (c *Client) NasdaqTickers(ctx context.Context) ([]Ticker, error) {
	return getPaged[Ticker](ctx, c, c.base+"/v3/reference/tickers?market=stocks&exchange=XNAS&type=CS&active=true&limit=1000")
}

type Float struct {
	Ticker    string `json:"ticker"`
	FreeFloat int64  `json:"free_float"`
}

// Floats returns free float for the given tickers, or for every ticker when none are given.
func (c *Client) Floats(ctx context.Context, tickers []string) ([]Float, error) {
	q := url.Values{"limit": {"5000"}, "sort": {"ticker.asc"}}
	if len(tickers) > 0 {
		q.Set("ticker.any_of", strings.Join(tickers, ","))
	}

	return getPaged[Float](ctx, c, c.base+"/stocks/vX/float?"+q.Encode())
}

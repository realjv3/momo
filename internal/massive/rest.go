package massive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type Client struct {
	key  string
	base string
	hc   *http.Client
	lim  *rate.Limiter
}

func New(key, base string, perMin int) *Client {
	return &Client{
		key:  key,
		base: base,
		hc:   &http.Client{Timeout: 60 * time.Second},
		lim:  rate.NewLimiter(rate.Every(time.Minute/time.Duration(perMin)), perMin),
	}
}

func (c *Client) get(ctx context.Context, url string, out any) error {
	for attempt := 0; ; attempt++ {
		if err := c.lim.Wait(ctx); err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		req.Header.Set("Authorization", "Bearer "+c.key)

		resp, err := c.hc.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode == http.StatusTooManyRequests && attempt < 8 {
			resp.Body.Close()

			select {
			case <-time.After(15 * time.Second):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err = decode(resp, out)
		resp.Body.Close()

		return err
	}
}

func decode(resp *http.Response, out any) error {
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 300))

		return fmt.Errorf("massive: %s: %s", resp.Status, b)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

// getPaged follows next_url until exhausted.
func getPaged[T any](ctx context.Context, c *Client, url string) ([]T, error) {
	var all []T

	for url != "" {
		var page struct {
			Results []T    `json:"results"`
			NextURL string `json:"next_url"`
		}

		if err := c.get(ctx, url, &page); err != nil {
			return nil, err
		}

		all = append(all, page.Results...)
		url = page.NextURL
	}

	return all, nil
}

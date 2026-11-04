package massive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/coder/websocket"

	"github.com/realjv3/momo/internal/market"
)

type Stream struct {
	URL      string
	Key      string
	OnStatus func(string)
}

type event struct {
	Ev      string  `json:"ev"`
	Status  string  `json:"status"`
	Message string  `json:"message"`
	Sym     string  `json:"sym"`
	V       float64 `json:"v"`
	OP      float64 `json:"op"`
	C       float64 `json:"c"`
	S       int64   `json:"s"`
	E       int64   `json:"e"`
}

// Run keeps an AM.* subscription alive, reconnecting with backoff, until ctx ends.
func (s *Stream) Run(ctx context.Context, out chan<- market.Bar) {
	backoff := time.Second

	for ctx.Err() == nil {
		authed, err := s.session(ctx, out)
		if ctx.Err() != nil {
			return
		}

		if authed {
			backoff = time.Second
		}

		s.status("disconnected: " + err.Error())
		log.Printf("ws: %v; retrying in %s", err, backoff)

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return
		}

		backoff = min(backoff*2, 30*time.Second)
	}
}

// session runs one connection and reports whether it authenticated before ending.
func (s *Stream) session(ctx context.Context, out chan<- market.Bar) (authed bool, err error) {
	conn, _, err := websocket.Dial(ctx, s.URL, nil)
	if err != nil {
		return false, err
	}
	defer conn.CloseNow()

	conn.SetReadLimit(32 << 20)

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return authed, err
		}

		var events []event
		if err := json.Unmarshal(data, &events); err != nil {
			return authed, fmt.Errorf("decode: %w", err)
		}

		for _, e := range events {
			switch e.Ev {
			case "status":
				if e.Status == "auth_success" {
					authed = true
				}

				if err := s.onStatus(ctx, conn, e); err != nil {
					return authed, err
				}
			case "AM":
				bar := market.Bar{Sym: e.Sym, C: e.C, V: e.V, OP: e.OP, Start: e.S, End: e.E}

				select {
				case out <- bar:
				case <-ctx.Done():
					return authed, ctx.Err()
				}
			}
		}
	}
}

func (s *Stream) onStatus(ctx context.Context, conn *websocket.Conn, e event) error {
	switch e.Status {
	case "connected":
		return conn.Write(ctx, websocket.MessageText, mustJSON(map[string]string{"action": "auth", "params": s.Key}))
	case "auth_success":
		s.status("connected")

		return conn.Write(ctx, websocket.MessageText, mustJSON(map[string]string{"action": "subscribe", "params": "AM.*"}))
	case "auth_failed", "max_connections":
		return errors.New(e.Status + ": " + e.Message)
	}

	return nil
}

func (s *Stream) status(msg string) {
	if s.OnStatus != nil {
		s.OnStatus(msg)
	}
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return b
}

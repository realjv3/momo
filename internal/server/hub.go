package server

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type hub struct {
	mu      sync.Mutex
	clients map[chan []byte]struct{}
}

func newHub() *hub { return &hub{clients: map[chan []byte]struct{}{}} }

func (h *hub) add() chan []byte {
	ch := make(chan []byte, 64)

	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()

	return ch
}

func (h *hub) remove(ch chan []byte) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
}

func envelope(msgType string, data any) ([]byte, error) {
	return json.Marshal(map[string]any{"type": msgType, "data": data})
}

// broadcast drops messages for clients that fall behind.
func (h *hub) broadcast(msgType string, data any) {
	raw, err := envelope(msgType, data)
	if err != nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	for ch := range h.clients {
		select {
		case ch <- raw:
		default:
		}
	}
}

func (h *hub) serve(ctx context.Context, conn *websocket.Conn, hello []byte) {
	ch := h.add()
	defer h.remove(ch)

	ctx = conn.CloseRead(ctx)

	if err := conn.Write(ctx, websocket.MessageText, hello); err != nil {
		return
	}

	for {
		select {
		case msg := <-ch:
			wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := conn.Write(wctx, websocket.MessageText, msg)

			cancel()

			if err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

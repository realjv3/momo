package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"

	"github.com/realjv3/momo/internal/engine"
	"github.com/realjv3/momo/internal/session"
	"github.com/realjv3/momo/internal/store"
)

type Server struct {
	st      *store.Store
	eng     *engine.Engine
	hub     *hub
	feed    atomic.Value
	webRoot string
	dev     bool
	unsaved atomic.Int64
}

func New(st *store.Store, eng *engine.Engine, webRoot string, dev bool) *Server {
	s := &Server{st: st, eng: eng, hub: newHub(), webRoot: webRoot, dev: dev}
	s.SetFeedStatus("connecting")

	return s
}

func (s *Server) SetFeedStatus(msg string) {
	s.feed.Store(msg)
	s.hub.broadcast("status", s.status())
}

// Emit saves sig then pushes it to every connected browser.
func (s *Server) Emit(ctx context.Context, sig engine.Signal) engine.Signal {
	if err := s.st.InsertSignal(ctx, &sig); err != nil {
		sig.ID = s.unsaved.Add(-1) // negative ids keep unsaved rows distinct in the UI

		log.Printf("insert signal %s: %v", sig.Symbol, err)
	}

	log.Printf("signal %s %s %.2f%% $%.2f", sig.Symbol, sig.Session, sig.ChangePct, sig.Price)
	s.hub.broadcast("signal", sig)

	return sig
}

// RunSignals emits every engine signal until ctx ends.
func (s *Server) RunSignals(ctx context.Context, signals <-chan engine.Signal) {
	for {
		select {
		case sig := <-signals:
			s.Emit(ctx, sig)
		case <-ctx.Done():
			return
		}
	}
}

type statusMsg struct {
	Session string `json:"session"`
	Feed    string `json:"feed"`
	Tracked int    `json:"tracked"`
}

func (s *Server) status() statusMsg {
	return statusMsg{
		Session: string(session.At(time.Now())),
		Feed:    s.feed.Load().(string),
		Tracked: s.eng.Tracked(),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/signals", s.getSignals)
	mux.HandleFunc("GET /api/filters", s.getFilters)
	mux.HandleFunc("PUT /api/filters", s.putFilters)
	mux.HandleFunc("DELETE /api/filters", s.resetFilters)
	mux.HandleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, s.status()) })
	mux.HandleFunc("GET /api/stream", s.stream)
	if s.dev {
		mux.HandleFunc("POST /api/dev/signal", s.devSignal)
	}

	mux.Handle("/", s.static())

	return mux
}

type signalPage struct {
	Signals []engine.Signal `json:"signals"`
	Total   int             `json:"total"`
}

func (s *Server) getSignals(w http.ResponseWriter, r *http.Request) {
	limit := min(max(queryInt(r, "limit", 50), 1), 200)
	offset := max(queryInt(r, "offset", 0), 0)

	sigs, total, err := s.st.ListSignals(r.Context(), limit, offset)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)

		return
	}

	writeJSON(w, signalPage{Signals: sigs, Total: total})
}

func queryInt(r *http.Request, key string, def int) int {
	n, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil {
		return def
	}

	return n
}

func (s *Server) getFilters(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.eng.Filters())
}

func (s *Server) putFilters(w http.ResponseWriter, r *http.Request) {
	f := s.eng.Filters()
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		httpError(w, err, http.StatusBadRequest)

		return
	}

	f = f.Normalize()

	if err := s.st.SaveFilters(r.Context(), f); err != nil {
		httpError(w, err, http.StatusInternalServerError)

		return
	}

	s.eng.SetFilters(f)
	writeJSON(w, s.eng.Filters())
}

func (s *Server) resetFilters(w http.ResponseWriter, r *http.Request) {
	if err := s.st.ResetFilters(r.Context()); err != nil {
		httpError(w, err, http.StatusInternalServerError)

		return
	}

	s.eng.SetFilters(engine.DefaultFilters())
	writeJSON(w, s.eng.Filters())
}

func (s *Server) devSignal(w http.ResponseWriter, r *http.Request) {
	var sig engine.Signal
	if err := json.NewDecoder(r.Body).Decode(&sig); err != nil {
		httpError(w, err, http.StatusBadRequest)

		return
	}

	if sig.Time.IsZero() {
		sig.Time = time.Now().UTC()
	}

	if sig.Session == "" {
		sig.Session = string(session.At(sig.Time))
	}

	writeJSON(w, s.Emit(r.Context(), sig))
}

func (s *Server) stream(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"localhost:*", "127.0.0.1:*"}})
	if err != nil {
		return
	}
	defer conn.CloseNow()

	hello, err := envelope("status", s.status())
	if err != nil {
		return
	}

	s.hub.serve(r.Context(), conn, hello)
}

// RunStatusTicker re-broadcasts status so the UI session badge stays current.
func (s *Server) RunStatusTicker(ctx context.Context) {
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			s.hub.broadcast("status", s.status())
		case <-ctx.Done():
			return
		}
	}
}

func (s *Server) static() http.Handler {
	if _, err := os.Stat(s.webRoot); err != nil {
		return http.NotFoundHandler()
	}

	fs := http.FileServer(http.Dir(s.webRoot))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(filepath.Join(s.webRoot, filepath.Clean(r.URL.Path))); err != nil {
			r.URL.Path = "/"
		}

		fs.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}

func httpError(w http.ResponseWriter, err error, code int) {
	http.Error(w, err.Error(), code)
}

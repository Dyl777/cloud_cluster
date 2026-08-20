package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gpuhub/cloud/internal/shared/events"
	"github.com/gpuhub/cloud/internal/shared/httpx"
)

// Handler receives internal events (via the event bus or POST /events) and is
// the hook point for email, push, dashboard feed and partner webhooks.
type Handler struct {
	mu   sync.Mutex
	feed []Event
}

// Event is an inbound notification.
type Event struct {
	Topic   string `json:"topic"`
	Payload string `json:"payload"`
}

// Routes registers the notify routes.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /events", h.ingest)
	mux.HandleFunc("GET /events", h.list)
}

// Handle ingests one event from the shared event bus.
func (h *Handler) Handle(ctx context.Context, evt events.Event) error {
	raw, err := json.Marshal(evt.Payload)
	if err != nil {
		raw = []byte{}
	}
	h.mu.Lock()
	h.feed = append(h.feed, Event{Topic: string(evt.Topic), Payload: string(raw)})
	h.mu.Unlock()
	return nil
}

func (h *Handler) ingest(w http.ResponseWriter, r *http.Request) {
	var e Event
	if err := httpx.Decode(r, &e); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	h.mu.Lock()
	h.feed = append(h.feed, e)
	h.mu.Unlock()
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	out := make([]Event, len(h.feed))
	copy(out, h.feed)
	h.mu.Unlock()
	httpx.WriteJSON(w, http.StatusOK, out)
}
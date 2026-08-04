package notify

import (
	"net/http"

	"github.com/gpuhub/cloud/internal/shared/httpx"
)

// Handler receives internal events (as webhooks) and could fan them out to
// user-facing notification channels (email, push, dashboard feed).
type Handler struct {
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

func (h *Handler) ingest(w http.ResponseWriter, r *http.Request) {
	var e Event
	if err := httpx.Decode(r, &e); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	h.feed = append(h.feed, e)
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, h.feed)
}

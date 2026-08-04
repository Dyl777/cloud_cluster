package payments

import (
	"net/http"

	"github.com/gpuhub/cloud/internal/shared/httpx"
	"github.com/gpuhub/cloud/internal/shared/id"
)

// Handler wires HTTP routes to the Orchestrator.
type Handler struct {
	orch *Orchestrator
}

// NewHandler returns a Handler backed by orch.
func NewHandler(orch *Orchestrator) *Handler {
	return &Handler{orch: orch}
}

// Routes registers all payment routes.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /payments/topup", h.startTopup)
	mux.HandleFunc("POST /payments/{topupID}/confirm", h.confirm)
	mux.HandleFunc("GET /payments/{topupID}", h.get)
}

func (h *Handler) startTopup(w http.ResponseWriter, r *http.Request) {
	var req TopupRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	req.ID = id.New("top")
	intent, err := h.orch.StartTopup(req)
	if err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "start_failed", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusAccepted, map[string]any{
		"topup_id": req.ID,
		"status":   "pending",
		"intent":   intent,
	})
}

func (h *Handler) confirm(w http.ResponseWriter, r *http.Request) {
	topupID := r.PathValue("topupID")
	if err := h.orch.ConfirmTopup(topupID); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "confirm_failed", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "confirmed"})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	topupID := r.PathValue("topupID")
	t, ok := h.orch.Get(topupID)
	if !ok {
		httpx.WriteErr(w, http.StatusNotFound, "not_found", "topup not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, t)
}

package simbridge

import (
	"net/http"

	"github.com/gpuhub/cloud/internal/shared/httpx"
)

// Handler exposes the local simbridge HTTP API for the webapp.
type Handler struct {
	svc *Service
}

// NewHandler returns a Handler backed by svc.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes registers simbridge routes. CORS is enabled for browser calls.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /simbridge/carriers", h.withCORS(h.listCarriers))
	mux.HandleFunc("GET /simbridge/sims", h.withCORS(h.listSIMs))
	mux.HandleFunc("POST /simbridge/dial", h.withCORS(h.dial))
}

func (h *Handler) withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func (h *Handler) listCarriers(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, Carriers)
}

func (h *Handler) listSIMs(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, h.svc.SIMs())
}

func (h *Handler) dial(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Carrier string `json:"carrier"`
		Phone   string `json:"phone"`
		Amount  int64  `json:"amount"`
		USSD    string `json:"ussd,omitempty"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	result, err := h.svc.Dial(body.Carrier, body.Phone, body.Amount, body.USSD)
	if err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "dial_failed", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

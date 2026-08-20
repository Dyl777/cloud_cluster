package paymgr

import (
	"net/http"

	"github.com/gpuhub/cloud/internal/shared/httpx"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /paymgr/system/config", h.systemConfig)
	mux.HandleFunc("POST /paymgr/route", h.route)
	mux.HandleFunc("POST /paymgr/topup", h.startTopup)
	mux.HandleFunc("GET /paymgr/nodes/live", h.liveNodes)
	mux.HandleFunc("POST /paymgr/topup/{topupID}/refund", h.refund)
	mux.HandleFunc("GET /paymgr/topup/{topupID}", h.get)
}

func (h *Handler) systemConfig(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, h.svc.SystemConfig())
}

func (h *Handler) route(w http.ResponseWriter, r *http.Request) {
	var in TopupInput
	if err := httpx.Decode(r, &in); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, h.svc.Route(in))
}

func (h *Handler) startTopup(w http.ResponseWriter, r *http.Request) {
	var in TopupInput
	if err := httpx.Decode(r, &in); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	res, err := h.svc.StartTopup(in)
	if err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "start_failed", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusAccepted, res)
}

func (h *Handler) liveNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.svc.LiveNodes()
	if err != nil {
		httpx.WriteErr(w, http.StatusBadGateway, "gateway_unreachable", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, nodes)
}

func (h *Handler) refund(w http.ResponseWriter, r *http.Request) {
	topupID := r.PathValue("topupID")
	if err := h.svc.Refund(topupID); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "refund_failed", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "refund_dispatched"})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	topupID := r.PathValue("topupID")
	t, ok := h.svc.Get(topupID)
	if !ok {
		httpx.WriteErr(w, http.StatusNotFound, "not_found", "topup not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, t)
}

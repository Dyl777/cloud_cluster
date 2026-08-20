package mobilevm

import (
	"net/http"

	"github.com/gpuhub/cloud/internal/shared/httpx"
	"github.com/gpuhub/cloud/internal/shared/id"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /mobilevm/transfer", h.start)
	mux.HandleFunc("GET /mobilevm/jobs/{jobID}", h.get)
	mux.HandleFunc("GET /mobilevm/jobs", h.list)
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	var req TransferRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	jobID := id.New("vm")
	job, err := h.svc.Start(jobID, req)
	if err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "transfer_failed", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusAccepted, job)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("jobID")
	job, ok := h.svc.Get(jobID)
	if !ok {
		httpx.WriteErr(w, http.StatusNotFound, "not_found", "job not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, job)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, h.svc.List())
}

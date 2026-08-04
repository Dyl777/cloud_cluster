package provision

import (
	"net/http"

	"github.com/gpuhub/cloud/internal/shared/httpx"
	"github.com/gpuhub/cloud/internal/shared/id"
)

// Handler wires HTTP routes to the Service.
type Handler struct {
	svc *Service
}

// NewHandler returns a Handler backed by svc.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes registers instance routes.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /instances", h.create)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserID   string `json:"user_id"`
		GPUName  string `json:"gpu_name"`
		NumGPUs  int    `json:"num_gpus"`
		Provider string `json:"provider"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	i := h.svc.Create(id.New("inst"), body.UserID, body.GPUName, body.NumGPUs, body.Provider)
	httpx.WriteJSON(w, http.StatusCreated, i)
}

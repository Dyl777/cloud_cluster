package identity

import (
	"net/http"
)

// Handler wires HTTP routes to the Service.
type Handler struct {
	svc *Service
}

// NewHandler returns a Handler backed by svc.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes registers all identity routes on the given mux.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /users", h.register)
	mux.HandleFunc("GET /users/current", h.current)
	mux.HandleFunc("GET /admin/users", h.adminList)
	mux.HandleFunc("POST /admin/users/{email}/role", h.adminSetRole)
}

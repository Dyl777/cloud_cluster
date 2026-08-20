package identity

import (
	"net/http"

	"github.com/gpuhub/cloud/internal/shared/httpx"
	"github.com/gpuhub/cloud/internal/shared/id"
)

// Register handles POST /users.
func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	u, err := h.svc.Register(id.New("usr"), body.Email, body.Name)
	if err != nil {
		httpx.WriteErr(w, http.StatusConflict, "conflict", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, u)
}

// current handles GET /users/current. The caller is resolved from the email
// query param (auth middleware would normally provide it).
func (h *Handler) current(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		httpx.WriteErr(w, http.StatusBadRequest, "bad_request", "email query param required")
		return
	}
	u, ok := h.svc.Get(email)
	if !ok {
		httpx.WriteErr(w, http.StatusNotFound, "not_found", "user not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, u)
}

// adminList handles GET /admin/users. Requires a superadmin actor.
func (h *Handler) adminList(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"users": h.svc.List()})
}

// adminSetRole handles POST /admin/users/{email}/role.
func (h *Handler) adminSetRole(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	email := r.PathValue("email")
	var body struct {
		Role Role `json:"role"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	u, err := h.svc.SetRole(email, body.Role)
	if err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, u)
}

// requireAdmin rejects the call unless the ?actor=<email> user holds an
// admin-or-higher role. In production this would come from real auth tokens.
func (h *Handler) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	actor := r.URL.Query().Get("actor")
	if actor == "" {
		httpx.WriteErr(w, http.StatusUnauthorized, "unauthorized", "actor email required")
		return false
	}
	u, ok := h.svc.Get(actor)
	if !ok || !u.Role.AtLeast(RoleAdmin) {
		httpx.WriteErr(w, http.StatusForbidden, "forbidden", "actor has no admin role")
		return false
	}
	return true
}

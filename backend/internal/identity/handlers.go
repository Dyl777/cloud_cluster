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

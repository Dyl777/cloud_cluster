package settlement

import (
	"net/http"

	"github.com/gpuhub/cloud/internal/shared/httpx"
	"github.com/gpuhub/cloud/internal/shared/id"
)

// Handler wires HTTP routes to the Settlement Service.
type Handler struct {
	svc *Service
}

// NewHandler returns a Handler backed by svc.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes registers the settlement routes.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /settlements", h.pay)
}

func (h *Handler) pay(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserID      string `json:"user_id"`
		Provider    string `json:"provider"`
		Amount      int64  `json:"amount"`
		Currency    string `json:"currency"`
		BankAccount string `json:"bank_account"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	p := h.svc.Pay(Payment{
		ID:          id.New("pay"),
		UserID:      body.UserID,
		Provider:    body.Provider,
		Amount:      body.Amount,
		Currency:    body.Currency,
		BankAccount: body.BankAccount,
	})
	httpx.WriteJSON(w, http.StatusOK, p)
}

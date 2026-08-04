package wallet

import (
	"net/http"
	"strings"

	"github.com/gpuhub/cloud/internal/shared/httpx"
	"github.com/gpuhub/cloud/internal/shared/id"
	"github.com/gpuhub/cloud/internal/shared/money"
)

// Handler wires HTTP routes to the wallet Service.
type Handler struct {
	svc *Service
}

// NewHandler returns a Handler backed by svc.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes registers all wallet routes.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /wallets/{userID}", h.balance)
	mux.HandleFunc("POST /wallets/{userID}/credit", h.credit)
	mux.HandleFunc("POST /wallets/{userID}/hold", h.hold)
	mux.HandleFunc("POST /wallets/holds/{holdID}/settle", h.settle)
	mux.HandleFunc("POST /wallets/holds/{holdID}/release", h.release)
	mux.HandleFunc("GET /wallets/{userID}/ledger", h.ledger)
}

func (h *Handler) balance(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	acct := h.svc.Balance(userID, "USD")
	httpx.WriteJSON(w, http.StatusOK, acct)
}

func (h *Handler) credit(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	var body struct {
		Subunits int64  `json:"subunits"`
		Currency string `json:"currency"`
		Ref      string `json:"ref"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	amt := money.Money{Subunits: body.Subunits, Currency: strings.ToUpper(body.Currency)}
	if err := h.svc.Credit(id.New("ldg"), userID, amt.Currency, amt, body.Ref); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "credit_failed", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, h.svc.Balance(userID, amt.Currency))
}

func (h *Handler) hold(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	var body struct {
		Subunits int64  `json:"subunits"`
		Currency string `json:"currency"`
		Ref      string `json:"ref"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	amt := money.Money{Subunits: body.Subunits, Currency: strings.ToUpper(body.Currency)}
	holdID := id.New("hold")
	hold, err := h.svc.Hold(holdID, userID, amt.Currency, amt, body.Ref)
	if err != nil {
		httpx.WriteErr(w, http.StatusPaymentRequired, "insufficient_balance", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, hold)
}

func (h *Handler) settle(w http.ResponseWriter, r *http.Request) {
	holdID := r.PathValue("holdID")
	if err := h.svc.Settle(holdID); err != nil {
		httpx.WriteErr(w, http.StatusNotFound, "no_hold", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "settled"})
}

func (h *Handler) release(w http.ResponseWriter, r *http.Request) {
	holdID := r.PathValue("holdID")
	if err := h.svc.Release(holdID); err != nil {
		httpx.WriteErr(w, http.StatusNotFound, "no_hold", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "released"})
}

func (h *Handler) ledger(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	httpx.WriteJSON(w, http.StatusOK, h.svc.Ledger(userID))
}

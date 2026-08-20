package compat

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gpuhub/cloud/internal/shared/httpx"
)

// showUser mirrors GET /users/current. The SDK strips "api_key" and prints
// fields like id, balance, email, api_key_hash.
func (s *Server) showUser(w http.ResponseWriter, r *http.Request) {
	acct := s.wal.Balance(s.userID, "USD")
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"id":           s.userID,
		"user_id":      s.userID,
		"email":        "demo@gpuhub.dev",
		"balance":      acct.Balance.Float(),
		"api_key":      "",
		"api_key_hash": s.apiKey,
		"is_verified":  true,
	})
}

// setUser mirrors PUT /users/ (settings like ssh_key). Accepts and ignores
// shape changes but persists nothing for now.
func (s *Server) setUser(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	_ = httpx.Decode(r, &body)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"success": true})
}

// showInvoices mirrors GET /users/me/invoices (billing history).
func (s *Server) showInvoices(w http.ResponseWriter, r *http.Request) {
	type row struct {
		Type      string  `json:"type"`
		Amount    float64 `json:"amount"`
		Timestamp int64   `json:"timestamp"`
	}
	inv := make([]row, 0)
	for _, e := range s.wal.Ledger(s.userID) {
		inv = append(inv, row{
			Type: e.Type, Amount: e.Amount.Float(), Timestamp: e.CreatedAt.Unix(),
		})
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"invoices": inv,
		"current":  map[string]any{"balance": s.wal.Balance(s.userID, "USD").Balance.Float()},
	})
}

// showCharges mirrors GET /charges/ (v1-style charge rows).
func (s *Server) showCharges(w http.ResponseWriter, r *http.Request) {
	charges := make([]map[string]any, 0)
	for _, e := range s.wal.Ledger(s.userID) {
		if e.Type == "credit" || e.Type == "topup" {
			continue
		}
		charges = append(charges, map[string]any{
			"type": e.Type, "amount": e.Amount.Float(), "timestamp": e.CreatedAt.Unix(),
		})
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"charges": charges})
}

// decodeStrict reads a small JSON body without draining.
func decodeStrict(r *http.Request, v any) error {
	buf, err := io.ReadAll(io.LimitReader(r.Body, 4096))
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, v)
}

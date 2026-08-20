// Package compat is a REST compatibility gateway that mirrors the
// console.vast.ai /api/v0 surface so unmodified official Vast tools
// (vast-cli, the SDK, skypilot's vast backend, vast-pyworker, ...) work
// against this platform instead of Vast.ai.
package compat

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gpuhub/cloud/internal/marketplace"
	"github.com/gpuhub/cloud/internal/provision"
	"github.com/gpuhub/cloud/internal/shared/httpx"
	"github.com/gpuhub/cloud/internal/shared/id"
	"github.com/gpuhub/cloud/internal/shared/money"
	"github.com/gpuhub/cloud/internal/wallet"
)

// Server wires the Vast-compatible routes to the platform services.
// It deliberately calls the internal services in-process so the whole
// surface is served by one binary: `cmd/compat`.
type Server struct {
	apiKey  string
	userID  string
	reg     *marketplace.Registry
	prov    *provision.Service
	wal     *wallet.Service
	mux     *http.ServeMux
	gateway string
}

// NewServer returns a compat Server rooted at the internal services.
func NewServer(apiKey, userID string, reg *marketplace.Registry, prov *provision.Service, wal *wallet.Service) *Server {
	s := &Server{apiKey: apiKey, userID: userID, reg: reg, prov: prov, wal: wal, mux: http.NewServeMux()}
	s.Routes()
	return s
}

// Mux returns the mounted route tree (for wrapping in httpx.NewServer.Run).
func (s *Server) Mux() *http.ServeMux { return s.mux }

// Routes registers every /api/v0-compatible route.
func (s *Server) Routes() {
	v0 := func(pattern string, fn http.HandlerFunc) {
		s.mux.HandleFunc("/api/v0"+pattern, fn)
		s.mux.HandleFunc("/api/v0"+pattern+"/", fn)
	}
	v0("/bundles", s.searchOffers)
	v0("/search/asks", s.searchOffers)
	v0("/template", s.searchTemplates)
	s.mux.HandleFunc("PUT /api/v0/asks/{id}", s.createInstance)
	s.mux.HandleFunc("POST /api/v0/asks/bulk", s.createInstancesBulk)
	s.mux.HandleFunc("PUT /api/v0/instances", s.updateInstances)
	s.mux.HandleFunc("DELETE /api/v0/instances", s.destroyInstances)
	s.mux.HandleFunc("PUT /api/v0/instances/{id}", s.updateInstance)
	s.mux.HandleFunc("DELETE /api/v0/instances/{id}", s.destroyInstance)
	s.mux.HandleFunc("GET /api/v0/instances/{id}", s.showInstance)
	s.mux.HandleFunc("PUT /api/v0/instances/reboot/{id}", s.rebootInstance)
	s.mux.HandleFunc("PUT /api/v0/instances/command/{id}", s.commandInstance)
	s.mux.HandleFunc("PUT /api/v0/instances/balance/{id}", s.instanceBalance)
	s.mux.HandleFunc("GET /api/v1/instances", s.showInstancesV1)
	s.mux.HandleFunc("GET /instances/balance/{id}", s.instanceBalance)
	v0("/instances", s.showInstances)
	v0("/users/current", s.showUser)
	v0("/users", s.setUser)
	v0("/users/me/invoices", s.showInvoices)
	v0("/charges", s.showCharges)
	s.mux.HandleFunc("GET /api/v0/healthz", s.healthz)
	s.mux.HandleFunc("GET /healthz", s.healthz)
	// top-up helper this platform adds (not in console.vast.ai)
	v0("/topup", s.topup)
	// superadmin surface (platform aggregates; not part of the vast API).
	v0("/admin/stats", s.adminStats)
}

// healthz is a simple liveness probe (not part of the vast surface).
func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// ServeHTTP runs auth then routes.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r) {
		httpx.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"success": false,
			"msg":     "unauthorized", "error": "invalid api_key",
		})
		return
	}
	// vast tools normalize trailing slashes inconsistently; drop one so both
	// /asks/{id} and /asks/{id}/ resolve to the same handler.
	r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}
	s.mux.ServeHTTP(w, r)
}

// authorized accepts the api_key via Authorization Bearer or JSON body,
// matching how the official SDK sends credentials (see client.py). The body
// is restored so downstream handlers can still decode it.
func (s *Server) authorized(r *http.Request) bool {
	key := ""
	if h := r.Header.Get("Authorization"); len(h) > 8 && h[:7] == "Bearer " {
		key = h[7:]
	}
	if key == "" && r.Body != nil {
		buf, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err == nil {
			var cred struct {
				APIKey string `json:"api_key"`
			}
			if json.Unmarshal(buf, &cred) == nil && cred.APIKey != "" {
				key = cred.APIKey
			}
			r.Body = io.NopCloser(bytes.NewReader(buf))
		}
	}
	return key != "" && key == s.apiKey
}

// topup credits the demo user's wallet via the deterministic mock rail.
func (s *Server) topup(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Amount   float64 `json:"amount"`
		Currency string  `json:"currency"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "msg": err.Error()})
		return
	}
	if body.Currency == "" {
		body.Currency = "USD"
	}
	cur, err := money.NormalizeCurrency(body.Currency)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "msg": "unknown currency"})
		return
	}
	s.wal.Credit(id.New("ldg"), s.userID, cur, money.FromUnits(body.Amount, cur), "topup")
	acct := s.wal.Balance(s.userID, cur)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"success": true, "balance": acct.Balance.Float(),
	})
}

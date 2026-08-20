package compat

import (
	"net/http"

	"github.com/gpuhub/cloud/internal/shared/httpx"
)

// adminStats mirrors GET /api/v0/admin/stats (platform-wide aggregates over
// the in-process registry, provisioner and wallet).
func (s *Server) adminStats(w http.ResponseWriter, r *http.Request) {
	offers := s.reg.ListAll()
	instances := s.prov.List("")
	ledger := s.wal.Ledger(s.userID)

	var offerGPUs, running int
	var creditSum, spendSum float64
	byGPU := map[string]int{}
	for _, o := range offers {
		offerGPUs += o.NumGPUs
		byGPU[o.GPUName] += o.NumGPUs
	}
	for _, i := range instances {
		if i.Status == "running" {
			running++
		}
	}
	for _, e := range ledger {
		f := e.Amount.Float()
		if e.Type == "credit" || e.Type == "topup" {
			creditSum += f
		} else if f < 0 {
			spendSum += -f
		}
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"users": map[string]any{
			"registered": 1,
			"active":     1,
		},
		"offers": map[string]any{
			"count":      len(offers),
			"total_gpus": offerGPUs,
			"by_gpu":     byGPU,
		},
		"instances": map[string]any{
			"count":   len(instances),
			"running": running,
		},
		"ledger": map[string]any{
			"credits": creditSum,
			"spend":   spendSum,
		},
	})
}
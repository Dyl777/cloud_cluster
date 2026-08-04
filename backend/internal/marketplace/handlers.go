package marketplace

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gpuhub/cloud/internal/shared/httpx"
)

// Handler wires HTTP routes to the registry.
type Handler struct {
	registry *Registry
}

// NewHandler returns a Handler backed by the registry.
func NewHandler(registry *Registry) *Handler {
	return &Handler{registry: registry}
}

// Routes registers the marketplace routes.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /offers", h.listOffers)
}

func (h *Handler) listOffers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	gpu := strings.ToLower(q.Get("gpu"))
	region := q.Get("region")
	maxPrice, _ := strconv.ParseFloat(q.Get("max_price"), 64)

	offers := h.registry.ListAll()
	filtered := make([]Offer, 0, len(offers))
	for _, o := range offers {
		if gpu != "" && !strings.Contains(strings.ToLower(o.GPUName), gpu) {
			continue
		}
		if region != "" && !strings.EqualFold(o.Region, region) {
			continue
		}
		if maxPrice > 0 && o.Dph > maxPrice {
			continue
		}
		filtered = append(filtered, o)
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"count": len(filtered), "offers": filtered})
}

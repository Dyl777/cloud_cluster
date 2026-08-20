package compat

import (
	"encoding/json"
	"net/http"

	"github.com/gpuhub/cloud/internal/marketplace"
	"github.com/gpuhub/cloud/internal/shared/httpx"
)

// searchOffers serves both POST /bundles/ and PUT /search/asks/. It translates
// the official SDK's rich filter language ({"gpu_name": {"eq": ...}},
// {"dph_total": {"lte": ...}}) and returns the {"offers": [...]} envelope.
//
// POST /bundles/ sends the filter dict directly as the JSON body; the newer
// PUT /search/asks/ sends {"select_cols": [...], "q": <filter dict>}.
func (s *Server) searchOffers(w http.ResponseWriter, r *http.Request) {
	var raw map[string]json.RawMessage
	if err := httpx.Decode(r, &raw); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "msg": err.Error()})
		return
	}
	q := raw
	if nested, ok := raw["q"]; ok {
		var inner map[string]json.RawMessage
		if json.Unmarshal(nested, &inner) == nil {
			q = inner
		}
	}
	limit := 0
	order := [][]any(nil)
	if b, ok := raw["limit"]; ok {
		_ = json.Unmarshal(b, &limit)
	}
	if b, ok := raw["order"]; ok {
		_ = json.Unmarshal(b, &order)
	}
	delete(q, "type")
	delete(q, "order")
	delete(q, "limit")
	delete(q, "allocated_storage")
	delete(q, "disable_bundling")
	delete(q, "select_cols")
	delete(q, "no_default")

	out := make([]offerView, 0)
	for _, o := range s.reg.ListAll() {
		if !offerMatches(o, q) {
			continue
		}
		v := offerViewFrom(o)
		if v.Verified && !v.Rented && v.Rentable {
			out = append(out, v)
		}
	}
	if len(order) > 0 && len(order[0]) >= 2 {
		dir, _ := order[0][1].(string)
		sortOffers(out, dir)
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"offers": out})
}

// searchTemplates mirrors GET /template/ (select_cols/select_filters query).
func (s *Server) searchTemplates(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"templates": []any{
			map[string]any{"id": "fastai", "name": "Fast.ai (fastai 1.0.61)", "instance_type": "1x RTX 3090", "disk_size": 32},
		},
	})
}

// offerView is the wire shape the official CLI prints for offer rows.
type offerView struct {
	ID           string  `json:"id"`
	GPUName      string  `json:"gpu_name"`
	GPUVRAM      int64   `json:"gpu_ram"`
	NumGPUs      int     `json:"num_gpus"`
	CPU          int     `json:"cpu_cores"`
	RAM          int     `json:"cpu_ram"`
	Disk         float64 `json:"disk_space"`
	Dph          float64 `json:"dph_total"`
	DisplayPrice float64 `json:"display_price"`
	InetDown     int     `json:"inet_down"`
	InetUp       int     `json:"inet_up"`
	Reliability  float64 `json:"reliability2"`
	Score        float64 `json:"score"`
	Region       string  `json:"region"`
	Datacenter   bool    `json:"datacenter"`
	Verified     bool    `json:"verified"`
	Rented       bool    `json:"rented"`
	Rentable     bool    `json:"rentable"`
	InetSpeed    struct {
		Down int `json:"down"`
		Up   int `json:"up"`
	} `json:"inet_speed"`
	GPUDisplay struct {
		Name string `json:"name"`
	} `json:"gpu_display"`
}

func offerViewFrom(o marketplace.Offer) offerView {
	v := offerView{
		ID: o.ID, GPUName: o.GPUName, GPUVRAM: o.GPUVRAM, NumGPUs: o.NumGPUs,
		CPU: o.CPU, RAM: o.RAM, Disk: o.Disk, Dph: o.Dph, DisplayPrice: o.Dph,
		InetDown: o.InetDown, InetUp: o.InetUp, Reliability: o.Reliability,
		Score: o.Score, Region: o.Region, Datacenter: o.Datacenter,
		Verified: o.Verified, Rented: o.Rented, Rentable: o.Rentable,
	}
	v.InetSpeed.Down = o.InetDown
	v.InetSpeed.Up = o.InetUp
	v.GPUDisplay.Name = o.GPUName
	return v
}

func sortOffers(v []offerView, dir string) {
	for i := 1; i < len(v); i++ {
		for j := i; j > 0; j-- {
			lt := v[j].Score < v[j-1].Score
			if dir == "desc" {
				lt = v[j].Score > v[j-1].Score
			}
			if !lt {
				break
			}
			v[j], v[j-1] = v[j-1], v[j]
		}
	}
}

// offerMatches applies the SDK filter block to one market offer. Unknown keys
// are ignored so future filters degrade gracefully.
func offerMatches(o marketplace.Offer, q map[string]json.RawMessage) bool {
	for key, raw := range q {
		var op map[string]json.RawMessage
		if json.Unmarshal(raw, &op) != nil {
			continue
		}
		switch key {
		case "type", "order", "allocated_storage", "disable_bundling", "limit", "verified", "external", "rentable", "rented":
			// handled globally or metadata only
		case "gpu_name":
			if !eqStr(op, o.GPUName) {
				return false
			}
		case "region", "country", "location":
			if !eqStr(op, o.Region) {
				return false
			}
		case "num_gpus":
			if !eqInt(op, o.NumGPUs) {
				return false
			}
		case "dph_total":
			if opEq, err := floatOp(op, "eq"); err == nil && opEq != 0 && o.Dph != opEq {
				return false
			}
			if opLte, err := floatOp(op, "lte"); err == nil && o.Dph > opLte {
				return false
			}
			if opGte, err := floatOp(op, "gte"); err == nil && o.Dph < opGte {
				return false
			}
		}
	}
	return true
}

// floatOp decodes a numeric operator value (eq/lte/gte) from a filter block.
// Returns err if the operator is absent or non-numeric.
func floatOp(op map[string]json.RawMessage, key string) (float64, error) {
	b, ok := op[key]
	if !ok {
		return 0, errMissing
	}
	var f float64
	if err := json.Unmarshal(b, &f); err != nil {
		return 0, err
	}
	return f, nil
}

var errMissing = errStr("missing operator")

type errStr string

func (e errStr) Error() string { return string(e) }

// eqStr matches a string filter like {"gpu_name": {"eq": "RTX 3090"}}.
func eqStr(op map[string]json.RawMessage, cur string) bool {
	b, ok := op["eq"]
	if !ok {
		return true
	}
	var t string
	if json.Unmarshal(b, &t) != nil {
		return true
	}
	return t == "" || t == cur
}

// eqInt matches an int filter like {"num_gpus": {"eq": 2}}.
func eqInt(op map[string]json.RawMessage, cur int) bool {
	b, ok := op["eq"]
	if !ok {
		return true
	}
	var n int
	if json.Unmarshal(b, &n) != nil {
		return true
	}
	return n == 0 || n == cur
}

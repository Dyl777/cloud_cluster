// Package marketplace aggregates GPU availability from several cloud
// providers via adapters and serves filtered offer listings.
package marketplace

// Offer is a rentable GPU listing surfaced from a cloud provider.
type Offer struct {
	ID          string  `json:"id"`
	Provider    string  `json:"provider"`
	GPUName     string  `json:"gpu_name"`
	GPUVRAM     int64   `json:"gpu_vram"` // MB
	NumGPUs     int     `json:"num_gpus"`
	CPU         int     `json:"cpu_cores"`
	RAM         int     `json:"cpu_ram"`    // GB
	Disk        float64 `json:"disk_space"` // TB
	Dph         float64 `json:"dph_total"`  // $/hour per offer
	Region      string  `json:"region"`
	Datacenter  bool    `json:"datacenter"`
	Verified    bool    `json:"verified"`
	Reliability float64 `json:"reliability2"`
}

// Provider is a cloud backend we can source GPUs from.
type Provider interface {
	Name() string
	ListOffers() ([]Offer, error)
}

// Registry holds the configured cloud providers.
type Registry struct {
	providers []Provider
}

// NewRegistry returns an empty Registry.
func NewRegistry(providers ...Provider) *Registry {
	return &Registry{providers: providers}
}

// ListAll aggregates offers from every provider.
func (r *Registry) ListAll() []Offer {
	var all []Offer
	for _, p := range r.providers {
		offers, err := p.ListOffers()
		if err != nil {
			continue
		}
		all = append(all, offers...)
	}
	return all
}

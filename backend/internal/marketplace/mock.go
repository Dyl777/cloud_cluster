package marketplace

import "fmt"

// mockProvider simulates a cloud provider so the system runs offline end to
// end. Real adapters (aws, azure, gcp, runpod, coreweave) implement the same
// Provider interface with live HTTP calls and the same to-low shape.
type mockProvider struct{}

// NewMockProvider returns a simulated provider.
func NewMockProvider(cloud string) Provider { return mockProvider{} }

func (mockProvider) Name() string { return "mock" }

var gpus = []struct {
	name  string
	bench int
	dph   float64
}{
	{"H100 SXM", 940, 1.85},
	{"H200 SXM", 1010, 2.40},
	{"A100 SXM4", 690, 1.10},
	{"RTX 4090", 420, 0.34},
	{"RTX 3090", 300, 0.19},
	{"L40S", 480, 0.72},
}

func (mockProvider) ListOffers() ([]Offer, error) {
	out := make([]Offer, 0, len(gpus))
	for i, g := range gpus {
		out = append(out, Offer{
			ID:          fmt.Sprintf("mock-%d", i),
			Provider:    "mock",
			GPUName:     g.name,
			GPUVRAM:     24564,
			NumGPUs:     1,
			CPU:         16,
			RAM:         64,
			Disk:        1.5,
			Dph:         g.dph,
			Region:      "US-West",
			Datacenter:  true,
			Verified:    true,
			Reliability: 0.99,
			Rented:      false,
			Rentable:    true,
			Score:       float64(g.bench) / 1000,
			InetDown:    25000,
			InetUp:      10000,
		})
	}
	return out, nil
}

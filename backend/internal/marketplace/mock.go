package marketplace

import "fmt"

// mockProvider simulates a cloud provider so the system runs offline end to
// end. Later adapters (aws, azure, gcp, runpod, coreweave) implement the
// same Provider interface with real HTTP calls and the same offer shape.
type mockProvider struct{}

// NewMockProvider returns a simulated provider named after cloud.
func NewMockProvider(cloud string) Provider { return mockProvider{} }

func (mockProvider) Name() string { return "mock" }

var gpuModels = []string{"H100 SXM", "A100 SXM4", "RTX 4090", "RTX 3090", "L40S"}

func (mockProvider) ListOffers() ([]Offer, error) {
	out := make([]Offer, 0, len(gpuModels))
	for i, g := range gpuModels {
		out = append(out, Offer{
			ID:          fmt.Sprintf("mock-%d", i),
			Provider:    "mock",
			GPUName:     g,
			GPUVRAM:     24564,
			NumGPUs:     1,
			CPU:         12,
			RAM:         64,
			Disk:        1,
			Dph:         float64(10-i) / 10.0,
			Region:      "US-West",
			Datacenter:  true,
			Verified:    true,
			Reliability: 0.99,
		})
	}
	return out, nil
}

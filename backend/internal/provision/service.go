package provision

import "sync"

// Service stores instances in memory.
type Service struct {
	mu        sync.Mutex
	instances map[string]Instance
}

// New returns an empty Service.
func New() *Service {
	return &Service{instances: make(map[string]Instance)}
}

// Create registers a new instance (simulated launch).
func (s *Service) Create(id, userID, gpuName string, numGPUs int, provider string) Instance {

	i := Instance{
		ID:       id,
		UserID:   userID,
		GPUName:  gpuName,
		NumGPUs:  numGPUs,
		Provider: provider,
		Status:   "running",
	}
	s.mu.Lock()
	s.instances[id] = i
	s.mu.Unlock()
	return i
}

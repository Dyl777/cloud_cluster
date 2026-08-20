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

// Launch creates an instance from a full spec (used by the compat gateway).
func (s *Service) Launch(i Instance) Instance {
	if i.Status == "" {
		i.Status = "running"
	}
	s.mu.Lock()
	s.instances[i.ID] = i
	s.mu.Unlock()
	return i
}

// Get returns one instance by id.
func (s *Service) Get(id string) (Instance, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i, ok := s.instances[id]
	return i, ok
}

// List returns all instances, optionally filtered by user.
func (s *Service) List(userID string) []Instance {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Instance, 0, len(s.instances))
	for _, i := range s.instances {
		if userID != "" && i.UserID != userID {
			continue
		}
		out = append(out, i)
	}
	return out
}

// Update mutates fields on an existing instance.
func (s *Service) Update(id string, fn func(*Instance)) (Instance, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i, ok := s.instances[id]
	if !ok {
		return i, false
	}
	fn(&i)
	s.instances[id] = i
	return i, true
}

// Destroy removes an instance.
func (s *Service) Destroy(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.instances[id]
	if !ok {
		return false
	}
	delete(s.instances, id)
	return true
}

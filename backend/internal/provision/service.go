package provision

import (
	"context"
	"log/slog"
	"sync"

	"github.com/gpuhub/cloud/internal/shared/events"
)

// Service manages instance lifecycle and persists to an InstanceStore.
type Service struct {
	mu    sync.Mutex
	store InstanceStore
	bus   events.Bus
}

// New returns an in-memory Service (single process).
func New() *Service { return &Service{store: newMemoryStore()} }

// NewPG returns a Service persisted to Postgres.
func NewPG(store InstanceStore) *Service { return &Service{store: store} }

// SetBus attaches an event bus; instance lifecycle events are fanned out.
func (s *Service) SetBus(b events.Bus) { s.bus = b }

// InstanceEvent is the payload published on instance.provisioned/destroyed.
type InstanceEvent struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	GPUName  string `json:"gpu_name"`
	NumGPUs  int    `json:"num_gpus"`
	Provider string `json:"provider"`
	Status   string `json:"status"`
}

func (s *Service) emit(topic events.Topic, e InstanceEvent) {
	if s.bus == nil {
		return
	}
	if err := s.bus.Publish(context.Background(), topic, e); err != nil {
		slog.Warn("provision event publish failed", "topic", topic, "err", err)
	}
}

// Create registers a new instance.
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
	defer s.mu.Unlock()
	_ = s.store.Put(&i)
	s.emit(events.TopicInstanceProvisioned, instanceEventFrom(i))
	return i
}

// Launch creates an instance from a full spec (used by the compat gateway).
func (s *Service) Launch(i Instance) Instance {
	if i.Status == "" {
		i.Status = "running"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.store.Put(&i)
	s.emit(events.TopicInstanceProvisioned, instanceEventFrom(i))
	return i
}

// Get returns one instance by id.
func (s *Service) Get(id string) (Instance, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i, err := s.store.Get(id)
	if err != nil {
		return Instance{}, false
	}
	return *i, true
}

// List returns all instances, optionally filtered by user.
func (s *Service) List(userID string) []Instance {
	s.mu.Lock()
	defer s.mu.Unlock()
	out, err := s.store.List(userID)
	if err != nil {
		return []Instance{}
	}
	return out
}

// Update mutates fields on an existing instance.
func (s *Service) Update(id string, fn func(*Instance)) (Instance, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i, err := s.store.Get(id)
	if err != nil {
		return Instance{}, false
	}
	fn(i)
	_ = s.store.Put(i)
	return *i, true
}

// Destroy removes an instance.
func (s *Service) Destroy(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.store.Delete(id); err != nil {
		return false
	}
	s.emit(events.TopicInstanceDestroyed, InstanceEvent{ID: id, Status: "destroyed"})
	return true
}

func instanceEventFrom(i Instance) InstanceEvent {
	return InstanceEvent{
		ID:       i.ID,
		UserID:   i.UserID,
		GPUName:  i.GPUName,
		NumGPUs:  i.NumGPUs,
		Provider: i.Provider,
		Status:   i.Status,
	}
}
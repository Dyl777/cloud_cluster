package settlement

import (
	"context"
	"log/slog"
	"sync"

	"github.com/gpuhub/cloud/internal/shared/events"
)

// Service pays providers from the pooled bank/fintech account.
type Service struct {
	mu       sync.Mutex
	payments map[string]Payment
	bus      events.Bus
}

// New returns an empty Settlement Service.
func New() *Service {
	return &Service{payments: make(map[string]Payment)}
}

// SetBus attaches an event bus; settlement.paid is fanned out.
func (s *Service) SetBus(b events.Bus) { s.bus = b }

// Pay records and marks a payment as paid (simulated transfer via the
// configured corporate bank account).
func (s *Service) Pay(p Payment) Payment {
	if p.Status == "" {
		p.Status = "paid"
	}
	s.mu.Lock()
	s.payments[p.ID] = p
	s.mu.Unlock()
	if s.bus != nil {
		if err := s.bus.Publish(context.Background(), events.TopicSettlementPaid, p); err != nil {
			slog.Warn("settlement event publish failed", "topic", events.TopicSettlementPaid, "err", err)
		}
	}
	return p
}
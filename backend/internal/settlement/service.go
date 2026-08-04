package settlement

import "sync"

// Service pays providers from the pooled bank/fintech account.
type Service struct {
	mu       sync.Mutex
	payments map[string]Payment
}

// New returns an empty Settlement Service.
func New() *Service {
	return &Service{payments: make(map[string]Payment)}
}

// Pay records and marks a payment as paid (simulated transfer via the
// configured corporate bank account).
func (s *Service) Pay(p Payment) Payment {
	if p.Status == "" {
		p.Status = "paid"
	}
	s.mu.Lock()
	s.payments[p.ID] = p
	s.mu.Unlock()
	return p
}

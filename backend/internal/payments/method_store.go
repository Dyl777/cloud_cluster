package payments

import (
	"errors"
	"sync"
)

var ErrNoMethod = errors.New("payment method not found")

// MethodStore keeps per-user saved payment methods in memory.
type MethodStore struct {
	mu      sync.Mutex
	methods map[string][]SavedMethod
}

func newMethodStore() *MethodStore {
	return &MethodStore{methods: make(map[string][]SavedMethod)}
}

func (s *MethodStore) List(userID string) []SavedMethod {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]SavedMethod(nil), s.methods[userID]...)
}

func (s *MethodStore) Add(m SavedMethod) SavedMethod {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.methods[m.UserID] = append(s.methods[m.UserID], m)
	return m
}

func (s *MethodStore) Get(userID, methodID string) (SavedMethod, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, m := range s.methods[userID] {
		if m.ID == methodID {
			return m, true
		}
	}
	return SavedMethod{}, false
}

func (s *MethodStore) Delete(userID, methodID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.methods[userID]
	for i, m := range list {
		if m.ID == methodID {
			s.methods[userID] = append(list[:i], list[i+1:]...)
			return true
		}
	}
	return false
}

func (o *Orchestrator) resolveTopup(req *TopupRequest) error {
	if req.PaymentMethodID == "" {
		return nil
	}
	m, ok := o.methods.Get(req.UserID, req.PaymentMethodID)
	if !ok {
		return ErrNoMethod
	}
	req.Method = m.Kind
	switch m.Kind {
	case MethodMobileMoney:
		req.Carrier = m.Carrier
		if req.Phone == "" {
			req.Phone = m.Phone
		}
	case MethodBank, MethodFintech:
		req.RailProvider = m.Provider
		if req.AccountRef == "" {
			req.AccountRef = m.AccountRef
		}
	}
	return nil
}

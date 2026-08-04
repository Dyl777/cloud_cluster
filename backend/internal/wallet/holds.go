package wallet

import (
	"time"

	"github.com/gpuhub/cloud/internal/shared/money"
)

// Hold reserves funds for a future charge.
func (s *Service) Hold(id, userID, currency string, amount money.Money, ref string) (*Hold, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a := s.ensure(userID, currency)
	if a.Balance.Subunits < amount.Subunits {
		return nil, ErrInsufficient
	}
	a.Balance, _ = a.Balance.Sub(amount)
	a.Held, _ = a.Held.Add(amount)
	h := &Hold{ID: id, UserID: userID, Amount: amount, Reference: ref, CreatedAt: time.Now()}
	s.holds[id] = h
	s.append(id, userID, "hold", amount, ref)
	return h, nil
}

// Settle moves a hold into spent funds.
func (s *Service) Settle(holdID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	h, ok := s.holds[holdID]
	if !ok {
		return ErrNoHold
	}
	a := s.ensure(h.UserID, h.Amount.Currency)
	a.Held, _ = a.Held.Sub(h.Amount)
	delete(s.holds, holdID)
	s.append(h.ID, h.UserID, "settle", h.Amount, h.Reference)
	return nil
}

// Release returns a hold to the balance.
func (s *Service) Release(holdID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	h, ok := s.holds[holdID]
	if !ok {
		return ErrNoHold
	}
	a := s.ensure(h.UserID, h.Amount.Currency)
	a.Held, _ = a.Held.Sub(h.Amount)
	a.Balance, _ = a.Balance.Add(h.Amount)
	delete(s.holds, holdID)
	s.append(h.ID, h.UserID, "release", h.Amount, h.Reference)
	return nil
}

// Ledger returns all entries for a user.
func (s *Service) Ledger(userID string) []LedgerEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]LedgerEntry, 0)
	for _, e := range s.ledger {
		if e.UserID == userID {
			out = append(out, e)
		}
	}
	return out
}

func (s *Service) append(id, userID, typ string, amount money.Money, ref string) {
	s.ledger = append(s.ledger, LedgerEntry{
		ID: id, UserID: userID, Type: typ, Amount: amount, Reference: ref, CreatedAt: time.Now(),
	})
}

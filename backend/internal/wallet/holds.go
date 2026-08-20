package wallet

import (
	"errors"
	"time"

	"github.com/gpuhub/cloud/internal/shared/money"
)

// Hold reserves funds for a future charge.
func (s *Service) Hold(id, userID, currency string, amount money.Money, ref string) (*Hold, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, err := s.acct(userID, currency)
	if err != nil {
		return nil, err
	}
	if a.Balance.Subunits < amount.Subunits {
		return nil, ErrInsufficient
	}
	a.Balance, _ = a.Balance.Sub(amount)
	a.Held, _ = a.Held.Add(amount)
	a.UpdatedAt = time.Now()
	h := &Hold{ID: id, UserID: userID, Amount: amount, Reference: ref, CreatedAt: time.Now()}

	if err := s.store.Upsert(a); err != nil {
		return nil, err
	}
	if err := s.store.PutHold(h); err != nil {
		return nil, err
	}
	if err := s.store.AppendLedger(ledgerEntry(id, userID, "hold", amount, ref)); err != nil {
		return nil, err
	}
	return h, nil
}

// Settle moves a hold into spent funds.
func (s *Service) Settle(holdID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	h, err := s.store.GetHold(holdID)
	if errors.Is(err, ErrNotFound) {
		return ErrNoHold
	}
	if err != nil {
		return err
	}
	a, err := s.acct(h.UserID, h.Amount.Currency)
	if err != nil {
		return err
	}
	a.Held, _ = a.Held.Sub(h.Amount)
	a.UpdatedAt = time.Now()

	if err := s.store.Upsert(a); err != nil {
		return err
	}
	if err := s.store.DeleteHold(holdID); err != nil {
		return err
	}
	return s.store.AppendLedger(ledgerEntry(h.ID+":settle", h.UserID, "settle", h.Amount, h.Reference))
}

// Release returns a hold to the balance.
func (s *Service) Release(holdID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	h, err := s.store.GetHold(holdID)
	if errors.Is(err, ErrNotFound) {
		return ErrNoHold
	}
	if err != nil {
		return err
	}
	a, err := s.acct(h.UserID, h.Amount.Currency)
	if err != nil {
		return err
	}
	a.Held, _ = a.Held.Sub(h.Amount)
	a.Balance, _ = a.Balance.Add(h.Amount)
	a.UpdatedAt = time.Now()

	if err := s.store.Upsert(a); err != nil {
		return err
	}
	if err := s.store.DeleteHold(holdID); err != nil {
		return err
	}
	return s.store.AppendLedger(ledgerEntry(h.ID+":release", h.UserID, "release", h.Amount, h.Reference))
}

// Ledger returns all entries for a user.
func (s *Service) Ledger(userID string) []LedgerEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out, err := s.store.ListLedger(userID)
	if err != nil {
		return []LedgerEntry{}
	}
	return out
}
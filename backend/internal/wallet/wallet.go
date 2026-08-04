// Package wallet owns user balances, holds and the immutable ledger.
// Money is stored as integer micro-units (see internal/shared/money).
package wallet

import (
	"errors"
	"sync"
	"time"

	"github.com/gpuhub/cloud/internal/shared/money"
)

var ErrInsufficient = errors.New("insufficient balance")
var ErrNoHold = errors.New("hold not found")

// Account is the balance of one user.
type Account struct {
	UserID    string      `json:"user_id"`
	Balance   money.Money `json:"balance"`
	Held      money.Money `json:"held"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// LedgerEntry records one irreversible movement.
type LedgerEntry struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Type      string      `json:"type"` // credit, hold, release, settle, topup
	Amount    money.Money `json:"amount"`
	Reference string      `json:"reference"`
	CreatedAt time.Time   `json:"created_at"`
}

// Hold reserves funds for a future charge.
type Hold struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Amount    money.Money `json:"amount"`
	Reference string      `json:"reference"`
	CreatedAt time.Time   `json:"created_at"`
}

// Service guards accounts with a mutex.
type Service struct {
	mu      sync.Mutex
	account map[string]*Account
	holds   map[string]*Hold
	ledger  []LedgerEntry
}

// New returns a wallet Service.
func New() *Service {
	return &Service{account: make(map[string]*Account), holds: make(map[string]*Hold)}
}

// Balance returns the current account (creates it if missing).
func (s *Service) Balance(userID, currency string) Account {
	s.mu.Lock()
	defer s.mu.Unlock()
	acct := s.ensure(userID, currency)
	return *acct
}

func (s *Service) ensure(userID, currency string) *Account {
	if a, ok := s.account[userID]; ok {
		return a
	}
	a := &Account{UserID: userID, Balance: money.Zero(currency), Held: money.Zero(currency)}
	s.account[userID] = a
	return a
}

// Credit adds funds (e.g. confirmed top-up) and records a ledger entry.
func (s *Service) Credit(id, userID, currency string, amount money.Money, ref string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	a := s.ensure(userID, currency)
	a.Balance, _ = a.Balance.Add(amount)
	a.UpdatedAt = time.Now()
	s.append(id, userID, "credit", amount, ref)
	return nil
}

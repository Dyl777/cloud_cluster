// Package wallet owns user balances, holds and the immutable ledger.
// Money is stored as integer micro-units (see internal/shared/money).
package wallet

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gpuhub/cloud/internal/shared/events"
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

// Service serializes wallet operations and persists them to a Store.
type Service struct {
	mu    sync.Mutex
	store Store
	bus   events.Bus
}

// New returns a wallet Service backed by in-memory storage (single process).
func New() *Service { return &Service{store: newMemoryStore()} }

// NewPG returns a wallet Service persisted to Postgres.
func NewPG(store Store) *Service { return &Service{store: store} }

// SetBus attaches an event bus; published lifecycle events are fanned out.
func (s *Service) SetBus(b events.Bus) { s.bus = b }

// WalletEvent is the payload published on wallet.credited / wallet.debited.
type WalletEvent struct {
	UserID    string      `json:"user_id"`
	Account   Account     `json:"account"`
	Amount    money.Money `json:"amount"`
	Reference string      `json:"reference"`
}

func (s *Service) emit(topic events.Topic, e WalletEvent) {
	if s.bus == nil {
		return
	}
	_ = s.bus.Publish(context.Background(), topic, e)
}

// acct returns the account for userID, creating a zero balance if missing.
func (s *Service) acct(userID, currency string) (*Account, error) {
	a, err := s.store.Get(userID)
	if err == nil {
		return a, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	a = &Account{
		UserID:    userID,
		Balance:   money.Zero(currency),
		Held:      money.Zero(currency),
		UpdatedAt: time.Now(),
	}
	if err := s.store.Upsert(a); err != nil {
		return nil, err
	}
	return a, nil
}

// Balance returns the current account (created if missing).
func (s *Service) Balance(userID, currency string) Account {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, err := s.acct(userID, currency)
	if err != nil {
		return Account{UserID: userID, Balance: money.Zero(currency), Held: money.Zero(currency)}
	}
	return *a
}

// Credit adds funds (e.g. confirmed top-up) and records a ledger entry.
func (s *Service) Credit(id, userID, currency string, amount money.Money, ref string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, err := s.acct(userID, currency)
	if err != nil {
		return err
	}
	a.Balance, _ = a.Balance.Add(amount)
	a.UpdatedAt = time.Now()
	if err := s.store.Upsert(a); err != nil {
		return err
	}
	if err := s.store.AppendLedger(ledgerEntry(id, userID, "credit", amount, ref)); err != nil {
		return err
	}
	s.emit(events.TopicWalletCredited, WalletEvent{UserID: userID, Account: *a, Amount: amount, Reference: ref})
	return nil
}

func ledgerEntry(id, userID, typ string, amount money.Money, ref string) LedgerEntry {
	return LedgerEntry{
		ID:        id,
		UserID:    userID,
		Type:      typ,
		Amount:    amount,
		Reference: ref,
		CreatedAt: time.Now(),
	}
}
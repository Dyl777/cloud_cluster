package wallet

import (
	"database/sql"
	"errors"
	"time"

	"github.com/gpuhub/cloud/internal/shared/money"
)

// ErrNotFound is returned by a Store when a record is absent.
var ErrNotFound = errors.New("wallet: not found")

// Store persists wallet state. memoryStore backs the default single-process
// behavior; pgStore persists to Postgres for containerized deployments.
type Store interface {
	Get(userID string) (*Account, error)
	Upsert(a *Account) error
	AppendLedger(e LedgerEntry) error
	ListLedger(userID string) ([]LedgerEntry, error)
	PutHold(h *Hold) error
	GetHold(id string) (*Hold, error)
	DeleteHold(id string) error
}

// memoryStore implements Store in-process (the original behavior).
type memoryStore struct {
	account map[string]*Account
	holds   map[string]*Hold
	ledger  []LedgerEntry
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		account: make(map[string]*Account),
		holds:   make(map[string]*Hold),
	}
}

func (m *memoryStore) Get(userID string) (*Account, error) {
	a, ok := m.account[userID]
	if !ok {
		return nil, ErrNotFound
	}
	return a, nil
}

func (m *memoryStore) Upsert(a *Account) error {
	m.account[a.UserID] = a
	return nil
}

func (m *memoryStore) AppendLedger(e LedgerEntry) error {
	m.ledger = append(m.ledger, e)
	return nil
}

func (m *memoryStore) ListLedger(userID string) ([]LedgerEntry, error) {
	out := make([]LedgerEntry, 0)
	for _, e := range m.ledger {
		if e.UserID == userID {
			out = append(out, e)
		}
	}
	return out, nil
}

func (m *memoryStore) PutHold(h *Hold) error {
	m.holds[h.ID] = h
	return nil
}

func (m *memoryStore) GetHold(id string) (*Hold, error) {
	h, ok := m.holds[id]
	if !ok {
		return nil, ErrNotFound
	}
	return h, nil
}

func (m *memoryStore) DeleteHold(id string) error {
	delete(m.holds, id)
	return nil
}

// pgStore implements Store on Postgres.
type pgStore struct{ db *sql.DB }

// NewPGStore returns wallet persistence backed by Postgres.
func NewPGStore(db *sql.DB) Store { return &pgStore{db: db} }

const upsertAccountSQL = `
	INSERT INTO wallet_account (user_id, balance_subunits, held_subunits, currency, updated_at)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (user_id) DO UPDATE SET
		balance_subunits = EXCLUDED.balance_subunits,
		held_subunits    = EXCLUDED.held_subunits,
		currency         = EXCLUDED.currency,
		updated_at       = EXCLUDED.updated_at`

func (p *pgStore) Get(userID string) (*Account, error) {
	cus, held := int64(0), int64(0)
	cur := ""
	var updated time.Time
	err := p.db.QueryRow(
		`SELECT balance_subunits, held_subunits, currency, updated_at
		   FROM wallet_account WHERE user_id = $1`, userID,
	).Scan(&cus, &held, &cur, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &Account{
		UserID:    userID,
		Balance:   money.Money{Subunits: cus, Currency: cur},
		Held:      money.Money{Subunits: held, Currency: cur},
		UpdatedAt: updated,
	}, nil
}

func (p *pgStore) Upsert(a *Account) error {
	if a.UpdatedAt.IsZero() {
		a.UpdatedAt = time.Now()
	}
	_, err := p.db.Exec(upsertAccountSQL,
		a.UserID, a.Balance.Subunits, a.Held.Subunits, a.Balance.Currency, a.UpdatedAt)
	return err
}

func (p *pgStore) AppendLedger(e LedgerEntry) error {
	_, err := p.db.Exec(
		`INSERT INTO wallet_ledger (id, user_id, type, amount_subunits, currency, reference, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		e.ID, e.UserID, e.Type, e.Amount.Subunits, e.Amount.Currency, e.Reference, e.CreatedAt)
	return err
}

func (p *pgStore) ListLedger(userID string) ([]LedgerEntry, error) {
	rows, err := p.db.Query(
		`SELECT id, user_id, type, amount_subunits, currency, reference, created_at
		   FROM wallet_ledger WHERE user_id = $1 ORDER BY created_at, id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []LedgerEntry{}
	for rows.Next() {
		var e LedgerEntry
		if err := rows.Scan(&e.ID, &e.UserID, &e.Type, &e.Amount.Subunits, &e.Amount.Currency, &e.Reference, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (p *pgStore) PutHold(h *Hold) error {
	_, err := p.db.Exec(
		`INSERT INTO wallet_hold (id, user_id, amount_subunits, currency, reference, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		h.ID, h.UserID, h.Amount.Subunits, h.Amount.Currency, h.Reference, h.CreatedAt)
	return err
}

func (p *pgStore) GetHold(id string) (*Hold, error) {
	var (
		h       Hold
		sub     int64
		cur, rf string
	)
	err := p.db.QueryRow(
		`SELECT id, user_id, amount_subunits, currency, reference, created_at
		   FROM wallet_hold WHERE id = $1`, id,
	).Scan(&h.ID, &h.UserID, &sub, &cur, &rf, &h.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	h.Amount = money.Money{Subunits: sub, Currency: cur}
	h.Reference = rf
	return &h, nil
}

func (p *pgStore) DeleteHold(id string) error {
	_, err := p.db.Exec(`DELETE FROM wallet_hold WHERE id = $1`, id)
	return err
}

var _ Store = (*memoryStore)(nil)
var _ Store = (*pgStore)(nil)
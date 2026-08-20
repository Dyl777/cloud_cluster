package identity

import (
	"database/sql"
	"errors"
	"strings"
)

// ErrNotFound is returned by a UserStore when an account is absent.
var ErrNotFound = errors.New("identity: user not found")

// UserStore persists registered accounts. memoryUserStore keeps the default
// single-process behavior; pgUserStore persists to Postgres.
type UserStore interface {
	GetByEmail(email string) (User, error)
	Insert(u User) error
	UpdateRole(email string, role Role) (User, error)
	List() ([]User, error)
}

// memoryUserStore implements UserStore in-process.
type memoryUserStore struct{ users map[string]User }

func newMemoryUserStore() *memoryUserStore {
	return &memoryUserStore{users: make(map[string]User)}
}

func (m *memoryUserStore) GetByEmail(email string) (User, error) {
	u, ok := m.users[strings.ToLower(email)]
	if !ok {
		return User{}, ErrNotFound
	}
	return u, nil
}

func (m *memoryUserStore) Insert(u User) error {
	m.users[u.Email] = u
	return nil
}

func (m *memoryUserStore) UpdateRole(email string, role Role) (User, error) {
	key := strings.ToLower(email)
	u, ok := m.users[key]
	if !ok {
		return User{}, ErrNotFound
	}
	u.Role = role
	m.users[key] = u
	return u, nil
}

func (m *memoryUserStore) List() ([]User, error) {
	out := make([]User, 0, len(m.users))
	for _, u := range m.users {
		out = append(out, u)
	}
	return out, nil
}

// pgUserStore implements UserStore on Postgres.
type pgUserStore struct{ db *sql.DB }

// NewPGUserStore returns identity persistence backed by Postgres.
func NewPGUserStore(db *sql.DB) UserStore { return &pgUserStore{db: db} }

func scanUser(row interface{ Scan(...any) error }) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	u.Email = strings.ToLower(u.Email)
	return u, nil
}

const userCols = "id, email, name, role, created_at"

func (p *pgUserStore) GetByEmail(email string) (User, error) {
	return scanUser(p.db.QueryRow(
		"SELECT "+userCols+" FROM identity_user WHERE email = $1", strings.ToLower(email)))
}

func (p *pgUserStore) Insert(u User) error {
	_, err := p.db.Exec(
		"INSERT INTO identity_user ("+userCols+") VALUES ($1, $2, $3, $4, $5)",
		u.ID, strings.ToLower(u.Email), u.Name, string(u.Role), u.CreatedAt)
	return err
}

func (p *pgUserStore) UpdateRole(email string, role Role) (User, error) {
	row := p.db.QueryRow(
		"UPDATE identity_user SET role = $1 WHERE email = $2 RETURNING "+userCols,
		string(role), strings.ToLower(email))
	return scanUser(row)
}

func (p *pgUserStore) List() ([]User, error) {
	rows, err := p.db.Query("SELECT " + userCols + " FROM identity_user ORDER BY created_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

var _ UserStore = (*memoryUserStore)(nil)
var _ UserStore = (*pgUserStore)(nil)
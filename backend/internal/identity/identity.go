package identity

import (
	"errors"
	"strings"
	"time"
)

var ErrDuplicateEmail = errors.New("email already registered")
var ErrUnknownEmail = errors.New("email not registered")
var ErrInvalidRole = errors.New("invalid role")

// Role identifies an account's privilege tier.
type Role string

const (
	RoleUser       Role = "user"
	RoleAdmin      Role = "admin"
	RoleSuperadmin Role = "superadmin"
)

// Roles is the set of assignable roles, ordered ascending by privilege.
var Roles = []Role{RoleUser, RoleAdmin, RoleSuperadmin}

// Valid reports whether r is a known role.
func (r Role) Valid() bool {
	switch r {
	case RoleUser, RoleAdmin, RoleSuperadmin:
		return true
	}
	return false
}

// AtLeast reports whether r has at least as much privilege as min.
func (r Role) AtLeast(min Role) bool {
	rank := map[Role]int{RoleUser: 0, RoleAdmin: 1, RoleSuperadmin: 2}
	return rank[r] >= rank[min]
}

// User is a registered account.
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// Service stores users behind a UserStore.
type Service struct {
	users UserStore
}

// New returns a Service backed by in-memory storage (single process).
func New() *Service { return &Service{users: newMemoryUserStore()} }

// NewPG returns a Service persisted to Postgres.
func NewPG(store UserStore) *Service { return &Service{users: store} }

func (s *Service) seedUser(id, email, name string, role Role) User {
	email = strings.ToLower(email)
	if u, err := s.users.GetByEmail(email); err == nil {
		return u
	}
	u := User{ID: id, Email: email, Name: name, Role: role, CreatedAt: time.Now()}
	_ = s.users.Insert(u)
	return u
}

// Seed registers the bootstrap superadmin (no-op if already present).
func (s *Service) Seed(id, email, name string) User {
	return s.seedUser(id, email, name, RoleSuperadmin)
}

// Register adds a user with the default user role and returns it.
func (s *Service) Register(id, email, name string) (User, error) {
	return s.RegisterRole(id, email, name, RoleUser)
}

// RegisterRole adds a user with an explicit role and returns it.
func (s *Service) RegisterRole(id, email, name string, role Role) (User, error) {
	if !role.Valid() {
		return User{}, ErrInvalidRole
	}
	email = strings.ToLower(email)
	if _, err := s.users.GetByEmail(email); err == nil {
		return User{}, ErrDuplicateEmail
	}
	u := User{ID: id, Email: email, Name: name, Role: role, CreatedAt: time.Now()}
	if err := s.users.Insert(u); err != nil {
		return User{}, err
	}
	return u, nil
}

// Get returns a user by email.
func (s *Service) Get(email string) (User, bool) {
	u, err := s.users.GetByEmail(email)
	if err != nil {
		return User{}, false
	}
	return u, true
}

// SetRole updates an existing user's role.
func (s *Service) SetRole(email string, role Role) (User, error) {
	if !role.Valid() {
		return User{}, ErrInvalidRole
	}
	u, err := s.users.UpdateRole(email, role)
	if errors.Is(err, ErrNotFound) {
		return User{}, ErrUnknownEmail
	}
	if err != nil {
		return User{}, err
	}
	return u, nil
}

// List returns every registered user.
func (s *Service) List() []User {
	out, err := s.users.List()
	if err != nil {
		return []User{}
	}
	return out
}

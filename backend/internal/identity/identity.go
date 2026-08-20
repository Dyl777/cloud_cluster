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

// Service stores users in memory.
type Service struct {
	users map[string]User
}

// New returns a Service seeded with the platform superadmin.
func New() *Service {
	return &Service{users: make(map[string]User)}
}

// Seed registers the bootstrap superadmin (no-op if already present).
func (s *Service) Seed(id, email, name string) User {
	key := strings.ToLower(email)
	if u, ok := s.users[key]; ok {
		return u
	}
	u := User{ID: id, Email: email, Name: name, Role: RoleSuperadmin, CreatedAt: time.Now()}
	s.users[key] = u
	return u
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
	key := strings.ToLower(email)
	if _, ok := s.users[key]; ok {
		return User{}, ErrDuplicateEmail
	}
	u := User{ID: id, Email: email, Name: name, Role: role, CreatedAt: time.Now()}
	s.users[key] = u
	return u, nil
}

// Get returns a user by email.
func (s *Service) Get(email string) (User, bool) {
	u, ok := s.users[strings.ToLower(email)]
	return u, ok
}

// SetRole updates an existing user's role.
func (s *Service) SetRole(email string, role Role) (User, error) {
	if !role.Valid() {
		return User{}, ErrInvalidRole
	}
	key := strings.ToLower(email)
	u, ok := s.users[key]
	if !ok {
		return User{}, ErrUnknownEmail
	}
	u.Role = role
	s.users[key] = u
	return u, nil
}

// List returns every registered user.
func (s *Service) List() []User {
	out := make([]User, 0, len(s.users))
	for _, u := range s.users {
		out = append(out, u)
	}
	return out
}

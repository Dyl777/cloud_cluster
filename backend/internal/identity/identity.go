package identity

import (
	"errors"
	"strings"
	"time"
)

var ErrDuplicateEmail = errors.New("email already registered")

// User is a registered account.
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Service stores users in memory.
type Service struct {
	users map[string]User
}

// New returns an empty Service.
func New() *Service {
	return &Service{users: make(map[string]User)}
}

// Register adds a user and returns it.
func (s *Service) Register(id, email, name string) (User, error) {
	key := strings.ToLower(email)
	if _, ok := s.users[key]; ok {
		return User{}, ErrDuplicateEmail
	}
	u := User{ID: id, Email: email, Name: name, CreatedAt: time.Now()}
	s.users[key] = u
	return u, nil
}

// Get returns a user by email.
func (s *Service) Get(email string) (User, bool) {
	u, ok := s.users[strings.ToLower(email)]
	return u, ok
}

package identity

import (
	"errors"
	"testing"
)

func TestRegisterGetSetRole(t *testing.T) {
	svc := New()
	svc.Seed("usr-1", "admin@gpuhub.dev", "SA")

	u, err := svc.Register("usr-2", "bob@example.com", "Bob")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if u.Role != RoleUser {
		t.Fatalf("role = %s, want user", u.Role)
	}

	if _, err := svc.Register("usr-3", "BOB@example.com", "Bob2"); !errors.Is(err, ErrDuplicateEmail) {
		t.Fatalf("want ErrDuplicateEmail, got %v", err)
	}

	g, ok := svc.Get("bob@example.com")
	if !ok || g.ID != "usr-2" || g.Role != RoleUser {
		t.Fatalf("get = %+v ok=%v", g, ok)
	}

	role, err := svc.SetRole("bob@example.com", RoleAdmin)
	if err != nil || role.Role != RoleAdmin {
		t.Fatalf("setrole = %+v err=%v", role, err)
	}

	if _, err := svc.SetRole("nobody@example.com", RoleAdmin); !errors.Is(err, ErrUnknownEmail) {
		t.Fatalf("want ErrUnknownEmail, got %v", err)
	}
	if _, err := svc.RegisterRole("usr-4", "x@example.com", "X", "guest"); !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("want ErrInvalidRole, got %v", err)
	}
	if _, err := svc.Register("usr-5", "x@example.com", "X"); err != nil {
		t.Fatalf("register should succeed, got %v", err)
	}
}

func TestSeedIsIdempotent(t *testing.T) {
	svc := New()
	got := []User{
		svc.Seed("usr-a", "admin@gpuhub.dev", "SA"),
		svc.Seed("usr-b", "admin@gpuhub.dev", "SA"),
	}
	if got[0].ID != got[1].ID || len(svc.List()) != 1 {
		t.Fatalf("seed not idempotent: %+v list=%d", got, len(svc.List()))
	}
}
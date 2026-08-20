package identity

import (
	"os"
	"testing"

	"github.com/gpuhub/cloud/internal/shared/db"
)

// TestPostgresUserStore exercises the pg UserStore against a live Postgres.
// Skipped unless DATABASE_URL is set.
func TestPostgresUserStore(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set")
	}
	sqlDB, err := db.Connect(url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer sqlDB.Close()
	if err := db.MigrateAll(sqlDB); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	svc := NewPG(NewPGUserStore(sqlDB))
	email := "pg-test-" + t.Name() + "@example.com"
	if _, err := svc.Register("u1", email, "PG User"); err != nil {
		t.Fatalf("register: %v", err)
	}
	got, ok := svc.Get(email)
	if !ok || got.ID != "u1" {
		t.Fatalf("reloaded get = %+v ok=%v", got, ok)
	}
	if _, err := svc.SetRole(email, RoleAdmin); err != nil {
		t.Fatalf("setrole: %v", err)
	}
	if u, _ := svc.Get(email); u.Role != RoleAdmin {
		t.Fatalf("role after reload = %s", u.Role)
	}
}
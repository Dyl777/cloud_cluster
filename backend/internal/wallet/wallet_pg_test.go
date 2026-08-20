package wallet

import (
	"os"
	"testing"

	"github.com/gpuhub/cloud/internal/shared/db"
	"github.com/gpuhub/cloud/internal/shared/money"
)

// TestPostgresPersistence exercises the pg Store end-to-end against a live
// Postgres. Skipped unless DATABASE_URL (schema owned by this platform) is set.
func TestPostgresPersistence(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set")
	}
	sqlDB, err := db.Connect(url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer sqlDB.Close()
	if err := db.MigrateAll(url); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	svc := NewPG(NewPGStore(sqlDB))
	uid := "pg-wallet-" + t.Name()
	if err := svc.Credit("l1", uid, "USD", money.FromUnits(7.5, "USD"), "topup"); err != nil {
		t.Fatalf("credit: %v", err)
	}
	hold, err := svc.Hold("h1", uid, "USD", money.FromUnits(2.5, "USD"), "rent")
	if err != nil {
		t.Fatalf("hold: %v", err)
	}
	if err := svc.Settle(hold.ID); err != nil {
		t.Fatalf("settle: %v", err)
	}

	// A brand-new Service over the same database must see the persisted state.
	again := NewPG(NewPGStore(sqlDB))
	b := again.Balance(uid, "USD")
	if b.Balance.Subunits != 5_000_000 || b.Held.Subunits != 0 {
		t.Fatalf("reloaded bal=%d held=%d", b.Balance.Subunits, b.Held.Subunits)
	}
	if lg := again.Ledger(uid); len(lg) != 3 {
		t.Fatalf("reloaded ledger len = %d, want 3", len(lg))
	}
}
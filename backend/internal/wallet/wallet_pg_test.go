package wallet

import (
	"fmt"
	"os"
	"testing"
	"time"

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
	t.Cleanup(func() { sqlDB.Close() }) // registered first, runs last
	if err := db.MigrateAll(url); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	svc := NewPG(NewPGStore(sqlDB))
	uid := fmt.Sprintf("pg-wallet-%d", time.Now().UnixNano())
	lid := "l-" + uid
	hid := "h-" + uid
	t.Cleanup(func() {
		for _, q := range []string{
			"DELETE FROM wallet_ledger WHERE user_id = $1",
			"DELETE FROM wallet_hold WHERE user_id = $1",
			"DELETE FROM wallet_account WHERE user_id = $1",
		} {
			_, _ = sqlDB.Exec(q, uid)
		}
	})

	if err := svc.Credit(lid, uid, "USD", money.FromUnits(7.5, "USD"), "topup"); err != nil {
		t.Fatalf("credit: %v", err)
	}
	hold, err := svc.Hold(hid, uid, "USD", money.FromUnits(2.5, "USD"), "rent")
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
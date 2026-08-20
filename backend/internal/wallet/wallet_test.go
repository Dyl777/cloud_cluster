package wallet

import (
	"errors"
	"testing"

	"github.com/gpuhub/cloud/internal/shared/money"
)

func TestCreditHoldSettle(t *testing.T) {
	svc := New()
	if err := svc.Credit("l1", "u1", "USD", money.FromUnits(10, "USD"), "topup"); err != nil {
		t.Fatalf("credit: %v", err)
	}
	if b := svc.Balance("u1", "USD"); b.Balance.Subunits != 10_000_000 {
		t.Fatalf("balance = %d, want 10000000", b.Balance.Subunits)
	}
	hold, err := svc.Hold("h1", "u1", "USD", money.FromUnits(4, "USD"), "rent")
	if err != nil {
		t.Fatalf("hold: %v", err)
	}
	b := svc.Balance("u1", "USD")
	if b.Balance.Subunits != 6_000_000 || b.Held.Subunits != 4_000_000 {
		t.Fatalf("after hold bal=%d held=%d", b.Balance.Subunits, b.Held.Subunits)
	}
	if err := svc.Settle(hold.ID); err != nil {
		t.Fatalf("settle: %v", err)
	}
	b = svc.Balance("u1", "USD")
	if b.Balance.Subunits != 6_000_000 || b.Held.Subunits != 0 {
		t.Fatalf("after settle bal=%d held=%d", b.Balance.Subunits, b.Held.Subunits)
	}
	if lg := svc.Ledger("u1"); len(lg) != 3 {
		t.Fatalf("ledger len = %d, want 3", len(lg))
	}
}

func TestHoldReleaseAndInsufficient(t *testing.T) {
	svc := New()
	if err := svc.Credit("l1", "u1", "USD", money.FromUnits(5, "USD"), "topup"); err != nil {
		t.Fatalf("credit: %v", err)
	}
	if _, err := svc.Hold("h1", "u1", "USD", money.FromUnits(99, "USD"), "rent"); !errors.Is(err, ErrInsufficient) {
		t.Fatalf("want ErrInsufficient, got %v", err)
	}
	hold, err := svc.Hold("h1", "u1", "USD", money.FromUnits(5, "USD"), "rent")
	if err != nil {
		t.Fatalf("hold: %v", err)
	}
	if err := svc.Release(hold.ID); err != nil {
		t.Fatalf("release: %v", err)
	}
	b := svc.Balance("u1", "USD")
	if b.Balance.Subunits != 5_000_000 || b.Held.Subunits != 0 {
		t.Fatalf("after release bal=%d held=%d", b.Balance.Subunits, b.Held.Subunits)
	}
	if err := svc.Release("missing-hold"); !errors.Is(err, ErrNoHold) {
		t.Fatalf("want ErrNoHold for missing hold, got %v", err)
	}
}

func TestLedgerIsPerUser(t *testing.T) {
	svc := New()
	_ = svc.Credit("l1", "u1", "USD", money.FromUnits(1, "USD"), "topup")
	_ = svc.Credit("l2", "u2", "USD", money.FromUnits(2, "USD"), "topup")
	if lg := svc.Ledger("u1"); len(lg) != 1 || lg[0].UserID != "u1" {
		t.Fatalf("u1 ledger = %+v", lg)
	}
	if lg := svc.Ledger("u2"); len(lg) != 1 || lg[0].Amount.Subunits != 2_000_000 {
		t.Fatalf("u2 ledger = %+v", lg)
	}
}
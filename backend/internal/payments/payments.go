// Package payments orchestrates user top-ups (cash-in) across multiple
// rails and returns intents that adapters (simulated) resolve. Confirmed
// payments credit the wallet via the REST /credit endpoint.
package payments

import (
	"time"
)

// Method is a payment rail a user can top up with.
type Method string

const (
	MethodMobileMoney Method = "mobile_money" // carrier global account (VM-bridged)
	MethodBank        Method = "bank"         // global bank account / fintech
	MethodFintech     Method = "fintech"      // third-party fintech app
)

// TopupRequest is a user-initiated cash-in.
type TopupRequest struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Method      Method    `json:"method"`
	Subunits    int64     `json:"subunits"`
	Currency    string    `json:"currency"`
	Phone       string    `json:"phone,omitempty"`
	Status      string    `json:"status"` // pending, confirmed, failed
	Provider    string    `json:"provider"`
	Reference   string    `json:"reference"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

// PaymentIntent is an external charge handle returned by an adapter.
type PaymentIntent struct {
	ChargeID string `json:"charge_id"`
	Provider string `json:"provider"`
	Raw      string `json:"raw,omitempty"`
}

// Provider abstracts a single payment rail.
type Provider interface {
	Name() string
	Create(req TopupRequest) (PaymentIntent, error)
	Confirm(chargeID string) (bool, error)
}

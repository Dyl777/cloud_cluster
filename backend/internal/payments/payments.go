// Package payments orchestrates user top-ups (cash-in) across multiple
// rails and returns intents that adapters resolve. Confirmed payments credit
// the wallet via the REST /credit endpoint.
package payments

import (
	"time"
)

// Method is a payment rail a user can top up with.
type Method string

const (
	MethodMobileMoney Method = "mobile_money" // user's line: Orange / MTN via USSD
	MethodBank        Method = "bank"
	MethodFintech     Method = "fintech"
)

// TopupRequest is a user-initiated cash-in.
type TopupRequest struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	Method          Method    `json:"method"`
	PaymentMethodID string    `json:"payment_method_id,omitempty"`
	Subunits        int64     `json:"subunits"`
	Currency        string    `json:"currency"`
	Carrier         string    `json:"carrier,omitempty"`
	Phone           string    `json:"phone,omitempty"`
	RailProvider    string    `json:"rail_provider,omitempty"`
	AccountRef      string    `json:"account_ref,omitempty"`
	Status          string    `json:"status"`
	Adapter         string    `json:"adapter,omitempty"`
	Reference       string    `json:"reference"`
	CreatedAt       time.Time `json:"created_at"`
	CompletedAt     time.Time `json:"completed_at,omitempty"`
}

// PaymentIntent is an external charge handle returned by an adapter.
type PaymentIntent struct {
	ChargeID          string            `json:"charge_id"`
	Provider          string            `json:"provider"`
	Raw               string            `json:"raw,omitempty"`
	USSDCode          string            `json:"ussd_code,omitempty"`
	Carrier           string            `json:"carrier,omitempty"`
	Phone             string            `json:"phone,omitempty"`
	UserPhone         string            `json:"user_phone,omitempty"`
	AmountUnits       int64             `json:"amount_units,omitempty"`
	RoutePath         string            `json:"route_path,omitempty"`
	NodeID            string            `json:"node_id,omitempty"`
	VMJobID           string            `json:"vm_job_id,omitempty"`
	CommandID         string            `json:"command_id,omitempty"`
	TransferType      string            `json:"transfer_type,omitempty"`
	RouteReason       string            `json:"route_reason,omitempty"`
	SystemDestination SystemDestination `json:"system_destination,omitempty"`
}

// Provider abstracts a single payment rail.
type Provider interface {
	Name() string
	Create(req TopupRequest) (PaymentIntent, error)
	Confirm(chargeID string) (bool, error)
}

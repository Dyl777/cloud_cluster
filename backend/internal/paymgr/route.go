package paymgr

import (
	"github.com/gpuhub/cloud/internal/shared/payconfig"
	"github.com/gpuhub/cloud/internal/shared/rail"
)

// Path is how a top-up is executed after routing.
type Path string

const (
	PathDirect        Path = "direct"
	PathMobileGateway Path = "mobilegateway"
	PathMobileVM      Path = "mobilevm"
	PathBank          Path = "bank"
	PathFintech       Path = "fintech"
)

// RouteDecision is paymgr's routing output for a top-up.
type RouteDecision struct {
	Target            string              `json:"target"`
	Path              Path                `json:"path"`
	TransferType      string              `json:"transfer_type,omitempty"`
	NodeID            string              `json:"node_id,omitempty"`
	SIMSlot           int                 `json:"sim_slot,omitempty"`
	VMJobID           string              `json:"vm_job_id,omitempty"`
	CommandID         string              `json:"command_id,omitempty"`
	USSDCode          string              `json:"ussd_code,omitempty"`
	Reason            string              `json:"reason,omitempty"`
	UserPhone         string              `json:"user_phone,omitempty"`
	SystemDestination payconfig.Destination `json:"system_destination"`
}

// TopupInput mirrors payments.TopupRequest for routing.
type TopupInput struct {
	ID              string `json:"id"`
	UserID          string `json:"user_id"`
	Method          string `json:"method"`
	PaymentMethodID string `json:"payment_method_id,omitempty"`
	Subunits        int64  `json:"subunits"`
	Currency        string `json:"currency"`
	Carrier         string `json:"carrier,omitempty"`
	Phone           string `json:"phone,omitempty"`
	RailProvider    string `json:"rail_provider,omitempty"`
	AccountRef      string `json:"account_ref,omitempty"`
	TransferHint    string `json:"transfer_hint,omitempty"`
}

// RouteResult is returned after routing + execution kick-off.
type RouteResult struct {
	TopupID  string        `json:"topup_id"`
	Decision RouteDecision `json:"decision"`
	Message  string        `json:"message,omitempty"`
	Status   string        `json:"status"`
}

func amountUnits(subunits int64) int64 {
	if subunits <= 0 {
		return 0
	}
	return subunits / 1_000_000
}

func isSpecialCarrier(carrier string) bool {
	for _, cr := range rail.Carriers {
		if cr.ID == carrier && cr.SpecialDial {
			return true
		}
	}
	return false
}

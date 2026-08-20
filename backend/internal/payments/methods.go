package payments

import "github.com/gpuhub/cloud/internal/shared/payconfig"

// RailOption describes a selectable bank or fintech provider in the user catalog.
type RailOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PaymentCatalog is the user-facing list of payment types they may configure.
// Mirrors supported rails from the system catalog.
type PaymentCatalog struct {
	MobileMoney []MobileCarrier `json:"mobile_money"`
	Banks       []RailOption    `json:"banks"`
	Fintechs    []RailOption    `json:"fintechs"`
}

// DefaultCatalog returns payment types a user can save as methods.
func DefaultCatalog() PaymentCatalog {
	return PaymentCatalog{
		MobileMoney: MobileCarriers,
		Banks: []RailOption{
			{ID: "wire", Name: "Bank wire / ACH"},
			{ID: "sepa", Name: "SEPA transfer"},
		},
		Fintechs: []RailOption{
			{ID: "stripe", Name: "Stripe"},
			{ID: "paypal", Name: "PayPal"},
			{ID: "wise", Name: "Wise"},
		},
	}
}

// SystemPaymentConfig is the platform collection configuration (system side).
type SystemPaymentConfig = payconfig.SystemPaymentConfig

// SystemDestination is where a top-up lands on the platform.
type SystemDestination = payconfig.Destination

// DefaultSystemConfig returns platform collection accounts.
func DefaultSystemConfig() SystemPaymentConfig {
	return payconfig.DefaultSystemConfig()
}

// SavedMethod is a user-configured payment source (user side).
type SavedMethod struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	Kind       Method `json:"kind"`
	Label      string `json:"label"`
	Carrier    string `json:"carrier,omitempty"`
	Phone      string `json:"phone,omitempty"`
	Provider   string `json:"provider,omitempty"`
	AccountRef string `json:"account_ref,omitempty"`
}

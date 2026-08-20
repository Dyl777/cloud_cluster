package payconfig

import (
	"errors"
	"strings"
)

var ErrNoDestination = errors.New("no system destination for rail")

// SystemMobileAccount is a platform-owned mobile-money collection line.
type SystemMobileAccount struct {
	ID      string `json:"id"`
	Carrier string `json:"carrier"` // mtn | orange
	Number  string `json:"number"`
	Label   string `json:"label"`
	Region  string `json:"region,omitempty"`
	Primary bool   `json:"primary"`
}

// SystemBankAccount is a platform pooled bank account.
type SystemBankAccount struct {
	ID         string `json:"id"`
	Provider   string `json:"provider"` // wire | sepa
	AccountRef string `json:"account_ref"`
	Label      string `json:"label"`
	Primary    bool   `json:"primary"`
}

// SystemFintechAccount is a platform fintech collection endpoint.
type SystemFintechAccount struct {
	ID         string `json:"id"`
	Provider   string `json:"provider"` // stripe | paypal | wise
	AccountRef string `json:"account_ref"`
	Label      string `json:"label"`
	Primary    bool   `json:"primary"`
}

// SystemPaymentConfig defines where top-ups land on the platform (system side).
type SystemPaymentConfig struct {
	MobileMoney []SystemMobileAccount  `json:"mobile_money"`
	Banks       []SystemBankAccount    `json:"banks"`
	Fintechs    []SystemFintechAccount `json:"fintechs"`
}

// Destination is the resolved system collection target for a top-up.
type Destination struct {
	Kind       string `json:"kind"` // mobile_money | bank | fintech
	AccountID  string `json:"account_id"`
	Label      string `json:"label"`
	Carrier    string `json:"carrier,omitempty"`
	Number     string `json:"number,omitempty"`
	Provider   string `json:"provider,omitempty"`
	AccountRef string `json:"account_ref,omitempty"`
}

// DefaultSystemConfig returns the platform collection configuration.
func DefaultSystemConfig() SystemPaymentConfig {
	return SystemPaymentConfig{
		MobileMoney: []SystemMobileAccount{
			{ID: "sys-mtn-primary", Carrier: "mtn", Number: "670000001", Label: "Platform MTN collection", Region: "CM", Primary: true},
			{ID: "sys-orange-primary", Carrier: "orange", Number: "690000001", Label: "Platform Orange collection", Region: "CM", Primary: true},
		},
		Banks: []SystemBankAccount{
			{ID: "sys-bank-wire", Provider: "wire", AccountRef: "GPUHub Operating ••••9001", Label: "Pooled operating account", Primary: true},
			{ID: "sys-bank-sepa", Provider: "sepa", AccountRef: "DE89 •••• 4321", Label: "SEPA settlement account", Primary: false},
		},
		Fintechs: []SystemFintechAccount{
			{ID: "sys-fintech-stripe", Provider: "stripe", AccountRef: "acct_platform_stripe", Label: "Stripe platform account", Primary: true},
			{ID: "sys-fintech-paypal", Provider: "paypal", AccountRef: "paypal@gpuhub.io", Label: "PayPal platform account", Primary: false},
		},
	}
}

// ResolveDestination picks the system account money moves into for a user rail choice.
func (c *SystemPaymentConfig) ResolveDestination(method, carrier, provider string) (Destination, error) {
	method = strings.ToLower(strings.TrimSpace(method))
	carrier = strings.ToLower(strings.TrimSpace(carrier))
	provider = strings.ToLower(strings.TrimSpace(provider))

	switch method {
	case "mobile_money":
		for _, a := range c.MobileMoney {
			if a.Carrier == carrier {
				return Destination{
					Kind: "mobile_money", AccountID: a.ID, Label: a.Label,
					Carrier: a.Carrier, Number: a.Number,
				}, nil
			}
		}
		for _, a := range c.MobileMoney {
			if a.Primary {
				return Destination{
					Kind: "mobile_money", AccountID: a.ID, Label: a.Label,
					Carrier: a.Carrier, Number: a.Number,
				}, nil
			}
		}

	case "bank":
		for _, a := range c.Banks {
			if provider != "" && a.Provider == provider {
				return Destination{
					Kind: "bank", AccountID: a.ID, Label: a.Label,
					Provider: a.Provider, AccountRef: a.AccountRef,
				}, nil
			}
		}
		for _, a := range c.Banks {
			if a.Primary {
				return Destination{
					Kind: "bank", AccountID: a.ID, Label: a.Label,
					Provider: a.Provider, AccountRef: a.AccountRef,
				}, nil
			}
		}

	case "fintech":
		for _, a := range c.Fintechs {
			if provider != "" && a.Provider == provider {
				return Destination{
					Kind: "fintech", AccountID: a.ID, Label: a.Label,
					Provider: a.Provider, AccountRef: a.AccountRef,
				}, nil
			}
		}
		for _, a := range c.Fintechs {
			if a.Primary {
				return Destination{
					Kind: "fintech", AccountID: a.ID, Label: a.Label,
					Provider: a.Provider, AccountRef: a.AccountRef,
				}, nil
			}
		}
	}

	return Destination{}, ErrNoDestination
}

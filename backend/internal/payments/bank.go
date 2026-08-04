package payments

import "time"

// BankProvider models cash-in via a global bank account or a third-party
// fintech app (wire/ACH/instant transfer). Simulated.
type BankProvider struct {
	kind string // "bank" or "fintech"
}

// NewBank builds a bank or fintech provider.
func NewBank(kind string) *BankProvider {
	return &BankProvider{kind: kind}
}

func (p *BankProvider) Name() string { return p.Kind() }

func (p *BankProvider) Kind() string {
	if p.kind == "" {
		return "bank"
	}
	return p.kind
}

func (p *BankProvider) Create(req TopupRequest) (PaymentIntent, error) {
	chargeID := "bk" + itoa(time.Now().UnixNano()) + "_" + reqID(req.UserID)
	raw := "simulated " + p.Kind() + " transfer intent; funds via provider show as available"
	return PaymentIntent{ChargeID: chargeID, Provider: p.Name(), Raw: raw}, nil
}

func (p *BankProvider) Confirm(chargeID string) (bool, error) {
	_ = chargeID
	return true, nil
}

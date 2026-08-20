package payments

import (
	"time"

	"github.com/gpuhub/cloud/internal/shared/payconfig"
)

// MobileMoneyProvider handles top-ups from the user's line to a platform collection account.
type MobileMoneyProvider struct {
	sys payconfig.SystemPaymentConfig
}

func NewMobileMoney() *MobileMoneyProvider {
	return &MobileMoneyProvider{sys: payconfig.DefaultSystemConfig()}
}

func (p *MobileMoneyProvider) Name() string { return "mobile-money" }

func (p *MobileMoneyProvider) Create(req TopupRequest) (PaymentIntent, error) {
	dest, err := p.sys.ResolveDestination(string(req.Method), req.Carrier, req.RailProvider)
	if err != nil {
		return PaymentIntent{}, err
	}
	units := amountUnits(req.Subunits)
	ussd, err := BuildUSSD(req.Carrier, dest.Number, units)
	if err != nil {
		return PaymentIntent{}, err
	}
	chargeID := "mm" + itoa(time.Now().UnixNano()) + reqID(req.UserID)
	return PaymentIntent{
		ChargeID:          chargeID,
		Provider:          p.Name(),
		USSDCode:          ussd,
		Carrier:           req.Carrier,
		Phone:             req.Phone,
		AmountUnits:       units,
		UserPhone:         req.Phone,
		SystemDestination: dest,
		RoutePath:         "direct",
		Raw:               "pay from user SIM → platform " + dest.Number,
	}, nil
}

func (p *MobileMoneyProvider) Confirm(chargeID string) (bool, error) {
	_ = chargeID
	return true, nil
}

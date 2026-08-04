package payments

import (
	"errors"
	"time"
)

// MobileMoneyProvider models cash-in via a carrier global mobile-money
// account. Requests are routed through a CarrierBridge, which may forward
// to a feeder VM that proxies commands onwards to the physical machine or
// phone holding the number. All of this is simulated for the skeleton.
type MobileMoneyProvider struct {
	bridge *CarrierBridge
}

// CarrierBridge is the proxying layer between us and the carrier rails.
type CarrierBridge struct {
	UseVMProxy bool // when true, commands round-trip through the feeder VM
}

// NewMobileMoney builds a provider with the given bridge.
func NewMobileMoney(bridge *CarrierBridge) *MobileMoneyProvider {
	if bridge == nil {
		bridge = &CarrierBridge{}
	}
	return &MobileMoneyProvider{bridge: bridge}
}

func (p *MobileMoneyProvider) Name() string { return "carrier-mobile-money" }

var ErrUnreachable = errors.New("carrier bridge unreachable")

func (p *MobileMoneyProvider) Create(req TopupRequest) (PaymentIntent, error) {
	if p.bridge.UseVMProxy {
		p.bridge.vmProxyWrite(req)
	}
	chargeID := "mm" + itoa(time.Now().UnixNano()) + reqID(req.UserID)
	return PaymentIntent{ChargeID: chargeID, Provider: p.Name(), Raw: "simulated mobile-money intent"}, nil
}

func (p *MobileMoneyProvider) Confirm(chargeID string) (bool, error) {
	_ = chargeID
	return true, nil
}

func (b *CarrierBridge) vmProxyWrite(req TopupRequest) {
	// Stand-in for: POST to bridge VM → VM issues a carrier request to
	// the physical phone/number. Simulated, no network hop.
	_ = req
}

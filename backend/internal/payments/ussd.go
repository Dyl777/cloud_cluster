package payments

import "github.com/gpuhub/cloud/internal/shared/rail"

// MobileCarrier re-exports carriers from shared/rail for the payment catalog.
type MobileCarrier = rail.Carrier

// MobileCarriers supported for mobile_money top-ups.
var MobileCarriers = rail.Carriers

// BuildUSSD delegates to shared/rail.
func BuildUSSD(carrierID, phone string, amountUnits int64) (string, error) {
	return rail.BuildUSSD(carrierID, phone, amountUnits)
}

func amountUnits(subunits int64) int64 {
	if subunits <= 0 {
		return 0
	}
	return subunits / 1_000_000
}

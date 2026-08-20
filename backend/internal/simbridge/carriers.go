package simbridge

import "github.com/gpuhub/cloud/internal/shared/rail"

// Carriers re-exported for the local bridge API.
var Carriers = rail.Carriers

// BuildUSSD delegates to shared/rail.
func BuildUSSD(carrierID, phone string, amount int64) (string, error) {
	return rail.BuildUSSD(carrierID, phone, amount)
}

package mobilevm

import "time"

// TransferType is a cross-rail movement handled by mobileVM.
type TransferType string

const (
	TransferNumberToNumber     TransferType = "number_to_number"
	TransferFintechToBank      TransferType = "fintech_to_bank"
	TransferBankToMobileMoney  TransferType = "bank_to_mobilemoney"
	TransferCarrierDialup      TransferType = "carrier_dialup"
)

// TransferRequest asks mobileVM to move funds or execute a special dial-up.
type TransferRequest struct {
	Type       TransferType `json:"type"`
	FromRef    string       `json:"from_ref"`
	ToRef      string       `json:"to_ref"`
	Carrier    string       `json:"carrier,omitempty"`
	Phone      string       `json:"phone,omitempty"`
	Amount     int64        `json:"amount"`
	TopupID    string       `json:"topup_id,omitempty"`
}

// Job tracks an in-flight VM transfer.
type Job struct {
	ID         string       `json:"id"`
	Type       TransferType `json:"type"`
	Status     string       `json:"status"` // pending | running | completed | failed
	USSD       string       `json:"ussd,omitempty"`
	Message    string       `json:"message,omitempty"`
	FromRef    string       `json:"from_ref"`
	ToRef      string       `json:"to_ref"`
	Carrier    string       `json:"carrier,omitempty"`
	Phone      string       `json:"phone,omitempty"`
	Amount     int64        `json:"amount"`
	CreatedAt  time.Time    `json:"created_at"`
	FinishedAt time.Time    `json:"finished_at,omitempty"`
}

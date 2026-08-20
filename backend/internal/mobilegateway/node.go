package mobilegateway

import "time"

// SIMSlot is a SIM line on a physical gateway node.
type SIMSlot struct {
	Slot     int    `json:"slot"`
	Number   string `json:"number"`
	Carrier  string `json:"carrier"`
	Capacity int    `json:"capacity,omitempty"` // max concurrent transfers
}

// Node is a live mobilegatewayNode device hosting SIMs.
type Node struct {
	ID          string    `json:"id"`
	SIMs        []SIMSlot `json:"sims"`
	Connected   bool      `json:"connected"`
	LastSeen    time.Time `json:"last_seen"`
	PendingJobs int       `json:"pending_jobs"`
	LatencyMs   int64     `json:"latency_ms"`
	Region      string    `json:"region,omitempty"`
	CanRefund   bool      `json:"can_refund"`
}

// Command is dispatched to a node for execution.
type Command struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"` // ussd_dial | transfer | refund
	USSD         string `json:"ussd,omitempty"`
	Carrier      string `json:"carrier,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Amount       int64  `json:"amount,omitempty"`
	SIMSlot      int    `json:"sim_slot,omitempty"`
	TopupID      string `json:"topup_id,omitempty"`
	TransferType string `json:"transfer_type,omitempty"`
}

// CommandResult is reported back by the node.
type CommandResult struct {
	CommandID string `json:"command_id"`
	Message   string `json:"message"`
	Status    string `json:"status"` // completed | failed
	LatencyMs int64  `json:"latency_ms,omitempty"`
}

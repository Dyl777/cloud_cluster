package settlement

import "time"

// Payment is an outgoing transfer to a cloud provider, funded from the
// pooled corporate bank/fintech account (money-out counterpart to top-ups).
type Payment struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Provider    string    `json:"provider"`
	Amount      int64     `json:"amount"`
	Currency    string    `json:"currency"`
	BankAccount string    `json:"bank_account"`
	Status      string    `json:"status"` // reserved, paid, failed
	CreatedAt   time.Time `json:"created_at"`
}

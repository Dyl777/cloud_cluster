package payments

import (
	"context"
	"log/slog"
	"time"

	"github.com/gpuhub/cloud/internal/shared/events"
)

// PaymentEvent is the payload published on payment.created / payment.completed.
type PaymentEvent struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Method      string    `json:"method"`
	Provider    string    `json:"provider"`
	ChargeID    string    `json:"charge_id"`
	AmountUnits int64     `json:"amount_units"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	Reference   string    `json:"reference"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

// SetBus attaches an event bus; top-up lifecycle events are fanned out.
func (o *Orchestrator) SetBus(b events.Bus) { o.bus = b }

func (o *Orchestrator) emit(topic events.Topic, e PaymentEvent) {
	if o.bus == nil {
		return
	}
	if err := o.bus.Publish(context.Background(), topic, e); err != nil {
		slog.Warn("payments event publish failed", "topic", topic, "err", err)
	}
}
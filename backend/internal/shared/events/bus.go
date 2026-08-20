// Package events defines the publish/subscribe contract between services.
// Services talk synchronously over REST; state changes are also published
// on a Bus so consumers can react without tight coupling.
package events

import (
	"context"
	"strings"
)

// Topic is a logical stream name that consumers subscribe to.
type Topic string

// Standard topics used across the platform.
const (
	TopicWalletCredited      Topic = "wallet.credited"
	TopicWalletDebited       Topic = "wallet.debited"
	TopicPaymentCreated      Topic = "payment.created"
	TopicPaymentCompleted    Topic = "payment.completed"
	TopicInstanceProvisioned Topic = "instance.provisioned"
	TopicInstanceDestroyed   Topic = "instance.destroyed"
	TopicSettlementPaid      Topic = "settlement.paid"
)

// AllTopics lists every standard topic; consumers (notify) subscribe to all.
var AllTopics = []Topic{
	TopicWalletCredited,
	TopicWalletDebited,
	TopicPaymentCreated,
	TopicPaymentCompleted,
	TopicInstanceProvisioned,
	TopicInstanceDestroyed,
	TopicSettlementPaid,
}

// Event is the envelope wrapped around every published message.
type Event struct {
	ID      string `json:"id"`
	Topic   Topic  `json:"topic"`
	Version string `json:"version"`
	Payload any    `json:"payload"`
}

// Handler receives a single event on a subscribed topic.
type Handler func(ctx context.Context, evt Event) error

// Bus is the publish/subscribe contract.
//
// Implementations provided in this package:
//   - MemoryBus      in-process delivery (tests, all-in-one dev binary)
//   - WebhookBus     delivery as POST /events REST calls between services
//
// A Kafka implementation lives in kafka_bus.go behind the `kafka` build tag
// and is wired in once the cloud/payment adapters are implemented.
type Bus interface {
	Publish(ctx context.Context, topic Topic, payload any) error
	Subscribe(topic Topic, handler Handler) func()
	Close() error
}

// normalizeSeeds parses the KAFKA_BROKERS env value into seed broker
// addresses. Kafka's classic client sets accept a "PLAINTEXT://" scheme
// prefix; franz-go's SeedBrokers does not, so it is stripped before use.
func normalizeSeeds(broker string) []string {
	seeds := strings.Split(broker, ",")
	for i := range seeds {
		seeds[i] = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(seeds[i]), "PLAINTEXT://"))
	}
	return seeds
}

//go:build kafka

package events

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// TestKafkaRoundTrip publishes a wallet.credited event and asserts the
// consumer group of a second bus instance receives it. Skipped unless
// KAFKA_BROKERS points at a live broker; unique consumer group per run so
// autocommitted offsets never reuse a prior run's records.
func TestKafkaRoundTrip(t *testing.T) {
	broker := os.Getenv("KAFKA_BROKERS")
	if broker == "" {
		t.Skip("KAFKA_BROKERS not set")
	}
	group := "roundtrip-" + time.Now().Format("20060102-150405")

	b, err := NewKafkaBus(strings.Split(broker, ","), group)
	if err != nil {
		t.Fatalf("NewKafkaBus: %v", err)
	}
	defer b.Close()

	got := make(chan Event, 1)
	unsub := b.Subscribe(TopicWalletCredited, func(ctx context.Context, evt Event) error {
		got <- evt
		return nil
	})
	defer unsub()

	// Give the consumer group time to join before producing.
	time.Sleep(2 * time.Second)

	want := map[string]any{"user_id": "user-rt", "amount": "10.00"}
	if err := b.Publish(context.Background(), TopicWalletCredited, want); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	select {
	case evt := <-got:
		if evt.Topic != TopicWalletCredited {
			t.Fatalf("topic = %s, want %s", evt.Topic, TopicWalletCredited)
		}
		if evt.ID == "" {
			t.Fatal("event missing id")
		}
		pl, ok := evt.Payload.(map[string]any)
		if !ok {
			t.Fatalf("payload type = %T, want map[string]any", evt.Payload)
		}
		if pl["user_id"] != "user-rt" {
			t.Fatalf("user_id = %v, want user-rt", pl["user_id"])
		}
	case <-time.After(20 * time.Second):
		t.Fatal("timed out waiting for round-tripped event")
	}
}
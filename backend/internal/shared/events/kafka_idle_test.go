//go:build kafka

package events

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestKafkaIdleReconnect creates a bus, idles past the client's idle
// connection prune window, then publishes and expects delivery. This mirrors
// the compat gateway, which creates its bus at startup and publishes later.
func TestKafkaIdleReconnect(t *testing.T) {
	broker := os.Getenv("KAFKA_BROKERS")
	if broker == "" {
		t.Skip("KAFKA_BROKERS not set")
	}
	b, err := NewKafkaBus(normalizeSeeds(broker), "idle-"+time.Now().Format("20060102-150405"))
	if err != nil {
		t.Fatalf("NewKafkaBus: %v", err)
	}
	defer b.Close()

	got := make(chan Event, 1)
	unsub := b.Subscribe(TopicInstanceProvisioned, func(ctx context.Context, evt Event) error {
		got <- evt
		return nil
	})
	defer unsub()

	time.Sleep(40 * time.Second) // idle past the 10s connection prune

	start := time.Now()
	if err := b.Publish(context.Background(), TopicInstanceProvisioned, map[string]any{"id": "idle-1"}); err != nil {
		t.Fatalf("Publish: %v (after %v)", err, time.Since(start))
	}
	t.Logf("publish ok in %v", time.Since(start))

	select {
	case evt := <-got:
		if evt.ID == "" {
			t.Fatal("event missing id")
		}
	case <-time.After(30 * time.Second):
		t.Fatal("timed out waiting for idle-delivery")
	}
}
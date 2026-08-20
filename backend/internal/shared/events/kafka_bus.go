//go:build kafka

package events

// Kafka transport backed by franz-go. This file is only compiled when the
// `kafka` build tag is set (see Dockerfile and `go build -tags kafka`).
//
// Semantics: at-least-once. Producers publish synchronously so callers see
// broker errors; consumers use a consumer group per service and rely on
// franz-go's periodic autocommit after each polled batch.

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// KafkaBus is a Bus implemented over Apache Kafka.
type KafkaBus struct {
	producer *kgo.Client
	seeds    []string
	group    string

	mu        sync.Mutex
	consumers []*busConsumer
	closed    bool
}

type busConsumer struct {
	client *kgo.Client
	cancel context.CancelFunc
}

// NewKafkaBus dials the given seed brokers and prepares a producer client.
// Consumers are created lazily per Subscribe with the shared consumer group.
func NewKafkaBus(seeds []string, group string) (*KafkaBus, error) {
	producer, err := kgo.NewClient(kgo.SeedBrokers(seeds...))
	if err != nil {
		return nil, fmt.Errorf("kafka producer: %w", err)
	}
	return &KafkaBus{producer: producer, seeds: seeds, group: group}, nil
}

// NewFromEnv returns the event bus for a running service. With the `kafka`
// build tag and KAFKA_BROKERS set it wires Kafka; otherwise it falls back to
// an in-process MemoryBus so every service still runs standalone.
func NewFromEnv(group string) Bus {
	broker := os.Getenv("KAFKA_BROKERS")
	if broker == "" {
		return NewMemoryBus()
	}
	seeds := strings.Split(broker, ",")
	for i := range seeds {
		seeds[i] = strings.TrimSpace(seeds[i])
	}
	b, err := NewKafkaBus(seeds, group)
	if err != nil {
		slog.Warn("kafka unavailable; falling back to memory bus", "err", err)
		return NewMemoryBus()
	}
	return b
}

// Publish produces one event to the topic named after the Topic constant.
func (b *KafkaBus) Publish(ctx context.Context, topic Topic, payload any) error {
	evt := Event{
		ID:      fmt.Sprintf("evt-%d-%d", os.Getpid(), time.Now().UnixNano()),
		Topic:   topic,
		Version: "1.0",
		Payload: payload,
	}
	value, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	rec := &kgo.Record{Topic: string(topic), Key: []byte(evt.ID), Value: value}
	return b.producer.ProduceSync(ctx, rec).FirstErr()
}

// Subscribe joins the shared consumer group on the topic and runs handler for
// each event. The returned func cancels the poll loop and leaves the group.
func (b *KafkaBus) Subscribe(topic Topic, handler Handler) func() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return func() {}
	}

	ctx, cancel := context.WithCancel(context.Background())
	client, err := kgo.NewClient(
		kgo.SeedBrokers(b.seeds...),
		kgo.ConsumerGroup(b.group),
		kgo.ConsumeTopics(string(topic)),
	)
	var c *busConsumer
	if err == nil {
		c = &busConsumer{client: client, cancel: cancel}
		b.consumers = append(b.consumers, c)
		go c.loop(ctx, handler)
	} else {
		cancel()
		slog.Error("kafka consumer setup failed", "topic", topic, "err", err)
	}

	return func() {
		if c != nil {
			c.stop()
		}
	}
}

func (c *busConsumer) loop(ctx context.Context, handler Handler) {
	defer c.client.Close()
	for {
		fetches := c.client.PollFetches(ctx)
		if ctx.Err() != nil {
			return
		}
		if fetches.IsClientClosed() {
			return
		}
		fetches.EachError(func(t string, p int32, err error) {
			slog.Error("kafka consume error", "topic", t, "partition", p, "err", err)
		})
		fetches.EachRecord(func(record *kgo.Record) {
			var evt Event
			if err := json.Unmarshal(record.Value, &evt); err != nil {
				slog.Error("kafka unparsable event", "err", err)
				return
			}
			if err := handler(ctx, evt); err != nil {
				slog.Error("kafka handler error", "topic", evt.Topic, "err", err)
			}
		})
	}
}

func (c *busConsumer) stop() {
	c.cancel()
	c.client.Close()
}

// Close shuts down the producer and every active consumer.
func (b *KafkaBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil
	}
	b.closed = true
	for _, c := range b.consumers {
		c.stop()
	}
	b.consumers = nil
	b.producer.Close()
	return nil
}
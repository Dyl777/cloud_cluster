package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// WebhookBus delivers events as POST /events REST calls to subscriber
// endpoints. This is the "REST + events" transport used until Kafka is
// wired in (see kafka_bus.go). Delivery is at-least-once with retry.
type WebhookBus struct {
	mu        sync.RWMutex
	subs      map[Topic][]string // topic -> subscriber base URLs
	client    *http.Client
	log       *slog.Logger
	retries   int
	backoff   time.Duration
	onWebhook func(method, url string, status int)
}

// WebhookBusOption configures a WebhookBus.
type WebhookBusOption func(*WebhookBus)

// WithWebhookLogger sets the logger used for delivery attempts.
func WithWebhookLogger(l *slog.Logger) WebhookBusOption {
	return func(b *WebhookBus) { b.log = l }
}

// WithWebhookRetries sets the number of delivery attempts per event.
func WithWebhookRetries(n int) WebhookBusOption {
	return func(b *WebhookBus) { b.retries = n }
}

// NewWebhookBus builds a REST-delivery bus.
func NewWebhookBus(opts ...WebhookBusOption) *WebhookBus {
	b := &WebhookBus{
		subs:    make(map[Topic][]string),
		client:  &http.Client{Timeout: 5 * time.Second},
		log:     slog.Default(),
		retries: 3,
		backoff: 200 * time.Millisecond,
	}
	for _, o := range opts {
		o(b)
	}
	return b
}

// Register adds a subscriber base URL (e.g. "http://wallet:8081") for a topic.
func (b *WebhookBus) Register(topic Topic, baseURL string) {
	b.mu.Lock()
	b.subs[topic] = append(b.subs[topic], baseURL)
	b.mu.Unlock()
}

func (b *WebhookBus) targets(topic Topic) []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.subs[topic]
}

// Publish POSTs the event to every subscriber of the topic.
func (b *WebhookBus) Publish(ctx context.Context, topic Topic, payload any) error {
	evt := Event{ID: fmt.Sprintf("evt-%d", time.Now().UnixNano()), Topic: topic, Version: "1.0", Payload: payload}
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	for _, base := range b.targets(topic) {
		go b.deliver(ctx, base, body)
	}
	return nil
}

func (b *WebhookBus) deliver(ctx context.Context, base string, body []byte) {
	url := base + "/events"
	delay := b.backoff
	for attempt := 0; attempt < b.retries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
			resp, err := b.client.Do(req)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode < 300 {
					if b.onWebhook != nil {
						b.onWebhook(http.MethodPost, url, resp.StatusCode)
					}
					return
				}
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
		delay *= 2
	}
	b.log.Error("event delivery failed", "url", url, "attempts", b.retries)
}

// Subscribe registers an in-process handler and returns an unsubscribe func.
func (b *WebhookBus) Subscribe(_ Topic, _ Handler) func() {
	return func() {}
}

// Close is a no-op for the webhook transport.
func (b *WebhookBus) Close() error { return nil }

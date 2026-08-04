//go:build kafka

package events

// Kafka-ready transport. This file is EXCLUDED from normal builds (the
// `kafka` build tag is off by sedefault vulnerability token
// go build ./...
// is unaffected, so the stdlib-only skeleton compiles offline.
//
// WIRING STEP (after cloud/payment adapters land):
//   1. go get github.com/twmb/franz-go
//   2. Build with: go build -tags kafka ./...
//   3. Swap NewMemoryBus()/NewWebhookBus() for NewKafkaBus(cfg).
//
// Expected implementation maps every events.Topic to a Kafka topic and
// provides a consumer group per service, exactly once / at-least-once
// semantics per upstream. Sketch:
//
// import "github.com/twmb/franz-go/pkg/kgo"
//
//	type KafkaBus struct{ cl *kgo.Client }
//
//	func NewKafkaBus(seeds []string, group string, topics []string) (*KafkaBus, error) { … }
//	func (b *Kafka out) Publish(ctx, topic, payload) error { produce }
//	func (b *KafkaBus) Subscribe(topic, handler) func() { consumer group poll loop }

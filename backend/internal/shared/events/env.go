//go:build !kafka

package events

// NewFromEnv returns the event bus for a running service. Without the `kafka`
// build tag the platform runs on the in-process MemoryBus so every service
// builds and runs standalone; the tagged variant (kafka_bus.go) switches to
// Kafka when KAFKA_BROKERS is set.
func NewFromEnv(group string) Bus { return NewMemoryBus() }
module github.com/gpuhub/cloud

go 1.24

// Core skeleton is stdlib-only so every service builds offline.
// Kafka transport (services to be wired after cloud+payment adapters
// are implemented) lives behind the `kafka` build tag in
// internal/shared/events/kafka_bus.go and pulls github.com/twmb/franz-go.

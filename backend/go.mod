module github.com/gpuhub/cloud

go 1.24

// The platform runs without a database by default (in-memory stores) so every
// service still builds and runs standalone. When a DATABASE_URL is supplied
// at runtime, the services it wires persist to Postgres (github.com/jackc/pgx
// is the only external dependency). Kafka transport lives behind the `kafka`
// build tag in internal/shared/events/kafka_bus.go and pulls franz-go.

require github.com/jackc/pgx/v5 v5.7.2

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)

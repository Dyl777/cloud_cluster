# GPU Cloud Platform — Backend (Go microservices)

A multi-rail GPU hosting platform. Users **top up** money through three
payment rails (user mobile-money via USSD on their SIM, bank, or fintech); the platform then **sources GPU
compute from existing cloud providers** via adapters and **pays those
providers from its pooled bank/fintech account**.

Each responsibility lives in its own microservice (see `cmd/<service>`).
Go stdlib only — the whole thing builds and runs offline with simulated
adapters.

```
backend/
  cmd/<service>/main.go     one binary per microservice
  internal/<service>/…       service domain + handlers + adapters
  internal/shared/…          events, money, ids, http kit
  Dockerfile · docker-compose.yml · Makefile
```

## Services

| Service      | Port | Responsibility |
|--------------|------|----------------|
| identity     | 8081 | accounts (register) |
| wallet       | 8082 | balances, holds, immutable ledger |
| payments     | 8083 | top-up orchestration (delegates routing to paymgr) |
| marketplace  | 8084 | GPU availability aggregation from cloud providers |
| provision    | 8085 | instance lifecycle (create/stop/destroy) |
| settlement   | 8086 | money-out: pays cloud providers from pooled bank account |
| paymgr       | 8089 | backend payment manager — routes top-ups to rails |
| mobilevm     | 8090 | cross-rail transfers (number↔number, fintech↔bank, dial-up) |
| mobilegateway| 8091 | physical SIM node registry, dispatch, load balance |
| notify       | 8088 | event/webhook ingestion |
| compat       | 8092 | Vast.ai-compatible `/api/v0` gateway (vast-cli / SDK / skypilot) |

**simbridge** (`go run ./cmd/simbridge`, port 9090) runs on the **user's
device** — not in docker-compose — to enter USSD codes on a phone SIM or
laptop USB modem.

## The two money/GPU loops

```mermaid
flowchart LR
  U[User] -->|top up| W[wallet]
  U -->|cash-in via| P[payments]
  P --> PM[paymgr]
  PM --> G[mobilegateway → nodes]
  PM --> VM[mobilevm]
  PM -->|direct| SB[simbridge on user device]
  MM & BK & FT -->|confirm| W
  W -->|Hold| PR[provision]
  PR --> MP[marketplace]
  MP --> CP[AWS / Azure / GCP / RunPod]
  PR -->|usage| ST[settlement]
  ST -->|pay provider| PP[Provider bank/fintech]
```

## Run

```
go build ./...
# or
make build
# postgres-backed (compose provisions `db` and sets DATABASE_URL):
docker compose up --build
```

Services that persist (identity, wallet, compat) run in-memory until
`DATABASE_URL` is set (compose does this), at which point they store to
Postgres and run schema migrations at startup (see `internal/shared/db`).

## Communication

Synchronous **REST** between services; state changes publish on an event
`Bus` (in-memory, or webhook POST `/events` between processes). A Kafka
transport ships behind the `kafka` build tag in
`internal/shared/events/kafka_bus.go` — wire it in after the payment + cloud
adapters are implemented (see its header comment).

Each service README documents its use cases, sequence, activity, deployment,
DB schema and ER diagram as Mermaid diagrams.
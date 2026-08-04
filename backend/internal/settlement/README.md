# settlement — Pay the Cloud Provider (money-out)

The counterpart to top-ups: user credits pool into our corporate
bank/fintech account; `settlement` pays each cloud provider from a specific
bank account for the underlying resource usage, on a metered schedule.
Simulated in the skeleton.

```
POST /settlements     pay provider (user_id, provider, amount, bank_account)
GET  /settlements
```

## Use case

```mermaid
flowchart LR
  PR[provision usage] --> M[Meter usage per provider]
  M --> P[Pay provider]
  P --> A[bank account A]
  P --> B[bank account B]
  A --> PF[Provider invoice paid]
  B --> PF
  M --> L[(Settlement ledger)]
```

## Sequence

```mermaid
sequenceDiagram
  participant PR as provision
  participant ST as settlement
  participant BK as corporate bank/fintech
  participant CP as cloud provider
  PR->>ST: usage meter event
  ST->>ST: compute amount due
  ST->>BK: transfer (bank account -> provider)
  BK-->>ST: settled
  ST->>CP: confirm invoice/payment applied
  CP-->>ST: receipt
  ST-->>PR: settlement.paid
```

## Activity

```mermaid
flowchart TD
  A[Usage event] --> B[Accumulate meter]
  B --> C[Schedule settlement run]
  C --> D{Provider bill due?}
  D -->|yes| E[Debit pooled bank account]
  D -->|no| F[Wait next cycle]
  E --> G[Post payment to provider]
  G --> H[Record settlement ledger]
  H --> I[Emit settlement.paid]
  style E fill:#2e7d32,color:#fff
```

## Deployment

```mermaid
flowchart LR
  PR[provision] --> ST[settlement]
  ST --> BK[corporate bank/fintech]
  BK --> CP[AWS / Azure / GCP / RunPod]
  ST -. settlement events .-> N[notify]
  W[wallet pool] --> ST
```

## DB schema

```sql
CREATE TABLE payment (
  id           TEXT PRIMARY KEY,
  user_id      TEXT NOT NULL,
  provider     TEXT NOT NULL,
  amount       INTEGER NOT NULL,      -- micro-units
  currency     TEXT NOT NULL DEFAULT 'USD',
  bank_account TEXT NOT NULL,
  status       TEXT NOT NULL,          -- reserved | paid | failed
  created_at   TIMESTAMP NOT NULL
);
CREATE INDEX idx_payment_provider ON payment(provider, created_at);

CREATE TABLE cash_pool (
  account TEXT PRIMARY KEY,
  balance INTEGER NOT NULL
);
```

## ER diagram

```mermaid
erDiagram
  CASH_POOL ||--o{ PAYMENT : "funds"
  INSTANCE }o--|| PAYMENT : "metered from"
  PROVIDER ||--o{ PAYMENT : "receives"
  CASH_POOL {
    text account PK
    int balance
  }
  PAYMENT {
    text id PK
    text provider
    int amount
    text bank_account
    text status
  }
  PROVIDER {
    text name PK
    text bank_account
  }
```

## Kafka

On Kafka wiring, `settlement` consumes `instance.provisioned/destroyed`
metering events and publishes `settlement.paid` for finance audit + notify.
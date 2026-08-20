# wallet — Balances, Holds & Ledger

Single source of truth for user money. Balances are integer micro-units
(via `internal/shared/money`). Movements are recorded to an immutable ledger.

```
GET  /wallets/{user}               balance
POST /wallets/{user}/credit        top-up/cash-in credit
POST /wallets/{user}/hold          reserve funds (rental)
POST /wallets/holds/{hold}/settle  charge a held amount
POST /wallets/holds/{hold}/release return a hold
GET  /wallets/{user}/ledger        history
```

## Use case

```mermaid
flowchart LR
  U[User] --> B[View balance]
  U --> H[Reserve funds for rental]
  U --> S[Settle usage]
  U --> R[Release un-used hold]
  P[payments] --> CR[Credit after topup] --> B
  B --> L[(Ledger)]
```

## Sequence — rental hold/settle

```mermaid
sequenceDiagram
  participant U as User
  participant PR as provision
  participant W as wallet
  U->>PR: start instance
  PR->>W: POST /wallets/u/hold
  W-->>PR: hold reserved
  PR->>W: POST /holds/{id}/settle
  W-->>PR: settled (spent)
  PR-->>U: instance running
```

## Activity

```mermaid
flowchart TD
  A[Request] --> B{Route}
  B -->|credit| C[Add funds + ledger]
  B -->|hold| D{Balance enough?}
  D -->|yes| E[Move to held + ledger]
  D -->|no| F[error insufficient_balance]
  B -->|settle| G[Released hold -> spent + ledger]
  B -->|release| H[Return to balance + ledger]
  style E fill:#2e7d32,color:#fff
  style F fill:#c62828,color:#fff
```

## Deployment

```mermaid
flowchart LR
  U[User] --> W[wallet]
  P[payments] -->|credit| W
  PR[provision] -->|hold/settle| W
  W -. ledger events .-> N[notify]
```

## DB schema (Postgres, applied by `internal/shared/db` migrations)

```sql
CREATE TABLE wallet_account (
  user_id          TEXT PRIMARY KEY,
  balance_subunits BIGINT NOT NULL,  -- micro-units
  held_subunits    BIGINT NOT NULL,
  currency         TEXT NOT NULL DEFAULT 'USD',
  updated_at       TIMESTAMPTZ NOT NULL
);

CREATE TABLE wallet_ledger (
  id         TEXT PRIMARY KEY,
  user_id    TEXT NOT NULL,
  type       TEXT NOT NULL,       -- credit | hold | release | settle | topup
  amount_subunits BIGINT NOT NULL,
  currency   TEXT NOT NULL DEFAULT 'USD',
  reference  TEXT,
  created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_wallet_ledger_user ON wallet_ledger(user_id, created_at);

CREATE TABLE wallet_hold (
  id         TEXT PRIMARY KEY,
  user_id    TEXT NOT NULL,
  amount_subunits BIGINT NOT NULL,
  currency   TEXT NOT NULL DEFAULT 'USD',
  reference  TEXT,
  created_at TIMESTAMPTZ NOT NULL
);
```

Without `DATABASE_URL` the service falls back to in-memory storage (`Store`
in `store.go`); set it to persist to Postgres.

## ER diagram

```mermaid
erDiagram
  ACCOUNT ||--o{ LEDGER_ENTRY : "posts"
  ACCOUNT ||--o{ HOLD : "reserves"
  PAYMENT ||--o{ LEDGER_ENTRY : "credits"
  HOLD ||--o{ LEDGER_ENTRY : "settles"
  ACCOUNT {
    text user_id PK
    int balance
    int held
    text currency
  }
  LEDGER_ENTRY {
    text id PK
    text user_id FK
    text type
    int amount
    timestamp created_at
  }
  HOLD {
    text id PK
    text user_id FK
    int amount
  }
```

## Kafka

`wallet` is the **credit ledger**; it consumes `payment.completed` and, on
new Kafka wiring, publishes `wallet.credited/debited` for `notify` and audit.
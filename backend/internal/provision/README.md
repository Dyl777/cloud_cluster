# provision — Instance Lifecycle

Creates, observes and destroys rented GPU instances on the cloud provider
behind the marketplace. In the skeleton, launches are simulated.

```
POST /instances        launch (user_id, gpu_name, num_gpus, provider)
GET  /instances        list
```

## Use case

```mermaid
flowchart LR
  U[User] --> C[Create instance]
  C --> O[Observe status]
  U --> S[Stop]
  U --> D[Destroy]
  C --> P[provider lease]
  O --> R[Running telemetry]
```

## Sequence — launch

```mermaid
sequenceDiagram
  participant U as User
  participant PR as provision
  participant CP as cloud provider
  participant W as wallet
  U->>PR: POST /instances
  PR->>W: hold funds
  W-->>PR: hold ok
  PR->>CP: create lease/instance
  CP-->>PR: lease id
  PR-->>U: instance (pending → running)
```

## Activity

```mermaid
flowchart TD
  A[Create] --> B[Hold wallet funds]
  B -->|ok| C[Call provider launch]
  B -->|insufficient| E[error]
  C --> D[Status pending]
  D --> F[Status running]
  F --> G{user action}
  G -->|stop| H[Status stopped]
  G -->|destroy| I[Release lease + release hold]
  H --> G
  I --> J[Event instance.destroyed]
  style I fill:#2e7d32,color:#fff
```

## Deployment

```mermaid
flowchart LR
  U[User] --> PR[provision]
  PR --> W[wallet hold]
  PR --> CP[AWS / Azure / GCP / RunPod]
  PR -. instance events .-> N[notify]
```

## DB schema

```sql
CREATE TABLE instance (
  id         TEXT PRIMARY KEY,
  user_id    TEXT NOT NULL,
  offer_id   TEXT,
  provider   TEXT NOT NULL,
  gpu_name   TEXT NOT NULL,
  num_gpus   INTEGER,
  dph_total  REAL,
  ssh_host   TEXT,
  ssh_port   INTEGER,
  status     TEXT NOT NULL,   -- pending | running | stopped | destroyed
  created_at TIMESTAMP NOT NULL
);
CREATE INDEX idx_instance_user ON instance(user_id);
```

## ER diagram

```mermaid
erDiagram
  USER ||--o{ INSTANCE : "rents"
  OFFER ||--o{ INSTANCE : "launched from"
  INSTANCE ||--|| HOLD : "funded by"
  INSTANCE {
    text id PK
    text user_id FK
    text provider
    text gpu_name
    text status
  }
  HOLD {
    text id PK
    int amount
  }
```

## Kafka

On Kafka wiring, `provision` publishes `instance.provisioned` and
`instance.destroyed`; `settlement` consumes them to meter and bill usage,
and `notify` to alert the user.
# marketplace — GPU Availability Aggregator

Sources GPU offers from existing cloud providers through adapters and
serves filtered, "in-stock" listings. Every cloud adapter returns the same
`Offer` shape, so new providers drop in behind the same interface.

```
GET /offers?gpu=H100&region=US&max_price=1.5
```

## Use case

```mermaid
flowchart LR
  U[User] --> S[Search GPU availability]
  S --> F[Filter by GPU/price/region]
  F --> P[Aggregate from providers]
  P --> A[AWS]
  P --> B[Azure]
  P --> C[GCP]
  P --> D[RunPod / CoreWeave / Lambda]
  F --> R[Pick offer and rent]
```

## Sequence

```mermaid
sequenceDiagram
  participant U as User
  participant M as marketplace
  participant AW as AWS adapter
  participant AZ as Azure adapter
  participant GC as GCP adapter
  participant RP as RunPod adapter

  U->>M: GET /offers?gpu=H100
  par fetch in parallel
    M->>AW: ListOffers()
    M->>AZ: ListOffers()
    M->>GC: ListOffers()
    M->>RP: ListOffers()
  end
  AW-->>M: offers
  AZ-->>M: offers
  GC-->>M: offers
  RP-->>M: offers
  M->>M: filter + normalize
  M-->>U: { count, offers }
```

## Activity

```mermaid
flowchart TD
  A[Query received] --> B{Query params}
  B --> C[Fetch each provider]
  C --> D[Apply gpu/region/price filters]
  D --> E{Any offers?}
  E -->|yes| F[Return aggregated list]
  E -->|no| G[Return empty list]
  F --> H[User rents -> provision-svc]
```

## Deployment

```mermaid
flowchart LR
  subgraph marketplace
    M[marketplace]
  end
  M --> AW[AWS API]
  M --> AZ[Azure API]
  M --> GC[GCP API]
  M --> RP[RunPod API]
  M -. publish availability changes .-> N[notify]
```

## DB schema

```sql
CREATE TABLE offer_snapshot (
  id            TEXT PRIMARY KEY,
  provider      TEXT NOT NULL,
  gpu_name      TEXT NOT NULL,
  gpu_vram      INTEGER,
  num_gpus      INTEGER,
  cpu_cores     INTEGER,
  cpu_ram       INTEGER,
  disk_space    REAL,
  dph_total     REAL,
  region        TEXT,
  datacenter    BOOLEAN,
  verified      BOOLEAN,
  reliability2  REAL,
  fetched_at    TIMESTAMP NOT NULL
);
CREATE INDEX idx_offer_gpu ON offer_snapshot(gpu_name, region);
```

## ER diagram

```mermaid
erDiagram
  PROVIDER ||--o{ OFFER_SNAPSHOT : "aggregates"
  OFFER_SNAPSHOT {
    text id PK
    text provider FK
    text gpu_name
    int gpu_vram
    real dph_total
    text region
    timestamp fetched_at
  }
  OFFER_SNAPSHOT }o--|| INSTANCE : "rented as"
```

## Kafka

With real adapters, `marketplace` publishes `offers.refreshed` to a Kafka
topic that `provision`/`notify` consume for spot-availability alerts.
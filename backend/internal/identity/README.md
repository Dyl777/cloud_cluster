# identity — Accounts

The credibility layer: registers users. In the skeleton this is an
in-memory store; swap in Postgres + hashed credentials + JWT issues for
production.

```
POST /users    register (email, name)
```

## Use case

```mermaid
flowchart LR
  U[User] --> R[Register account]
  R --> L[Login / session]
  U --> A[Manage API keys]
  L --> S[Access wallet & instances]
```

## Sequence — register

```mermaid
sequenceDiagram
  participant U as User
  participant I as identity
  participant DB as accounts store
  U->>I: POST /users {email, name}
  I->>DB: email exists?
  DB-->I: no
  I->>DB: insert user
  DB-->>I: user
  I-->>U: 201 user
```

## Activity

```mermaid
flowchart TD
  A[Register] --> B{Email unique?}
  B -->|no| C[409 conflict]
  B -->|yes| D[Create user]
  D --> E[Return user]
  C --> F[error]
```

## DB schema / ER

```mermaid
erDiagram
  USER ||--o{ ACCOUNT : "owns wallet"
  USER ||--o{ API_KEY : "uses"
  USER {
    text id PK
    text email UK
    text name
    timestamp created_at
  }
  API_KEY {
    text id PK
    text user_id FK
    text scopes
  }
```

```sql
CREATE TABLE app_user (
  id         TEXT PRIMARY KEY,
  email      TEXT UNIQUE NOT NULL,
  name       TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL
);
CREATE TABLE api_key (
  id      TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES app_user(id),
  token   TEXT UNIQUE NOT NULL,
  scopes  TEXT NOT NULL
);
```

## Kafka

`identity` publishes `user.registered` on Kafka wiring so onboarding
services (welcome notifications) react without coupling.
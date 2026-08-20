-- wallet: balances, holds, immutable ledger.
-- Money is stored as BIGINT micro-units (1 unit = 1/1,000,000 of the currency).
CREATE TABLE IF NOT EXISTS wallet_account (
    user_id         TEXT PRIMARY KEY,
    balance_subunits BIGINT NOT NULL DEFAULT 0,
    held_subunits    BIGINT NOT NULL DEFAULT 0,
    currency         TEXT   NOT NULL DEFAULT 'USD',
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS wallet_ledger (
    id               TEXT PRIMARY KEY,
    user_id          TEXT NOT NULL,
    type             TEXT NOT NULL, -- credit | hold | release | settle | topup
    amount_subunits  BIGINT NOT NULL,
    currency         TEXT NOT NULL DEFAULT 'USD',
    reference        TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_wallet_ledger_user ON wallet_ledger(user_id, created_at);

CREATE TABLE IF NOT EXISTS wallet_hold (
    id               TEXT PRIMARY KEY,
    user_id          TEXT NOT NULL,
    amount_subunits  BIGINT NOT NULL,
    currency         TEXT NOT NULL DEFAULT 'USD',
    reference        TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
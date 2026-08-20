-- identity: registered accounts.
CREATE TABLE IF NOT EXISTS identity_user (
    id         TEXT PRIMARY KEY,
    email      TEXT NOT NULL UNIQUE, -- stored lower-cased
    name       TEXT NOT NULL,
    role       TEXT NOT NULL,        -- user | admin | superadmin
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
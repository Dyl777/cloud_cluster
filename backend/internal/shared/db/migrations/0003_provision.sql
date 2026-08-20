-- provision: rented GPU instances.
CREATE TABLE IF NOT EXISTS provision_instance (
    id            TEXT PRIMARY KEY,
    user_id       TEXT NOT NULL,
    gpu_name      TEXT NOT NULL,
    gpu_vram      BIGINT NOT NULL DEFAULT 0,   -- MB
    num_gpus      INTEGER NOT NULL DEFAULT 1,
    cpu_cores     INTEGER NOT NULL DEFAULT 0,
    cpu_ram       INTEGER NOT NULL DEFAULT 0,  -- GB
    disk_space    DOUBLE PRECISION NOT NULL DEFAULT 0, -- TB
    region        TEXT NOT NULL DEFAULT '',
    provider      TEXT NOT NULL DEFAULT '',
    image         TEXT NOT NULL DEFAULT '',
    label         TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'running',
    price         DOUBLE PRECISION NOT NULL DEFAULT 0, -- $/hour
    ssh_port      INTEGER NOT NULL DEFAULT 0,
    public_ip     TEXT NOT NULL DEFAULT '',
    start_date    BIGINT NOT NULL DEFAULT 0,   -- unix seconds
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_provision_user ON provision_instance(user_id, created_at);
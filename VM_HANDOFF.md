# VM Handoff / Dev-Loop Guide

Everything needed to recreate the dev VM and continue the integration work.
Written at the point the original VM (`13.49.49.51`) was deleted. **All code is in
the repo (GitHub, public: `Dyl777/cloud_cluster`); nothing of value lived only
on the VM.** The only things lost on VM deletion:

- the docker volume `pgdata` (wallet balances, ledger, identities, instances —
  demo data; re-create via the demo flows below)
- the ephemeral SSH tunnel for local browsing (re-create below)

## The workflow (write → push → VM test)

1. Edit code locally (Windows, VS Code).
2. `git add backend/ … && git commit -m "…" && git push origin feature/production-integrations`
3. Run the loop helper on the VM (below) to pull, rebuild, restart, and verify.

Current branch: **`feature/production-integrations`** (15 commits ahead of `main`).
REPO: https://github.com/Dyl777/cloud_cluster.git (public, so VMs clone without keys;
switch to a PAT/deploy key before this product is in production).

## 1. Spin up a new VM (same shape as before)

Anything ≥ 2 vCPU / 4 GB / Ubuntu 22.04+ works; original was t3.large-ish
(2 vCPU, 7.6 GB, Ubuntu 26.04 LTS, ~35 GB disk). Allow inbound TCP 22, 4173,
8081–8092, 9092 in the security group.

SSH key (Windows): `C:\Users\AMBE\.ssh\deploy-sv.pem`
Set once per PowerShell shell: `$env:AWS_SSH_KEY = "C:\Users\AMBE\.ssh\deploy-sv.pem"`

Test: `ssh -i $env:AWS_SSH_KEY -o ConnectTimeout=15 ubuntu@<NEW_IP> "uname -a"`

## 2. Provision toolchain (idempotent, one-shot)

The repo ships `scripts/provision.sh` (installs docker, compose v2, node, git,
Go from the official tarball to `/usr/local/go`, adds `~/.profile` PATH, adds
user to `docker` group, clones the repo to `~/cloud_cluster`).

```powershell
# run it on the VM (quoted through PowerShell)
ssh -i $env:AWS_SSH_KEY ubuntu@<NEW_IP> "export PATH=\$PATH:/usr/local/go/bin; curl -fsSL -o /tmp/provision.sh https://raw.githubusercontent.com/Dyl777/cloud_cluster/main/scripts/provision.sh && bash /tmp/provision.sh"
```

Notes:
- After provisioning, **log out and back in** (or `newgrp docker`) so the
  `docker` group applies.
- `provision.sh` clones `main`; you must checkout the feature branch.

## 3. Check out the feature branch

```bash
cd ~/cloud_cluster
git checkout feature/production-integrations
git pull
```

## 4. Run the stack

```bash
cd ~/cloud_cluster/backend
docker compose up -d --build   # builds with -tags kafka, starts db + kafka + all services
docker compose ps              # all 12 services Up, db + kafka "healthy"
```

Services & ports (see `backend/README.md`): identity 8081, wallet 8082,
payments 8083, marketplace 8084, provision 8085, settlement 8086, paymgr 8089,
mobilevm 8090, mobilegateway 8091, notify 8088, compat 8092.
`simbridge` (9090) runs on the user's device, not in compose.
Compose also runs `postgres:16-alpine` (db, 5432) and `apache/kafka:3.8.1`
(kafka, 9092).

## 5. Loop helper (Windows)

`scripts/deploy.ps1 "<remote bash command>"` wraps ssh with the key. Default
host `ubuntu@13.49.49.51` — override with `-Host_`:

```powershell
. .\scripts\deploy.ps1 "cd ~/cloud_cluster/backend && git pull --ff-only && docker compose up -d --build"
```

**Quoting gotcha (critical):** PowerShell + ssh mangle quotes, `|`, `$`, and
`;`. For anything beyond one simple command, write the script to a temp file,
`scp` it, and run it:

```powershell
scp -i "C:\Users\AMBE\.ssh\deploy-sv.pem" C:\Users\AMBE\AppData\Local\Temp\opencode\myscript.sh ubuntu@<NEW_IP>:/tmp/myscript.sh
. .\scripts\deploy.ps1 "bash /tmp/myscript.sh"
```

Prefer LF line endings in those scripts (strip CRLF with `sed -i 's/\r$//'` or
write them with the Write tool).

## 6. Local browser tunnel (optional, for frontend)

```powershell
Start-Process ssh -ArgumentList "-i","C:\Users\AMBE\.ssh\deploy-sv.pem","-N","-L","4173:localhost:4173","-L","8083:localhost:8083","ubuntu@<NEW_IP>"
# browser: http://localhost:4173 (vite preview) / http://localhost:8083 (payments)
```

## 7. Verification commands (use `curl` on the VM)

Demo auth for compat: `Authorization: Bearer demo-api-key` (`COMPAT_API_KEY`),
user id `user-demo` (`COMPAT_USER_ID`).

```bash
# wallet credit / debit via Kafka → notify feed
curl -s -X PUT http://localhost:8092/api/v0/topup -H "Authorization: Bearer demo-api-key" -H "Content-Type: application/json" -d '{"amount":7,"currency":"USD"}'

# instance lifecycle (offers are mock-0..mock-N from GET /api/v0/search/asks)
curl -s -X PUT http://localhost:8092/api/v0/asks/mock-0 -H "Authorization: Bearer demo-api-key" -H "Content-Type: application/json" -d '{"image":"pytorch/pytorch","disk":16,"price":0.5,"label":"demo"}'
curl -s -X DELETE http://localhost:8092/api/v0/instances/mock-0 -H "Authorization: Bearer demo-api-key" -H "Content-Type: application/json" -d '{}'

# payments (units = subunits/1_000_000, so use >= 1,000,000)
curl -s -X POST http://localhost:8083/payments/topup -H "Content-Type: application/json" -d '{"user_id":"user-demo","method":"mobile_money","subunits":2000000,"currency":"USD","carrier":"MTN","phone":"+255700000001"}'
curl -s -X POST http://localhost:8083/payments/<topup_id>/confirm

# settlement money-out
curl -s -X POST http://localhost:8086/settlements -H "Content-Type: application/json" -d '{"user_id":"user-demo","provider":"nebius","amount":120000,"currency":"USD","bank_account":"corp-main"}'

# notify feed (should show wallet.credited/debited, payment.*, instance.*, settlement.paid)
curl -s http://localhost:8088/events
```

## 8. Kafka specifics (learned the hard way)

- Broker advertises `kafka:9092`; only resolves **inside the docker network**.
  - Running the kafka tests **from the VM host fails** with
    `lookup kafka on 127.0.0.53: server misbehaving` / `context deadline exceeded`.
  - Run kafka tests **in-network** instead:
    ```bash
    cd ~/cloud_cluster/backend
    docker run --rm --network backend_default -e KAFKA_BROKERS=kafka:9092 \
      -v $PWD:/src -w /src golang:1.25-alpine \
      go test -tags kafka ./internal/shared/events/ -run 'TestKafkaRoundTrip|TestKafkaIdleReconnect' -v -count=1
    ```
- `KAFKA_BROKERS` must be plain `host:port`, **no `PLAINTEXT://` scheme prefix**
  (franz-go's `SeedBrokers` rejects it). `normalizeSeeds` in
  `backend/internal/shared/events/bus.go` strips the prefix defensively.
- `ConsumeResetOffset(AtStart())` is set so a fresh consumer group replays
  events published before it joined.
- `Publish` uses a 10s bounded context so a down broker can't wedge a
  request handler (a goroutine dump showed it holding the wallet mutex).
- PG-backed tests: `go test ./...` with `DATABASE_URL` set run migrations and
  are idempotent (unique ids + `t.Cleanup`); pool close is registered first so
  cleanup runs before the pool closes.

## 9. Current state (as of handoff)

Postgres persistence (identity + wallet + provision) and Kafka event bus are
done and VM-verified. All 7 standard topics publish and notify consumes them:
`wallet.credited`, `wallet.debited`, `payment.created`, `payment.completed`,
`instance.provisioned`, `instance.destroyed`, `settlement.paid`.

Commits on `feature/production-integrations` (oldest→newest):
`aa20521` scripts/provision → `7a401bf` identity+wallet PG → `00a61cf` migration
fixes → `d519ab6` idempotent PG tests → `e992120` settle/release ledger ids →
`df79f2f` identity test cleanup → `d71ede1` pool close ordering → `60962ed`
provision PG → `b970d6d` Kafka transport + wallet/provision emits → `f50fada`
compose dup fix → `08f7ce7` produce timeout → `946d39b` earliest offsets →
`77486b7` seed scheme fix → `b275cc9` doc → `59e9812` payments/settlement emits.

## 10. Next steps / open work

- Add a compose `test` service (runs suites against db+kafka on the VM) so the
  full suite runs in-network automatically.
- Outbox pattern in wallet/provision/payments/settlement for true
  at-least-once delivery (emit currently best-effort; failures are logged).
- Dedupe notify feed on `Event.ID` (feed is currently a plain append).
- Wire more services to the bus (e.g. identity events), add a topic per
  important state change.
- Re-verify on a fresh VM: provision → checkout branch → compose up → run the
  section-7 verification commands → top-up persists across `docker compose
  restart wallet`.
- Eventually: merge `feature/production-integrations` into `main`, and move the
  git remote to require auth for production.
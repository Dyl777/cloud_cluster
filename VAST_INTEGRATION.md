# Integrating official Vast.ai open-source tools with our platform

Our backend speaks the **console.vast.ai `/api/v0` REST protocol** through the
`compat` gateway (`backend/cmd/compat`, port 8092). Any official Vast
open-source tool that talks to `console.vast.ai` can be pointed at our
platform with two environment variables — no source changes, no forks.

## The tools

| Project | Repo | How it talks to Vast |
|---------|------|----------------------|
| `vastai` CLI | `vast-ai/vast-cli` | HTTPS to `console.vast.ai/api/v0` |
| `vastai` SDK | `vast-ai/vast-sdk` | same `/api/v0` protocol |
| `vast-pyworker` | `vast-ai/vast-pyworker` | calls instance APIs to boot worker containers |
| `skypilot` (Vast backend) | `vast-ai/skypilot` | provider driver uses `/api/v0` and `/api/v1` |
| `vast-tools` | `vast-ai/vast-tools` | misc CLI helpers over the same API |
| `vast-console` (web) | — | browser hits `/api/v0` too |

Every one of these resolves its endpoint from the same env var:

- `VAST_URL` (the SDK/CLI default to `https://console.vast.ai`)
- `VAST_API_KEY`

So setting `VAST_URL=http://<our-host>:8092` and `VAST_API_KEY=<our key>`
redirects them all to our platform.

## What is implemented today (`backend/internal/compat`)

| Capability | Route | Verified against official `vastai` 1.5.4? |
|------------|-------|:---:|
| Offer search | `POST /api/v0/bundles/` | ✅ `search_offers` |
| Offer search (new) | `PUT /api/v0/search/asks/` | ✅ |
| Templates | `GET /api/v0/template/` | ✅ |
| Create instance | `PUT /api/v0/asks/{id}` | ✅ `create_instance` |
| Bulk create | `POST /api/v0/asks/bulk` | ✅ |
| List instances | `GET /api/v0/instances/` + `GET /api/v1/instances/` | ✅ `show_instances` |
| Show instance | `GET /api/v0/instances/{id}` | ✅ |
| Start/stop | `PUT /api/v0/instances` (state) | ✅ |
| Update/label | `PUT /api/v0/instances/{id}` | ✅ |
| Destroy | `DELETE /api/v0/instances/{id}` | ✅ `destroy_instance` |
| Reboot | `PUT /api/v0/instances/reboot/{id}` | ✅ |
| Exec | `PUT /api/v0/instances/command/{id}` | ✅ |
| Instance balance | `GET /instances/balance/{id}` | ✅ |
| Current user | `GET /users/current` | ✅ `show user` |
| Billing history | `GET /users/me/invoices`, `GET /api/v0/charges/` | ✅ |
| Top-up helper | `POST /api/v0/topup` (our addition) | — |

## Verified end-to-end

We installed the **official** `vastai` package from PyPI (`pip install vastai`)
and ran it unmodified against `VAST_URL=http://localhost:8092`:

```text
vastai show user                       # balance, id, email — OK
vastai search offers 'gpu_name="RTX 3090"'   # offers — OK (via SDK)
create_instance('mock-4', image=..., disk=32, price=0.19)  # OK
show_instances()                       # row incl. ssh_port, ip — OK
destroy_instance('mock-4')             # OK
```

## How to run it

```bash
cd backend
go run ./cmd/compat
# gateway on :8092, api_key=demo-api-key
```

```bash
pip install vastai
export VAST_URL="http://localhost:8092"
export VAST_API_KEY="demo-api-key"
vastai show user
vastai search offers 'gpu_name="RTX 3090"'
vastai create instance <offer-id> --image pytorch/pytorch:latest
vastai show instances
vastai destroy instance <id>
```

## Wiring each project

### vast-cli / vast-sdk
Env var only:
```bash
export VAST_URL="http://<host>:8092"
export VAST_API_KEY="<our-key>"
```

### skypilot
SkyPilot's Vast backend reads the same env vars (`VAST_URL` /
`VAST_API_KEY`) when `--cloud vast` is selected; the driver calls
`/api/v0/bundles/` for offers, `/api/v0/asks/{id}` to launch, and
`/api/v1/instances/` for status. No code changes needed.

### vast-pyworker
The worker bootstraps instances by calling the instance API (launch, SSH
keys, env). Point `VAST_URL`/`VAST_API_KEY` at our gateway as above.

## Design notes

- `compat` maps every Vast verb onto our domain services in-process; see
  `backend/internal/compat/README.md` for the routing table and a sequence
  diagram.
- The SDK's offer search sends the filter dict as the JSON body for
  `/bundles/` but wraps it in `{"q": ...}` for `/search/asks/`; `compat`
  handles both.
- Responses follow the Vast envelope (`{"success": bool, "msg": ...}`, offer
  row field names) so CLI printers and the SDK's `raise_for_status`
  work unchanged.

## Next steps (not yet implemented)

- Real provider adapters (AWS/Azure/GCP/RunPod) feeding `marketplace.Registry`
  instead of the mock snapshot.
- Real top-up rails (`payments`, `paymgr`, `mobilegateway`) crediting the
  wallet that `compat` reads for balance.
- Kafka-backed event bus (`-tags kafka`) so instance lifecycle events fan out.
- SSH key endpoints (`/ssh-keys/`) and endpoint/proxy routes for PyWorker and
  `vastai ssh-url`.

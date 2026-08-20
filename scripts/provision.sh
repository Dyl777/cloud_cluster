#!/usr/bin/env bash
# One-shot provisioning for the dev VM (Ubuntu).
# Idempotent: safe to re-run.
set -euo pipefail

sudo apt-get update -qq
sudo apt-get install -y -qq curl ca-certificates git docker.io docker-compose-v2 nodejs npm

# Go via official tarball (installs to /usr/local/go)
if ! command -v go >/dev/null 2>&1 && [ ! -x /usr/local/go/bin/go ]; then
    GOVER=$(curl -fsSL https://go.dev/VERSION?m=text | head -1)
    curl -fsSLo /tmp/go.tgz "https://go.dev/dl/${GOVER}.linux-amd64.tar.gz"
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf /tmp/go.tgz
fi
grep -q "usr/local/go/bin" "$HOME/.profile" || echo 'export PATH=$PATH:/usr/local/go/bin' >> "$HOME/.profile"

sudo systemctl enable --now docker >/dev/null 2>&1
sudo usermod -aG docker "$USER" || true

# Repo
[ -d "$HOME/cloud_cluster/.git" ] || git clone -q https://github.com/Dyl777/cloud_cluster.git "$HOME/cloud_cluster"

echo "=== toolchain ==="
go version
docker --version
docker compose version
node --version
npm --version
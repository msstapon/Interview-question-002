#!/usr/bin/env bash
#
# One-shot VPS prep for interview-question-002.
# Idempotent — safe to re-run (won't overwrite an existing .env).
#
# Run as root on the target server:
#   sudo bash server-setup.sh
#
# Or in one line with options:
#   sudo GHCR_OWNER=msstapon SITE_ADDRESS=:80 \
#        DEPLOY_PUBKEY="ssh-ed25519 AAAA... gh-deploy" bash server-setup.sh
#
# Optional overrides (env vars):
#   GHCR_OWNER     GitHub owner hosting the GHCR images   (default: msstapon)
#   SITE_ADDRESS   domain for auto-HTTPS, or ":80" for HTTP by IP (default: :80)
#   HTTP_PORT      Caddy host HTTP port  — set if 80 is taken (e.g. keyvc on this VPS)
#   HTTPS_PORT     Caddy host HTTPS port — set if 443 is taken
#   DEPLOY_PUBKEY  CI deploy SSH *public* key to authorize (whole line)
#   APP_DIR        install dir                              (default: /opt/interview-question-002)
#   DB_USER        (default: keyvc)   DB_NAME (default: test_tcc)
#
set -euo pipefail

GHCR_OWNER="${GHCR_OWNER:-msstapon}"
SITE_ADDRESS="${SITE_ADDRESS:-:80}"
APP_DIR="${APP_DIR:-/opt/interview-question-002}"
DB_USER="${DB_USER:-keyvc}"
DB_NAME="${DB_NAME:-test_tcc}"
HTTP_PORT="${HTTP_PORT:-}"
HTTPS_PORT="${HTTPS_PORT:-}"
DEPLOY_PUBKEY="${DEPLOY_PUBKEY:-}"

log()  { printf '\n\033[1;32m==>\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33m!  %s\033[0m\n' "$*"; }

[ "$(id -u)" -eq 0 ] || { echo "Please run as root:  sudo bash server-setup.sh"; exit 1; }
command -v curl >/dev/null 2>&1 || { echo "curl is required — install it first (apt-get install -y curl)"; exit 1; }

# ── 1) Docker + compose plugin ───────────────────────────────────────────────
if command -v docker >/dev/null 2>&1; then
  log "Docker already installed: $(docker --version)"
else
  log "Installing Docker (get.docker.com)..."
  curl -fsSL https://get.docker.com | sh
fi
if ! docker compose version >/dev/null 2>&1; then
  warn "docker compose plugin not found — install docker-compose-plugin for your distro."
fi
systemctl enable --now docker >/dev/null 2>&1 || true

# ── 2) app directories ───────────────────────────────────────────────────────
log "Creating $APP_DIR ..."
mkdir -p "$APP_DIR/deployments/docker" "$APP_DIR/deployments/caddy"

# ── 3) .env (create only if missing so secrets survive re-runs) ───────────────
ENV_FILE="$APP_DIR/deployments/docker/.env"
if [ -f "$ENV_FILE" ]; then
  log ".env already exists — leaving it untouched ($ENV_FILE)"
else
  log "Generating $ENV_FILE (with a random DB_PASSWORD) ..."
  DB_PASSWORD="$(openssl rand -hex 24 2>/dev/null || head -c 48 /dev/urandom | base64 | tr -dc 'A-Za-z0-9' | head -c 32)"
  {
    echo "APP_ENV=production"
    echo "APP_PORT=8080"
    echo "APP_NAME=interview-question-002"
    echo "APP_ALLOWED_ORIGINS=*"
    echo "DB_HOST=postgres"
    echo "DB_PORT=5432"
    echo "DB_USER=${DB_USER}"
    echo "DB_PASSWORD=${DB_PASSWORD}"
    echo "DB_NAME=${DB_NAME}"
    echo "DB_SSLMODE=disable"
    echo "JWT_PRIVATE_KEY_PATH=./secrets/jwt_private.pem"
    echo "JWT_PUBLIC_KEY_PATH=./secrets/jwt_public.pem"
    echo "JWT_ACCESS_TTL=60m"
    echo "JWT_ISSUER=example.com"
    echo "LOG_LEVEL=info"
    echo "LOG_FORMAT=json"
    echo "SITE_ADDRESS=${SITE_ADDRESS}"
    [ -n "$HTTP_PORT" ]  && echo "HTTP_PORT=${HTTP_PORT}"
    [ -n "$HTTPS_PORT" ] && echo "HTTPS_PORT=${HTTPS_PORT}"
    echo "GHCR_OWNER=${GHCR_OWNER}"
    echo "TAG=latest"
  } > "$ENV_FILE"
  chmod 600 "$ENV_FILE"
fi

# ── 4) authorize the CI deploy public key (optional) ─────────────────────────
if [ -n "$DEPLOY_PUBKEY" ]; then
  mkdir -p /root/.ssh && chmod 700 /root/.ssh
  touch /root/.ssh/authorized_keys && chmod 600 /root/.ssh/authorized_keys
  if grep -qF "$DEPLOY_PUBKEY" /root/.ssh/authorized_keys; then
    log "Deploy key already authorized."
  else
    echo "$DEPLOY_PUBKEY" >> /root/.ssh/authorized_keys
    log "Added deploy public key to /root/.ssh/authorized_keys"
  fi
else
  warn "DEPLOY_PUBKEY not set — add your CI deploy PUBLIC key to /root/.ssh/authorized_keys manually."
fi

# ── done ─────────────────────────────────────────────────────────────────────
log "VPS is ready."
cat <<EOF

Next steps
  1) (optional) review ${ENV_FILE}  — SITE_ADDRESS / ports. DB_PASSWORD is already generated.
  2) GitHub repo → Settings → Secrets and variables → Actions → add:
       VPS_HOST         = this server's IP
       VPS_USER         = root
       SSH_PRIVATE_KEY  = the PRIVATE key matching the authorized public key
       GHCR_TOKEN       = a PAT with 'read:packages'  (skip if GHCR packages are public)
  3) git push origin main   → the deploy workflow builds images and deploys here.

Manual first deploy (optional — after the workflow has synced the compose files here):
  cd ${APP_DIR}/deployments/docker
  # if images are private:  echo <GHCR_TOKEN> | docker login ghcr.io -u ${GHCR_OWNER} --password-stdin
  docker compose pull && docker compose up -d
  docker compose ps
EOF

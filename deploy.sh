#!/usr/bin/env sh
# Usage:
#   ./deploy.sh          # local dev (docker compose up)
#   ./deploy.sh dev      # same as above
#   ./deploy.sh prod     # production deploy
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
cd "$ROOT_DIR"

MODE="${1:-dev}"

log() { printf '\n==> %s\n' "$1"; }
fail() { printf 'FAIL: %s\n' "$1" >&2; exit 1; }

has_command() { command -v "$1" >/dev/null 2>&1; }

compose() {
  if docker compose version >/dev/null 2>&1; then
    docker compose "$@"
  elif has_command docker-compose; then
    docker-compose "$@"
  else
    fail "docker compose or docker-compose is required"
  fi
}

# ---------------------------------------------------------------------------
# dev mode: env setup + docker compose up
# ---------------------------------------------------------------------------
deploy_dev() {
  log "Setup local environment"
  if [ ! -f .env.local ]; then
    fail ".env.local not found. Copy .env.example to .env.local and fill in values."
  fi
  cp .env.local .env
  printf 'Loaded .env.local\n'

  log "Start Docker services"
  compose up -d --build

  log "Wait for API to be healthy"
  PORT="${SERVER_PORT:-8081}"
  attempts=0
  max=30
  while [ "$attempts" -lt "$max" ]; do
    if curl -sf "http://localhost:${PORT}/health" >/dev/null 2>&1; then
      break
    fi
    attempts=$((attempts + 1))
    printf '.'
    sleep 2
  done
  printf '\n'
  if [ "$attempts" -eq "$max" ]; then
    fail "API did not become healthy after $((max * 2))s"
  fi

  printf '\nDev environment ready.\n'
  printf '  API:  http://localhost:%s\n' "$PORT"
  printf '  Docs: http://localhost:%s/swagger/index.html\n' "$PORT"
}

# ---------------------------------------------------------------------------
# prod mode: delegate to full production deploy script
# ---------------------------------------------------------------------------
deploy_prod() {
  exec env ROOT_DIR="$ROOT_DIR" "$ROOT_DIR/backend/scripts/deploy-production.sh"
}

# ---------------------------------------------------------------------------
# dispatch
# ---------------------------------------------------------------------------
case "$MODE" in
  dev|local) deploy_dev ;;
  prod|production) deploy_prod ;;
  *) fail "Unknown mode '$MODE'. Use: dev | prod" ;;
esac

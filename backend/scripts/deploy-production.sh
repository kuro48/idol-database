#!/usr/bin/env sh
set -eu

ROOT_DIR="${ROOT_DIR:-$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)}"
cd "$ROOT_DIR"

ENV_FILE="${ENV_FILE:-.env}"
override_names="ENV_FILE BASE_URL FRONTEND_URL DRY_RUN"
for name in $override_names; do
  eval "is_set=\${$name+x}"
  if [ "$is_set" = "x" ]; then
    eval "deploy_override_$name=\${$name}"
  fi
done

if [ -f "$ENV_FILE" ]; then
  set -a
  # shellcheck disable=SC1090
  . "$ENV_FILE"
  set +a
fi

for name in $override_names; do
  eval "is_set=\${deploy_override_$name+x}"
  if [ "$is_set" = "x" ]; then
    eval "$name=\${deploy_override_$name}"
  fi
done

BASE_URL="${BASE_URL:-}"
FRONTEND_URL="${FRONTEND_URL:-}"
FRONTEND_DEPLOY_DIR="$ROOT_DIR/frontend-deploy"
DRY_RUN="${DRY_RUN:-0}"

log() {
  printf '\n==> %s\n' "$1"
}

fail() {
  printf 'FAIL: %s\n' "$1" >&2
  exit 1
}

has_command() {
  command -v "$1" >/dev/null 2>&1
}

run() {
  printf '+ %s\n' "$*"
  if [ "$DRY_RUN" = "1" ]; then
    return 0
  fi
  "$@"
}

compose() {
  if [ "$DRY_RUN" = "1" ]; then
    run docker compose "$@"
    return
  fi

  if docker compose version >/dev/null 2>&1; then
    docker compose "$@"
  elif has_command docker-compose; then
    docker-compose "$@"
  else
    fail "docker compose or docker-compose is required"
  fi
}

frontend_build() {
  if has_command pnpm && [ -f frontend/pnpm-lock.yaml ]; then
    run pnpm --dir frontend build
  else
    run npm --prefix frontend run build
  fi
}

check_git_state() {
  if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    return
  fi

  if [ -n "$(git status --porcelain)" ]; then
    fail "working tree is dirty. Pull deploy branch cleanly before running deploy"
  fi
}

check_env() {
  log "Check production env"
  run env ENV_FILE="$ENV_FILE" ./backend/scripts/check-production-env.sh
}

build_frontend() {
  log "Build frontend"
  frontend_build

  log "Publish frontend dist"
  run mkdir -p "$FRONTEND_DEPLOY_DIR"
  if has_command rsync; then
    run rsync -a --delete frontend/dist/ "$FRONTEND_DEPLOY_DIR"/
  else
    printf 'WARN: rsync not found; copying without deleting stale files.\n' >&2
    run cp -R frontend/dist/. "$FRONTEND_DEPLOY_DIR"/
  fi
}

deploy_docker() {
  log "Deploy backend with Docker Compose"
  compose --env-file "$ENV_FILE" up -d --build
}

run_smoke() {
  if [ -z "$BASE_URL" ]; then
    fail "BASE_URL is required for smoke test"
  fi

  log "Run production smoke"
  run env BASE_URL="$BASE_URL" FRONTEND_URL="$FRONTEND_URL" ./backend/scripts/smoke-production.sh
}

check_git_state
check_env
build_frontend
deploy_docker
run_smoke

printf '\nDeploy finished.\n'

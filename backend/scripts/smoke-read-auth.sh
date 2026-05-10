#!/usr/bin/env sh
set -eu

BASE_URL="${BASE_URL:-http://localhost:8081}"
API_KEY="${API_KEY:-}"

status() {
  path="$1"
  shift
  curl -sS -o /dev/null -w "%{http_code}" "$@" "${BASE_URL}${path}"
}

health_status="$(status /health/ready)"
if [ "$health_status" != "200" ]; then
  echo "health readiness failed: expected 200, got ${health_status}" >&2
  exit 1
fi

anonymous_status="$(status /api/v1/idols)"
if [ "$anonymous_status" != "401" ]; then
  echo "anonymous read auth failed: expected 401, got ${anonymous_status}" >&2
  exit 1
fi

if [ -n "$API_KEY" ]; then
  authenticated_status="$(status /api/v1/idols -H "Authorization: Bearer ${API_KEY}")"
  case "$authenticated_status" in
    200|404)
      ;;
    *)
      echo "authenticated read failed: expected 200 or 404, got ${authenticated_status}" >&2
      exit 1
      ;;
  esac
fi

echo "smoke read auth passed"

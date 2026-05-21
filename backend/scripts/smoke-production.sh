#!/usr/bin/env sh
set -eu

BASE_URL="${BASE_URL:-http://localhost:8081}"
FRONTEND_URL="${FRONTEND_URL:-}"
ACCESS_TOKEN="${ACCESS_TOKEN:-}"
ID_TOKEN="${ID_TOKEN:-}"
ADMIN_ACCESS_TOKEN="${ADMIN_ACCESS_TOKEN:-}"
RUN_MUTATION="${RUN_MUTATION:-0}"
RUN_REMOVAL="${RUN_REMOVAL:-0}"
REMOVAL_TARGET_ID="${REMOVAL_TARGET_ID:-}"
REMOVAL_TARGET_TYPE="${REMOVAL_TARGET_TYPE:-idol}"

tmp_dir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT INT TERM

fail() {
  printf 'FAIL: %s\n' "$1" >&2
  if [ -f "${tmp_dir}/last-body" ]; then
    printf '%s\n' '--- response body ---' >&2
    sed -n '1,80p' "${tmp_dir}/last-body" >&2
  fi
  exit 1
}

pass() {
  printf 'OK: %s\n' "$1"
}

request_status() {
  method="$1"
  path="$2"
  body="${3:-}"
  shift 3 || true
  body_file="${tmp_dir}/last-body"
  if [ -n "$body" ]; then
    printf '%s' "$body" | curl -sS -o "$body_file" -w '%{http_code}' -X "$method" "$@" \
      -H 'Content-Type: application/json' \
      --data-binary @- \
      "${BASE_URL}${path}"
  else
    curl -sS -o "$body_file" -w '%{http_code}' -X "$method" "$@" "${BASE_URL}${path}"
  fi
}

expect_status() {
  label="$1"
  expected="$2"
  method="$3"
  path="$4"
  body="${5:-}"
  shift 5 || true
  status="$(request_status "$method" "$path" "$body" "$@")"
  if [ "$status" != "$expected" ]; then
    fail "$label expected HTTP $expected, got $status"
  fi
  pass "$label"
}

expect_status_one_of() {
  label="$1"
  expected_csv="$2"
  method="$3"
  path="$4"
  body="${5:-}"
  shift 5 || true
  status="$(request_status "$method" "$path" "$body" "$@")"
  old_ifs="$IFS"
  IFS=","
  for expected in $expected_csv; do
    IFS="$old_ifs"
    if [ "$status" = "$expected" ]; then
      pass "$label"
      return
    fi
    IFS=","
  done
  IFS="$old_ifs"
  fail "$label expected HTTP one of [$expected_csv], got $status"
}

printf 'Target API: %s\n' "$BASE_URL"

expect_status "health" "200" "GET" "/health" ""
expect_status "readiness" "200" "GET" "/health/ready" ""

expect_status_one_of "anonymous idols read" "200" "GET" "/api/v1/idols?limit=1" ""
expect_status_one_of "anonymous groups read" "200" "GET" "/api/v1/groups?limit=1" ""
expect_status_one_of "anonymous agencies read" "200" "GET" "/api/v1/agencies?limit=1" ""
expect_status_one_of "anonymous events read" "200" "GET" "/api/v1/events?limit=1" ""
expect_status_one_of "anonymous releases read" "200" "GET" "/api/v1/releases?limit=1" ""
expect_status_one_of "anonymous tags read" "200" "GET" "/api/v1/tags?limit=1" ""

submission_body='{"target_type":"idol","payload":{"name":"Smoke Test Idol"},"source_urls":["https://example.com/smoke"]}'
expect_status "anonymous /me is blocked" "401" "GET" "/api/v1/me" ""
expect_status "anonymous submission create is blocked" "401" "POST" "/api/v1/submissions" "$submission_body"

if [ -n "$FRONTEND_URL" ]; then
  frontend_body="${tmp_dir}/frontend-body"
  frontend_status="$(curl -sS -o "$frontend_body" -w '%{http_code}' "$FRONTEND_URL")"
  if [ "$frontend_status" != "200" ]; then
    cp "$frontend_body" "${tmp_dir}/last-body"
    fail "frontend expected HTTP 200, got $frontend_status"
  fi
  pass "frontend is reachable"

  cors_status="$(curl -sS -o "${tmp_dir}/last-body" -w '%{http_code}' -X OPTIONS \
    -H "Origin: ${FRONTEND_URL}" \
    -H 'Access-Control-Request-Method: GET' \
    -H 'Access-Control-Request-Headers: Authorization,X-ID-Token' \
    "${BASE_URL}/api/v1/me")"
  if [ "$cors_status" != "204" ]; then
    fail "CORS preflight expected HTTP 204, got $cors_status"
  fi
  pass "CORS preflight allows frontend auth headers"
fi

if [ -n "$ACCESS_TOKEN" ] || [ -n "$ID_TOKEN" ]; then
  if [ -z "$ACCESS_TOKEN" ] || [ -z "$ID_TOKEN" ]; then
    fail "ACCESS_TOKEN and ID_TOKEN must be set together for authenticated checks"
  fi

  expect_status "authenticated /me" "200" "GET" "/api/v1/me" "" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -H "X-ID-Token: ${ID_TOKEN}"
  expect_status "authenticated submission history" "200" "GET" "/api/v1/me/submissions" "" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -H "X-ID-Token: ${ID_TOKEN}"
  expect_status "authenticated removal history" "200" "GET" "/api/v1/me/removal-requests" "" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -H "X-ID-Token: ${ID_TOKEN}"

  if [ "$RUN_MUTATION" = "1" ]; then
    timestamp="$(date +%Y%m%d%H%M%S)"
    mutation_body='{"target_type":"idol","payload":{"name":"Smoke Test Idol '"$timestamp"'"},"source_urls":["https://example.com/smoke"]}'
    expect_status "authenticated submission create" "201" "POST" "/api/v1/submissions" "$mutation_body" \
      -H "Authorization: Bearer ${ACCESS_TOKEN}" \
      -H "X-ID-Token: ${ID_TOKEN}"
  else
    printf 'SKIP: authenticated submission create. Set RUN_MUTATION=1 to create a real pending submission.\n'
  fi

  if [ "$RUN_REMOVAL" = "1" ]; then
    if [ -z "$REMOVAL_TARGET_ID" ]; then
      fail "REMOVAL_TARGET_ID is required when RUN_REMOVAL=1"
    fi
    removal_body='{"target_type":"'"$REMOVAL_TARGET_TYPE"'","target_id":"'"$REMOVAL_TARGET_ID"'","requester_type":"third_party","reason":"Smoke test removal reason","description":"Smoke test removal request for production flow verification."}'
    expect_status "authenticated removal request create" "201" "POST" "/api/v1/removal-requests" "$removal_body" \
      -H "Authorization: Bearer ${ACCESS_TOKEN}" \
      -H "X-ID-Token: ${ID_TOKEN}"
  else
    printf 'SKIP: authenticated removal request create. Set RUN_REMOVAL=1 REMOVAL_TARGET_ID=... to create a real request.\n'
  fi
fi

if [ -n "$ADMIN_ACCESS_TOKEN" ]; then
  expect_status "admin submissions list" "200" "GET" "/api/v1/submissions?limit=1" "" \
    -H "Authorization: Bearer ${ADMIN_ACCESS_TOKEN}"
else
  printf 'SKIP: admin checks. Set ADMIN_ACCESS_TOKEN to test admin-only APIs.\n'
fi

printf '\nProduction smoke passed.\n'

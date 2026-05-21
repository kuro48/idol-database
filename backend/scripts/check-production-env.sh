#!/usr/bin/env sh
set -eu

ENV_FILE="${ENV_FILE:-.env}"

if [ -f "$ENV_FILE" ]; then
  set -a
  # shellcheck disable=SC1090
  . "$ENV_FILE"
  set +a
fi

failures=0

fail() {
  failures=$((failures + 1))
  printf 'FAIL: %s\n' "$1" >&2
}

ok() {
  printf 'OK: %s\n' "$1"
}

require_var() {
  name="$1"
  eval "value=\${$name:-}"
  if [ -z "$value" ]; then
    fail "$name is required"
    return
  fi
  ok "$name is set"
}

require_https_url() {
  name="$1"
  eval "value=\${$name:-}"
  if [ -z "$value" ]; then
    fail "$name is required"
    return
  fi
  case "$value" in
    https://*)
      ok "$name uses https"
      ;;
    *)
      fail "$name must start with https://"
      ;;
  esac
  case "$value" in
    *localhost*|*127.0.0.1*)
      fail "$name must not point to localhost in production"
      ;;
  esac
}

if [ "${GIN_MODE:-}" != "release" ]; then
  fail "GIN_MODE must be release"
else
  ok "GIN_MODE is release"
fi

require_var MONGODB_URI
require_var MONGODB_DATABASE
require_https_url IDOL_AUTH_URL
require_https_url IDOL_AUTH_ISSUER_URL
require_var IDOL_AUTH_CLIENT_ID
require_var CORS_ALLOWED_ORIGINS

if [ -n "${CORS_ALLOWED_ORIGINS:-}" ]; then
  old_ifs="$IFS"
  IFS=","
  for origin in $CORS_ALLOWED_ORIGINS; do
    IFS="$old_ifs"
    origin="$(printf '%s' "$origin" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
    if [ -z "$origin" ]; then
      fail "CORS_ALLOWED_ORIGINS contains an empty origin"
    fi
    case "$origin" in
      "*")
        fail "CORS_ALLOWED_ORIGINS must not contain *"
        ;;
      http://*)
        fail "CORS origin must use https://: $origin"
        ;;
      https://*)
        ok "CORS origin is https: $origin"
        ;;
      *)
        fail "CORS origin must be a full https:// origin: $origin"
        ;;
    esac
    case "$origin" in
      *localhost*|*127.0.0.1*)
        fail "CORS origin must not be localhost in production: $origin"
        ;;
      */*)
        without_scheme="${origin#https://}"
        case "$without_scheme" in
          */*)
            fail "CORS origin must not include a path: $origin"
            ;;
        esac
        ;;
    esac
    IFS=","
  done
  IFS="$old_ifs"
fi

if [ -n "${TRUSTED_PROXIES:-}" ]; then
  ok "TRUSTED_PROXIES is set"
else
  printf 'WARN: TRUSTED_PROXIES is empty. This is OK only when the app is not behind a trusted reverse proxy.\n' >&2
fi

if [ "$failures" -ne 0 ]; then
  printf '\nProduction env check failed with %s issue(s).\n' "$failures" >&2
  exit 1
fi

printf '\nProduction env check passed.\n'

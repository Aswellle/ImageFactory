#!/usr/bin/env bash
#
# ImageForge deployment smoke test.
#
# Boots the stack with docker compose, waits for the backend health endpoint,
# then exercises the key HTTP endpoints with curl and reports pass/fail.
#
# Usage:
#   bash deploy/tests/smoke.sh
#
# The script is idempotent: it brings the stack up, runs checks, and leaves
# the stack running so you can inspect it. Run `docker compose down -v` to
# clean up. Set KEEP_STACK=0 to tear down automatically on exit.
#
# Exit codes:
#   0 - all checks passed
#   1 - one or more checks failed (or prerequisites missing)

set -euo pipefail

# --- Configuration (overridable via environment) ---
COMPOSE_CMD="${COMPOSE_CMD:-docker compose}"
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DEPLOY_DIR="${PROJECT_DIR}/deploy"
BACKEND_URL="${IF_BASE_URL:-http://localhost:8080}"
HEALTH_URL="${BACKEND_URL}/v1/health"
MAX_WAIT="${IF_MAX_WAIT:-120}"     # seconds to wait for health
SLEEP_INTERVAL="${IF_SLEEP:-5}"    # seconds between health polls
KEEP_STACK="${KEEP_STACK:-1}"      # 1 = leave stack up, 0 = tear down

PASS=0
FAIL=0
FAILED_NAMES=()

log()  { printf '[smoke] %s\n' "$*"; }
ok()   { printf '[smoke] PASS: %s\n' "$*"; }
fail() { printf '[smoke] FAIL: %s\n' "$*"; }

# check runs a named assertion; it does NOT abort the whole script on failure.
check() {
    local name="$1" expected="$2" actual="$3"
    if [[ "$actual" == "$expected" ]]; then
        ok "$name (got $actual)"
        PASS=$((PASS + 1))
    else
        fail "$name (expected $expected, got $actual)"
        FAIL=$((FAIL + 1))
        FAILED_NAMES+=("$name")
    fi
}

# curl_status issues a request and prints only the HTTP status code.
curl_status() {
    local method="$1" url="$2" data="${3:-}" token="${4:-}"
    local -a args=(--silent --output /dev/null --write-out '%{http_code}' \
                    --max-time 20 --location --request "$method" "$url")
    if [[ -n "$data" ]]; then
        args+=(--header 'Content-Type: application/json' --data "$data")
    fi
    if [[ -n "$token" ]]; then
        args+=(--header "Authorization: Bearer ${token}")
    fi
    curl "${args[@]}" || echo "000"
}

# curl_body issues a request and prints the HTTP status code and body.
curl_body() {
    local method="$1" url="$2" data="${3:-}" token="${4:-}"
    local -a args=(--silent --max-time 20 --location --request "$method" "$url")
    if [[ -n "$data" ]]; then
        args+=(--header 'Content-Type: application/json' --data "$data")
    fi
    if [[ -n "$token" ]]; then
        args+=(--header "Authorization: Bearer ${token}")
    fi
    curl "${args[@]}"
}

# require_cmd aborts if a prerequisite CLI is missing.
require_cmd() {
    if ! command -v "$1" >/dev/null 2>&1; then
        log "ERROR: required command '$1' not found on PATH"
        exit 1
    fi
}

cleanup() {
    if [[ "$KEEP_STACK" == "1" ]]; then
        log "leaving stack running (KEEP_STACK=1). Clean up with: ${COMPOSE_CMD} -f ${DEPLOY_DIR}/docker-compose.yml down -v"
    else
        log "tearing down stack (KEEP_STACK=0)..."
        ( cd "$DEPLOY_DIR" && $COMPOSE_CMD -f docker-compose.yml down -v ) || true
    fi
}

main() {
    require_cmd docker
    require_cmd curl

    log "using compose command: ${COMPOSE_CMD}"
    log "project dir: ${PROJECT_DIR}"
    log "backend url: ${BACKEND_URL}"

    # --- Bring up the stack ---
    log "starting stack from ${DEPLOY_DIR}/docker-compose.yml ..."
    ( cd "$DEPLOY_DIR" && $COMPOSE_CMD -f docker-compose.yml up -d --build ) || {
        log "ERROR: docker compose up failed"
        exit 1
    }
    trap cleanup EXIT

    # --- Wait for the health endpoint ---
    log "waiting up to ${MAX_WAIT}s for ${HEALTH_URL} ..."
    local waited=0
    local health_status="000"
    while (( waited < MAX_WAIT )); do
        health_status="$(curl_status GET "$HEALTH_URL")"
        if [[ "$health_status" == "200" ]]; then
            break
        fi
        log "  health=${health_status} (${waited}s/${MAX_WAIT}s) ..."
        sleep "$SLEEP_INTERVAL"
        waited=$((waited + SLEEP_INTERVAL))
    done
    check "health endpoint reachable" "200" "$health_status"

    if [[ "$health_status" != "200" ]]; then
        log "backend never became healthy; skipping remaining checks"
        print_summary
        exit 1
    fi

    # --- Health body contract ---
    health_body="$(curl_body GET "$HEALTH_URL")"
    if echo "$health_body" | grep -q '"status"[[:space:]]*:[[:space:]]*"ok"'; then
        ok "health body contains status:ok"
        PASS=$((PASS + 1))
    else
        fail "health body missing status:ok (got: ${health_body})"
        FAIL=$((FAIL + 1))
        FAILED_NAMES+=("health body contract")
    fi

    # --- User registration ---
    local email="smoke-$(date +%s)@imageforge.test"
    local reg_body
    reg_body="$(curl_body POST "${BACKEND_URL}/v1/auth/register" \
                "{\"email\":\"${email}\",\"password\":\"supersecret123\",\"name\":\"Smoke\"}")"
    local reg_status
    reg_status="$(curl_status POST "${BACKEND_URL}/v1/auth/register" \
                  "{\"email\":\"${email}-2@imageforge.test\",\"password\":\"supersecret123\"}")"
    check "user registration" "201" "$reg_status"

    # Extract the token from the registration response.
    local token
    token="$(echo "$reg_body" | grep -o '"token"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"token"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')"
    if [[ -n "$token" ]]; then
        ok "registration returned a token"
        PASS=$((PASS + 1))
    else
        fail "registration did not return a token (body: ${reg_body})"
        FAIL=$((FAIL + 1))
        FAILED_NAMES+=("registration token present")
    fi

    # --- User login ---
    local login_status
    login_status="$(curl_status POST "${BACKEND_URL}/v1/auth/login" \
                    "{\"email\":\"${email}\",\"password\":\"supersecret123\"}")"
    check "user login" "200" "$login_status"

    # --- Create API key (authenticated) ---
    if [[ -n "$token" ]]; then
        local key_status
        key_status="$(curl_status POST "${BACKEND_URL}/v1/api-keys" '{"name":"smoke-key"}' "$token")"
        check "create api key (authed)" "201" "$key_status"

        # --- List API keys (authenticated) ---
        local list_status
        list_status="$(curl_status GET "${BACKEND_URL}/v1/api-keys" "" "$token")"
        check "list api keys (authed)" "200" "$list_status"

        # --- Usage endpoint (authenticated) ---
        local usage_status
        usage_status="$(curl_status GET "${BACKEND_URL}/v1/usage" "" "$token")"
        check "usage endpoint (authed)" "200" "$usage_status"
    else
        log "skipping authenticated checks (no token available)"
    fi

    # --- Unauthenticated access is rejected ---
    local me_status
    me_status="$(curl_status GET "${BACKEND_URL}/v1/auth/me")"
    check "authed endpoint rejects anonymous" "401" "$me_status"

    print_summary
    if (( FAIL > 0 )); then
        exit 1
    fi
}

print_summary() {
    log "-------------------------------------------"
    log "smoke results: ${PASS} passed, ${FAIL} failed"
    if (( FAIL > 0 )); then
        log "failed checks:"
        for name in "${FAILED_NAMES[@]}"; do
            log "  - ${name}"
        done
    fi
    log "-------------------------------------------"
}

main "$@"

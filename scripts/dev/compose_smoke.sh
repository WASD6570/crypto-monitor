#!/bin/sh

set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
TMP_DIR=$(mktemp -d)

log() {
  printf '[compose-smoke] %s\n' "$*"
}

cleanup() {
  log "Stopping Compose stack"
  docker compose down --remove-orphans >/dev/null 2>&1 || true
  rm -rf "$TMP_DIR"
}

show_compose_state() {
  docker compose ps || true
}

wait_for_http() {
  name=$1
  url=$2
  allowed_statuses=$3
  body_file=$4
  status_file=$5
  attempts=30
  attempt=1

  while [ "$attempt" -le "$attempts" ]; do
    if curl -sS --max-time 5 -o "$body_file" -w '%{http_code}' "$url" >"$status_file"; then
      status=$(cat "$status_file")
      for allowed in $allowed_statuses; do
        if [ "$status" = "$allowed" ]; then
          return 0
        fi
      done
      log "Waiting for $name at $url (attempt $attempt/$attempts, status $status)"
    else
      log "Waiting for $name at $url (attempt $attempt/$attempts, request failed)"
    fi
    attempt=$((attempt + 1))
    sleep 2
  done

  log "Timed out waiting for $name at $url"
  show_compose_state
  return 1
}

trap cleanup EXIT INT TERM

cd "$ROOT_DIR"

runtime_body="$TMP_DIR/runtime-status.json"
runtime_status="$TMP_DIR/runtime-status.code"
global_body="$TMP_DIR/current-global.json"
global_status="$TMP_DIR/current-global.code"
health_body="$TMP_DIR/healthz.json"

log "Rendering Compose configuration"
docker compose config >/dev/null

log "Resetting any prior Compose stack"
docker compose down --remove-orphans >/dev/null 2>&1 || true

log "Starting Compose stack"
docker compose up --build -d

log "Checking same-origin runtime-status route"
wait_for_http "runtime-status" "http://127.0.0.1:4173/api/runtime-status" "200" "$runtime_body" "$runtime_status"
grep -Eq '"symbols"[[:space:]]*:' "$runtime_body"
grep -Eq '"BTC-USD"' "$runtime_body"
grep -Eq '"ETH-USD"' "$runtime_body"
grep -Eq '"readiness"[[:space:]]*:[[:space:]]*"(READY|NOT_READY)"' "$runtime_body"
runtime_readiness=READY
if grep -Eq '"readiness"[[:space:]]*:[[:space:]]*"NOT_READY"' "$runtime_body"; then
  runtime_readiness=NOT_READY
  log "Runtime status reports warm-up via readiness=NOT_READY"
else
  log "Runtime status is readable and not in initial warm-up"
fi

log "Checking same-origin current-state route"
wait_for_http "market-state-global" "http://127.0.0.1:4173/api/market-state/global" "200 500" "$global_body" "$global_status"
current_status=$(cat "$global_status")
if [ "$current_status" = "200" ]; then
  grep -Eq '"schemaVersion"[[:space:]]*:' "$global_body"
  grep -Eq '"symbols"[[:space:]]*:' "$global_body"
  log "Current-state route returned a readable payload"
else
  if [ "$runtime_readiness" != "NOT_READY" ]; then
    log "Current-state route returned 500 after runtime status left warm-up"
    show_compose_state
    exit 1
  fi
  grep -Eq '"error"[[:space:]]*:' "$global_body"
  log "Current-state route is reachable and still warming up"
fi

log "Checking internal process health from market-state-api"
docker compose exec -T market-state-api wget -qO- http://127.0.0.1:8080/healthz >"$health_body"
grep -Eq '"status"[[:space:]]*:[[:space:]]*"ok"' "$health_body"

log "Compose rollout smoke passed"

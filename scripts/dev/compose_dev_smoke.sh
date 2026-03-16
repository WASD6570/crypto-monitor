#!/bin/sh

set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
TMP_DIR=$(mktemp -d)
COMPOSE_FILES="-f docker-compose.yml -f docker-compose.dev.yml"
WATCHED_FILE="$ROOT_DIR/cmd/market-state-api/main.go"
WATCHED_BACKUP="$TMP_DIR/main.go.before"

log() {
  printf '[compose-dev-smoke] %s\n' "$*"
}

cleanup() {
  if [ -f "$WATCHED_BACKUP" ]; then
    cp "$WATCHED_BACKUP" "$WATCHED_FILE"
  fi
  log "Stopping Compose dev stack"
  docker compose -f "$ROOT_DIR/docker-compose.yml" -f "$ROOT_DIR/docker-compose.dev.yml" down --remove-orphans >/dev/null 2>&1 || true
  rm -rf "$TMP_DIR"
}

wait_for_http() {
  name=$1
  url=$2
  allowed_statuses=$3
  body_file=$4
  status_file=$5
  attempts=40
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
  return 1
}

wait_for_log_match() {
  pattern=$1
  since_ts=$2
  attempts=30
  attempt=1

  while [ "$attempt" -le "$attempts" ]; do
    if docker compose -f "$ROOT_DIR/docker-compose.yml" -f "$ROOT_DIR/docker-compose.dev.yml" logs --since "$since_ts" market-state-api | grep -Eq "$pattern"; then
      return 0
    fi
    attempt=$((attempt + 1))
    sleep 2
  done

  log "Timed out waiting for market-state-api log pattern: $pattern"
  return 1
}

trap cleanup EXIT INT TERM

cd "$ROOT_DIR"

dashboard_body="$TMP_DIR/dashboard.html"
dashboard_status="$TMP_DIR/dashboard.code"
runtime_body="$TMP_DIR/runtime-status.json"
runtime_status="$TMP_DIR/runtime-status.code"
global_body="$TMP_DIR/current-global.json"
global_status="$TMP_DIR/current-global.code"
health_body="$TMP_DIR/healthz.json"

log "Rendering Compose dev overlay"
docker compose $COMPOSE_FILES config >/dev/null

log "Resetting any prior Compose dev stack"
docker compose $COMPOSE_FILES down --remove-orphans >/dev/null 2>&1 || true

log "Starting Compose dev stack"
docker compose $COMPOSE_FILES up --build -d

log "Checking Vite dashboard shell"
wait_for_http "dashboard" "http://127.0.0.1:4173/dashboard" "200" "$dashboard_body" "$dashboard_status"
grep -Eq '/@vite/client' "$dashboard_body"

log "Checking same-origin runtime-status route"
wait_for_http "runtime-status" "http://127.0.0.1:4173/api/runtime-status" "200" "$runtime_body" "$runtime_status"
grep -Eq '"BTC-USD"' "$runtime_body"
grep -Eq '"ETH-USD"' "$runtime_body"

log "Checking same-origin current-state route"
wait_for_http "market-state-global" "http://127.0.0.1:4173/api/market-state/global" "200 500" "$global_body" "$global_status"
grep -Eq '"(schemaVersion|error)"' "$global_body"

log "Checking internal process health from market-state-api"
docker compose $COMPOSE_FILES exec -T market-state-api wget -qO- http://127.0.0.1:8080/healthz >"$health_body"
grep -Eq '"status"[[:space:]]*:[[:space:]]*"ok"' "$health_body"

log "Triggering watcher restart with a benign file touch"
cp "$WATCHED_FILE" "$WATCHED_BACKUP"
restart_since=$(date -Iseconds)
printf '\n' >>"$WATCHED_FILE"
wait_for_log_match 'cmd/market-state-api/main.go has changed|building\.\.\.|running\.\.\.' "$restart_since"

log "Re-checking process health after watcher restart"
docker compose $COMPOSE_FILES exec -T market-state-api wget -qO- http://127.0.0.1:8080/healthz >"$health_body"
grep -Eq '"status"[[:space:]]*:[[:space:]]*"ok"' "$health_body"

log "Compose dev smoke passed"

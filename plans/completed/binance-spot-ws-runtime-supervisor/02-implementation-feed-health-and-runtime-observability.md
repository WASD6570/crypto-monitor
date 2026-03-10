# Implementation Module 2: Feed Health And Runtime Observability

## Scope

- Plan how the Spot supervisor emits adapter-scoped feed-health and exposes runtime state needed for debugging and later integration.
- Cover connection-state transitions, message freshness, reconnect-loop degradation, and clock-health visibility.
- Exclude trade/top-of-book canonical event generation and exclude depth-specific gap or snapshot degradation.

## Target Repo Areas

- `services/venue-binance/runtime.go`
- `services/venue-binance` new files such as `spot_feed_health.go` or `spot_ws_supervisor.go`
- `services/normalizer` only as a consumed boundary for later feed-health normalization, not for new logic in this feature
- `tests/integration/ingestion_smoke_test.go` or a new focused Binance integration test file

## Requirements

- Supervisor health decisions must be built from the Spot-filtered runtime config so snapshot freshness stays `NOT_APPLICABLE` in this slice.
- Emit machine-readable feed-health when the supervisor becomes healthy, degraded, stale, reconnecting, or rolled over.
- Preserve connection state, freshness, reconnect count, clock state, and degradation reasons using the shared `libs/go/ingestion` vocabulary.
- Treat missing or late data after a healthy connection as message-freshness degradation, not as a silent socket issue.
- Surface reconnect-loop degradation once configured thresholds are crossed, even if the next reconnect eventually succeeds.
- Preserve Wave 1 source-record ID and timestamp rules when the supervisor publishes feed-health messages for later normalization.

## Key Decisions

- Reuse the existing `Runtime`, `AdapterLoop`, and `FeedHealthStatus` helpers as the policy engine; add supervisor-facing adapters rather than duplicating health logic.
- Keep sequence-gap and resync counters untouched in this feature and leave them at their neutral values because they belong to later depth work.
- Treat `last frame seen` as the primary freshness input for this slice; `last pong` is diagnostic state, not a second health contract.
- Publish feed-health on meaningful state transitions and on periodic stale checks so downstream consumers can see both recovery and degradation.
- Keep logs secondary to machine-readable output; tests should assert feed-health fields, not log strings.

## Unit Test Expectations

- Healthy startup produces a `HEALTHY` decision with `CONNECTED` state and no degradation reasons.
- Message staleness after the configured timeout produces `STALE` with `message-stale` preserved.
- Reconnect-loop thresholds produce `DEGRADED` with `connection-not-ready` and `reconnect-loop` reasons.
- Local clock offset at warning vs degraded thresholds preserves the shared clock-state behavior.
- Recovery after a successful reconnect and fresh traffic returns feed-health to healthy without stale reasons lingering.

## Contract / Fixture / Replay Impacts

- No new contract families are expected.
- Add supervisor-focused feed-health fixtures or expected-output snippets only where they improve deterministic lifecycle coverage.
- Replay remains downstream of this feature, but feed-health publication must preserve enough timestamp and source-record stability for later replay acceptance.
- If existing fixture manifests need new runtime-control entries, keep them narrow and limited to supervisor/feed-health scenarios.

## Summary

This module makes supervisor health observable as a product output, not just a transport detail, while keeping the emitted health semantics aligned with shared ingestion vocabulary and isolated from later depth behavior.

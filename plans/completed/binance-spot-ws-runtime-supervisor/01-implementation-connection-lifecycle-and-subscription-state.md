# Implementation Module 1: Connection Lifecycle And Subscription State

## Scope

- Plan the concrete supervisor state and lifecycle flow for one Binance Spot websocket connection covering BTC/ETH `trade` and `bookTicker` only.
- Include connect, subscribe, control-frame handling, proactive rollover, reconnect, and resubscribe.
- Exclude trade payload parsing, `bookTicker` payload parsing, order-book snapshot/bootstrap, and normalizer logic.

## Target Repo Areas

- `services/venue-binance/runtime.go`
- `services/venue-binance/runtime_test.go`
- `services/venue-binance` new supervisor-focused files such as `spot_ws_supervisor.go` and `spot_ws_supervisor_test.go`
- `configs/local/ingestion.v1.json`

## Requirements

- Derive a Spot-only runtime config that includes only `trades` and `top-of-book` stream definitions for `marketType == spot`.
- Build the desired subscription set from repo-configured symbols and Binance-native stream names, limited to BTC/ETH in local config.
- Keep one supervisor-owned websocket session for the bounded stream set.
- Send subscribe commands at connect time and after every reconnect when `resubscribeOnReconnect` is enabled.
- Track pending subscribe command IDs and treat subscribe completion as a supervisor concern.
- Respond to Binance ping/pong requirements without pulling parsing logic into the supervisor.
- Trigger proactive reconnect before Binance's 24h forced disconnect and route that reconnect through the same bounded recovery path as transport failures.
- Apply existing backoff min/max and connect-per-minute limits from config before each reconnect attempt.

## Key Decisions

- Use one combined Spot connection instead of per-stream sockets because the child feature seed explicitly wants one shared runtime owner for `trade` and `bookTicker`.
- Introduce a supervisor state struct with at least: desired subscriptions, active subscriptions, pending command IDs, connection opened time, last frame time, last pong time, reconnect attempt count, and next rollover deadline.
- Keep frame classification shallow: the supervisor only needs enough structure to tell control/ack/error frames from later market payload frames.
- Treat rollover as an intentional reconnect cause, not a special second lifecycle, so the same reconnection accounting and feed-health transitions stay testable.
- Do not add depth subscriptions, snapshot timers, or resync state here; if implementation needs them, stop and move that work to `plans/epics/binance-spot-depth-bootstrap-and-recovery/`.

## Unit Test Expectations

- Subscription-set construction is deterministic from config order and symbol inventory.
- Connect startup sends one subscribe command covering the four expected stream names.
- Successful reconnect reuses the same desired subscriptions and clears stale pending command state.
- Control-frame handling refreshes liveness without requiring trade or `bookTicker` payload decoding.
- Proactive rollover triggers before the 24h venue cutoff and increments reconnect accounting only once per rollover.
- Backoff and connect-rate enforcement are deterministic under synthetic time.

## Contract / Fixture / Replay Impacts

- No schema changes are expected.
- No canonical trade or top-of-book fixtures are owned by this module.
- Add deterministic supervisor-script fixtures or testdata for subscribe ack, ping, pong, close, and reconnect timing so later implementation does not depend on live Binance access.
- Replay impact is limited to ensuring later feed-health events reuse the already-locked Wave 1 source-record pattern.

## Summary

This module defines the supervisor as the single owner of Spot websocket lifecycle and subscription state for BTC/ETH `trade` plus `bookTicker`, with explicit boundaries that keep parsing and depth recovery out of scope.

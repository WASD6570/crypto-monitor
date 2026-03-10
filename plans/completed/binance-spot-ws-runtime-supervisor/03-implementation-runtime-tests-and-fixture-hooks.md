# Implementation Module 3: Runtime Tests And Fixture Hooks

## Scope

- Plan the deterministic test harnesses and fixture hooks needed to implement and validate the Spot supervisor without live Binance access.
- Cover scripted websocket behavior, synthetic time control, and focused integration seams.
- Exclude end-to-end market-state API cutover and exclude parser-level fixture expansion for trade or top-of-book payloads.

## Target Repo Areas

- `services/venue-binance/spot_ws_supervisor_test.go`
- `services/venue-binance/runtime_test.go`
- `tests/integration` new focused Binance supervisor test file
- `tests/fixtures/events/binance` or `services/venue-binance/testdata` for scripted runtime fixtures

## Requirements

- Provide a fake websocket peer or scripted frame driver that can emit open, ping, subscribe-ack, idle, close, and reconnect sequences deterministically.
- Use synthetic time so heartbeat timeout, staleness checks, backoff windows, connect-rate limits, and rollover deadlines are testable without sleeps.
- Keep fixtures small and scenario-based rather than mirroring full market payload corpora.
- Give later parser child features a reusable supervisor harness that can hand off accepted raw payload frames without re-testing socket orchestration from scratch.
- Add one focused integration seam that proves supervisor feed-health can be normalized downstream while stream parsing remains stubbed.

## Key Decisions

- Prefer service-local `testdata` or scenario structs for websocket control-flow scripts; use shared `tests/fixtures/events/binance` only when the artifact should be reused across later integration work.
- Separate unit and integration responsibilities:
  - unit tests own lifecycle sequencing, timers, and state transitions
  - integration smoke owns supervisor plus normalizer feed-health handoff
- Keep failure injection explicit: subscribe ack missing, ping unanswered, close during rollover, and reconnect budget exceeded should each map to one test case.
- Do not make tests depend on live Binance timing quirks beyond the fixed lifecycle rules already chosen in this plan.

## Unit Test Expectations

- Healthy startup script reaches subscribed/connected state and emits healthy feed-health.
- Missing traffic past `heartbeatTimeoutMs` and `messageStaleAfterMs` degrades exactly as planned.
- Reconnect path applies bounded exponential backoff and resubscribe-on-reconnect behavior.
- Connect-rate limit prevents reconnect storm behavior under repeated failures.
- Proactive rollover script reconnects cleanly before the 24h deadline and preserves the desired subscription set.
- Feed-health normalization smoke proves the emitted message remains consumable by `services/normalizer` with Wave 1 provenance fields intact.

## Contract / Fixture / Replay Impacts

- No replay engine changes belong here.
- New fixture hooks should focus on supervisor control flow and expected feed-health outputs, not canonical trade/top-of-book outputs.
- Any reused fixture artifacts should document that they are lifecycle fixtures for the Spot supervisor so later parser slices do not misread them as payload-contract fixtures.

## Summary

This module defines the deterministic harnesses that make the supervisor implementation safe to build and review without live network dependence, and it leaves reusable hooks for the later trade and top-of-book child features.

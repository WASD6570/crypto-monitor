# Refinement Map: Binance Live Market State API Cutover

## Current Status

This Wave 5 epic is newly materialized from the initiative handoff and still needs decomposition before `feature-planning`.

## What Is Already Covered

- `plans/completed/binance-spot-trade-canonical-handoff/`, `plans/completed/binance-spot-top-of-book-canonical-handoff/`, `plans/completed/binance-spot-depth-bootstrap-and-buffering/`, and `plans/completed/binance-spot-depth-resync-and-snapshot-health/` already settle the accepted Binance Spot live families this epic must consume instead of redefining.
- `plans/completed/binance-live-raw-append-and-feed-health-provenance/` and `plans/completed/binance-live-replay-binance-family-determinism/` already settle replay-safe identity, degraded feed-health retention, and deterministic audit behavior for the Binance inputs that back this cutover.
- `plans/completed/market-state-api-compose-integration/` already created the stable Go API boundary, same-origin `/api` path, health endpoint, and local compose/browser seam that this epic must preserve.
- `services/feature-engine/service.go`, `services/regime-engine/service.go`, and `libs/go/features/market_state_current.go` already provide the current-state and global-state builders once a live query source exists.
- `apps/web/src/api/dashboard/dashboardClient.ts` and `apps/web/src/api/dashboard/dashboardDecoders.ts` already assume the current consumer contract and should remain stable through the cutover.

## What Remains

- replace the deterministic current-state bundle builder in `services/market-state-api` with an explicit live source seam
- define the first live current-state assembly posture for `BTC-USD` and `ETH-USD` using accepted Binance Spot inputs while keeping the existing contract honest about unavailable or degraded sections
- preserve symbol and global current-state response stability while the backing provider changes from deterministic builders to live runtime state
- wire `cmd/market-state-api`, compose, and browser verification to the live-backed provider and document operator-visible limits and degradation behavior

## What This Epic Must Not Absorb

- new venue parsing, canonical schema work, raw append changes, or replay-engine redesign
- broad frontend redesign or client-owned market logic
- a validation-only smoke or integration child feature detached from implementation
- immediate USD-M weighting or contract redesign unless a later bounded slice is explicitly opened for that purpose

## Refinement Waves

### Wave 6A

- `binance-live-current-state-query-assembly`
- Why first: the API cutover is unsafe until one service-owned live read model exists for symbol and global current-state assembly.

### Wave 6B

- `binance-live-market-state-api-provider-cutover`
- Why next: command wiring, compose validation, browser checks, and operator docs depend on the live query assembly and degradation posture from Wave 6A being settled.

## Notes For Future Planning

- keep the first cutover Spot-driven unless a later bounded slice explicitly brings USD-M into current-state or regime inputs
- keep slow-context optional and non-blocking just as the current feature-engine seam already behaves
- preserve the existing dashboard decoder contract instead of introducing a transitional API shape
- no safe parallel child planning exists until the live query assembly posture is settled
- no MCP servers are configured in the current session, but refinement is not blocked because all required context is repo-local

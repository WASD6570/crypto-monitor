# Binance Live Market State API Cutover

## Epic Summary

Replace the deterministic `services/market-state-api` provider with a live Binance-backed current-state source while keeping the existing Go API boundary, same-origin `/api` behavior, and dashboard response contracts stable for `BTC-USD` and `ETH-USD`.

## In Scope

- replace package-local deterministic provider behavior behind `services/market-state-api`
- assemble live `SymbolCurrentStateQuery` and `GlobalCurrentStateQuery` inputs from accepted Binance Spot state for `BTC-USD` and `ETH-USD`
- preserve the current `/api/market-state/global` and `/api/market-state/:symbol` response shape, health endpoint, and same-origin browser path
- keep slow-context behavior non-blocking while documenting operator-visible degradation and first-cutover limits
- update compose and browser verification so the web app reads the live-backed Go API path rather than deterministic bundle builders

## Out Of Scope

- frontend-owned venue logic, browser-side market-state derivation, or a second API boundary
- reopening Binance parsing, depth recovery, raw append, or replay behavior already settled by completed initiative slices
- broad current-state schema redesign, including redefinition of the existing `world` and `usa` contract sections
- making USD-M context a regime or composite input before a bounded follow-on slice proves the weighting and freshness rules

## Target Repo Areas

- `services/market-state-api`
- `services/feature-engine`
- `services/regime-engine`
- `cmd/market-state-api`
- `apps/web/tests/e2e`
- `docker-compose.yml`
- `services/market-state-api/README.md`
- `README.md`

## Validation Shape

- targeted Go tests for the live provider seam, query assembly behavior, unsupported-symbol handling, and degradation paths
- deterministic integration checks that pinned Binance accepted inputs produce stable symbol and global current-state responses across repeated runs
- compose and browser checks against the same-origin `/api` path served by the live-backed `market-state-api`
- direct degradation and contract-stability checks attached to the owning implementation slices, not a validation-only child feature

## Major Constraints

- preserve the existing current-state consumer contract, including `world`, `usa`, 30s/2m/5m bucket sections, slow-context shape, and reserved history/audit seam
- treat the first live cutover as Spot-driven for current-state and regime inputs; keep absent `usa` or USD-M influence explicit rather than filling it with new deterministic placeholders
- preserve explicit feed-health and timestamp-degradation visibility from the completed Binance live and replay slices
- keep `BTC-USD` and `ETH-USD` as the only supported symbols for this epic unless a later plan expands the contract deliberately
- keep Go as the live runtime path; Python remains offline-only

# Child Plan Seeds

## `binance-spot-runtime-read-model-owner`

- Outcome: add one sustained process-owned Spot runtime owner that composes the completed websocket supervisor, depth bootstrap, and depth recovery behavior into continuously updated current-state observations for `BTC-USD` and `ETH-USD`
- Primary repo area: `cmd/market-state-api`, `services/venue-binance`
- Dependencies: completed `binance-spot-ws-runtime-supervisor`, `binance-spot-depth-bootstrap-and-buffering`, and `binance-spot-depth-resync-and-snapshot-health`
- Validation shape: targeted Go tests for runtime startup, accepted frame progression, bootstrap success, gap-triggered resync, stale/degraded state, and reconnect carry-forward
- Why it stands alone: this is the missing runtime plumbing seam; without it, the command still has no long-lived owner to feed current-state queries

## `binance-market-state-live-reader-cutover`

- Outcome: replace the command-local polling snapshot reader with the sustained read model, keep the existing live provider contract unchanged, and prove current-state responses still preserve warm-up and degradation semantics through the real API path
- Primary repo area: `cmd/market-state-api`, `services/market-state-api`, `tests/integration`, `tests/replay`
- Dependencies: `binance-spot-runtime-read-model-owner`
- Validation shape: targeted provider and command tests, focused integration checks against `/api/market-state/global` and `/api/market-state/:symbol`, and repeated-input determinism proof for the read model outputs used by current-state assembly
- Why it stands alone: it is the consumer-facing cutover slice and should stay separate from lower-level runtime orchestration so API-contract proof remains focused and reviewable

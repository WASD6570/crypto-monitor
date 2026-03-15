# Binance Live Market Data Handoff

## Refined Epic Queue

No active queue remains. This initiative is archived and superseded by `initiatives/crypto-market-copilot-binance-integration-completion/`.

## Execution State

- Initiative status: `archived`
- Historical reference only: `plans/epics/binance-live-contract-seams-and-fixtures/`, `plans/epics/binance-spot-trades-and-top-of-book-runtime/`, `plans/epics/binance-usdm-context-sensors/`, `plans/epics/binance-spot-depth-bootstrap-and-recovery/`, `plans/epics/binance-live-raw-storage-and-replay/`, `plans/epics/binance-live-market-state-api-cutover/`
- Next recommended epic: none
- Parallel-safe now: none

| Epic | Status | Depends On | Parallel With | Next Action | Notes |
|---|---|---|---|---|---|
| `plans/epics/binance-live-contract-seams-and-fixtures/` | `archived` | Completed | - | Use archived feature evidence only | Historical prerequisite for later Binance work |
| `plans/epics/binance-spot-trades-and-top-of-book-runtime/` | `archived` | Completed | - | Use archived feature evidence only | Historical prerequisite for later Binance work |
| `plans/epics/binance-usdm-context-sensors/` | `archived` | Completed | - | Use archived feature evidence only | Historical prerequisite for later Binance work |
| `plans/epics/binance-spot-depth-bootstrap-and-recovery/` | `archived` | Completed | - | Use archived feature evidence only | Historical prerequisite for later Binance work |
| `plans/epics/binance-live-raw-storage-and-replay/` | `archived` | Completed | - | Use archived feature evidence only | Historical prerequisite for later Binance work |
| `plans/epics/binance-live-market-state-api-cutover/` | `archived` | Completed | - | Use archived feature evidence only | Superseded by the later integration-completion initiative |

## Planning Waves

### Wave 1

- `binance-live-contract-seams-and-fixtures`
- Why now: every later slice depends on one stable answer for canonical symbols, timestamp resolution, source-record identity, and fixture vocabulary.

### Wave 2

- `binance-spot-trades-and-top-of-book-runtime`
- `binance-usdm-context-sensors`
- Why parallel: both consume the same contract decisions but touch different live surfaces and do not need to redefine the same recovery path.

### Wave 3

- `binance-spot-depth-bootstrap-and-recovery`
- Why later: depth recovery inherits Spot runtime behavior but adds sequence, snapshot, and resync risk that should not block initial live trade/top-of-book and USD-M context planning.

### Wave 4

- `binance-live-raw-storage-and-replay`
- Why later: replay and audit acceptance should lock in after the live Spot and USD-M semantics are stable enough to avoid duplicate identity drift.

### Wave 5

- `binance-live-market-state-api-cutover`
- Why later: consumer cutover should happen only after live ingestion, resync behavior, and replay-safe identities are already understood.

## Refined Epics

No active refined epics remain for this archived initiative. Use `## Execution State` for historical references and the later Binance initiatives plus `plans/completed/` for the implementation history.

## Historical Open Questions

These questions are retained only as historical context for later Binance initiatives:

- Spot-only versus USD-M-influenced current-state semantics were deferred to later initiative work.
- `openInterest` freshness and polling defaults remained rollout-sensitive and were deferred to later environment work.
- Connection-shape tradeoffs between combined versus stream-family-specific sockets remained historical design context.
- Any schema-gap evaluation for live Binance semantics remained historical context for later contract work.

# Binance USD-M Output Application And Replay Proof

## Ordered Implementation Plan

1. Add one conservative application policy that consumes the settled internal USD-M influence signal and caps symbol/global regime output only when the posture is `DEGRADE_CAP`, while leaving `AUXILIARY`, `NO_CONTEXT`, and `DEGRADED_CONTEXT` as spot-derived output plus provenance.
2. Add the smallest additive current-state provenance seam needed to explain when USD-M influence was evaluated and whether it changed the result, without changing `/healthz` or widening `/api/market-state/*` beyond an optional machine-readable summary.
3. Wire the live and deterministic market-state assembly paths to evaluate and apply the signal from the already-landed venue-side input owner, keeping Binance USD-M acquisition inside Go-owned services and reusing the existing tracked symbol set.
4. Add focused API, integration, and replay proof that repeated pinned Spot plus USD-M inputs yield identical current-state and regime outputs and that the route shape stays backward-compatible apart from the planned additive provenance.
5. Keep this feature active under `plans/binance-usdm-output-application-and-replay-proof/` until implementation and validation complete, then move the full directory and `testing-report.md` to `plans/completed/binance-usdm-output-application-and-replay-proof/`.

## Requirements

- Scope is limited to applying the already-settled USD-M influence signal to current-state and regime assembly plus the proof required to trust the resulting consumer-facing behavior.
- Keep Go as the live runtime path; Python remains offline-only.
- Keep the tracked symbols fixed to `BTC-USD` and `ETH-USD`.
- Reuse the completed internal signal contract and venue-side input seam from `plans/completed/binance-usdm-influence-policy-and-signal/`; do not reopen new USD-M acquisition scope beyond the smallest live wiring needed to feed that existing seam.
- Keep `/healthz` process-only and outside this child.
- Keep `GET /api/market-state/global` and `GET /api/market-state/:symbol` backward-compatible by default; if added metadata is required for operator honesty, keep it optional, additive, and machine-readable.
- Apply only the conservative negative posture that was already settled in the first child: `DEGRADE_CAP` may cap output to `WATCH`, while `AUXILIARY`, `NO_CONTEXT`, and `DEGRADED_CONTEXT` do not silently worsen spot-derived regime output.
- Keep replay determinism and repeated-run stability explicit for both the internal signal application and the public current-state responses.

## Design Notes

### Application posture

- Treat the settled USD-M signal as a bounded cap, not a new scoring system.
- When a symbol signal is `DEGRADE_CAP`, cap the symbol regime to `WATCH` and attach one explicit regime reason that the API can expose without reverse-engineering evaluator internals.
- When a symbol signal is `AUXILIARY`, `NO_CONTEXT`, or `DEGRADED_CONTEXT`, preserve the existing spot-derived symbol and global outcomes and expose the signal only through provenance.
- Let global regime reflect the adjusted symbol snapshots rather than inventing a second independent USD-M policy layer.

### Provenance seam

- Add one small optional provenance summary on symbol current-state output that makes USD-M evaluation visible when this child is active.
- Keep the public summary narrower than the internal signal schema: posture, primary reason, whether output was capped, observed timestamp, and version metadata are enough.
- Prefer adding detailed trigger metrics only if implementation proves the smaller summary is not sufficient for operator honesty.
- Keep the top-level route shape stable; avoid a broad schema-family reset for one additive seam.

### Live-path boundary

- Keep USD-M data acquisition and feed-health ownership inside `services/venue-binance` plus `cmd/market-state-api`.
- Add the smallest read-only seam needed for `services/market-state-api` to obtain either the raw evaluator input snapshot or the already-evaluated signal set for the current bundle.
- Do not add browser-side Binance logic, new Python dependencies, or runtime-health work in this child.

## ASCII Flow

```text
Binance Spot current-state bundle
  - composites
  - buckets
  - spot-derived symbol regime
            |
            +------------------------------+
                                           |
existing USD-M input owner                 |
  - funding / mark-index / liquidation     |
  - open interest                          |
  - feed-health freshness                  |
            |                              |
            v                              v
services/feature-engine evaluates settled USD-M influence signal
            |
            v
bounded application policy
  - AUXILIARY / NO_CONTEXT / DEGRADED_CONTEXT => provenance only
  - DEGRADE_CAP => cap symbol/global output to WATCH as needed
            |
            v
services/market-state-api current-state responses
  - GET /api/market-state/:symbol
  - GET /api/market-state/global

unchanged in this child
  - GET /healthz
  - runtime-health ownership and operator route policy
```

## Archive Intent

- Keep this feature active under `plans/binance-usdm-output-application-and-replay-proof/` while implementation and validation are in progress.
- When complete, move the full directory and `testing-report.md` to `plans/completed/binance-usdm-output-application-and-replay-proof/`.

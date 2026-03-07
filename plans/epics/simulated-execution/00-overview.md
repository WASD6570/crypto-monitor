# Simulated Execution Overview

## Ordered Implementation Plan

1. Define deterministic simulation inputs, instrument/mode coverage, and fill-model boundaries from emitted alerts plus outcome-evaluation path data.
2. Add cost, latency, leverage, and confidence-label logic so spot long and perp long/short runs produce conservative, auditable net results.
3. Persist append-only simulation runs and expose saved-run review surfaces without crossing into live trading or exchange-auth flows.
4. Validate replay determinism, degraded L2 refusal vs fallback behavior, storage auditability, and negative-path safety.

## Problem Statement

Outcome evaluation tells the user whether the market moved in the alert's favor, but it does not answer whether a realistic execution would still have been worth attention after latency, slippage, fees, and leverage rules. Initiative 2 needs a separate, deterministic simulation layer that reuses trusted alert and outcome context, stays clearly labeled as hypothetical, and never blurs into live order submission.

## Role In Initiative 2

- This is slice 5 of `crypto-market-copilot-alerting-and-evaluation`.
- It extends the trust loop after `outcome-evaluation`: first the system records what the market did, then it estimates what an execution would likely have done under explicit assumptions.
- It is the feature that directly supports the Initiative-2 success gate that simulated execution must show whether edge survives costs and latency.
- Later review, notes, analytics, and baseline-comparison work should read persisted simulation runs instead of recomputing assumptions in the UI.

## Relationship To Outcome Evaluation

- `outcome-evaluation` remains the source of truth for market-path facts: target ordering, invalidation ordering, MAE, MFE, and horizon windows.
- `simulated-execution` consumes the same alert and post-alert path, then adds execution-only assumptions: venue choice, order start delay, slippage, fees, leverage, and confidence labeling.
- Outcome records may expose conservative net-viability boundary fields, but final simulated fill logic, partial fills, fallback rules, and liquidation-aware perp behavior belong here.
- Simulation runs must store references to the originating `alertId`, `outcomeRecordId`, replay manifest or live data window, and pinned config versions so both layers stay comparable.

## First-User Outcome

The user should be able to open a recent alert and see a clearly labeled answer to: "If I had tried to take this with conservative defaults, would spot long or perp long/short still look viable after realistic delay and cost assumptions?"

## In Scope

- deterministic simulation runs seeded from stored alerts and outcome-evaluation inputs
- support for `spot-long`, `perp-long`, and `perp-short` modes
- leverage options for perps, including a safe supported set that explicitly includes `x5`
- latency defaults and operator-selectable conservative presets
- slippage behavior using L2 data when healthy and clearly labeled fallback assumptions when L2 is unavailable but simulation can still be approximated safely
- fee defaults by instrument mode and maker/taker assumption policy
- confidence labeling when assumptions weaken realism
- saved simulation runs, append-only storage, and operator review retrieval
- strict separation from live trading, private endpoints, and exchange credentials

## Out Of Scope

- real order placement, paper-trading broker sessions, or any exchange API credentials
- portfolio sizing, balance management, margin transfers, or account-risk state
- multi-leg strategies, options, hedging baskets, or cross-exchange routing optimization
- automatic threshold mutation from simulation results alone
- client-side simulation math that bypasses service-owned truth

## Safe Defaults For Ambiguous Areas

- Default symbols: plan first for BTC and ETH because Initiative 1 tradeability defaults already center on them.
- Default venue choice: use the primary trusted composite-driving venue configured for the alert's symbol and mode; if no mode-specific venue is pinned, choose the most liquid healthy venue from the trusted venue set and record the selection reason.
- Default spot instrument: simulate long-only spot on the canonical quote pair already used for alert evaluation; no synthetic shorting.
- Default perp leverage menu: `x1`, `x2`, `x3`, and `x5`; no leverage above `x5` in MVP planning.
- Default latency posture: apply at least one conservative non-zero delay from alert emission to order-start time; never assume instantaneous fill.
- Default fee posture: assume taker entry and taker exit unless a future saved-run preset explicitly opts into maker logic.
- Default low-confidence posture: degrade labels before claiming optimism; if assumptions are materially weak, mark the run low confidence even if PnL is positive.
- Default live-boundary posture: every payload, route, and UI surface must say `SIMULATED` or equivalent and must not share code paths with live order submission because no live order submission exists in scope.

## Requirements

- Keep execution math and decision ownership in Go services or shared Go helpers; `apps/web` only requests, renders, filters, and compares stored simulation runs.
- Reuse trusted service-side alert and outcome identifiers; do not let the UI reconstruct price paths or cost assumptions.
- Preserve deterministic results for the same alert, path inputs, config snapshot, selected preset, and code version wherever the underlying market data is deterministic.
- Store explicit confidence and assumption fields whenever fallbacks, degraded timestamps, missing L2 depth, or venue substitutions reduce realism.
- Keep simulation runs append-only and auditable with `configVersion`, `algorithmVersion`, schema version, timestamp-source metadata, and operator/request provenance.
- Refuse runs that would imply live trading integration, missing core identifiers, or unsupported instrument/mode combinations.
- Maintain a hard boundary from exchange credentials, private REST/WebSocket calls, or anything that could mutate external trading state.

## Target Repo Areas

- `services/simulation-api`
- `services/outcome-engine`
- `services/alert-engine`
- `libs/go`
- `apps/web`
- `schemas/json/simulation`
- `schemas/json/outcomes`
- `schemas/json/alerts`
- `schemas/json/replay`
- `configs/*`
- `tests/fixtures`
- `tests/replay`
- `tests/integration`

## Design Overview

### Boundary: Simulation Is Not Trading

- Simulation estimates hypothetical fills from trusted market data and stored assumptions.
- It must never require exchange authentication, order intent signing, account state, or balance reads.
- If the product later adds a paper-trading experiment, that must remain a separate feature with separate contracts and service boundaries.

### Coverage Model

- `spot-long` answers the simplest question: would buying spot after alert latency and exiting on objective rules still look worthwhile?
- `perp-long` and `perp-short` reuse the same market-path and fee/slippage framework, but they must additionally capture leverage choice, liquidation-distance guardrails, and funding or carry assumptions only if deterministic inputs exist.
- MVP simulation should prefer conservative underfitting over optimistic realism theater.

### L2 Refusal vs Fallback Policy

- If healthy L2 order book snapshots exist around the simulated entry or exit windows, use them for depth-walk slippage estimates.
- If L2 is degraded but top-of-book plus trusted trade path still exists, allow a fallback slippage model only when the run is clearly labeled `LOW_CONFIDENCE` or similar and the fallback formula is pinned by config.
- Refuse the run instead of falling back when depth is critically degraded, the selected venue is unhealthy, or the fallback would require inventing unsupported liquidity assumptions.
- Store the exact reason code: `L2_FALLBACK_USED`, `L2_REFUSED_CRITICAL_DEGRADATION`, `VENUE_SUBSTITUTED`, and similar stable labels.

### Storage And Audit Expectations

- Keep simulation records hot for at least the operating-default alert/outcome/simulation retention window and cold for at least the same 2-year minimum.
- Persist the full assumption envelope: symbol, instrument mode, side, leverage, venue selection, entry/exit latency, fee preset, slippage method, confidence label, and why the label was assigned.
- Link every run to the originating alert, outcome record, and replay/live source context so an operator can audit a result without reconstructing the path manually.
- Never overwrite a prior saved run; new presets or reruns append new records with provenance.

## ASCII Flow

```text
emitted alert -----------> outcome evaluation record
- alertId                 - outcomeRecordId
- setup/direction         - MAE/MFE/horizon path
- target/invalidation     - regime + degradation context
         \                         /
          \                       /
           +---- simulation request/preset ----+
           |    - mode: spot-long/perp-long/perp-short
           |    - leverage: x1/x2/x3/x5 for perps
           |    - latency/fee/slippage preset
           v
      simulation engine
      - choose venue/instrument safely
      - read L2 if healthy
      - fallback or refuse if degraded
      - apply fees, slippage, latency
      - assign confidence label
           |
           +------------------+
           |                  |
           v                  v
    append-only run      refused run record
    - simulated fills    - reason codes
    - gross/net PnL      - audit metadata
    - confidence         - no trading side effects
    - assumption envelope
           |
           v
      operator surfaces
      - saved runs list
      - alert drill-down comparison
      - explicit `SIMULATED` labeling
```

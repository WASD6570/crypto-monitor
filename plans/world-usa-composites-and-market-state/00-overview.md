# World USA Composites And Market State

## Ordered Implementation Plan

1. Build composite inputs, venue eligibility, and weighting/clamping policy in Go service-owned logic
2. Build 30s, 2m, and 5m feature buckets plus fragmentation and market-quality metrics from composite outputs
3. Classify 5m symbol and global market state as `TRADEABLE`, `WATCH`, or `NO-OPERATE`
4. Expose versioned query and contract surfaces for UI and later alert consumers without client recomputation
5. Validate deterministic replay, degraded-feed behavior, and negative-path regime transitions

## Problem Statement

The visibility initiative needs a single trusted answer for BTC and ETH across offshore and USA liquidity. Raw venue rows are necessary for auditability, but the first user workflow depends on service-owned WORLD and USA composites that turn fragmented multi-venue feeds into a readable market-quality and tradeability gate.

## Role In The Visibility Initiative

- This is slice 4 of `crypto-market-copilot-visibility-foundation` and the first place where the operator gets a coherent cross-venue market-state answer instead of separate health and replay primitives.
- It converts canonical events plus feed-health inputs into outputs the first user can trust within 60 seconds: composite direction, USA vs WORLD alignment, degraded conditions, and whether the symbol is tradeable now.
- It is the bridge between the ingestion/replay foundation and later UI and alerting work; later surfaces should explain these outputs, not recreate them.

## First-User Workflow

1. Open the product and select BTC or ETH.
2. Read current WORLD composite, USA composite, and divergence status.
3. See whether feed quality or venue fragmentation is degrading confidence.
4. Read the current 5m symbol regime and the global regime ceiling.
5. Decide whether to watch the market, ignore it, or trust later alerts to escalate.

## In Scope

- WORLD and USA composite construction for BTC and ETH from canonical venue inputs
- venue eligibility, weighting, clamping, and degraded-input handling policy
- stablecoin normalization policy for WORLD spot/perp inputs at planning level
- divergence and fragmentation metrics between WORLD and USA
- market-quality metrics that incorporate feed health and microstructure trust
- 30s, 2m, and 5m feature buckets derived in service-owned logic
- 5m `TRADEABLE`, `WATCH`, and `NO-OPERATE` classification for symbol and global state
- versioned consumer contracts and query surfaces for dashboards and later alert engines
- replay and test expectations for deterministic and negative-path behavior

## Out Of Scope

- alert setup A/B/C trigger logic, delivery, and outcomes
- UI layout or client-side derived market-state math
- CME and ETF slow-context computation beyond consuming future inputs as optional context
- storage-engine implementation details already covered by raw storage and replay foundation
- speculative business logic beyond safe, explicit recommendations for implementation

## Safe Defaults For Vague Areas

- Default composite price basis: quote-normalized mid or last-trade-derived price from the canonical event stream, with the implementation choosing one service-wide basis and carrying it explicitly in config.
- Default weight basis: bounded liquidity-quality weighting using recent notional/turnover and feed-health penalties, rather than equal weighting or opaque scoring.
- Default degraded-input behavior: penalize, clamp, or exclude at venue level before a composite is marked unusable; do not silently keep stale venues at full influence.
- Default regime posture: when uncertainty is high, degrade toward `WATCH`, then `NO-OPERATE`, never toward `TRADEABLE`.
- Default stablecoin policy: normalize trusted USD proxies for WORLD composition through explicit config and mark proxy usage in outputs for audit.

## Requirements

- Keep all live composite, feature, and regime logic in Go services or shared Go helpers.
- Consume canonical contracts, feed-health semantics, and replay rules from earlier visibility-foundation plans.
- Use service-owned outputs as the source of truth; `apps/web` renders values and reasons but does not recompute composite weights, feature buckets, or regimes.
- Treat feed degradation, timestamp degradation, and venue gaps as first-class inputs to market-quality and regime logic.
- Preserve deterministic replay for the same raw inputs, config version, and code version.
- Keep config version, algorithm version, and contract version attached to outputs so historical states remain auditable.
- Keep Python optional and offline-only for parity or research; never required for live composite or regime computation.

## Target Repo Areas

- `services/feature-engine`
- `services/regime-engine`
- `libs/go`
- `schemas/json/features`
- `schemas/json/alerts` for future regime references only if needed by shared envelopes
- `configs/*`
- `tests/fixtures`
- `tests/replay`
- `tests/integration`
- `tests/parity` for optional offline parity fixtures only

## Design Overview

### Composite Intent

- WORLD answers: what offshore and globally liquid crypto venues imply right now from Binance and Bybit spot/perp sensors.
- USA answers: what regulated USA spot venues imply right now from Coinbase and Kraken.
- Composite outputs should keep provenance: contributing venues, excluded venues, penalties, and degradation reasons.

### Stablecoin Normalization Policy

- Canonical symbols remain `BTC-USD` and `ETH-USD`, but WORLD sources may quote in `USDT` or `USDC`.
- The composite plan assumes normalization happens before final weighting using configured quote-proxy factors supplied by canonical events or a trusted normalization helper.
- Safe default for MVP: treat `USD`, `USDC`, and `USDT` as separate quote contexts at ingest, then map approved stablecoin quotes into a USD-equivalent composite input only through explicit config.
- If quote normalization confidence degrades, lower market-quality and expose the reason rather than fabricating a precise USD value.

### Fragmented Regime Handling

- Fragmentation is not only price spread; it also includes disagreement in short-horizon direction, missing venues, asymmetric staleness, and unstable weight leadership.
- High fragmentation should bias the system toward `WATCH` even if one composite still trends cleanly.
- Severe fragmentation plus degraded feed health should force `NO-OPERATE` because the operator cannot trust apparent alignment.

### Config And Versioning

- Keep thresholds, venue membership, quote-normalization allowlists, weight caps, and state transition thresholds in versioned config.
- Output payloads should include at least `configVersion`, `algorithmVersion`, and schema version references.
- Replays must pin the exact config snapshot used for bucket assignment, weighting, and regime classification.

## ASCII Flow

```text
canonical events + feed health + replay/config snapshot
                     |
                     v
         service-owned venue eligibility filter
         - symbol + venue membership
         - timestamp plausibility
         - stablecoin quote normalization gate
         - stale/gap/degraded penalties
                     |
                     v
         WORLD composite           USA composite
      (Binance + Bybit)       (Coinbase + Kraken)
         - bounded weights        - bounded weights
         - clamping               - clamping
         - contribution reasons   - contribution reasons
                 \                  /
                  \                /
                   v              v
                divergence + fragmentation
                market-quality metrics
                30s / 2m / 5m feature buckets
                          |
                          v
          symbol 5m regime + global regime ceiling
          `TRADEABLE` / `WATCH` / `NO-OPERATE`
                          |
             +------------+-------------+
             |                          |
             v                          v
   apps/web read-only views     later alert/risk consumers
```

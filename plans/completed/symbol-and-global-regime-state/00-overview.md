# Symbol And Global Regime State

## Ordered Implementation Plan

1. Implement conservative 5m symbol regime classification that reads completed bucket summaries and emits `TRADEABLE`, `WATCH`, or `NO-OPERATE` with explicit reason outputs.
2. Implement global ceiling classification plus downgrade and recovery hysteresis that caps symbol state without introducing query-contract or storage concerns.
3. Add deterministic Go unit, integration, and replay coverage and document evidence in `plans/completed/symbol-and-global-regime-state/testing-report.md`.

## Problem Statement

The epic now has deterministic WORLD and USA composite snapshots plus bucketed divergence, fragmentation, coverage, timestamp-trust, and market-quality summaries. What is still missing is the conservative trust gate that converts those 5m summaries into the service-owned symbol regime and global ceiling state later readers can trust without reassembling market logic.

## Bounded Scope

- 5m symbol regime classification for `BTC-USD` and `ETH-USD` only
- symbol state outputs limited to `TRADEABLE`, `WATCH`, and `NO-OPERATE`
- conservative downgrade and slower recovery behavior driven by deterministic bucket windows
- explicit reason outputs, transition metadata, and version provenance for every emitted state
- global ceiling classification that caps symbol state per operating defaults
- config-versioned thresholds, hysteresis windows, and tie-break rules needed for regime classification
- replay-safe validation of downgrade, recovery, fragmentation, and degraded-feed behavior

## Out Of Scope

- current-state query contracts, API routes, pagination, or dashboard response envelopes
- storage layout, persistence schema design, audit-read retrieval, or replay-history lookups
- recomputing composite weighting, bucket math, or feed-health semantics already completed upstream
- alert setup logic, risk decisions, or UI copy/presentation behavior
- slow-context dependencies such as CME or ETF inputs as hard requirements for realtime classification

## Requirements

- Build directly on completed bucket outputs from `plans/completed/market-quality-and-divergence-buckets/` and preserve the completed composite and ingestion trust seams.
- Keep the live path in Go under `services/regime-engine` and shared Go helpers; Python remains offline-only.
- Use 5m bucket summaries as the only classification inputs for this slice; do not pull current-state query contracts into scope beyond naming downstream seams.
- Degrade quickly on critical trust loss, severe fragmentation, unavailable composite state, or coverage collapse.
- Recover more conservatively than downgrade, using explicit persistence across consecutive closed windows rather than wall-clock timers.
- Emit explicit reason families, trigger metrics, `configVersion`, `algorithmVersion`, and schema/version seams so replay differences are explainable.
- Stay storage-neutral and UI-neutral: define live runtime behavior and service-owned outputs only.

## Target Repo Areas

- `services/regime-engine`
- `libs/go/features`
- `configs/local`
- `configs/dev`
- `configs/prod`
- `schemas/json/features` only if a regime output seam must be reserved now
- `tests/fixtures`
- `tests/integration`
- `tests/replay`

## Module Breakdown

### 1. Symbol Regime Classification

- Own 5m symbol classification inputs, threshold evaluation, downgrade/recovery hysteresis, and explicit symbol-level reason outputs.
- Keep bucket math upstream and query/read-model packaging downstream.

### 2. Global Ceiling And Recovery Rules

- Own global market ceiling state, cross-symbol ceiling application, and deterministic transition rules when one or both symbols degrade.
- Keep storage, current-state API design, and alert/risk consumer policy out of scope except as named downstream seams.

## Acceptance Criteria

- Another agent can implement the regime slice without reopening the parent epic or redoing bucket logic.
- Repo areas for Go runtime logic, config, fixtures, and validation are explicit.
- Out-of-scope boundaries clearly hold query-contract and history/audit work for later slices.
- Validation commands are deterministic and specific enough for another agent to run without extra interpretation.

## ASCII Flow

```text
closed 5m bucket summaries + config snapshot + replay provenance
                       |
                       v
         symbol regime evaluator per symbol
         - fragmentation severity
         - market-quality caps
         - coverage completeness
         - timestamp-trust loss
         - unavailable/degraded reasons
                       |
                       v
      BTC 5m state               ETH 5m state
  `TRADEABLE/WATCH/NO-OPERATE` `TRADEABLE/WATCH/NO-OPERATE`
      + reason outputs            + reason outputs
               \                  /
                \                /
                 v              v
            global ceiling evaluator
            - shared trust failures
            - cross-symbol persistence
            - downgrade / recovery hysteresis
                       |
                       v
          service-owned regime outputs only
              +--------------------------+
              |                          |
              v                          v
  later current-state query slice   later alert/risk consumers
   packages outputs as read models   read authoritative state
```

## Live-Path Boundary

- This feature stops at deterministic symbol and global regime output production in Go.
- Later query work may expose these outputs, but it must not redefine thresholds, hysteresis, or reason assembly.
- Later history/audit work may retrieve emitted regime artifacts, but this plan does not choose storage or retrieval mechanics.

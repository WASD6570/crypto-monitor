# Market Quality And Divergence Buckets

## Ordered Implementation Plan

1. Implement deterministic bucket assignment, watermark handling, and 30s-to-2m-to-5m window rollups for WORLD and USA composite inputs.
2. Implement service-owned bucket outputs for divergence, fragmentation, coverage, timestamp trust, and market-quality summaries without embedding final regime decisions.
3. Add targeted Go unit, integration, and replay coverage for bucket determinism, degraded-path behavior, and late-event handling.

## Problem Statement

`world-usa-composite-snapshots` now provides deterministic WORLD and USA composite seams, but later regime and current-state work still lacks a stable bucket layer. This feature creates the bounded live-path bucket families that summarize whether WORLD and USA are aligned, fragmented, well-covered, and timestamp-trustworthy across 30s, 2m, and 5m windows.

## Bounded Scope

- deterministic 30s, 2m, and 5m bucket families only
- service-owned divergence summaries between WORLD and USA composites
- fragmentation summaries based on disagreement, contributor churn, and asymmetric degradation
- coverage summaries derived from configured versus contributing composite members
- timestamp-trust summaries based on `exchangeTs` primary assignment and `recvTs` fallback markers
- market-quality summaries that cap trust conservatively for later regime/query readers
- config-versioned thresholds, watermarks, and downgrade caps needed for the bucket layer
- replay-safe live behavior for on-time, slightly late, and too-late events

## Out Of Scope

- final `TRADEABLE`, `WATCH`, or `NO-OPERATE` classification
- global ceiling logic and symbol state transition hysteresis
- current-state query contracts, API shapes, or UI payload design beyond naming downstream seams
- storage layout, audit-read retrieval, backfill orchestration, or replay-history lookup behavior
- new composite weighting or quote-normalization policy beyond consuming the completed snapshot seam

## Requirements

- Keep the live path in Go under `services/feature-engine` and `libs/go`; Python remains offline-only.
- Build only on completed prerequisite seams from canonical contracts, ingestion/feed health, and composite snapshots.
- Follow operating defaults exactly for UTC bucket assignment, `exchangeTs` precedence, `recvTs` fallback, and lateness windows.
- Keep thresholds, watermarks, bucket closure rules, and severity cutoffs config-versioned and replay-pinned.
- Preserve enough provenance for later regime and current-state features to consume bucket outputs without recomputing divergence or quality math.
- Stay storage-neutral and UI-neutral: define runtime behavior and output seams only.

## Target Repo Areas

- `services/feature-engine`
- `libs/go/features`
- `configs/local`
- `configs/dev`
- `configs/prod`
- `schemas/json/features` only if a bucket output contract seam must be reserved now
- `tests/fixtures`
- `tests/integration`
- `tests/replay`

## Module Breakdown

### 1. Bucket Assignment And Windowing

- Own canonical 30s bucket assignment, 2m and 5m rollup construction, watermark rules, and late-event disposition.
- Keep bucket closure deterministic and auditable with explicit bucket-source and lateness metadata.

### 2. Divergence, Fragmentation, And Quality Outputs

- Own the emitted bucket payloads and metric families for divergence, fragmentation, coverage, timestamp trust, and market quality.
- Expose only downstream seams for later regime and query work; do not classify or shape read APIs here.

## Acceptance Criteria

- Another agent can implement the bucket layer without reopening the epic or the completed composite feature.
- Repo areas for runtime logic, config, fixtures, and validation are explicit.
- The plan keeps regime/query work out of scope while naming the bucket summaries those later slices will read.
- Validation commands cover deterministic bucket assignment, degraded-path summaries, watermark behavior, and replay reproduction for the same fixture window.

## ASCII Flow

```text
canonical events + feed health + composite snapshots + config snapshot
                             |
                             v
                timestamp choice and bucket assignment
                - exchangeTs primary
                - recvTs fallback with degraded marker
                - 30s watermark handling
                - explicit late-event disposition
                             |
                             v
                   canonical 30s bucket state
                - WORLD composite view
                - USA composite view
                - coverage/timestamp inputs
                             |
                +------------+------------+
                |                         |
                v                         v
           2m rollups                 5m rollups
      (4 closed 30s buckets)     (10 closed 30s buckets)
                \                         /
                 \                       /
                  v                     v
      divergence + fragmentation + coverage + timestamp trust
                    + market-quality summaries
                             |
               +-------------+--------------+
               |                            |
               v                            v
  later symbol/global regime slice   later current-state query slice
        reads bucket outputs only      formats bucket outputs only
```

## Live-Path Boundary

- This feature stops at deterministic bucket production inside Go service-owned logic.
- Later regime work may read 5m summaries and severity flags, but it must not move bucket math into `services/regime-engine` or `apps/web`.
- Later query work may package these outputs for consumers, but it must not redefine divergence, fragmentation, coverage, or quality calculations.

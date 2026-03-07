# Market State Current Query Contracts

## Ordered Implementation Plan

1. Define versioned current-state read-model schemas that package existing composite, bucket, and regime outputs without reopening upstream logic.
2. Implement service-owned current-state query surfaces and bounded recent-context assembly for dashboard and service consumers.
3. Validate schema compatibility, current-state query completeness, consumer trust boundaries, and replay-stable version metadata; record evidence in `plans/completed/market-state-current-query-contracts/testing-report.md`.

## Problem Statement

The repo now has deterministic WORLD and USA composite seams, bucket summaries, and a separate symbol/global regime slice. What is still missing is one service-owned current-state contract family that dashboard and service consumers can read directly, without joining internal service outputs or recomputing tradeability, fragmentation, or trust semantics on the client side.

## Bounded Scope

- versioned current-state read models only
- current query surfaces for symbol state and global state
- a bounded recent-context window packaged with current state for dashboard context and service-side explanation
- schema/version seams that preserve `configVersion`, `algorithmVersion`, bucket timestamps, and provenance metadata
- consumer mapping for `apps/web` and later service consumers as read-only readers of service-owned outputs
- deterministic validation for schema shape, integration response assembly, and replay/version stability

## Out Of Scope

- arbitrary history, audit, or bucket-range retrieval beyond the bounded recent-context window included in current responses
- replay-storage lookup mechanics, pagination, or historical artifact indexing
- recomputing composite weighting, bucket math, or regime thresholds already owned upstream
- UI layout, view composition, or presentation logic inside `apps/web`
- alert setup logic, mutation endpoints, or config-management workflows

## Requirements

- Build on completed seams from `plans/completed/world-usa-composite-snapshots/`, `plans/completed/market-quality-and-divergence-buckets/`, and `plans/completed/symbol-and-global-regime-state/` regime outputs.
- Keep the live path in Go service-owned code; Python remains offline-only.
- Keep contracts storage-neutral and UI-neutral: the plan defines response families and assembly boundaries, not persistence shape or dashboard components.
- Include enough provenance and version metadata for consumers to explain current state without joining raw venue rows or internal bucket tables.
- Preserve operating-default trust rules: services are the source of truth, global ceiling caps symbol state, and degraded conditions never drift toward a more permissive client interpretation.
- Keep history/audit reads explicitly as a later seam under `market-state-history-and-audit-reads` so current-state delivery does not inherit replay-query complexity.

## Target Repo Areas

- `schemas/json/features`
- `services/feature-engine`
- `services/regime-engine`
- `libs/go`
- `tests/fixtures`
- `tests/integration`
- `tests/replay`
- `apps/web` as a consuming surface only

## Module Breakdown

### 1. Current-State Contracts And Schema Seams

- Define the versioned schema family for packaged current market state and the bounded recent-context sections consumers receive.
- Reserve explicit seams for future history/audit retrieval without mixing history query semantics into the current-state contract family.

### 2. Service Query Surfaces And Consumer Mapping

- Define how `services/feature-engine` and `services/regime-engine` assemble and serve the read models for symbol/global consumers.
- Map how dashboard and later service consumers read, cache, and trust the service response without recomputing market logic.

## Acceptance Criteria

- Another agent can implement the current-state contract slice without reopening the parent epic or the completed composite/bucket features.
- The plan names exact repo areas for schema work, service query assembly, fixtures, and validation.
- Out-of-scope boundaries clearly hold history/audit retrieval for the later `market-state-history-and-audit-reads` slice.
- Validation commands are deterministic, concrete, and runnable by another agent without extra interpretation.

## ASCII Flow

```text
composite snapshots + bucket summaries + regime outputs + config/version context
                                 |
                                 v
                  current-state contract assembler in Go
                  - schema-versioned envelopes
                  - provenance and degraded reasons
                  - bounded recent-context slices
                  - symbol + global read models
                                 |
                 +---------------+----------------+
                 |                                |
                 v                                v
      current symbol market-state query   current global-state query
      - composite and divergence view     - global ceiling state
      - quality and regime summary        - affected symbols summary
      - recent closed-window context      - version/trust metadata
                 |                                |
                 +---------------+----------------+
                                 |
                                 v
                read-only consumers (`apps/web`, later services)
                format and react to service-owned outputs only

future seam kept separate:
historical/audit retrieval by bucket/version context
belongs to `market-state-history-and-audit-reads`
```

## Live-Path Boundary

- This feature stops at versioned current-state read models and service query assembly in Go.
- `apps/web` remains read-only and presentational.
- Later history/audit work may reuse the same contract family and metadata, but this plan does not choose storage, indexing, or replay retrieval mechanics.

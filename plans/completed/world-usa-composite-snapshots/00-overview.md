# World USA Composite Snapshots

## Ordered Implementation Plan

1. Implement contributor eligibility, quote-normalization gating, weighting inputs, and clamp policy for WORLD and USA composites in Go.
2. Implement deterministic composite snapshot outputs with contributor provenance, degradation reasons, and explicit seams for later bucket/regime/query consumers.
3. Validate deterministic weighting, degraded/excluded contributor handling, and replay-stable snapshot reproduction.

## Problem Statement

The visibility foundation needs one service-owned snapshot boundary for `BTC-USD` and `ETH-USD` before later bucket, regime, and read-model work can stay bounded. That boundary must decide which venues are eligible, how contributors are weighted, when degraded venues are penalized or excluded, and what deterministic snapshot payload downstream Go services can trust without re-deriving market logic.

## Bounded Scope

- WORLD and USA composite snapshots only.
- Contributor eligibility for configured venue members.
- Quote-normalization eligibility and confidence handling needed to include or exclude WORLD contributors.
- Deterministic weighting, clamping, and degradation/exclusion behavior.
- Snapshot output fields and service-side seams required by later bucket/regime/query work.
- Focused Go unit, integration, and replay validation for deterministic snapshot behavior.

## Out Of Scope

- 30s, 2m, or 5m divergence bucket math.
- `TRADEABLE`, `WATCH`, or `NO-OPERATE` regime classification.
- Current-state query contracts, storage design, or UI rendering concerns.
- Replay manifests, backfill orchestration, or historical audit reads beyond preserving snapshot provenance seams.
- New business logic outside contributor eligibility, weighting, degradation, and deterministic snapshot emission.

## Requirements

- Keep the live path in Go under service-owned logic; Python remains offline-only.
- Build on canonical symbol, venue, timestamp, degraded-marker, and replay expectations already established in completed prerequisite plans.
- Use `exchangeTs` first and `recvTs` fallback semantics from operating defaults, while preserving timestamp trust in contributor decisions.
- Keep venue membership, quote-proxy allowlists, weight bounds, and degradation penalties config-versioned and replay-pinned.
- Emit enough provenance for later bucket, regime, and query features to consume snapshots without recomputing contributor decisions.
- Keep the slice storage-neutral and UI-neutral: this plan defines service-owned runtime behavior and deterministic outputs, not persistence or presentation.

## Target Repo Areas

- `services/feature-engine`
- `libs/go`
- `configs/*`
- `schemas/json/features` only if a snapshot contract update is required for the service-owned output seam
- `tests/fixtures`
- `tests/integration`
- `tests/replay`

## Module Breakdown

### 1. Contributor Eligibility And Weighting

- Own WORLD and USA membership, quote-normalization gates, penalty composition, weight normalization, and clamping.
- Keep deterministic ordering and replay-pinned config as first-class requirements.

### 2. Snapshot Output And Degradation Seams

- Define the emitted snapshot shape, contributor provenance, unavailable/degraded semantics, and downstream seams for later bucket/regime/query work.
- Keep later divergence, regime, and read-surface fields explicitly out of scope except as named consumers of the snapshot boundary.

## Acceptance Criteria

- Another agent can implement WORLD and USA composite snapshots without reopening the parent epic.
- Repo areas for logic, config, fixtures, and validation are explicit.
- The plan keeps later bucket/regime/query work out of scope while naming the snapshot fields those later slices will read.
- Validation commands cover contributor eligibility, weighting/clamping, degraded exclusion reasons, and deterministic replay of the same fixture window.

## ASCII Flow

```text
canonical events + feed health + config snapshot
                    |
                    v
      contributor eligibility and quote gate
      - venue membership
      - timestamp trust
      - quote proxy allow/deny
      - degraded exclusion rules
                    |
                    v
          deterministic weighting pipeline
          - raw quality weights
          - penalty multipliers
          - normalize
          - clamp
          - renormalize in stable order
               /                    \
              v                      v
     WORLD composite snapshot   USA composite snapshot
      - contributors             - contributors
      - composite price          - composite price
      - coverage and health      - coverage and health
      - degraded reasons         - degraded reasons
               \                    /
                \                  /
                 v                v
          later bucket / regime / query readers
          consume snapshot outputs only
```

## Live-Path Boundary

- This feature stops at service-owned snapshot production in Go.
- Later features may read these snapshots, but they must not move weighting, degradation, or contributor-trust logic into `apps/web`, storage layers, or Python tooling.

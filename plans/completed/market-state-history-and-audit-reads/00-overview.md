# Market State History And Audit Reads

## Ordered Implementation Plan

1. Extend the current-state contract family with versioned history and audit envelopes that reuse the same symbol/global market-state sections and version metadata.
2. Implement service-owned historical retrieval and version-pinned lookup surfaces in Go for closed bucket windows and replay contexts.
3. Implement replay-corrected audit provenance assembly so operators can explain why a historical state changed and which replay/config context produced the authoritative answer.
4. Validate focused Go integration and replay coverage for historical retrieval, version-pinned bucket/context lookup, and replay-corrected audit provenance; record evidence in `plans/completed/market-state-history-and-audit-reads/testing-report.md`.

## Problem Statement

The repo now has deterministic current market-state contracts for symbol and global reads, but it still lacks a bounded way to retrieve older current-state answers and the audit evidence behind them. Operators and downstream services need to ask what the authoritative state was for a closed bucket or replay context without inventing a second read model or manually joining raw replay artifacts.

## Bounded Scope

- extend the completed current-state contract family with historical and audit response variants
- historical retrieval for closed symbol and global market-state snapshots keyed by symbol, bucket window, and version context
- version-pinned lookup by `configVersion`, `algorithmVersion`, replay run identity, and closed bucket timestamps or keys
- audit provenance sections that explain source artifacts, replay corrections, and superseded state where applicable
- focused Go service query surfaces and deterministic test fixtures for historical and audit reads

## Out Of Scope

- new market-state computation logic, threshold changes, or client-side joins
- unbounded analytics queries, ad hoc SQL-style filtering, or warehouse/reporting APIs
- mutation flows, operator overrides, retention policy changes, or backfill orchestration
- `apps/web` implementation work beyond noting it remains a read-only consumer of service-owned outputs
- Python runtime paths, notebooks, or research-only tooling

## Requirements

- Reuse the `market-state-current-query-contracts` family so current and historical reads share the same core symbol/global sections, reason codes, and version fields.
- Keep the live path in Go-owned service surfaces under `services/*`; history lookup may consult replay-owned artifacts, but consumers still read one service-owned contract family.
- Restrict retrieval to closed windows and replay-stable contexts only; never expose in-flight bucket state as authoritative history.
- Carry authoritative provenance for composite snapshots, bucket summaries, regime outputs, replay run identity, and correction lineage when a replay supersedes an earlier answer.
- Preserve deterministic results for the same pinned symbol, bucket key, replay context, and version tuple.
- Keep `apps/web` read-only; if it consumes these reads later, it formats the response and never recomputes state.

## Target Repo Areas

- `schemas/json/features`
- `schemas/json/replay`
- `services/feature-engine`
- `services/regime-engine`
- `services/replay-engine`
- `libs/go`
- `tests/fixtures`
- `tests/integration`
- `tests/replay`

## Module Breakdown

### 1. History Contracts And Lookup Keys

- Add versioned history and audit envelopes that embed the completed current-state response sections and add closed-window lookup metadata.
- Define the request and response seams for bucket-key, `asOf`, and version-pinned retrieval without creating a parallel market-state schema family.

### 2. Historical Retrieval Surfaces

- Define the Go query surfaces that resolve historical symbol/global state from closed feature/regime artifacts and replay-owned storage metadata.
- Keep lookup rules conservative: exact match on requested closed window, explicit unavailable responses when artifacts are missing, and no consumer-side joining.

### 3. Replay-Corrected Audit Provenance

- Define how audit responses show whether a historical answer came from the original run, a replay correction, or a superseding config/version context.
- Surface provenance as machine-readable lineage so investigations can explain state corrections without reopening raw event streams by hand.

## Acceptance Criteria

- Another agent can implement the feature without reopening the epic or inventing a second market-state read model.
- The plan names the exact repo areas for schema work, service lookup surfaces, replay provenance, fixtures, and validation.
- Ordered steps are bounded to one child feature and stop short of broader reporting, retention, or UI work.
- Testing commands are concrete and cover Go integration plus replay determinism for historical retrieval and audit provenance.

## ASCII Flow

```text
closed composite snapshots + closed bucket summaries + regime outputs + replay manifest/context
                                      |
                                      v
                 Go-owned history lookup and audit assembler
                 - request keyed by symbol/global scope + bucket key/asOf
                 - exact version pin: configVersion + algorithmVersion + replay run
                 - authoritative current-state sections reused from v1 family
                 - correction lineage and provenance
                                      |
                    +-----------------+------------------+
                    |                                    |
                    v                                    v
        historical market-state read            audit provenance read
        - symbol/global closed snapshot         - source artifact ids
        - version-pinned state sections         - replay correction status
        - unavailable markers on gaps           - superseded-by lineage
                    |                                    |
                    +-----------------+------------------+
                                      |
                                      v
                        read-only consumers and investigations
                        format service output only
```

## Live-Path Boundary

- This feature stays service-owned and Go-only in the live/runtime path.
- Replay artifacts may be read for authoritative historical answers, but Python remains out of scope and out of runtime.
- `apps/web` is only a future consumer of these reads; this plan does not add UI behavior.

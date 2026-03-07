# Canonical Contracts And Fixtures

## Ordered Implementation Plan

1. Define contract families, naming rules, and versioning boundaries
2. Define canonical field semantics for symbols, venues, timestamps, and degraded-state markers
3. Build deterministic fixture corpus and replay seed layout
4. Add schema validation, consumer validation, and parity scaffolding
5. Run targeted contract and fixture validation checks

## Problem Statement

Visibility and later alerting cannot be trusted if events, derived-state payloads, and replay fixtures are ambiguous or inconsistent across services and languages.

This feature creates the durable contract and fixture foundation for:

- live Go services in `services/*`
- the React/Vite operator UI in `apps/web`
- optional offline Python research and parity work in `apps/research` and `libs/python`

## Requirements

- Use `schemas/json/` as the only canonical home for shared JSON payload definitions.
- Reserve and define contract families for:
  - `schemas/json/events`
  - `schemas/json/features`
  - `schemas/json/alerts`
  - `schemas/json/outcomes`
  - `schemas/json/replay`
  - `schemas/json/simulation`
- Normalize tradable symbols to `BTC-USD` and `ETH-USD` while preserving source quote context such as `USD`, `USDT`, or `USDC`.
- Standardize venue identifiers and instrument context so WORLD, USA, spot, and perp data can be joined without guesswork.
- Define canonical time semantics that respect the program defaults:
  - `exchangeTs` is primary event time when plausible
  - `recvTs` is always stored
  - timestamp degradation must be representable in contracts and fixtures
- Define where downstream payloads carry config version, regime tags, degraded-feed markers, and replay provenance.
- Create deterministic fixtures for happy-path, degraded-path, and replay-sensitive scenarios.
- Keep Python out of the live runtime path; parity support is optional and offline only.

## Out Of Scope

- Implementing service logic or storage engines
- Final threshold values, score weights, or alert formulas
- UI rendering logic
- Live delivery or outcome computation logic

## Design Notes

### Contract Strategy

- Treat each payload family as versioned and additive by default.
- Prefer explicit, boring filenames that include version identity.
- Keep shared metadata fields consistent across families where that improves auditability, but do not force a speculative universal envelope if it obscures payload meaning.

### Canonical Identity Rules

- Symbol identity is canonicalized at ingest time and used unchanged downstream.
- Venue identity and market type remain explicit in every contract family that needs provenance.
- Composite payloads must distinguish source venue rows from WORLD and USA aggregate views.

### Timestamp And Replay Rules

- Event-time and processing-time semantics must be documented once here and reused everywhere else.
- Fixtures must include timestamp edge cases: normal ordering, late events, out-of-order messages, and timestamp-degraded cases.
- Replay payloads must preserve the information needed to reproduce the same ordering and feature bucket assignment later.

### Fixture Strategy

- Build a small, high-signal corpus first.
- Prefer fixtures that represent operator-relevant cases over large generic dumps.
- Include at least one fixture set per venue family plus a small number of cross-venue replay windows that exercise fragmentation and feed degradation.

### Live vs Research Boundary

- Go services and the UI consume the canonical contracts directly.
- Python may consume the same fixtures and contracts for offline analysis or parity checks, but it must not become a live dependency.

## Target Repo Areas

- `schemas/json`
- `tests/fixtures`
- `tests/replay`
- `tests/parity`
- `docs/specs`
- optional validation helpers in `libs/go`, `libs/ts`, and `libs/python`

## ASCII Flow

```text
source venue payloads
        |
        v
canonical event contracts (`schemas/json/events`)
        |
        +------> deterministic fixtures (`tests/fixtures`)
        |
        +------> replay payloads (`schemas/json/replay` + `tests/replay`)
        |
        +------> derived contracts (`features`, `alerts`, `outcomes`, `simulation`)
        |
        +------> consumer validation (Go / TS / optional Python parity)
        |
        v
later services and dashboards consume one shared contract vocabulary
```

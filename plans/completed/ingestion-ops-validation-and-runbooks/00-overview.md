# Ingestion Ops Validation And Runbooks

## Ordered Implementation Plan

1. Add a focused ingestion smoke harness for happy path, gap, stale, and retry scenarios.
2. Define the ops-facing metric inventory and alert-condition matrix for feed degradation.
3. Write degraded-feed investigation runbook material using the shared health vocabulary.
4. Produce the high-signal validation report for future agents and operators.

## Problem Statement

The umbrella ingestion plan requires the system to be operable, but the remaining work is now mostly integration validation and operator-facing documentation rather than parser/runtime building blocks.

## Requirements

- Keep tests deterministic and local.
- Cover gap, stale, reconnect, resync, and timestamp-trust scenarios.
- Use the same health vocabulary in tests, docs, and logs.
- Prefer targeted integration tests and runbooks over ad hoc prose.

## Out Of Scope

- Live exchange testing as a requirement for completion.
- Dashboard work.

## Target Repo Areas

- `tests/integration`
- `docs/runbooks`
- `services/venue-*`
- `services/normalizer`

## ASCII Flow

```text
deterministic harness inputs
        |
        v
venue adapters + normalizer
        |
        +--> smoke report
        +--> ops metric inventory
        +--> degraded-feed runbook
```

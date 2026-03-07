# Implementation: Alert Payloads And Consumer Contracts

## Module Requirements And Scope

- Define the alert record shape expected by later delivery, outcome, simulation, and UI consumers.
- Make payload expectations specific enough for downstream implementation while avoiding premature concrete schema files.
- Preserve deterministic replay, service-owned truth, and compatibility across future slices.

## Target Repo Areas

- `services/alert-engine`
- `services/outcome-engine`
- `apps/web`
- `schemas/json/alerts`
- `schemas/json/outcomes`
- `tests/fixtures`
- `tests/integration`
- `tests/replay`

## Planning Guidance

### 1. Alert Record Layers

- Plan for three related records:
  - `alert_decision_record`: internal emitted or suppressed decision with full reasoning
  - `alert_delivery_payload`: consumer-safe emitted subset for UI, Telegram, and webhook delivery
  - `alert_outcome_seed`: stable identifiers and timing fields needed by later outcome evaluation
- Suppressed decisions should stay internal unless a later ops surface needs them.

### 2. Required Core Fields

- Stable identifiers:
  - `alertId`
  - `clusterId`
  - `symbol`
  - `setupKey`
  - `configVersion`
  - `algorithmVersion`
  - contract/schema version reference
- Timing:
  - `eventTimeFirstSeen`
  - `eventTimeValidated`
  - `decisionTime`
  - `bucket30s`
  - `bucket2m`
  - `bucket5m`
- Decision:
  - final severity
  - emission status (`EMITTED`, `SUPPRESSED`, `DOWNGRADED`)
  - market-state inputs and applied cap
  - reserved risk-state fields for later compatibility
- Explanation:
  - bounded reason codes
  - degraded reason codes
  - fragmentation flags
  - summary evidence references

### 3. Evidence Expectations

- Include enough evidence metadata for consumers to explain the alert without recomputing logic.
- Safe default evidence sections:
  - setup evidence summary from 30s nomination
  - validator evidence summary from 2m confirmation
  - 5m market-state snapshot reference
  - feed-quality and fragmentation summary
- Use numeric or enum outputs already produced by services where possible; avoid embedding raw venue payloads in the delivery payload.

### 4. Delivery Consumer Expectations

- `apps/web` needs: setup family, severity, why-fired summary, why-allowed summary, degraded flags, timestamps, and drill-down identifiers.
- Telegram and webhook need: short explanation, severity, symbol, setup, relevant horizon snapshot, and a durable link or identifier for later review.
- Delivery surfaces should not infer severity from raw fields; the alert engine must send the final capped severity.
- Informational no-operate alerts must remain clearly labeled as non-actionable.

### 5. Outcome Consumer Expectations

- Outcome evaluation later needs stable seeds, not a rewritten alert object.
- Preserve: symbol, setup, direction intent if applicable, alert timestamps, config/version references, and market/risk gate snapshots.
- Keep the alert record explicit about whether it was actionable or informational so outcome reports do not mix unlike alerts.
- Ensure cluster linkage survives into outcomes so later analytics can compare one market episode against one outcome context.

### 6. Compatibility And Versioning

- Additive evolution should be the default for delivery payloads.
- Contract version references must travel with every emitted payload and replay artifact.
- Schema work later should keep internal decision richness separate from external delivery contracts so transport surfaces do not depend on debugging-only fields.

### 7. Storage And Query Notes

- Persist emitted alerts and decision records long enough to support the operating defaults for alerts and later review flows.
- Plan read models that let consumers fetch recent alerts and a single alert detail without recomputing the setup logic.
- Keep raw canonical-event provenance as references or IDs, not duplicated blobs, so payload size stays bounded.

## Unit And Integration Test Expectations

- contract-focused tests for emitted payload completeness and version references
- integration tests for suppressed decisions remaining internal while emitted alerts produce delivery-safe payloads
- integration tests for outcome seed generation preserving cluster and gate fields
- replay tests verifying identical input/config emit byte-stable payload fields where ordering matters

## Summary

This module defines the boundary between alert generation and its consumers. It keeps the alert engine as source of truth, gives delivery and outcome slices enough detail to act without recomputation, and preserves compatibility through explicit IDs, timestamps, reasons, and version references.

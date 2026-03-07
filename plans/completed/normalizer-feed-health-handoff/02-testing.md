# Testing Plan: Normalizer Feed Health Handoff

Expected output artifact: `plans/completed/normalizer-feed-health-handoff/testing-report.md`

## Smoke Matrix

| Case | Flow | Expected | Evidence |
|---|---|---|---|
| Trade handoff | Venue trade input -> normalizer | canonical trade output unchanged | `go test` output |
| Book handoff | Venue book input -> normalizer | canonical book output unchanged | `go test` output |
| Feed-health handoff | Degraded feed status -> normalizer | canonical feed-health output preserved | `go test` output |

## Required Commands

- `"/usr/local/go/bin/go" test ./services/normalizer/... ./libs/go/...`

## Verification Checklist

- Normalizer is now the explicit canonical boundary.
- Degradation metadata remains visible through the handoff.

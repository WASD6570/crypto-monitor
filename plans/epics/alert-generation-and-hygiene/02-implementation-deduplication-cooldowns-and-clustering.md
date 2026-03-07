# Implementation: Deduplication, Cooldowns, And Clustering

## Module Requirements And Scope

- Prevent one market event from producing repeated user interruptions across venues, horizons, or setup reevaluations.
- Define deterministic dedupe keys, cooldown policy, clustering windows, and fragmentation-aware suppression behavior.
- Cover degraded feeds and noisy transitions without inventing probabilistic or AI-driven ranking.

## Target Repo Areas

- `services/alert-engine`
- `libs/go`
- `configs/*`
- `tests/fixtures`
- `tests/integration`
- `tests/replay`

## Planning Guidance

### 1. Deduplication Model

- Deduplicate on the underlying alert intent, not only raw event identity.
- Safe default dedupe dimensions: symbol, setup family, side or directional intent if later present, decision tier, cluster window, and config version.
- Keep source-event references alongside the dedupe key so replay can show which canonical events contributed to the first accepted alert.
- Distinguish between exact duplicates and near-duplicates caused by recalculation across adjacent buckets.

### 2. Cooldown Policy

- Use cooldowns to limit repeated emissions after an alert has already been accepted for the same underlying move.
- Cooldowns should be setup-specific and severity-aware in config, with conservative defaults.
- A stronger later alert may bypass a weaker earlier cooldown only if the config explicitly allows severity escalation and the payload explains the escalation reason.
- Recovery from cooldown should require materially new evidence, not only the passage of wall-clock time.

### 3. Clustering Policy

- Model an alert cluster as a bounded episode representing one market move or one failed move sequence.
- Cluster windows should absorb repeated 30s nominations that all roll up to the same 2m-validated context.
- The cluster should own summary fields such as first candidate time, first emitted time, last contributing event time, and emission count.
- Delivery consumers later need cluster identifiers so repeated surfaces can thread related alerts together rather than show independent noise.

### 4. Fragmentation And Chatter Control

- Elevated WORLD vs USA divergence should increase suppression or cluster absorption, especially for `A` and `B` style setups that can flicker under fragmented leadership.
- Repeated trigger flips inside fragmentation windows should prefer one informational cluster update or full suppression over alternating bullish and bearish interruptions.
- If both directions qualify within a short bounded window, treat that as likely noise unless higher-level state explicitly supports one side.

### 5. Degraded Feed Policy

- Critical venue loss, timestamp degradation, or stale key inputs should not create bursty retries that turn one degraded period into many alerts.
- Safe default: emit at most one bounded informational degradation-related alert per symbol and cluster window when user awareness matters.
- Suppression decisions during degradation should be persisted with stable reason codes so replay explains silence instead of hiding it.

### 6. Fragmentation vs Genuine New Opportunity

- The hygiene layer must not erase clearly new opportunities after conditions materially reset.
- Require explicit reset conditions in config, such as validator failure followed by a fresh setup path, cluster expiry, state improvement, or materially different evidence bundle.
- Keep reset logic deterministic and auditable; do not rely on fuzzy similarity scoring.

### 7. Versioning And Auditability

- Dedupe, cooldown, and cluster decisions need the same version context as emitted alerts.
- Replays must be able to reconstruct why the second and third candidate were suppressed under one config but emitted under another.
- Record both `decisionTs` and underlying evidence times to prevent confusion when replay processes events later than live.

## Unit And Integration Test Expectations

- unit tests for dedupe-key construction and cluster rollover behavior
- unit tests for severity-escalation cooldown exceptions
- integration tests for repeated 30s flickers collapsing into one emitted alert
- integration tests for fragmented opposite-direction candidates suppressing each other
- replay tests proving degraded periods emit one bounded informational event rather than a stream

## Summary

This module bounds alert volume. It defines how identical or near-identical candidates collapse into a single episode, how cooldowns prevent repetition, how fragmentation reduces chatter, and how degraded feeds stay explainable without becoming noise. The next slice can then define payload contracts on top of stable emitted and suppressed decision records.

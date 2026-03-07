# Testing

## Testing Goals

- prove setup A/B/C nomination, validation, and market-state gating are deterministic for the same inputs and config snapshot
- prove dedupe, cooldown, and clustering bound alert volume during noisy and fragmented conditions
- prove degraded feeds and timestamp issues downgrade or suppress alerts instead of inflating urgency
- prove emitted payloads are complete for delivery and outcome consumers without client-side recomputation
- prove Go remains the live source of truth and replay reproduces the same decisions

## Output Artifact

- Write the implementation-phase test report to `plans/epics/alert-generation-and-hygiene/testing-report.md`.

## Test Matrix

### 1. Setup Registry And Gating Ladder

- Purpose: verify `A`, `B`, and `C` candidates move deterministically through nomination, validation, permissioning, and emission.
- Fixtures:
  - one clean `A` path with 30s nomination, 2m validation, and 5m `TRADEABLE`
  - one `B` path where 30s interest fails 2m confirmation
  - one `C` path that validates but is capped by 5m `WATCH`
- Validation commands:
  - `go test ./services/alert-engine/... -run 'Test(SetupRegistry|AlertDecisionLadder|MarketStateSeverityCaps)'`
  - `go test ./tests/integration/... -run 'TestAlertSetupGatingScenarios'`
- Verify:
  - setup family and horizon evidence are recorded consistently
  - 30s-only signals do not emit actionable alerts
  - 5m state ceilings cap severity exactly as planned

### 2. Dedupe, Cooldown, And Cluster Behavior

- Purpose: verify repeated candidates collapse into bounded episodes.
- Fixtures:
  - repeated identical 30s triggers for one `A` move
  - stronger follow-up alert inside cooldown window
  - separate genuine second move after cluster expiry
- Validation commands:
  - `go test ./services/alert-engine/... -run 'Test(AlertDeduplication|AlertCooldownPolicy|AlertClusterRollup)'`
  - `go test ./tests/integration/... -run 'TestAlertBurstSuppression'`
- Verify:
  - exact duplicates emit once
  - cooldown prevents same-episode spam
  - config-controlled severity escalation behaves deterministically
  - a genuinely new cluster can emit after reset conditions are satisfied

### 3. Fragmentation And Opposite-Direction Noise

- Purpose: verify fragmentation reduces chatter and prevents conflicting interruptions.
- Fixtures:
  - elevated WORLD vs USA divergence with fast flip-flopping 30s candidates
  - one-side confirmation under mild fragmentation
  - severe fragmentation with unstable venue leadership
- Validation commands:
  - `go test ./services/alert-engine/... -run 'Test(FragmentationSuppression|OppositeDirectionSuppression|ClusteredNoiseHandling)'`
  - `go test ./tests/replay/... -run 'TestFragmentedAlertReplayDeterminism'`
- Verify:
  - alternating opposite-direction candidates do not produce alternating alerts
  - mild fragmentation downgrades or delays as configured
  - severe fragmentation caps to `INFO` or suppresses fully

### 4. Degraded Feed And Timestamp Handling

- Purpose: verify stale or degraded inputs reduce trust instead of creating urgency.
- Fixtures:
  - key venue missing during otherwise valid setup
  - timestamp fallback from `exchangeTs` to `recvTs`
  - critical degraded period with repeated reevaluation attempts
- Validation commands:
  - `go test ./services/alert-engine/... -run 'Test(DegradedFeedSuppression|TimestampDegradedAlertDecision|SingleInformationalDegradedAlert)'`
  - `go test ./tests/integration/... -run 'TestAlertBehaviorUnderFeedDegradation'`
- Verify:
  - degraded conditions are visible in reason codes
  - one degraded interval does not create repeated alerts
  - missing or implausible timestamps do not silently bypass gating rules

### 5. Payload And Consumer Contract Coverage

- Purpose: verify emitted payloads are complete for downstream consumers.
- Fixtures:
  - one emitted actionable alert
  - one downgraded watch alert
  - one suppressed decision with internal-only details
- Validation commands:
  - `go test ./services/alert-engine/... -run 'Test(AlertPayloadShape|AlertReasonCodes|OutcomeSeedFields)'`
  - `go test ./tests/integration/... -run 'TestAlertDeliveryContractCompatibility'`
- Verify:
  - emitted payload contains IDs, timestamps, version fields, and reason summaries
  - suppressed-only fields do not leak into delivery payloads
  - outcome seed fields remain stable and complete

### 6. Replay Determinism And Config Versioning

- Purpose: prove the same fixture window yields identical alert decisions under the same snapshot and auditable changes under a new snapshot.
- Fixtures:
  - pinned BTC day with healthy conditions
  - pinned ETH day with fragmented and degraded intervals
  - same replay window under two config versions
- Validation commands:
  - `go test ./tests/replay/... -run 'TestAlertGenerationReplayDeterminism'`
  - `go test ./tests/replay/... -run 'TestAlertGenerationConfigVersionPinnedReplay'`
  - `go test ./tests/integration/... -run 'TestAlertDecisionAuditFields'`
- Verify:
  - identical input plus config emits identical decisions and payload fields
  - config changes alter outputs only through versioned, auditable policy differences
  - replay output preserves suppressed-decision reasons and cluster context

## Required Negative Cases

- 30s candidate exists but 2m validator window never confirms
- 2m validator confirms but 5m symbol state is `NO-OPERATE`
- global state `WATCH` tries to allow `ACTIONABLE`
- repeated identical candidate across adjacent buckets attempts to bypass dedupe
- opposite-direction candidates appear in the same cluster window
- degraded key venue plus stale timestamps attempt to escalate severity
- config version changes dedupe windows and replay must show why behavior changed
- missing reserved risk-state fields must default safely without breaking payload consumers

## Determinism And Parity Notes

- Go is the only live source of truth for setup, gating, hygiene, and payload tests.
- Optional parity tests under `tests/parity` may compare offline analysis helpers against Go-produced fixtures, but parity is advisory and never a live dependency.
- Validation paths must not rely on live venue connections, wall-clock randomness, or unordered map iteration.

## Exit Criteria For Implementation

- targeted unit, integration, and replay commands pass
- negative cases cover suppression, downgrade, fragmentation, and degraded-feed behavior
- repeated replay runs match exactly for the same config snapshot
- emitted payloads are sufficient for delivery and outcome slices without client recomputation

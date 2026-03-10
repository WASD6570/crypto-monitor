# Testing: Binance Live Replay Binance Family Determinism

## Validation Matrix

| Case | Flow | Expected Result |
|---|---|---|
| Binance manifest acceptance | Binance raw partition manifest records -> replay manifest build -> resolved partition validation | manifest includes the expected Binance logical partitions in stable order with no guessed storage paths |
| Spot shared-vs-dedicated partition replay | Spot trade/top-of-book shared partition plus depth/feed-health dedicated partitions -> replay execute | ordered output stays deterministic across mixed Binance partition shapes |
| USD-M mixed-surface replay | websocket funding/mark-index/liquidation plus REST open-interest/feed-health -> replay execute | replay preserves distinct source identity, partition selection, and deterministic digest behavior across surfaces |
| Duplicate-input determinism | identical Binance raw entries replayed repeatedly | ordered IDs, input counters, and compare/rebuild digests remain identical across repeated runs |
| Degraded evidence retention | Binance raw entries with `recvTs` bucket fallback, `Late`, and degraded feed-health linkage -> replay inspect/rebuild | degraded timestamp counters and degraded-feed evidence survive unchanged |
| Replay compare drift proof | baseline Binance replay artifact vs modified input set | compare mode reports deterministic drift classification and stable first mismatch facts |

## Commands

### Replay Engine Boundary

- `"/usr/local/go/bin/go" test ./services/replay-engine -run 'TestReplay(Manifest|Request|Run|Deterministic|Stable|Retention|Partition).*' -v`

Expected coverage:
- Binance manifest resolution and validation
- deterministic ordering and counter behavior
- compare/rebuild audit stability

### Binance Replay Proof

- `"/usr/local/go/bin/go" test ./tests/replay -run 'TestReplay.*(Binance|Determinism|Retention)' -v`

Expected coverage:
- representative Binance raw-entry replay across Spot and USD-M families
- duplicate-input, degraded timestamp, and degraded feed-health retention

### End-To-End Audit Proof

- `"/usr/local/go/bin/go" test ./tests/integration -run 'TestIngestionBinance.*Replay' -v`

Expected coverage:
- raw append output consumed by replay paths without identity or digest drift
- mixed partition and surface proof for representative Binance families

## Inputs / Env

- Default deterministic suite requires no secrets.
- This feature should not require live Binance network access.
- If new replay fixtures or manifests are added, keep them deterministic and local.

## Verification Checklist

- replay manifest build resolves the expected Binance logical partitions and no others
- repeated replay runs over the same Binance raw inputs produce identical ordered IDs and output digests
- duplicate, late, and degraded timestamp counters remain stable for representative Binance families
- degraded feed-health linkage and source identity survive replay unchanged
- any new replay fixtures or helpers are captured in the feature testing report

## Testing Report Output

- While active: `plans/binance-live-replay-binance-family-determinism/testing-report.md`
- After implementation and validation complete: move the full directory to `plans/completed/binance-live-replay-binance-family-determinism/`

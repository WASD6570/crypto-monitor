# Testing Report: Binance Live Replay Binance Family Determinism

## Result

- Status: passed
- Date: 2026-03-10

## Commands

- `"/usr/local/go/bin/go" test ./services/replay-engine -run 'TestReplay(Manifest|Request|Run|Deterministic|Stable|Retention|Partition).*' -v`
- `"/usr/local/go/bin/go" test ./tests/replay -run 'TestReplay.*(Binance|Determinism|Retention)' -v`
- `"/usr/local/go/bin/go" test ./tests/integration -run 'TestIngestionBinance.*Replay' -v`

## Coverage Notes

- Replay manifest build now accepts the settled Binance raw partition posture, including shared Spot partitions for `trades` and `top-of-book` plus dedicated partitions for `order-book`, `feed-health`, `funding-rate`, and `open-interest`.
- Replay execution digests now include persisted replay evidence such as degraded timestamp facts, degraded feed references, and duplicate audit data, so compare mode catches evidence drift that did not change canonical IDs alone.
- Replay tests now prove deterministic repeated-run behavior for representative Binance Spot and USD-M families, including duplicate inputs and degraded evidence retention.
- Integration proof shows actual Binance raw append outputs replay deterministically across Spot shared partitions, Spot depth degraded feed-health, and USD-M mixed websocket and REST surfaces.

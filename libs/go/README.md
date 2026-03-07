# Go Shared Contracts

- Put canonical contract decoding and version guards in `libs/go` so live services share one interpretation path.
- New Go consumers should be fixture-backed and reject unsupported `schemaVersion` values before business logic runs.
- Keep Go as the live-path consumer of canonical market state; do not route live validation through Python.
- `libs/go/ingestion` now holds the shared venue-adapter health/config primitives used before payloads reach `services/normalizer`.
- Venue adapters own connection lifecycle, snapshot bootstrap, gap detection, and health evaluation; `services/normalizer` owns canonical symbol/timestamp/degraded-marker handoff after adapter trust has been assessed.
- Use `LoadEnvironmentConfig` in `libs/go/ingestion` to parse the checked-in `configs/*/ingestion.v1.json` files into adapter-ready runtime config instead of re-parsing thresholds inside each venue service.
- Use `OrderBookSequencer` in `libs/go/ingestion` for snapshot-plus-delta integrity so adapters can deterministically accept in-order deltas, ignore stale repeats, and force resync on gaps instead of guessing.
- Use `ResolveCanonicalTimestamp` with `StrictTimestampPolicy()` to keep both `exchangeTs` and `recvTs` while treating `exchangeTs` as canonical only when its skew from `recvTs` stays within the strict 2s window.
- `NormalizeTradeMessage` is the first fixture-backed normalization helper; it shows the adapter-to-canonical handoff pattern for preserving source metadata while applying strict timestamp fallback rules.
- `NormalizeOrderBookMessage` extends that pattern for snapshot/delta book updates by using `OrderBookSequencer` to emit canonical `order-book-top` events on good sequences and canonical degraded `feed-health` events on deterministic resync paths.

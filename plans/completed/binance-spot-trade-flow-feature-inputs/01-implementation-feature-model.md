# Implementation: Feature Model

## Scope

- Target areas: `libs/go/features`, `services/feature-engine`, and focused unit tests beside those packages.
- Define the internal Spot trade-flow observation and bucket contract that later runtime, replay, API, and alerting work can consume.
- Keep the model generic enough to represent Spot trade flow, but do not generalize beyond Binance or the current `BTC-USD`/`ETH-USD` need if that would add names or dependencies.

## Requirements

- Consume normalized trade facts, not native websocket payloads: symbol, venue, market type, source symbol, source record ID, trade side, price, size, exchange time, receive time, timestamp status, and feed-health status.
- Accept only Binance Spot observations for supported symbols in this feature; invalid symbols, empty IDs, non-positive price/size, and unsupported sides must fail fast in tests.
- Assign observations to bucket families using the same deterministic time posture as existing market-quality buckets: canonical event time first, receive time only when timestamp policy degrades.
- Deduplicate by stable source identity before aggregation.
- Sort output deterministically by symbol, bucket family, bucket start/end, then source identity when needed.

## Planned Types

Implementation can adjust exact names, but the plan expects one small surface close to this shape:

```go
type SpotTradeFlowObservation struct {
    Symbol string
    Venue ingestion.Venue
    MarketType string
    SourceSymbol string
    SourceRecordID string
    Side string
    Price float64
    Size float64
    ExchangeTs time.Time
    RecvTs time.Time
    TimestampStatus ingestion.CanonicalTimestampStatus
    FeedHealthState ingestion.FeedHealthState
    FeedHealthReasons []ingestion.DegradationReason
}

type SpotTradeFlowBucket struct {
    SchemaVersion string
    Symbol string
    Family BucketFamily
    BucketStart string
    BucketEnd string
    TradeCount int
    DuplicateCount int
    BuyTradeCount int
    SellTradeCount int
    BuyNotional float64
    SellNotional float64
    NetAggressorNotional float64
    TotalNotional float64
    VWAP float64
    FirstPrice float64
    LastPrice float64
    PriceChangeBps float64
    TimestampFallbackCount int
    FeedHealthState ingestion.FeedHealthState
    FeedHealthReasons []ingestion.DegradationReason
    ConfigVersion string
    AlgorithmVersion string
}
```

## Aggregation Rules

- Notional is `price * size`, rounded using the existing feature metric rounding style where practical.
- `buy` trades increase buy count/notional and positive net aggressor notional; `sell` trades increase sell count/notional and negative net aggressor notional.
- VWAP is `totalNotional / totalSize` and should be zero only when no valid trades exist.
- First/last price are ordered by canonical event time, then receive time, then source record ID.
- Price-change bps is `(last - first) / first * 10000`, with zero when a bucket has fewer than two accepted unique trades.
- Feed health should use worst state across accepted observations and carry stable sorted reasons.
- Timestamp fallback count increments when canonical timestamp status is degraded or the observation bucket source falls back to receive time.

## Unit Test Expectations

- `libs/go/features`: happy buy/sell aggregation, duplicate suppression, timestamp fallback accounting, invalid input rejection, stable sorting, and repeated-run output equality.
- `services/feature-engine`: service wrapper initializes the processor, observes trade-flow inputs, advances buckets, and returns deterministic snapshots without exposing public API changes.
- Include both `BTC-USD` and `ETH-USD` examples so symbol filtering and ordering are explicit.

## Summary For Next Agent

Implement the narrow internal feature model first. Do not wire runtime or replay before the pure feature tests prove deterministic aggregation, dedupe, timestamp posture, and stable ordering.

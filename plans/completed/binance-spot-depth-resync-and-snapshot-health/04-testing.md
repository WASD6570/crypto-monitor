# Testing: Binance Spot Depth Resync And Snapshot Health

## Validation Matrix

| Case | Flow | Expected Result |
|---|---|---|
| Gap-triggered resync | synchronized depth detects a `U/u` gap and enters recovery | depth stops claiming synchronized state, sequence-gap stays visible, and recovery state becomes explicit |
| Cooldown-blocked retry | recovery starts before `snapshotCooldownMs` has elapsed | no snapshot request is made, remaining cooldown is surfaced, and health stays degraded or stale rather than silently healthy |
| Rate-limit-blocked retry | repeated snapshot recoveries exhaust `snapshotRecoveryPerMinuteLimit` | retry is blocked with explicit allowance and `retryAfter` visibility |
| Successful replacement snapshot | a later eligible snapshot plus aligned deltas arrives | synchronized state is restored, sequence-gap clears, and depth outputs resume through the shared sequencer |
| Snapshot refresh due | refresh interval elapses without an explicit gap | recovery owner requests a replacement snapshot on schedule and preserves explicit recovery visibility during the refresh window |
| Snapshot stale precedence | snapshot age exceeds `snapshotStaleAfterMs` while messages may still be fresh | final health state is `STALE` with `snapshot-stale` preserved alongside any other degraded reasons |
| Resync-loop visibility | repeated failed recoveries exceed the configured threshold | `resync-loop` remains machine-visible in the final feed-health output |
| Live Binance shape check | public REST snapshot and optional live recovery probe | implemented parser and recovery assumptions still match the live venue surface |

## Commands

### Venue Runtime And Recovery

- `"/usr/local/go/bin/go" test ./services/venue-binance -run 'Test(Runtime|SpotDepth).*' -v`

Expected coverage:
- recovery owner state transitions
- cooldown and rate-limit blocking
- refresh cadence and snapshot staleness
- resync-loop and connection-state composition

### Shared Sequencing And Canonical Handoff

- `"/usr/local/go/bin/go" test ./libs/go/ingestion -run 'TestOrderBook|TestNormalizeOrderBookMessage|TestFeedHealth' -v`
- `"/usr/local/go/bin/go" test ./services/normalizer -run 'TestServiceNormalize(OrderBook|FeedHealth)' -v`

Expected coverage:
- replacement snapshot and resumed deltas remain compatible with shared sequencing
- canonical feed-health output preserves recovery reasons and freshness posture

### Fixture-Backed Integration

- `"/usr/local/go/bin/go" test ./tests/integration -run 'TestIngestionBinanceSpotDepth(Recovery|Bootstrap)' -v`

Expected coverage:
- bootstrap prerequisites still hold
- successful resync and blocked recovery states are proven through supervisor-fed depth input

### Parity / Fixture Validation

- `"/usr/local/go/bin/go" test ./tests/parity -run 'TestGoConsumerValidatesFixtures' -v`

Expected coverage:
- new Binance recovery fixtures decode and satisfy the shared manifest rules

### Direct Live Validation

- `curl -fsS "https://api.binance.com/api/v3/depth?symbol=BTCUSDT&limit=5"`
- `BINANCE_LIVE_VALIDATION=1 "/usr/local/go/bin/go" test ./tests/integration -run 'TestIngestionBinanceSpotDepthLive' -v`

Expected coverage:
- current live REST snapshot shape still matches parser assumptions
- the attached live recovery probe exercises the real API boundary when network access is available

## Inputs / Env

- Default deterministic suite requires no secrets.
- Direct live validation should use public Binance endpoints only.
- Gate any live-network Go test behind `BINANCE_LIVE_VALIDATION=1` so default CI or local runs remain deterministic.

## Verification Checklist

- successful replacement snapshot clears prior sequence-gap and blocked-recovery posture
- blocked retry states preserve remaining cooldown or retry-after facts
- snapshot-stale and message-stale precedence remain correct in final health output
- resync-loop visibility survives into canonical feed-health output
- fixture manifest remains valid after any new recovery fixtures are added

## Testing Report Output

- While active: `plans/binance-spot-depth-resync-and-snapshot-health/testing-report.md`
- After implementation and validation complete: move the full directory to `plans/completed/binance-spot-depth-resync-and-snapshot-health/`

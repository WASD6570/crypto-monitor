# Binance USD-M Context Sensors

## Scope

- Covers the mixed Binance USD-M context surface for `markPrice@1s`, `forceOrder`, and `/fapi/v1/openInterest`.
- Treat websocket-derived and REST-derived signals as related but distinct acquisition modes.

## Fixture Inventory

- `tests/fixtures/events/binance/BTC-USD/happy-funding-usdt.fixture.v1.json`
- `tests/fixtures/events/binance/BTC-USD/happy-liquidation-usdt.fixture.v1.json`
- `tests/fixtures/events/binance/BTC-USD/happy-open-interest-usdt.fixture.v1.json`
- `tests/fixtures/events/binance/ETH-USD/happy-mark-index-usdt.fixture.v1.json`
- `tests/fixtures/events/binance/ETH-USD/happy-open-interest-usdt.fixture.v1.json`
- `tests/fixtures/events/binance/BTC-USD/edge-timestamp-degraded-funding-usdt.fixture.v1.json`
- `tests/fixtures/events/binance/ETH-USD/edge-timestamp-degraded-open-interest-usdt.fixture.v1.json`

## Verification Sequence

1. Run `"/usr/local/go/bin/go" test ./tests/integration -run 'TestIngestionBinanceUSDM' -v`.
2. Confirm websocket-derived canonical events still preserve `sourceSymbol`, `quoteCurrency`, `venue`, `marketType`, `exchangeTs`, and `recvTs`.
3. Confirm REST-derived `open-interest-snapshot` events preserve the same provenance and keep `recvTs` when Binance `time` is missing.
4. Confirm feed-health can emit `HEALTHY`, `DEGRADED`, or `STALE` independently for websocket and REST acquisition modes.

## Mixed-Surface Reason Checks

- websocket reconnect pressure stays visible as `reconnect-loop`
- websocket inactivity stays visible as `message-stale`
- REST poll throttling stays visible as `rate-limit`
- both surfaces still preserve `connection-not-ready`, `snapshot-stale`, `sequence-gap`, `resync-loop`, and `clock-degraded` when applicable

## Operator Notes

- Use the websocket feed-health source IDs with the `runtime:binance-usdm-ws:` prefix to inspect `markPrice@1s` freshness-bearing behavior.
- Use the REST feed-health source IDs with the `runtime:binance-usdm-open-interest:` prefix to inspect polling freshness and rate-limit posture.
- Do not collapse websocket and REST degradation into one generic sensor label; preserve the emitted reason strings exactly.

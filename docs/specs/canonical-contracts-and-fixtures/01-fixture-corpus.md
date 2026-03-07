# Fixture Corpus And Replay Seeds

## Fixture Layout

- Fixture files live under `tests/fixtures/<family>/<venue>/<symbol>/<scenario>.fixture.v1.json`.
- Replay seed files live under `tests/replay/seeds/<seed-name>.seed.v1.json`.
- `tests/fixtures/manifest.v1.json` is the indexed source of truth for fixture discovery.
- `tests/replay/manifest.v1.json` is the indexed source of truth for deterministic replay seed discovery.

## Fixture File Shape

Each fixture file carries one scenario with:

- source metadata: `id`, `family`, `category`, `venue`, `symbol`, `quoteCurrency`, `scenarioClass`
- contract intent: `targetSchema`, `checks`
- deterministic inputs: `rawMessages`
- deterministic expected outputs: `expectedCanonical`

Fixtures pair small raw source examples with the expected canonical result so Go, TypeScript, and optional Python tooling can validate the same scenario meaning.

## Scenario Catalog

### Happy-path categories

- `happy-trades`
- `happy-top-of-book`
- `happy-order-book-snapshot-delta`
- `happy-funding`
- `happy-open-interest`
- `happy-mark-index`
- `happy-liquidation`

### Edge-case categories

- `edge-sequence-gap`
- `edge-forced-resync`
- `edge-stale-feed`
- `edge-timestamp-degraded`
- `edge-late-out-of-order`
- `edge-quote-variant`

The starter corpus intentionally covers `USD`, `USDT`, and `USDC` quote contexts and includes sequence, timestamp, and feed-health degradation paths.

## Replay Seed Catalog

- `normal-microstructure`: normal cross-venue ordering for a short BTC visibility window
- `fragmented-world-usa`: a small BTC window that mixes WORLD and USA venues so downstream consumers can prove fragmented-state handling deterministically
- `degraded-feed`: an ETH window that combines stale feed and timestamp degradation signals

Each replay seed must record:

- `targetSchema` pointing to the concrete replay seed schema file
- `fixtureRefs` pointing to fixture ids from `tests/fixtures/manifest.v1.json`
- a fixed processing order in `expectedDeterminism.orderedSourceRecordIds`
- a stable `eventCount`
- scenario tags explaining why the seed exists

## Consumer Notes

- Fixtures stay intentionally small so `make fixtures-validate` and later replay smoke runs remain fast locally.
- Optional offline Python parity work must reuse these files rather than maintaining a duplicate corpus.
- Later concrete JSON Schemas should reference these categories and seed ids instead of inventing new scenario names.

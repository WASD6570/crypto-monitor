import { readFileSync } from 'node:fs';

const requiredFieldsBySchema = new Map([
  ['market-trade.v1.schema.json', ['schemaVersion', 'eventType', 'symbol', 'sourceSymbol', 'quoteCurrency', 'venue', 'marketType', 'exchangeTs', 'recvTs', 'timestampStatus', 'sourceRecordId']],
  ['order-book-top.v1.schema.json', ['schemaVersion', 'eventType', 'symbol', 'sourceSymbol', 'quoteCurrency', 'venue', 'marketType', 'exchangeTs', 'recvTs', 'timestampStatus', 'sourceRecordId']],
  ['feed-health.v1.schema.json', ['schemaVersion', 'eventType', 'symbol', 'sourceSymbol', 'quoteCurrency', 'venue', 'marketType', 'exchangeTs', 'recvTs', 'timestampStatus', 'feedHealthState', 'sourceRecordId']],
  ['funding-rate.v1.schema.json', ['schemaVersion', 'eventType', 'symbol', 'sourceSymbol', 'quoteCurrency', 'venue', 'marketType', 'exchangeTs', 'recvTs', 'timestampStatus', 'sourceRecordId']],
  ['open-interest-snapshot.v1.schema.json', ['schemaVersion', 'eventType', 'symbol', 'sourceSymbol', 'quoteCurrency', 'venue', 'marketType', 'exchangeTs', 'recvTs', 'timestampStatus', 'sourceRecordId']],
  ['mark-index.v1.schema.json', ['schemaVersion', 'eventType', 'symbol', 'sourceSymbol', 'quoteCurrency', 'venue', 'marketType', 'exchangeTs', 'recvTs', 'timestampStatus', 'sourceRecordId']],
  ['liquidation-print.v1.schema.json', ['schemaVersion', 'eventType', 'symbol', 'sourceSymbol', 'quoteCurrency', 'venue', 'marketType', 'exchangeTs', 'recvTs', 'timestampStatus', 'sourceRecordId']],
]);

const expectedEventTypeBySchema = new Map([
  ['market-trade.v1.schema.json', 'market-trade'],
  ['order-book-top.v1.schema.json', 'order-book-top'],
  ['feed-health.v1.schema.json', 'feed-health'],
  ['funding-rate.v1.schema.json', 'funding-rate'],
  ['open-interest-snapshot.v1.schema.json', 'open-interest-snapshot'],
  ['mark-index.v1.schema.json', 'mark-index'],
  ['liquidation-print.v1.schema.json', 'liquidation-print'],
]);

export function loadJson(filePath) {
  return JSON.parse(readFileSync(filePath, 'utf8'));
}

export function validateFixtureForTsConsumer(fixture) {
  if (fixture.fixtureVersion !== 'v1') {
    throw new Error('fixtureVersion must be v1');
  }
  const requiredFields = requiredFieldsBySchema.get(fixture.targetSchema);
  if (!requiredFields) {
    throw new Error(`unsupported target schema ${fixture.targetSchema}`);
  }
  if (!Array.isArray(fixture.checks) || fixture.checks.length === 0) {
    throw new Error('fixture checks must not be empty');
  }
  if (!Array.isArray(fixture.expectedCanonical) || fixture.expectedCanonical.length === 0) {
    throw new Error('expectedCanonical must not be empty');
  }

  const expectedEventType = expectedEventTypeBySchema.get(fixture.targetSchema);
  fixture.expectedCanonical.forEach((payload, index) => {
    requiredFields.forEach((field) => {
      if (!(field in payload)) {
        throw new Error(`canonical payload ${index} missing field ${field}`);
      }
    });
    if (payload.schemaVersion !== 'v1') {
      throw new Error(`canonical payload ${index} has unsupported schemaVersion ${payload.schemaVersion}`);
    }
    if (payload.eventType !== expectedEventType) {
      throw new Error(`canonical payload ${index} has eventType ${payload.eventType}, expected ${expectedEventType}`);
    }
  });
}

export function validateReplaySeedForTsConsumer(seed) {
  if (seed.schemaVersion !== 'v1') {
    throw new Error('schemaVersion must be v1');
  }
  if (seed.targetSchema !== 'replay-seed.v1.schema.json') {
    throw new Error(`unsupported replay seed target schema ${seed.targetSchema}`);
  }
  if (!Array.isArray(seed.fixtureRefs) || seed.fixtureRefs.length === 0) {
    throw new Error('fixtureRefs must not be empty');
  }
  if (!Array.isArray(seed.tags) || seed.tags.length === 0) {
    throw new Error('tags must not be empty');
  }
  if (!seed.expectedDeterminism || seed.expectedDeterminism.eventCount <= 0) {
    throw new Error('expectedDeterminism.eventCount must be positive');
  }
  if (!Array.isArray(seed.expectedDeterminism.orderedSourceRecordIds)) {
    throw new Error('orderedSourceRecordIds must be an array');
  }
  if (seed.expectedDeterminism.orderedSourceRecordIds.length !== seed.expectedDeterminism.eventCount) {
    throw new Error('orderedSourceRecordIds length must match eventCount');
  }
}

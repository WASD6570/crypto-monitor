import {
  DASHBOARD_AVAILABILITIES,
  DASHBOARD_SLOW_CONTEXT_AVAILABILITIES,
  DASHBOARD_SLOW_CONTEXT_FRESHNESS,
  DASHBOARD_SLOW_CONTEXT_METRIC_FAMILIES,
  type DashboardAvailability,
  type DashboardBucketContract,
  type DashboardBucketSectionContract,
  type DashboardBucketWindowContract,
  type DashboardCompositeSideContract,
  type DashboardGlobalStateContract,
  type DashboardGlobalStateSummaryContract,
  type DashboardSlowContextAvailability,
  type DashboardSlowContextContract,
  type DashboardSlowContextEntryContract,
  type DashboardSlowContextFreshness,
  type DashboardSlowContextMetricFamily,
  type DashboardSlowContextThresholdBasisContract,
  type DashboardSlowContextValueContract,
  type DashboardRecentContextFamilyContract,
  type DashboardSymbolStateContract,
  type DashboardVersionContract,
} from './dashboardContracts'
import { DASHBOARD_SYMBOLS, type DashboardSymbol } from '../../features/dashboard-shell/model/dashboardShellModel'

const validAvailabilitySet = new Set<DashboardAvailability>(DASHBOARD_AVAILABILITIES)
const validSlowContextAvailabilitySet = new Set<DashboardSlowContextAvailability>(DASHBOARD_SLOW_CONTEXT_AVAILABILITIES)
const validSlowContextFreshnessSet = new Set<DashboardSlowContextFreshness>(DASHBOARD_SLOW_CONTEXT_FRESHNESS)
const validSlowContextMetricFamilySet = new Set<DashboardSlowContextMetricFamily>(DASHBOARD_SLOW_CONTEXT_METRIC_FAMILIES)
const validSymbolSet = new Set<DashboardSymbol>(DASHBOARD_SYMBOLS)

export class DashboardDecodeError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'DashboardDecodeError'
  }
}

export function decodeDashboardSymbolState(value: unknown): DashboardSymbolStateContract {
  const root = readObject(value, 'symbol response')

  return {
    schemaVersion: readString(root.schemaVersion, 'schemaVersion'),
    symbol: readSymbol(root.symbol, 'symbol'),
    asOf: readString(root.asOf, 'asOf'),
    version: readVersion(root.version, 'version'),
    slowContext: readSlowContext(root.slowContext, readSymbol(root.symbol, 'symbol')),
    composite: {
      availability: readAvailability(readObject(root.composite, 'composite').availability, 'composite.availability'),
      reasonCodes: readStringArray(readObject(root.composite, 'composite').reasonCodes, 'composite.reasonCodes'),
      world: readCompositeSide(readObject(root.composite, 'composite').world, 'composite.world'),
      usa: readCompositeSide(readObject(root.composite, 'composite').usa, 'composite.usa'),
    },
    buckets: readBuckets(root.buckets),
    regime: readRegime(root.regime),
    recentContext: readRecentContext(root.recentContext),
    provenance: {
      historySeam: {
        reservedSchemaFamily: readString(
          readObject(readObject(root.provenance, 'provenance').historySeam, 'provenance.historySeam')
            .reservedSchemaFamily,
          'provenance.historySeam.reservedSchemaFamily',
        ),
      },
    },
  }
}

export function decodeDashboardGlobalState(value: unknown): DashboardGlobalStateContract {
  const root = readObject(value, 'global response')
  const globalRoot = readObject(root.global, 'global')
  const symbolsValue = root.symbols
  const symbols = Array.isArray(symbolsValue)
    ? symbolsValue.map((entry, index) => readGlobalSummary(entry, `symbols[${index}]`))
    : fail('symbols must be an array')

  return {
    schemaVersion: readString(root.schemaVersion, 'schemaVersion'),
    asOf: readString(root.asOf, 'asOf'),
    version: readVersion(root.version, 'version'),
    global: {
      state: readString(globalRoot.state, 'global.state'),
      reasons: readStringArray(globalRoot.reasons, 'global.reasons'),
      effectiveBucketEnd: readString(globalRoot.effectiveBucketEnd, 'global.effectiveBucketEnd'),
    },
    symbols,
    provenance: {
      historySeam: {
        reservedSchemaFamily: readString(
          readObject(readObject(root.provenance, 'provenance').historySeam, 'provenance.historySeam')
            .reservedSchemaFamily,
          'provenance.historySeam.reservedSchemaFamily',
        ),
      },
    },
  }
}

function readBuckets(value: unknown): DashboardSymbolStateContract['buckets'] {
  const buckets = readObject(value, 'buckets')

  return {
    thirtySeconds: readBucketSection(buckets.thirtySeconds, 'buckets.thirtySeconds'),
    twoMinutes: readBucketSection(buckets.twoMinutes, 'buckets.twoMinutes'),
    fiveMinutes: readBucketSection(buckets.fiveMinutes, 'buckets.fiveMinutes'),
  }
}

function readSlowContext(value: unknown, symbol: DashboardSymbol): DashboardSlowContextContract {
  if (value === undefined || value === null) {
    return createMissingSlowContext(symbol)
  }

  const slowContext = readObject(value, 'slowContext')
  const asset = readString(slowContext.asset, 'slowContext.asset')
  const contextsValue = slowContext.contexts
  const contexts = Array.isArray(contextsValue)
    ? contextsValue.map((entry, index) => readSlowContextEntry(entry, asset, `slowContext.contexts[${index}]`))
    : fail('slowContext.contexts must be an array')

  return {
    asset,
    queriedAt: readOptionalString(slowContext.queriedAt, 'slowContext.queriedAt'),
    contexts,
  }
}

function readSlowContextEntry(value: unknown, asset: string, name: string): DashboardSlowContextEntryContract {
  const entry = readObject(value, name)

  return {
    sourceFamily: readOptionalString(entry.sourceFamily, `${name}.sourceFamily`),
    metricFamily: readSlowContextMetricFamily(entry.metricFamily, `${name}.metricFamily`),
    asset: readOptionalString(entry.asset, `${name}.asset`) ?? asset,
    availability: readSlowContextAvailability(entry.availability, `${name}.availability`),
    freshness: readSlowContextFreshness(entry.freshness, `${name}.freshness`),
    expectedCadence: readOptionalString(entry.expectedCadence, `${name}.expectedCadence`),
    asOfTs: readOptionalString(entry.asOfTs, `${name}.asOfTs`),
    publishedTs: readOptionalString(entry.publishedTs, `${name}.publishedTs`),
    ingestTs: readOptionalString(entry.ingestTs, `${name}.ingestTs`),
    revision: readOptionalString(entry.revision, `${name}.revision`),
    value: readOptionalSlowContextValue(entry.value, `${name}.value`),
    previousValue: readOptionalSlowContextValue(entry.previousValue, `${name}.previousValue`),
    thresholdBasis: readOptionalThresholdBasis(entry.thresholdBasis, `${name}.thresholdBasis`),
    messageKey: readString(entry.messageKey, `${name}.messageKey`),
    message: readString(entry.message, `${name}.message`),
    error: readOptionalString(entry.error, `${name}.error`),
  }
}

function readRecentContext(value: unknown): DashboardSymbolStateContract['recentContext'] {
  const recentContext = readObject(value, 'recentContext')

  return {
    schemaVersion: readString(recentContext.schemaVersion, 'recentContext.schemaVersion'),
    thirtySeconds: readRecentContextFamily(recentContext.thirtySeconds, 'recentContext.thirtySeconds'),
    twoMinutes: readRecentContextFamily(recentContext.twoMinutes, 'recentContext.twoMinutes'),
    fiveMinutes: readRecentContextFamily(recentContext.fiveMinutes, 'recentContext.fiveMinutes'),
  }
}

function readRegime(value: unknown): DashboardSymbolStateContract['regime'] {
  const regime = readObject(value, 'regime')
  const symbol = readObject(regime.symbol, 'regime.symbol')
  const global = readObject(regime.global, 'regime.global')

  return {
    availability: readAvailability(regime.availability, 'regime.availability'),
    effectiveState: readString(regime.effectiveState, 'regime.effectiveState'),
    reasonCodes: readStringArray(regime.reasonCodes, 'regime.reasonCodes'),
    symbol: {
      state: readString(symbol.state, 'regime.symbol.state'),
      reasons: readStringArray(symbol.reasons, 'regime.symbol.reasons'),
      effectiveBucketEnd: readString(symbol.effectiveBucketEnd, 'regime.symbol.effectiveBucketEnd'),
    },
    global: {
      state: readString(global.state, 'regime.global.state'),
      reasons: readStringArray(global.reasons, 'regime.global.reasons'),
      effectiveBucketEnd: readString(global.effectiveBucketEnd, 'regime.global.effectiveBucketEnd'),
    },
  }
}

function readVersion(value: unknown, name: string): DashboardVersionContract {
  const version = readObject(value, name)

  return {
    schemaFamilyVersion: readString(version.schemaFamilyVersion, `${name}.schemaFamilyVersion`),
    configVersion: readString(version.configVersion, `${name}.configVersion`),
    algorithmVersion: readString(version.algorithmVersion, `${name}.algorithmVersion`),
  }
}

function readOptionalSlowContextValue(value: unknown, name: string): DashboardSlowContextValueContract | undefined {
  if (value === undefined || value === null) {
    return undefined
  }

  const parsed = readObject(value, name)

  return {
    amount: readString(parsed.amount, `${name}.amount`),
    unit: readString(parsed.unit, `${name}.unit`),
  }
}

function readOptionalThresholdBasis(value: unknown, name: string): DashboardSlowContextThresholdBasisContract | undefined {
  if (value === undefined || value === null) {
    return undefined
  }

  const parsed = readObject(value, name)

  return {
    expectedCadence: readString(parsed.expectedCadence, `${name}.expectedCadence`),
    delayedAfterTs: readOptionalString(parsed.delayedAfterTs, `${name}.delayedAfterTs`),
    staleAfterTs: readOptionalString(parsed.staleAfterTs, `${name}.staleAfterTs`),
    ageReference: readOptionalString(parsed.ageReference, `${name}.ageReference`),
  }
}

function readCompositeSide(value: unknown, name: string): DashboardCompositeSideContract {
  const side = readObject(value, name)

  return {
    bucketTs: readString(side.bucketTs, `${name}.bucketTs`),
    compositePrice: readOptionalNumber(side.compositePrice, `${name}.compositePrice`),
    coverageRatio: readOptionalNumber(side.coverageRatio, `${name}.coverageRatio`),
    healthScore: readOptionalNumber(side.healthScore, `${name}.healthScore`),
    degraded: readOptionalBoolean(side.degraded, `${name}.degraded`),
    unavailable: readOptionalBoolean(side.unavailable, `${name}.unavailable`),
  }
}

function readBucketSection(value: unknown, name: string): DashboardBucketSectionContract {
  const section = readObject(value, name)

  return {
    availability: readAvailability(section.availability, `${name}.availability`),
    reasonCodes: readStringArray(section.reasonCodes, `${name}.reasonCodes`),
    bucket: readBucket(section.bucket, `${name}.bucket`),
  }
}

function readBucket(value: unknown, name: string): DashboardBucketContract {
  const bucket = readObject(value, name)

  return {
    window: readBucketWindow(bucket.window, `${name}.window`),
  }
}

function readBucketWindow(value: unknown, name: string): DashboardBucketWindowContract {
  const windowValue = readObject(value, name)

  return {
    family: readString(windowValue.family, `${name}.family`),
    end: readString(windowValue.end, `${name}.end`),
    missingBucketCount: readOptionalInteger(windowValue.missingBucketCount, `${name}.missingBucketCount`) ?? 0,
    expectedBucketCount: readOptionalInteger(windowValue.expectedBucketCount, `${name}.expectedBucketCount`) ?? 0,
  }
}

function readRecentContextFamily(value: unknown, name: string): DashboardRecentContextFamilyContract {
  const family = readObject(value, name)
  const bucketsValue = family.buckets
  const buckets = Array.isArray(bucketsValue)
    ? bucketsValue.map((entry, index) => readBucket(entry, `${name}.buckets[${index}]`))
    : fail(`${name}.buckets must be an array`)

  return {
    availability: readAvailability(family.availability, `${name}.availability`),
    complete: readBoolean(family.complete, `${name}.complete`),
    missingBucketCount: readOptionalInteger(family.missingBucketCount, `${name}.missingBucketCount`) ?? 0,
    buckets,
  }
}

function readGlobalSummary(value: unknown, name: string): DashboardGlobalStateSummaryContract {
  const summary = readObject(value, name)

  return {
    symbol: readSymbol(summary.symbol, `${name}.symbol`),
    asOf: readString(summary.asOf, `${name}.asOf`),
    effectiveState: readString(summary.effectiveState, `${name}.effectiveState`),
    availability: readAvailability(summary.availability, `${name}.availability`),
    reasonCodes: readStringArray(summary.reasonCodes, `${name}.reasonCodes`),
    configVersion: readOptionalString(summary.configVersion, `${name}.configVersion`),
    algorithmVersion: readOptionalString(summary.algorithmVersion, `${name}.algorithmVersion`),
  }
}

function readObject(value: unknown, name: string): Record<string, unknown> {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) {
    fail(`${name} must be an object`)
  }

  return value as Record<string, unknown>
}

function readString(value: unknown, name: string): string {
  if (typeof value !== 'string' || value.length === 0) {
    fail(`${name} must be a non-empty string`)
  }

  return value
}

function readOptionalString(value: unknown, name: string): string | undefined {
  if (value === undefined || value === null) {
    return undefined
  }

  return readString(value, name)
}

function readStringArray(value: unknown, name: string): string[] {
  if (value === undefined || value === null) {
    return []
  }
  if (!Array.isArray(value)) {
    fail(`${name} must be an array of strings`)
  }

  return value.map((entry, index) => readString(entry, `${name}[${index}]`))
}

function readNumber(value: unknown, name: string): number {
  if (typeof value !== 'number' || Number.isNaN(value)) {
    fail(`${name} must be a number`)
  }

  return value
}

function readOptionalNumber(value: unknown, name: string): number | undefined {
  if (value === undefined || value === null) {
    return undefined
  }

  return readNumber(value, name)
}

function readOptionalInteger(value: unknown, name: string): number | undefined {
  if (value === undefined || value === null) {
    return undefined
  }

  const parsed = readNumber(value, name)
  if (!Number.isInteger(parsed)) {
    fail(`${name} must be an integer`)
  }
  return parsed
}

function readBoolean(value: unknown, name: string): boolean {
  if (typeof value !== 'boolean') {
    fail(`${name} must be a boolean`)
  }

  return value
}

function readOptionalBoolean(value: unknown, name: string): boolean | undefined {
  if (value === undefined || value === null) {
    return undefined
  }

  return readBoolean(value, name)
}

function readAvailability(value: unknown, name: string): DashboardAvailability {
  const parsed = readString(value, name)
  if (!validAvailabilitySet.has(parsed as DashboardAvailability)) {
    fail(`${name} must be one of ${DASHBOARD_AVAILABILITIES.join(', ')}`)
  }

  return parsed as DashboardAvailability
}

function readSlowContextAvailability(value: unknown, name: string): DashboardSlowContextAvailability {
  const parsed = readString(value, name)
  if (!validSlowContextAvailabilitySet.has(parsed as DashboardSlowContextAvailability)) {
    fail(`${name} must be one of ${DASHBOARD_SLOW_CONTEXT_AVAILABILITIES.join(', ')}`)
  }

  return parsed as DashboardSlowContextAvailability
}

function readSlowContextFreshness(value: unknown, name: string): DashboardSlowContextFreshness {
  const parsed = readString(value, name)
  if (!validSlowContextFreshnessSet.has(parsed as DashboardSlowContextFreshness)) {
    fail(`${name} must be one of ${DASHBOARD_SLOW_CONTEXT_FRESHNESS.join(', ')}`)
  }

  return parsed as DashboardSlowContextFreshness
}

function readSlowContextMetricFamily(value: unknown, name: string): DashboardSlowContextMetricFamily {
  const parsed = readString(value, name)
  if (!validSlowContextMetricFamilySet.has(parsed as DashboardSlowContextMetricFamily)) {
    fail(`${name} must be one of ${DASHBOARD_SLOW_CONTEXT_METRIC_FAMILIES.join(', ')}`)
  }

  return parsed as DashboardSlowContextMetricFamily
}

function readSymbol(value: unknown, name: string): DashboardSymbol {
  const parsed = readString(value, name)
  if (!validSymbolSet.has(parsed as DashboardSymbol)) {
    fail(`${name} must be one of ${DASHBOARD_SYMBOLS.join(', ')}`)
  }

  return parsed as DashboardSymbol
}

function createMissingSlowContext(symbol: DashboardSymbol): DashboardSlowContextContract {
  const asset = symbol.split('-')[0]

  return {
    asset,
    contexts: DASHBOARD_SLOW_CONTEXT_METRIC_FAMILIES.map((metricFamily) => ({
      metricFamily,
      asset,
      availability: 'unavailable',
      freshness: 'unavailable',
      messageKey: `${metricFamily}_unavailable`,
      message: 'Slow context is unavailable',
      error: 'slow-context block missing from symbol payload',
    })),
  }
}

function fail(message: string): never {
  throw new DashboardDecodeError(message)
}

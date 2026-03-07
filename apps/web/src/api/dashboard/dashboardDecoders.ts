import {
  DASHBOARD_AVAILABILITIES,
  type DashboardAvailability,
  type DashboardBucketContract,
  type DashboardBucketSectionContract,
  type DashboardBucketWindowContract,
  type DashboardCompositeSideContract,
  type DashboardGlobalStateContract,
  type DashboardGlobalStateSummaryContract,
  type DashboardRecentContextFamilyContract,
  type DashboardSymbolStateContract,
  type DashboardVersionContract,
} from './dashboardContracts'
import { DASHBOARD_SYMBOLS, type DashboardSymbol } from '../../features/dashboard-shell/model/dashboardShellModel'

const validAvailabilitySet = new Set<DashboardAvailability>(DASHBOARD_AVAILABILITIES)
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

function readSymbol(value: unknown, name: string): DashboardSymbol {
  const parsed = readString(value, name)
  if (!validSymbolSet.has(parsed as DashboardSymbol)) {
    fail(`${name} must be one of ${DASHBOARD_SYMBOLS.join(', ')}`)
  }

  return parsed as DashboardSymbol
}

function fail(message: string): never {
  throw new DashboardDecodeError(message)
}

import type { DashboardSymbol } from '../../features/dashboard-shell/model/dashboardShellModel'

export const DASHBOARD_AVAILABILITIES = [
  'available',
  'degraded',
  'partial',
  'unavailable',
] as const

export type DashboardAvailability = (typeof DASHBOARD_AVAILABILITIES)[number]

export const DASHBOARD_SLOW_CONTEXT_AVAILABILITIES = [
  'available',
  'unavailable',
] as const

export type DashboardSlowContextAvailability = (typeof DASHBOARD_SLOW_CONTEXT_AVAILABILITIES)[number]

export const DASHBOARD_SLOW_CONTEXT_FRESHNESS = [
  'fresh',
  'delayed',
  'stale',
  'unavailable',
] as const

export type DashboardSlowContextFreshness = (typeof DASHBOARD_SLOW_CONTEXT_FRESHNESS)[number]

export const DASHBOARD_SLOW_CONTEXT_METRIC_FAMILIES = [
  'cme_volume',
  'cme_open_interest',
  'etf_daily_flow',
] as const

export type DashboardSlowContextMetricFamily = (typeof DASHBOARD_SLOW_CONTEXT_METRIC_FAMILIES)[number]

export type DashboardVersionContract = {
  schemaFamilyVersion: string
  configVersion: string
  algorithmVersion: string
}

export type DashboardSlowContextValueContract = {
  amount: string
  unit: string
}

export type DashboardSlowContextThresholdBasisContract = {
  expectedCadence: string
  delayedAfterTs?: string
  staleAfterTs?: string
  ageReference?: string
}

export type DashboardSlowContextEntryContract = {
  sourceFamily?: string
  metricFamily: DashboardSlowContextMetricFamily
  asset: string
  availability: DashboardSlowContextAvailability
  freshness: DashboardSlowContextFreshness
  expectedCadence?: string
  asOfTs?: string
  publishedTs?: string
  ingestTs?: string
  revision?: string
  value?: DashboardSlowContextValueContract
  previousValue?: DashboardSlowContextValueContract
  thresholdBasis?: DashboardSlowContextThresholdBasisContract
  messageKey: string
  message: string
  error?: string
}

export type DashboardSlowContextContract = {
  asset: string
  queriedAt?: string
  contexts: DashboardSlowContextEntryContract[]
}

export type DashboardCompositeSideContract = {
  bucketTs: string
  compositePrice?: number
  coverageRatio?: number
  healthScore?: number
  degraded?: boolean
  unavailable?: boolean
}

export type DashboardBucketWindowContract = {
  family: string
  end: string
  missingBucketCount: number
  expectedBucketCount: number
}

export type DashboardBucketContract = {
  window: DashboardBucketWindowContract
}

export type DashboardBucketSectionContract = {
  availability: DashboardAvailability
  reasonCodes: string[]
  bucket: DashboardBucketContract
}

export type DashboardRecentContextFamilyContract = {
  availability: DashboardAvailability
  complete: boolean
  missingBucketCount: number
  buckets: DashboardBucketContract[]
}

export type DashboardSymbolStateContract = {
  schemaVersion: string
  symbol: DashboardSymbol
  asOf: string
  version: DashboardVersionContract
  slowContext: DashboardSlowContextContract
  composite: {
    availability: DashboardAvailability
    reasonCodes: string[]
    world: DashboardCompositeSideContract
    usa: DashboardCompositeSideContract
  }
  buckets: {
    thirtySeconds: DashboardBucketSectionContract
    twoMinutes: DashboardBucketSectionContract
    fiveMinutes: DashboardBucketSectionContract
  }
  regime: {
    availability: DashboardAvailability
    effectiveState: string
    reasonCodes: string[]
    symbol: {
      state: string
      reasons: string[]
      effectiveBucketEnd: string
    }
    global: {
      state: string
      reasons: string[]
      effectiveBucketEnd: string
    }
  }
  recentContext: {
    schemaVersion: string
    thirtySeconds: DashboardRecentContextFamilyContract
    twoMinutes: DashboardRecentContextFamilyContract
    fiveMinutes: DashboardRecentContextFamilyContract
  }
  provenance: {
    historySeam: {
      reservedSchemaFamily: string
    }
  }
}

export type DashboardGlobalStateSummaryContract = {
  symbol: DashboardSymbol
  asOf: string
  effectiveState: string
  availability: DashboardAvailability
  reasonCodes: string[]
  configVersion?: string
  algorithmVersion?: string
}

export type DashboardGlobalStateContract = {
  schemaVersion: string
  asOf: string
  version: DashboardVersionContract
  global: {
    state: string
    reasons: string[]
    effectiveBucketEnd: string
  }
  symbols: DashboardGlobalStateSummaryContract[]
  provenance: {
    historySeam: {
      reservedSchemaFamily: string
    }
  }
}

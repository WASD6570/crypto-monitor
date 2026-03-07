import type { DashboardSymbol } from '../../features/dashboard-shell/model/dashboardShellModel'

export const DASHBOARD_AVAILABILITIES = [
  'available',
  'degraded',
  'partial',
  'unavailable',
] as const

export type DashboardAvailability = (typeof DASHBOARD_AVAILABILITIES)[number]

export type DashboardVersionContract = {
  schemaFamilyVersion: string
  configVersion: string
  algorithmVersion: string
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

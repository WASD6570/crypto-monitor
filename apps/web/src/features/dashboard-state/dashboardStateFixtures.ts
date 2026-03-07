import type { DashboardClient } from '../../api/dashboard/dashboardClient'
import type { DashboardGlobalStateContract, DashboardSymbolStateContract } from '../../api/dashboard/dashboardContracts'
import { type DashboardSymbol } from '../dashboard-shell/model/dashboardShellModel'

export type DashboardResponseSet = {
  global: DashboardGlobalStateContract
  symbols: Record<DashboardSymbol, DashboardSymbolStateContract>
}

const symbolVersion = {
  schemaFamilyVersion: 'market_state_current_response_v1',
  configVersion: 'regime-engine.market-state.v1',
  algorithmVersion: 'symbol-global-regime.v1',
}

const globalVersion = {
  schemaFamilyVersion: 'market_state_current_global_v1',
  configVersion: 'regime-engine.market-state.v1',
  algorithmVersion: 'symbol-global-regime.v1',
}

function createBucket(family: string, end: string, availability: DashboardSymbolStateContract['buckets']['thirtySeconds']['availability'], reasonCodes: string[], missingBucketCount = 0, expectedBucketCount = 1) {
  return {
    availability,
    reasonCodes,
    bucket: {
      window: {
        family,
        end,
        missingBucketCount,
        expectedBucketCount,
      },
    },
  }
}

function createRecentContext(family: string, end: string, availability: DashboardSymbolStateContract['recentContext']['thirtySeconds']['availability']) {
  return {
    availability,
    complete: availability === 'available',
    missingBucketCount: availability === 'partial' ? 1 : 0,
    buckets: [
      {
        window: {
          family,
          end,
          missingBucketCount: availability === 'partial' ? 1 : 0,
          expectedBucketCount: 1,
        },
      },
    ],
  }
}

function createSymbolState(symbol: DashboardSymbol, asOf: string, overrides?: Partial<DashboardSymbolStateContract>): DashboardSymbolStateContract {
  const isBtc = symbol === 'BTC-USD'

  return {
    schemaVersion: 'market_state_current_response_v1',
    symbol,
    asOf,
    version: symbolVersion,
    composite: {
      availability: 'available',
      reasonCodes: [],
      world: {
        bucketTs: asOf,
        compositePrice: isBtc ? 64_000 : 3_240,
        coverageRatio: 1,
        healthScore: 0.99,
      },
      usa: {
        bucketTs: asOf,
        compositePrice: isBtc ? 63_996 : 3_236,
        coverageRatio: isBtc ? 0.98 : 0.86,
        healthScore: isBtc ? 0.97 : 0.84,
      },
    },
    buckets: {
      thirtySeconds: createBucket('30s', asOf, 'available', []),
      twoMinutes: createBucket('2m', asOf, isBtc ? 'available' : 'degraded', isBtc ? [] : ['timestamp-trust-loss']),
      fiveMinutes: createBucket('5m', asOf, isBtc ? 'available' : 'degraded', isBtc ? [] : ['global-shared-watch']),
    },
    regime: {
      availability: isBtc ? 'available' : 'degraded',
      effectiveState: isBtc ? 'TRADEABLE' : 'WATCH',
      reasonCodes: isBtc ? ['healthy'] : ['global-shared-watch', 'timestamp-trust-loss'],
      symbol: {
        state: isBtc ? 'TRADEABLE' : 'WATCH',
        reasons: isBtc ? ['healthy'] : ['timestamp-trust-loss'],
        effectiveBucketEnd: asOf,
      },
      global: {
        state: isBtc ? 'TRADEABLE' : 'WATCH',
        reasons: isBtc ? ['healthy'] : ['global-shared-watch'],
        effectiveBucketEnd: asOf,
      },
    },
    recentContext: {
      schemaVersion: 'market_state_recent_context_v1',
      thirtySeconds: createRecentContext('30s', asOf, 'available'),
      twoMinutes: createRecentContext('2m', asOf, 'available'),
      fiveMinutes: createRecentContext('5m', asOf, 'available'),
    },
    provenance: {
      historySeam: {
        reservedSchemaFamily: 'market-state-history-and-audit-reads',
      },
    },
    ...overrides,
  }
}

export const healthyDashboardResponses: DashboardResponseSet = {
  global: {
    schemaVersion: 'market_state_current_global_v1',
    asOf: '2026-01-15T14:32:10Z',
    version: globalVersion,
    global: {
      state: 'WATCH',
      reasons: ['global-shared-watch'],
      effectiveBucketEnd: '2026-01-15T14:32:10Z',
    },
    symbols: [
      {
        symbol: 'BTC-USD',
        asOf: '2026-01-15T14:32:10Z',
        effectiveState: 'TRADEABLE',
        availability: 'available',
        reasonCodes: ['healthy'],
        configVersion: symbolVersion.configVersion,
        algorithmVersion: symbolVersion.algorithmVersion,
      },
      {
        symbol: 'ETH-USD',
        asOf: '2026-01-15T14:32:10Z',
        effectiveState: 'WATCH',
        availability: 'degraded',
        reasonCodes: ['global-shared-watch', 'timestamp-trust-loss'],
        configVersion: symbolVersion.configVersion,
        algorithmVersion: symbolVersion.algorithmVersion,
      },
    ],
    provenance: {
      historySeam: {
        reservedSchemaFamily: 'market-state-history-and-audit-reads',
      },
    },
  },
  symbols: {
    'BTC-USD': createSymbolState('BTC-USD', '2026-01-15T14:31:54Z'),
    'ETH-USD': createSymbolState('ETH-USD', '2026-01-15T14:31:50Z'),
  },
}

export const degradedDashboardResponses: DashboardResponseSet = {
  global: healthyDashboardResponses.global,
  symbols: {
    ...healthyDashboardResponses.symbols,
    'ETH-USD': createSymbolState('ETH-USD', '2026-01-15T14:31:23Z', {
      composite: {
        availability: 'degraded',
        reasonCodes: ['timestamp-fallback', 'feed-health-degraded'],
        world: {
          bucketTs: '2026-01-15T14:31:23Z',
          compositePrice: 3_242,
          coverageRatio: 0.98,
          healthScore: 0.94,
        },
        usa: {
          bucketTs: '2026-01-15T14:31:23Z',
          compositePrice: 3_236,
          coverageRatio: 0.72,
          healthScore: 0.64,
          degraded: true,
        },
      },
      regime: {
        availability: 'degraded',
        effectiveState: 'WATCH',
        reasonCodes: ['global-shared-watch', 'timestamp-trust-loss'],
        symbol: {
          state: 'WATCH',
          reasons: ['timestamp-trust-loss'],
          effectiveBucketEnd: '2026-01-15T14:31:23Z',
        },
        global: {
          state: 'WATCH',
          reasons: ['global-shared-watch'],
          effectiveBucketEnd: '2026-01-15T14:31:23Z',
        },
      },
      buckets: {
        thirtySeconds: createBucket('30s', '2026-01-15T14:31:23Z', 'degraded', ['timestamp-trust-loss']),
        twoMinutes: createBucket('2m', '2026-01-15T14:31:23Z', 'degraded', ['feed-health-degraded', 'timestamp-trust-loss']),
        fiveMinutes: createBucket('5m', '2026-01-15T14:31:23Z', 'degraded', ['global-shared-watch']),
      },
    }),
  },
}

export const partialDashboardResponses: DashboardResponseSet = {
  global: healthyDashboardResponses.global,
  symbols: {
    ...healthyDashboardResponses.symbols,
    'ETH-USD': createSymbolState('ETH-USD', '2026-01-15T14:31:50Z', {
      composite: {
        availability: 'partial',
        reasonCodes: ['missing-input'],
        world: {
          bucketTs: '2026-01-15T14:31:50Z',
          compositePrice: 3_240,
          coverageRatio: 1,
          healthScore: 0.97,
        },
        usa: {
          bucketTs: '2026-01-15T14:31:50Z',
          coverageRatio: 0.4,
          healthScore: 0.45,
          unavailable: true,
        },
      },
      regime: {
        availability: 'degraded',
        effectiveState: 'WATCH',
        reasonCodes: ['composite-unavailable'],
        symbol: {
          state: 'WATCH',
          reasons: ['composite-unavailable'],
          effectiveBucketEnd: '2026-01-15T14:31:50Z',
        },
        global: {
          state: 'WATCH',
          reasons: ['global-shared-watch'],
          effectiveBucketEnd: '2026-01-15T14:31:50Z',
        },
      },
      buckets: {
        thirtySeconds: createBucket('30s', '2026-01-15T14:31:50Z', 'partial', ['missing-input'], 1, 2),
        twoMinutes: createBucket('2m', '2026-01-15T14:31:50Z', 'degraded', ['coverage-low']),
        fiveMinutes: createBucket('5m', '2026-01-15T14:31:50Z', 'degraded', ['global-shared-watch']),
      },
      recentContext: {
        schemaVersion: 'market_state_recent_context_v1',
        thirtySeconds: createRecentContext('30s', '2026-01-15T14:31:50Z', 'partial'),
        twoMinutes: createRecentContext('2m', '2026-01-15T14:31:50Z', 'available'),
        fiveMinutes: createRecentContext('5m', '2026-01-15T14:31:50Z', 'available'),
      },
    }),
  },
}

export function createStaticDashboardClient(responses: DashboardResponseSet): DashboardClient {
  return {
    async getGlobalState() {
      return structuredClone(responses.global)
    },
    async getSymbolState(symbol) {
      return structuredClone(responses.symbols[symbol])
    },
  }
}

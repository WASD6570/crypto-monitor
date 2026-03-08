import type { DashboardClient } from '../../api/dashboard/dashboardClient'
import type {
  DashboardGlobalStateContract,
  DashboardSlowContextContract,
  DashboardSlowContextEntryContract,
  DashboardSymbolStateContract,
} from '../../api/dashboard/dashboardContracts'
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

function createSlowContextEntry(
  asset: string,
  metricFamily: DashboardSlowContextEntryContract['metricFamily'],
  overrides?: Partial<DashboardSlowContextEntryContract>,
): DashboardSlowContextEntryContract {
  const defaults: Record<DashboardSlowContextEntryContract['metricFamily'], DashboardSlowContextEntryContract> = {
    cme_volume: {
      sourceFamily: 'CME',
      metricFamily: 'cme_volume',
      asset,
      availability: 'available',
      freshness: 'fresh',
      expectedCadence: 'session',
      asOfTs: '2026-01-14T00:00:00Z',
      publishedTs: '2026-01-14T20:45:00Z',
      ingestTs: '2026-01-14T20:48:00Z',
      revision: '1',
      value: {
        amount: asset === 'BTC' ? '18342.55' : '9284.00',
        unit: 'contracts',
      },
      previousValue: {
        amount: asset === 'BTC' ? '17880.10' : '9012.25',
        unit: 'contracts',
      },
      thresholdBasis: {
        expectedCadence: 'session',
        delayedAfterTs: '2026-01-15T21:00:00Z',
        staleAfterTs: '2026-01-17T09:00:00Z',
        ageReference: 'as_of',
      },
      messageKey: 'cme_volume_fresh',
      message: 'CME volume is current',
    },
    cme_open_interest: {
      sourceFamily: 'CME',
      metricFamily: 'cme_open_interest',
      asset,
      availability: 'available',
      freshness: 'fresh',
      expectedCadence: 'session',
      asOfTs: '2026-01-14T00:00:00Z',
      publishedTs: '2026-01-14T20:45:00Z',
      ingestTs: '2026-01-14T20:48:00Z',
      revision: '1',
      value: {
        amount: asset === 'BTC' ? '9284.00' : '6120.25',
        unit: 'contracts',
      },
      previousValue: {
        amount: asset === 'BTC' ? '9151.90' : '6012.80',
        unit: 'contracts',
      },
      thresholdBasis: {
        expectedCadence: 'session',
        delayedAfterTs: '2026-01-15T21:00:00Z',
        staleAfterTs: '2026-01-17T09:00:00Z',
        ageReference: 'as_of',
      },
      messageKey: 'cme_open_interest_fresh',
      message: 'CME open interest is current',
    },
    etf_daily_flow: {
      sourceFamily: 'ETF',
      metricFamily: 'etf_daily_flow',
      asset,
      availability: asset === 'BTC' ? 'available' : 'unavailable',
      freshness: asset === 'BTC' ? 'fresh' : 'unavailable',
      expectedCadence: 'daily',
      asOfTs: asset === 'BTC' ? '2026-01-14T00:00:00Z' : undefined,
      publishedTs: asset === 'BTC' ? '2026-01-14T22:15:00Z' : undefined,
      ingestTs: asset === 'BTC' ? '2026-01-14T22:18:00Z' : undefined,
      revision: asset === 'BTC' ? '1' : undefined,
      value: asset === 'BTC'
        ? {
            amount: '245000000.00',
            unit: 'usd',
          }
        : undefined,
      previousValue: asset === 'BTC'
        ? {
            amount: '198000000.00',
            unit: 'usd',
          }
        : undefined,
      thresholdBasis: asset === 'BTC'
        ? {
            expectedCadence: 'daily',
            delayedAfterTs: '2026-01-15T23:00:00Z',
            staleAfterTs: '2026-01-17T23:00:00Z',
            ageReference: 'as_of',
          }
        : undefined,
      messageKey: asset === 'BTC' ? 'etf_daily_flow_fresh' : 'etf_daily_flow_unavailable',
      message: asset === 'BTC' ? 'ETF daily flow is current' : 'ETF daily flow is unavailable',
      error: asset === 'BTC' ? undefined : 'No trusted ETF daily flow is tracked for this focused asset.',
    },
  }

  return {
    ...defaults[metricFamily],
    ...overrides,
    value: overrides?.value === undefined ? defaults[metricFamily].value : overrides.value,
    previousValue: overrides?.previousValue === undefined ? defaults[metricFamily].previousValue : overrides.previousValue,
    thresholdBasis: overrides?.thresholdBasis === undefined ? defaults[metricFamily].thresholdBasis : overrides.thresholdBasis,
  }
}

function createSlowContext(asset: string, overrides?: Partial<Record<DashboardSlowContextEntryContract['metricFamily'], Partial<DashboardSlowContextEntryContract>>>): DashboardSlowContextContract {
  return {
    asset,
    queriedAt: '2026-01-15T14:32:10Z',
    contexts: [
      createSlowContextEntry(asset, 'cme_volume', overrides?.cme_volume),
      createSlowContextEntry(asset, 'cme_open_interest', overrides?.cme_open_interest),
      createSlowContextEntry(asset, 'etf_daily_flow', overrides?.etf_daily_flow),
    ],
  }
}

function createSymbolState(symbol: DashboardSymbol, asOf: string, overrides?: Partial<DashboardSymbolStateContract>): DashboardSymbolStateContract {
  const isBtc = symbol === 'BTC-USD'
  const asset = isBtc ? 'BTC' : 'ETH'

  return {
    schemaVersion: 'market_state_current_response_v1',
    symbol,
    asOf,
    version: symbolVersion,
    slowContext: createSlowContext(asset),
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

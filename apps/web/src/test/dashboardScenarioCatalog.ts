import type { DashboardClient } from '../api/dashboard/dashboardClient'
import type {
  DashboardGlobalStateContract,
  DashboardSymbolStateContract,
} from '../api/dashboard/dashboardContracts'
import {
  DASHBOARD_SYMBOLS,
  type DashboardSymbol,
} from '../features/dashboard-shell/model/dashboardShellModel'
import {
  DASHBOARD_STALE_AFTER_MS,
  createInitialDashboardDataState,
  type DashboardDataState,
} from '../features/dashboard-state/dashboardQueryState'
import {
  degradedDashboardResponses,
  healthyDashboardResponses,
  partialDashboardResponses,
  type DashboardResponseSet,
} from '../features/dashboard-state/dashboardStateFixtures'

export const DASHBOARD_SCENARIO_NAMES = [
  'healthy',
  'degraded',
  'stale',
  'partial',
  'unavailable',
] as const

export type DashboardScenarioName = (typeof DASHBOARD_SCENARIO_NAMES)[number]

export const DASHBOARD_SLOW_CONTEXT_SCENARIOS = [
  'healthy',
  'delayed',
  'stale',
  'partial',
  'unavailable',
] as const

export type DashboardSlowContextScenarioName = (typeof DASHBOARD_SLOW_CONTEXT_SCENARIOS)[number]

export type DashboardScenarioMockStep<T> = {
  responses: T[]
  error?: string
}

export type DashboardScenarioMockPlan = {
  focusedSymbol: DashboardSymbol
  global: DashboardScenarioMockStep<DashboardGlobalStateContract>
  symbols: Record<DashboardSymbol, DashboardScenarioMockStep<DashboardSymbolStateContract>>
}

export type DashboardScenarioState = {
  focusedSymbol: DashboardSymbol
  nowMs: number
  state: DashboardDataState
}

type DashboardScenarioClientOptions = {
  baseMs?: number
  slowContextVariant?: DashboardSlowContextScenarioName
}

type DashboardScenarioResponseOptions = {
  baseMs?: number
  slowContextVariant?: DashboardSlowContextScenarioName
}

const SCENARIO_BASE_RESPONSES: Record<DashboardScenarioName, DashboardResponseSet> = {
  healthy: createHealthyScenarioResponses(),
  degraded: degradedDashboardResponses,
  stale: createHealthyScenarioResponses(),
  partial: partialDashboardResponses,
  unavailable: createHealthyScenarioResponses(),
}

const SCENARIO_FOCUSED_SYMBOL: Record<DashboardScenarioName, DashboardSymbol> = {
  healthy: 'BTC-USD',
  degraded: 'ETH-USD',
  stale: 'BTC-USD',
  partial: 'ETH-USD',
  unavailable: 'ETH-USD',
}

const DEFAULT_BASE_MS = Date.parse('2026-01-15T14:32:20Z')
const DEFAULT_REQUEST_ERROR = 'current-state request failed: 503 Service Unavailable'
const RESPONSE_AGE_OFFSETS_MS: Record<'global' | DashboardSymbol, number> = {
  global: 10_000,
  'BTC-USD': 16_000,
  'ETH-USD': 20_000,
}

export function createDashboardScenarioResponses(
  name: DashboardScenarioName,
  options?: DashboardScenarioResponseOptions,
): DashboardResponseSet {
  const responses = freshenDashboardResponses(
    structuredClone(SCENARIO_BASE_RESPONSES[name]),
    options?.baseMs ?? DEFAULT_BASE_MS,
  )

  applySlowContextVariant(
    responses,
    SCENARIO_FOCUSED_SYMBOL[name],
    options?.slowContextVariant ?? defaultSlowContextVariant(name),
  )

  return responses
}

export function createDashboardScenarioMockPlan(
  name: DashboardScenarioName,
  options?: DashboardScenarioResponseOptions,
): DashboardScenarioMockPlan {
  const responses = createDashboardScenarioResponses(name, options)
  const successPlan = createSuccessPlan(responses)

  if (name === 'stale') {
    return {
      focusedSymbol: SCENARIO_FOCUSED_SYMBOL[name],
      global: { responses: [successPlan.global.responses[0]], error: DEFAULT_REQUEST_ERROR },
      symbols: {
        'BTC-USD': { responses: [successPlan.symbols['BTC-USD'].responses[0]], error: DEFAULT_REQUEST_ERROR },
        'ETH-USD': successPlan.symbols['ETH-USD'],
      },
    }
  }

  if (name === 'unavailable') {
    return {
      focusedSymbol: SCENARIO_FOCUSED_SYMBOL[name],
      global: successPlan.global,
      symbols: {
        'BTC-USD': successPlan.symbols['BTC-USD'],
        'ETH-USD': { responses: [], error: DEFAULT_REQUEST_ERROR },
      },
    }
  }

  return {
    focusedSymbol: SCENARIO_FOCUSED_SYMBOL[name],
    ...successPlan,
  }
}

export function createDashboardScenarioClient(
  name: DashboardScenarioName,
  options?: DashboardScenarioClientOptions,
): DashboardClient {
  const plan = createDashboardScenarioMockPlan(name, options)
  let globalCalls = 0
  const symbolCalls = DASHBOARD_SYMBOLS.reduce<Record<DashboardSymbol, number>>((accumulator, symbol) => {
    accumulator[symbol] = 0
    return accumulator
  }, {} as Record<DashboardSymbol, number>)

  return {
    async getGlobalState() {
      const response = nextScenarioResponse(plan.global, globalCalls)
      globalCalls += 1
      return response
    },
    async getSymbolState(symbol) {
      const response = nextScenarioResponse(plan.symbols[symbol], symbolCalls[symbol])
      symbolCalls[symbol] += 1
      return response
    },
  }
}

export function createDashboardScenarioState(
  name: DashboardScenarioName,
  options?: DashboardScenarioResponseOptions,
): DashboardScenarioState {
  const responses = createDashboardScenarioResponses(name, options)
  const state = createInitialDashboardDataState()

  state.global = {
    data: structuredClone(responses.global),
    pending: false,
    lastSuccessAt: Date.parse(responses.global.asOf),
  }

  for (const symbol of DASHBOARD_SYMBOLS) {
    state.symbols[symbol] = {
      data: structuredClone(responses.symbols[symbol]),
      pending: false,
      lastSuccessAt: Date.parse(responses.symbols[symbol].asOf),
    }
  }

  if (name === 'stale') {
    state.global.error = DEFAULT_REQUEST_ERROR
    state.global.lastFailureAt = Date.parse('2026-01-15T14:33:55Z')
    state.symbols['BTC-USD'].error = DEFAULT_REQUEST_ERROR
    state.symbols['BTC-USD'].lastFailureAt = Date.parse('2026-01-15T14:33:55Z')

    return {
      focusedSymbol: SCENARIO_FOCUSED_SYMBOL[name],
      nowMs: Date.parse('2026-01-15T14:31:54Z') + DASHBOARD_STALE_AFTER_MS + 5_000,
      state,
    }
  }

  if (name === 'unavailable') {
    state.symbols['ETH-USD'] = {
      pending: false,
      error: DEFAULT_REQUEST_ERROR,
      lastFailureAt: Date.parse('2026-01-15T14:32:10Z'),
    }
  }

  return {
    focusedSymbol: SCENARIO_FOCUSED_SYMBOL[name],
    nowMs: Date.parse('2026-01-15T14:32:20Z'),
    state,
  }
}

function createSuccessPlan(responses: DashboardResponseSet) {
  return {
    global: {
      responses: [responses.global],
    },
    symbols: {
      'BTC-USD': { responses: [responses.symbols['BTC-USD']] },
      'ETH-USD': { responses: [responses.symbols['ETH-USD']] },
    } satisfies Record<DashboardSymbol, DashboardScenarioMockStep<DashboardSymbolStateContract>>,
  }
}

function createHealthyScenarioResponses(): DashboardResponseSet {
  const responses = structuredClone(healthyDashboardResponses)

  responses.global.symbols = responses.global.symbols.map((summary) => {
    if (summary.symbol !== 'ETH-USD') {
      return summary
    }

    return {
      ...summary,
      availability: 'available',
      reasonCodes: ['global-shared-watch'],
    }
  })

  responses.symbols['ETH-USD'] = {
    ...responses.symbols['ETH-USD'],
    composite: {
      ...responses.symbols['ETH-USD'].composite,
      availability: 'available',
      reasonCodes: [],
    },
    buckets: {
      thirtySeconds: {
        ...responses.symbols['ETH-USD'].buckets.thirtySeconds,
        availability: 'available',
        reasonCodes: [],
      },
      twoMinutes: {
        ...responses.symbols['ETH-USD'].buckets.twoMinutes,
        availability: 'available',
        reasonCodes: [],
      },
      fiveMinutes: {
        ...responses.symbols['ETH-USD'].buckets.fiveMinutes,
        availability: 'available',
        reasonCodes: [],
      },
    },
    regime: {
      ...responses.symbols['ETH-USD'].regime,
      availability: 'available',
      reasonCodes: ['global-shared-watch'],
      symbol: {
        ...responses.symbols['ETH-USD'].regime.symbol,
        reasons: ['global-shared-watch'],
      },
      global: {
        ...responses.symbols['ETH-USD'].regime.global,
        reasons: ['global-shared-watch'],
      },
    },
  }

  return responses
}

function nextScenarioResponse<T>(step: DashboardScenarioMockStep<T>, callIndex: number): T {
  if (callIndex < step.responses.length) {
    return structuredClone(step.responses[callIndex])
  }

  if (step.error) {
    throw new Error(step.error)
  }

  if (step.responses.length > 0) {
    return structuredClone(step.responses[step.responses.length - 1])
  }

  throw new Error(DEFAULT_REQUEST_ERROR)
}

function freshenDashboardResponses(responses: DashboardResponseSet, baseMs: number): DashboardResponseSet {
  const globalAsOf = new Date(baseMs - RESPONSE_AGE_OFFSETS_MS.global).toISOString()
  const btcAsOf = new Date(baseMs - RESPONSE_AGE_OFFSETS_MS['BTC-USD']).toISOString()
  const ethAsOf = new Date(baseMs - RESPONSE_AGE_OFFSETS_MS['ETH-USD']).toISOString()

  responses.global.asOf = globalAsOf
  responses.global.global.effectiveBucketEnd = globalAsOf
  responses.global.symbols = responses.global.symbols.map((summary) => ({
    ...summary,
    asOf: summary.symbol === 'BTC-USD' ? btcAsOf : ethAsOf,
  }))

  updateSymbolTimestamps(responses.symbols['BTC-USD'], btcAsOf)
  updateSymbolTimestamps(responses.symbols['ETH-USD'], ethAsOf)
  updateSlowContextTimestamps(responses.symbols['BTC-USD'], baseMs, btcAsOf)
  updateSlowContextTimestamps(responses.symbols['ETH-USD'], baseMs, ethAsOf)

  return responses
}

function updateSymbolTimestamps(symbol: DashboardSymbolStateContract, asOf: string) {
  symbol.asOf = asOf
  symbol.composite.world.bucketTs = asOf
  symbol.composite.usa.bucketTs = asOf
  symbol.regime.symbol.effectiveBucketEnd = asOf
  symbol.regime.global.effectiveBucketEnd = asOf
  symbol.buckets.thirtySeconds.bucket.window.end = asOf
  symbol.buckets.twoMinutes.bucket.window.end = asOf
  symbol.buckets.fiveMinutes.bucket.window.end = asOf
  symbol.recentContext.thirtySeconds.buckets[0].window.end = asOf
  symbol.recentContext.twoMinutes.buckets[0].window.end = asOf
  symbol.recentContext.fiveMinutes.buckets[0].window.end = asOf
}

function updateSlowContextTimestamps(symbol: DashboardSymbolStateContract, baseMs: number, queriedAt: string) {
  symbol.slowContext.queriedAt = queriedAt

  for (const context of symbol.slowContext.contexts) {
    if (!context.asOfTs) {
      continue
    }

    const asOfMs = context.metricFamily === 'etf_daily_flow'
      ? baseMs - 36 * 60 * 60 * 1000
      : baseMs - 18 * 60 * 60 * 1000
    const publishedOffsetMs = context.metricFamily === 'etf_daily_flow' ? 2 * 60 * 60 * 1000 : 90 * 60 * 1000

    context.asOfTs = new Date(asOfMs).toISOString()
    context.publishedTs = new Date(asOfMs + publishedOffsetMs).toISOString()
    context.ingestTs = new Date(asOfMs + publishedOffsetMs + 3 * 60 * 1000).toISOString()

    if (context.thresholdBasis) {
      context.thresholdBasis = {
        ...context.thresholdBasis,
        delayedAfterTs: new Date(asOfMs + 24 * 60 * 60 * 1000).toISOString(),
        staleAfterTs: new Date(
          asOfMs + (context.metricFamily === 'etf_daily_flow' ? 72 : 60) * 60 * 60 * 1000,
        ).toISOString(),
      }
    }
  }
}

function applySlowContextVariant(
  responses: DashboardResponseSet,
  focusedSymbol: DashboardSymbol,
  variant: DashboardSlowContextScenarioName,
) {
  const slowContext = responses.symbols[focusedSymbol].slowContext

  switch (variant) {
    case 'healthy':
      return
    case 'delayed':
      updateSlowContextContext(slowContext, 'cme_open_interest', {
        freshness: 'delayed',
        messageKey: 'cme_open_interest_delayed',
        message: 'CME open interest is delayed',
      })
      return
    case 'stale':
      updateSlowContextContext(slowContext, 'cme_volume', {
        freshness: 'stale',
        messageKey: 'cme_volume_stale',
        message: 'CME volume is stale',
      })
      return
    case 'partial':
      updateSlowContextContext(
        slowContext,
        focusedSymbol === 'BTC-USD' ? 'etf_daily_flow' : 'cme_open_interest',
        unavailableSlowContextOverrides('partial slow-context metric unavailable'),
      )
      return
    case 'unavailable':
      for (const context of slowContext.contexts) {
        updateSlowContextContext(
          slowContext,
          context.metricFamily,
          unavailableSlowContextOverrides('slow context reader unavailable'),
        )
      }
  }
}

function defaultSlowContextVariant(name: DashboardScenarioName): DashboardSlowContextScenarioName {
  switch (name) {
    case 'degraded':
      return 'delayed'
    case 'stale':
      return 'stale'
    case 'partial':
      return 'partial'
    default:
      return 'healthy'
  }
}

function updateSlowContextContext(
  slowContext: DashboardSymbolStateContract['slowContext'],
  metricFamily: DashboardSymbolStateContract['slowContext']['contexts'][number]['metricFamily'],
  overrides: Partial<DashboardSymbolStateContract['slowContext']['contexts'][number]>,
) {
  slowContext.contexts = slowContext.contexts.map((context) => {
    if (context.metricFamily !== metricFamily) {
      return context
    }

    return {
      ...context,
      ...overrides,
    }
  })
}

function unavailableSlowContextOverrides(error: string): Partial<DashboardSymbolStateContract['slowContext']['contexts'][number]> {
  return {
    availability: 'unavailable',
    freshness: 'unavailable',
    asOfTs: undefined,
    publishedTs: undefined,
    ingestTs: undefined,
    revision: undefined,
    value: undefined,
    previousValue: undefined,
    thresholdBasis: undefined,
    messageKey: 'slow_context_unavailable',
    message: 'Slow context is unavailable',
    error,
  }
}

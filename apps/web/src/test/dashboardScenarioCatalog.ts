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
}

type DashboardScenarioResponseOptions = {
  baseMs?: number
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
  return freshenDashboardResponses(
    structuredClone(SCENARIO_BASE_RESPONSES[name]),
    options?.baseMs ?? DEFAULT_BASE_MS,
  )
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

export function createDashboardScenarioState(name: DashboardScenarioName): DashboardScenarioState {
  const responses = structuredClone(SCENARIO_BASE_RESPONSES[name])
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

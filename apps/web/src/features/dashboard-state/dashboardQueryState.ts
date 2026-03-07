import type {
  DashboardGlobalStateContract,
  DashboardSymbolStateContract,
} from '../../api/dashboard/dashboardContracts'
import { DASHBOARD_SYMBOLS, type DashboardSymbol } from '../dashboard-shell/model/dashboardShellModel'

export type DashboardClock = {
  now: () => number
}

export const defaultDashboardClock: DashboardClock = {
  now: () => Date.now(),
}

export const DASHBOARD_SYMBOL_REVALIDATE_MS = 15_000
export const DASHBOARD_STALE_AFTER_MS = 90_000
export const DASHBOARD_SEVERE_STALE_AFTER_MS = 5 * 60_000

export type DashboardSurfaceRecord<T> = {
  data?: T
  pending: boolean
  error?: string
  lastSuccessAt?: number
  lastFailureAt?: number
}

export type DashboardDataState = {
  global: DashboardSurfaceRecord<DashboardGlobalStateContract>
  symbols: Record<DashboardSymbol, DashboardSurfaceRecord<DashboardSymbolStateContract>>
}

export function createInitialDashboardDataState(): DashboardDataState {
  return {
    global: {
      pending: true,
    },
    symbols: DASHBOARD_SYMBOLS.reduce<Record<DashboardSymbol, DashboardSurfaceRecord<DashboardSymbolStateContract>>>(
      (accumulator, symbol) => {
        accumulator[symbol] = { pending: true }
        return accumulator
      },
      {} as Record<DashboardSymbol, DashboardSurfaceRecord<DashboardSymbolStateContract>>,
    ),
  }
}

export function toDashboardErrorMessage(error: unknown): string {
  if (error instanceof Error && error.message.length > 0) {
    return error.message
  }

  return 'current-state request failed'
}

import { useCallback, useEffect, useRef, useState } from 'react'
import { dashboardClient, type DashboardClient } from '../../api/dashboard/dashboardClient'
import { DASHBOARD_SYMBOLS, type DashboardSymbol, type DashboardViewModel } from '../dashboard-shell/model/dashboardShellModel'
import {
  DASHBOARD_SYMBOL_REVALIDATE_MS,
  createInitialDashboardDataState,
  defaultDashboardClock,
  toDashboardErrorMessage,
  type DashboardClock,
  type DashboardDataState,
} from './dashboardQueryState'
import { deriveDashboardViewModel } from './dashboardStateMapper'

type UseDashboardDataOptions = {
  client?: DashboardClient
  clock?: DashboardClock
  enabled?: boolean
  focusedSymbol: DashboardSymbol
}

export type UseDashboardDataResult = {
  viewModel: DashboardViewModel
  isInitialUnavailable: boolean
  initialUnavailableReason?: string
  retry: () => void
}

export function useDashboardData({
  client = dashboardClient,
  clock = defaultDashboardClock,
  enabled = true,
  focusedSymbol,
}: UseDashboardDataOptions): UseDashboardDataResult {
  const [state, setState] = useState<DashboardDataState>(() => createInitialDashboardDataState())
  const lastFocusedSymbolRef = useRef<DashboardSymbol>(focusedSymbol)

  const loadGlobal = useCallback(async () => {
    setState((current) => ({
      ...current,
      global: {
        ...current.global,
        pending: true,
        error: undefined,
      },
    }))

    try {
      const data = await client.getGlobalState()
      const nowMs = clock.now()
      setState((current) => ({
        ...current,
        global: {
          data,
          pending: false,
          lastSuccessAt: nowMs,
          lastFailureAt: current.global.lastFailureAt,
        },
      }))
    } catch (error) {
      const nowMs = clock.now()
      setState((current) => ({
        ...current,
        global: {
          ...current.global,
          pending: false,
          error: toDashboardErrorMessage(error),
          lastFailureAt: nowMs,
        },
      }))
    }
  }, [client, clock])

  const loadSymbol = useCallback(async (symbol: DashboardSymbol) => {
    setState((current) => ({
      ...current,
      symbols: {
        ...current.symbols,
        [symbol]: {
          ...current.symbols[symbol],
          pending: true,
          error: undefined,
        },
      },
    }))

    try {
      const data = await client.getSymbolState(symbol)
      const nowMs = clock.now()
      setState((current) => ({
        ...current,
        symbols: {
          ...current.symbols,
          [symbol]: {
            data,
            pending: false,
            lastSuccessAt: nowMs,
            lastFailureAt: current.symbols[symbol].lastFailureAt,
          },
        },
      }))
    } catch (error) {
      const nowMs = clock.now()
      setState((current) => ({
        ...current,
        symbols: {
          ...current.symbols,
          [symbol]: {
            ...current.symbols[symbol],
            pending: false,
            error: toDashboardErrorMessage(error),
            lastFailureAt: nowMs,
          },
        },
      }))
    }
  }, [client, clock])

  const refreshAll = useCallback(() => {
    void loadGlobal()
    for (const symbol of DASHBOARD_SYMBOLS) {
      void loadSymbol(symbol)
    }
  }, [loadGlobal, loadSymbol])

  useEffect(() => {
    if (!enabled) {
      return
    }

    refreshAll()
  }, [enabled, refreshAll])

  useEffect(() => {
    if (!enabled) {
      return
    }

    if (lastFocusedSymbolRef.current === focusedSymbol) {
      return
    }

    lastFocusedSymbolRef.current = focusedSymbol
    const record = state.symbols[focusedSymbol]
    const shouldRevalidate =
      !record.data ||
      record.error !== undefined ||
      (record.lastSuccessAt !== undefined && clock.now() - record.lastSuccessAt >= DASHBOARD_SYMBOL_REVALIDATE_MS)

    if (shouldRevalidate) {
      void loadSymbol(focusedSymbol)
    }
  }, [clock, enabled, focusedSymbol, loadSymbol, state.symbols])

  const derived = deriveDashboardViewModel({
    state,
    focusedSymbol,
    nowMs: clock.now(),
  })

  return {
    viewModel: derived.viewModel,
    isInitialUnavailable: !derived.hasRenderableShell && !derived.viewModel.isRefreshing && !hasPending(state),
    initialUnavailableReason: derived.initialUnavailableReason,
    retry: refreshAll,
  }
}

function hasPending(state: DashboardDataState): boolean {
  return state.global.pending || DASHBOARD_SYMBOLS.some((symbol) => state.symbols[symbol].pending)
}

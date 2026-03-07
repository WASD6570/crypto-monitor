import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  normalizeDashboardSection,
  normalizeDashboardSymbol,
  type DashboardSectionKey,
  type DashboardSymbol,
} from '../model/dashboardShellModel'

export type DashboardShellRouteState = {
  symbol: DashboardSymbol
  section: DashboardSectionKey
  setSymbol: (symbol: DashboardSymbol) => void
  setSection: (section: DashboardSectionKey) => void
}

function readRouteState() {
  const params = new URLSearchParams(window.location.search)

  return {
    symbol: normalizeDashboardSymbol(params.get('symbol')),
    section: normalizeDashboardSection(params.get('section')),
  }
}

function writeRouteState(next: { symbol?: DashboardSymbol; section?: DashboardSectionKey }) {
  const params = new URLSearchParams(window.location.search)

  if (next.symbol) {
    params.set('symbol', next.symbol)
  }

  if (next.section) {
    params.set('section', next.section)
  }

  window.history.replaceState({}, '', `${window.location.pathname}?${params.toString()}`)
}

function isNormalizedRouteState(symbol: DashboardSymbol, section: DashboardSectionKey) {
  const params = new URLSearchParams(window.location.search)

  return params.get('symbol') === symbol && params.get('section') === section
}

export function useDashboardShellRouteState(): DashboardShellRouteState {
  const [routeState, setRouteState] = useState(readRouteState)

  useEffect(() => {
    if (!isNormalizedRouteState(routeState.symbol, routeState.section)) {
      writeRouteState(routeState)
    }

    const onPopState = () => {
      setRouteState(readRouteState())
    }

    window.addEventListener('popstate', onPopState)

    return () => {
      window.removeEventListener('popstate', onPopState)
    }
  }, [])

  const setSymbol = useCallback((symbol: DashboardSymbol) => {
    const nextState = {
      symbol,
      section: routeState.section,
    }

    writeRouteState(nextState)
    setRouteState(nextState)
  }, [routeState.section])

  const setSection = useCallback((section: DashboardSectionKey) => {
    const nextState = {
      symbol: routeState.symbol,
      section,
    }

    writeRouteState(nextState)
    setRouteState(nextState)
  }, [routeState.symbol])

  return useMemo(() => ({
    symbol: routeState.symbol,
    section: routeState.section,
    setSymbol,
    setSection,
  }), [routeState.section, routeState.symbol, setSection, setSymbol])
}

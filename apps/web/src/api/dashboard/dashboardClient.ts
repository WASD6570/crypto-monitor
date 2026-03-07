import type { DashboardSymbol } from '../../features/dashboard-shell/model/dashboardShellModel'
import { decodeDashboardGlobalState, decodeDashboardSymbolState } from './dashboardDecoders'
import type { DashboardGlobalStateContract, DashboardSymbolStateContract } from './dashboardContracts'

export type DashboardClient = {
  getGlobalState: () => Promise<DashboardGlobalStateContract>
  getSymbolState: (symbol: DashboardSymbol) => Promise<DashboardSymbolStateContract>
}

type FetchLike = typeof fetch

export function createDashboardClient(fetcher: FetchLike = fetch): DashboardClient {
  return {
    async getGlobalState() {
      const payload = await fetchJson(fetcher, '/api/market-state/global')
      return decodeDashboardGlobalState(payload)
    },
    async getSymbolState(symbol) {
      const payload = await fetchJson(fetcher, `/api/market-state/${encodeURIComponent(symbol)}`)
      return decodeDashboardSymbolState(payload)
    },
  }
}

export const dashboardClient = createDashboardClient()

async function fetchJson(fetcher: FetchLike, url: string): Promise<unknown> {
  const response = await fetcher(url, {
    headers: {
      Accept: 'application/json',
    },
  })

  if (!response.ok) {
    throw new Error(`current-state request failed: ${response.status} ${response.statusText}`)
  }

  return response.json()
}

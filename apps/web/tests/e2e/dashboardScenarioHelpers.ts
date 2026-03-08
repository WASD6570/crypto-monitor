import type { Page, Route } from '@playwright/test'
import {
  createDashboardScenarioMockPlan,
  type DashboardScenarioMockStep,
  type DashboardScenarioName,
} from '../../src/test/dashboardScenarioCatalog'
import { DASHBOARD_SYMBOLS, type DashboardSymbol } from '../../src/features/dashboard-shell/model/dashboardShellModel'

export async function mockDashboardScenario(page: Page, name: DashboardScenarioName) {
  const plan = createDashboardScenarioMockPlan(name, { baseMs: Date.now() })
  let globalCalls = 0
  const symbolCalls = DASHBOARD_SYMBOLS.reduce<Record<DashboardSymbol, number>>((accumulator, symbol) => {
    accumulator[symbol] = 0
    return accumulator
  }, {} as Record<DashboardSymbol, number>)

  await page.route('**/api/market-state/global', async (route) => {
    await fulfillDashboardRoute(route, plan.global, globalCalls)
    globalCalls += 1
  })
  await page.route('**/api/market-state/BTC-USD', async (route) => {
    await fulfillDashboardRoute(route, plan.symbols['BTC-USD'], symbolCalls['BTC-USD'])
    symbolCalls['BTC-USD'] += 1
  })
  await page.route('**/api/market-state/ETH-USD', async (route) => {
    await fulfillDashboardRoute(route, plan.symbols['ETH-USD'], symbolCalls['ETH-USD'])
    symbolCalls['ETH-USD'] += 1
  })
}

async function fulfillDashboardRoute<T>(route: Route, step: DashboardScenarioMockStep<T>, callIndex: number) {
  const response = callIndex < step.responses.length ? step.responses[callIndex] : undefined

  if (response) {
    await route.fulfill({
      contentType: 'application/json',
      body: JSON.stringify(response),
    })
    return
  }

  if (step.error) {
    await route.fulfill({
      status: 503,
      contentType: 'application/json',
      body: JSON.stringify({
        error: step.error,
      }),
    })
    return
  }

  if (step.responses.length > 0) {
    await route.fulfill({
      contentType: 'application/json',
      body: JSON.stringify(step.responses[step.responses.length - 1]),
    })
    return
  }

  await route.fulfill({
    status: 503,
    contentType: 'application/json',
    body: JSON.stringify({
      error: step.error ?? 'current-state request failed: 503 Service Unavailable',
    }),
  })
}

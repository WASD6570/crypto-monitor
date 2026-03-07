import { expect, test } from '@playwright/test'
import {
  healthyDashboardResponses,
  partialDashboardResponses,
} from '../../src/features/dashboard-state/dashboardStateFixtures'

function freshenResponses<T extends typeof healthyDashboardResponses>(responses: T): T {
  const next = structuredClone(responses)
  const baseMs = Date.now()
  const globalAsOf = new Date(baseMs - 10_000).toISOString()
  const btcAsOf = new Date(baseMs - 16_000).toISOString()
  const ethAsOf = new Date(baseMs - 20_000).toISOString()

  next.global.asOf = globalAsOf
  next.global.global.effectiveBucketEnd = globalAsOf
  next.global.symbols = next.global.symbols.map((summary) => ({
    ...summary,
    asOf: summary.symbol === 'BTC-USD' ? btcAsOf : ethAsOf,
  }))

  updateSymbolTimestamps(next.symbols['BTC-USD'], btcAsOf)
  updateSymbolTimestamps(next.symbols['ETH-USD'], ethAsOf)

  return next
}

function updateSymbolTimestamps(symbol: (typeof healthyDashboardResponses)['symbols']['BTC-USD'], asOf: string) {
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

test('renders adapter-backed dashboard data and keeps symbol switching populated', async ({ page }) => {
  const responses = freshenResponses(healthyDashboardResponses)

  await page.route('**/api/market-state/global', async (route) => {
    await route.fulfill({
      contentType: 'application/json',
      body: JSON.stringify(responses.global),
    })
  })
  await page.route('**/api/market-state/BTC-USD', async (route) => {
    await route.fulfill({
      contentType: 'application/json',
      body: JSON.stringify(responses.symbols['BTC-USD']),
    })
  })
  await page.route('**/api/market-state/ETH-USD', async (route) => {
    await route.fulfill({
      contentType: 'application/json',
      body: JSON.stringify(responses.symbols['ETH-USD']),
    })
  })

  await page.goto('/dashboard')

  await expect(page.getByTestId('status-rail')).toBeVisible()
  await expect(page.getByTestId('summary-card-BTC-USD')).toBeVisible()
  await expect(page.getByTestId('summary-card-ETH-USD')).toBeVisible()
  await expect(page.getByTestId('focused-symbol-heading')).toContainText('BTC-USD')

  await page.getByRole('button', { name: /ETH-USD/i }).click()

  await expect(page).toHaveURL(/symbol=ETH-USD/)
  await expect(page.getByTestId('focused-symbol-heading')).toContainText('ETH-USD')
  await expect(page.getByText(/Derivatives Context is unavailable/i)).toBeVisible()
})

test('keeps partial service failures explicit without blanking safe surfaces', async ({ page }) => {
  const responses = freshenResponses(partialDashboardResponses)

  await page.route('**/api/market-state/global', async (route) => {
    await route.fulfill({
      contentType: 'application/json',
      body: JSON.stringify(responses.global),
    })
  })
  await page.route('**/api/market-state/BTC-USD', async (route) => {
    await route.fulfill({
      contentType: 'application/json',
      body: JSON.stringify(responses.symbols['BTC-USD']),
    })
  })
  await page.route('**/api/market-state/ETH-USD', async (route) => {
    await route.fulfill({
      contentType: 'application/json',
      body: JSON.stringify(responses.symbols['ETH-USD']),
    })
  })

  await page.goto('/dashboard?symbol=ETH-USD')

  await expect(page.getByTestId('summary-card-BTC-USD')).toBeVisible()
  await expect(page.getByTestId('summary-card-ETH-USD')).toBeVisible()
  await expect(page.getByTestId('section-slot-microstructure')).toContainText('degraded')
  await expect(page.getByText(/Derivatives Context is unavailable/i)).toBeVisible()
  await expect(page.getByTestId('section-slot-microstructure')).toContainText('Missing Input')
})

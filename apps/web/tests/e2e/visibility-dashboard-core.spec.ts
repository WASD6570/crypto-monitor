import { expect, test, type Page } from '@playwright/test'
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

async function mockDashboardResponses(page: Page, responses: typeof healthyDashboardResponses) {
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
}

test('renders focused panels for the default symbol and keeps detail composition visible', async ({ page }) => {
  await mockDashboardResponses(page, freshenResponses(healthyDashboardResponses))

  await page.goto('/dashboard')

  await expect(page.getByTestId('summary-card-BTC-USD')).toBeVisible()
  await expect(page.getByTestId('summary-card-ETH-USD')).toBeVisible()
  await expect(page.getByTestId('section-slot-overview')).toContainText('BTC-USD is TRADEABLE')
  await expect(page.getByTestId('section-slot-microstructure')).toContainText('Trusted bucket')
  await expect(page.getByTestId('section-slot-health')).toContainText('Global ceiling')
  await expect(page.getByTestId('section-slot-derivatives')).toContainText('BTC-USD derivatives unavailable')
})

test('switches symbols without hiding peer summaries and preserves route focus on reload', async ({ page }) => {
  await mockDashboardResponses(page, freshenResponses(healthyDashboardResponses))

  await page.goto('/dashboard')
  await page.getByRole('button', { name: /ETH-USD/i }).click()

  await expect(page).toHaveURL(/symbol=ETH-USD/)
  await expect(page.getByTestId('focused-symbol-heading')).toContainText('ETH-USD')
  await expect(page.getByTestId('summary-card-BTC-USD')).toBeVisible()
  await expect(page.getByTestId('summary-card-ETH-USD')).toBeVisible()
  await expect(page.getByTestId('section-slot-overview')).toContainText('ETH-USD is WATCH')
  await expect(page.getByTestId('section-slot-derivatives')).toContainText('ETH-USD derivatives unavailable')

  await page.reload()

  await expect(page).toHaveURL(/symbol=ETH-USD/)
  await expect(page.getByTestId('focused-symbol-heading')).toContainText('ETH-USD')
})

test('keeps degraded panel cues explicit when one symbol response is partial', async ({ page }) => {
  await mockDashboardResponses(page, freshenResponses(partialDashboardResponses))

  await page.goto('/dashboard?symbol=ETH-USD')

  await expect(page.getByTestId('section-slot-overview')).toContainText('ETH-USD is WATCH')
  await expect(page.getByTestId('section-slot-microstructure')).toContainText('Missing Input')
  await expect(page.getByTestId('section-slot-microstructure')).toContainText('degraded')
  await expect(page.getByTestId('section-slot-derivatives')).toContainText('Current-state derivatives context not yet exposed')
  await expect(page.getByTestId('section-slot-health')).toContainText('Global ceiling')
})

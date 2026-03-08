import { expect, test } from '@playwright/test'
import { mockDashboardScenario } from './dashboardScenarioHelpers'

test('renders adapter-backed dashboard data and keeps symbol switching populated', async ({ page }) => {
  await mockDashboardScenario(page, 'healthy')

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
  await mockDashboardScenario(page, 'partial')

  await page.goto('/dashboard?symbol=ETH-USD')

  await expect(page.getByTestId('summary-card-BTC-USD')).toBeVisible()
  await expect(page.getByTestId('summary-card-ETH-USD')).toBeVisible()
  await expect(page.getByTestId('section-slot-microstructure')).toContainText('degraded')
  await expect(page.getByText(/Derivatives Context is unavailable/i)).toBeVisible()
  await expect(page.getByTestId('section-slot-microstructure')).toContainText('Missing Input')
})

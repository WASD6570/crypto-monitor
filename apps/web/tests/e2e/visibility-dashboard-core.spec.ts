import { expect, test } from '@playwright/test'
import { mockDashboardScenario } from './dashboardScenarioHelpers'

test('renders focused panels for the default symbol and keeps detail composition visible', async ({ page }) => {
  await mockDashboardScenario(page, 'healthy')

  await page.goto('/dashboard')

  await expect(page.getByTestId('summary-card-BTC-USD')).toBeVisible()
  await expect(page.getByTestId('summary-card-ETH-USD')).toBeVisible()
  await expect(page.getByTestId('section-slot-overview')).toContainText('BTC-USD is TRADEABLE')
  await expect(page.getByTestId('section-slot-microstructure')).toContainText('Trusted bucket')
  await expect(page.getByTestId('section-slot-health')).toContainText('Global ceiling')
  await expect(page.getByTestId('section-slot-derivatives')).toContainText('BTC-USD derivatives unavailable')
})

test('switches symbols without hiding peer summaries and preserves route focus on reload', async ({ page }) => {
  await mockDashboardScenario(page, 'healthy')

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

test('keeps degraded timestamp-trust warnings explicit without blanking peer summaries', async ({ page }) => {
  await mockDashboardScenario(page, 'degraded')

  await page.goto('/dashboard?symbol=ETH-USD')

  await expect(page.getByTestId('summary-card-BTC-USD')).toBeVisible()
  await expect(page.getByTestId('summary-card-ETH-USD')).toBeVisible()
  await expect(page.getByTestId('route-warning')).toContainText('ETH-USD trust reduced')
  await expect(page.getByTestId('focused-symbol-warning')).toContainText('ETH-USD trust reduced')
  await expect(page.getByTestId('route-warning')).toContainText('Timestamp trust is reduced')
  await expect(page.getByTestId('section-slot-health')).toContainText('Timestamp Trust Loss')
})

test('keeps degraded panel cues explicit when one symbol response is partial', async ({ page }) => {
  await mockDashboardScenario(page, 'partial')

  await page.goto('/dashboard?symbol=ETH-USD')

  await expect(page.getByTestId('section-slot-overview')).toContainText('ETH-USD is WATCH')
  await expect(page.getByTestId('section-slot-microstructure')).toContainText('Missing Input')
  await expect(page.getByTestId('section-slot-microstructure')).toContainText('degraded')
  await expect(page.getByTestId('section-slot-derivatives')).toContainText('Current-state derivatives context not yet exposed')
  await expect(page.getByTestId('section-slot-health')).toContainText('Global ceiling')
})

test('keeps unavailable focused-symbol fallback explicit without blanking healthy neighbors', async ({ page }) => {
  await mockDashboardScenario(page, 'unavailable')

  await page.goto('/dashboard?symbol=ETH-USD&section=overview')

  await expect(page.getByTestId('summary-card-BTC-USD')).toBeVisible()
  await expect(page.getByTestId('summary-card-ETH-USD')).toContainText('Unavailable')
  await expect(page.getByTestId('route-warning')).toContainText('ETH-USD current state unavailable')
  await expect(page.getByTestId('focused-symbol-warning')).toContainText('ETH-USD current state unavailable')
  await expect(page.getByTestId('section-slot-overview')).toContainText('ETH-USD overview is unavailable')
  await expect(page.getByTestId('section-slot-health')).toContainText('Global ceiling')
})

test('keeps keyboard symbol and section navigation route-backed on desktop', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'chromium', 'desktop keyboard smoke only')

  await mockDashboardScenario(page, 'partial')

  await page.goto('/dashboard')

  const ethButton = page.getByTestId('summary-card-ETH-USD')
  await ethButton.focus()
  await page.keyboard.press('Enter')

  await expect(page).toHaveURL(/symbol=ETH-USD/)
  await expect(page.getByTestId('focused-symbol-heading')).toContainText('ETH-USD')
  await expect(page.getByTestId('summary-card-BTC-USD')).toBeVisible()
  await expect(page.getByTestId('summary-card-ETH-USD')).toHaveAttribute('aria-current', 'true')
  await expect(page.getByTestId('focused-symbol-warning')).toContainText('ETH-USD partial inputs')

  const microstructureButton = page.getByRole('button', { name: 'Microstructure' })
  await microstructureButton.focus()
  await page.keyboard.press('Enter')

  await expect(page).toHaveURL(/section=microstructure/)
  await expect(microstructureButton).toHaveAttribute('aria-current', 'true')
  await expect(page.getByTestId('section-slot-microstructure')).toContainText('Missing Input')

  await page.reload()

  await expect(page).toHaveURL(/symbol=ETH-USD/)
  await expect(page).toHaveURL(/section=microstructure/)
  await expect(page.getByRole('button', { name: 'Microstructure' })).toHaveAttribute('aria-current', 'true')
})

test('keeps stale last-known-good copy explicit after retry failure on desktop', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'chromium', 'desktop stale smoke only')

  await mockDashboardScenario(page, 'stale')

  await page.goto('/dashboard')

  await expect(page.getByTestId('summary-card-BTC-USD')).toBeVisible()
  await page.getByRole('button', { name: /Retry current state/i }).click()

  await expect(page.getByTestId('route-warning')).toContainText('Global trust stale')
  await expect(page.getByTestId('summary-card-BTC-USD')).toContainText('stale')
  await expect(page.getByTestId('section-slot-overview')).toContainText('last-known-good overview')
  await expect(page.getByTestId('section-slot-health')).toContainText('last successful payload')
})

test('keeps warning hierarchy and route context readable on mobile', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'mobile-chrome', 'mobile smoke only')

  await mockDashboardScenario(page, 'partial')

  await page.goto('/dashboard?symbol=ETH-USD&section=overview')

  await expect(page.getByTestId('route-warning')).toBeVisible()
  await expect(page.getByTestId('route-warning')).toContainText('ETH-USD partial inputs')
  await expect(page.getByTestId('focused-symbol-warning')).toBeVisible()
  await expect(page.getByTestId('summary-card-BTC-USD')).toBeVisible()
  await expect(page.getByTestId('summary-card-ETH-USD')).toBeVisible()
  await expect(page.getByTestId('summary-card-ETH-USD')).toHaveAttribute('aria-current', 'true')

  const healthButton = page.getByRole('button', { name: 'Feed Health And Regime' })
  await healthButton.click()

  await expect(page).toHaveURL(/section=health/)
  await expect(healthButton).toHaveAttribute('aria-current', 'true')
  await expect(page.getByTestId('section-slot-health')).toContainText('Global ceiling')

  await page.reload()

  await expect(page).toHaveURL(/symbol=ETH-USD/)
  await expect(page).toHaveURL(/section=health/)
  await expect(page.getByTestId('focused-symbol-heading')).toContainText('ETH-USD')
  await expect(page.getByRole('button', { name: 'Feed Health And Regime' })).toHaveAttribute('aria-current', 'true')
})

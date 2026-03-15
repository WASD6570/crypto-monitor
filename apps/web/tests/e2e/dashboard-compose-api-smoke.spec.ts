import { expect, test, type Page } from '@playwright/test'

test('loads dashboard data from the Go market-state API without frontend route mocks', async ({ page }) => {
  const baseURL = test.info().project.use.baseURL ?? 'http://127.0.0.1:4173'
  const appOrigin = new URL(baseURL).origin
  const sameOriginMarketStateResponse = page.waitForResponse(
    (response) =>
      response.url().startsWith(`${appOrigin}/api/market-state/`) && response.request().method() === 'GET',
  )

  await page.goto('/dashboard')
  await sameOriginMarketStateResponse

  await waitForReadableShell(page)

  await expect(page.getByTestId('summary-card-BTC-USD')).toBeVisible()
  await expect(page.getByTestId('summary-card-ETH-USD')).toBeVisible()
  await expect(page.getByTestId('focused-symbol-heading')).toContainText('BTC-USD')
  await expect(page.getByTestId('summary-strip')).toBeVisible()
  await expect(page.getByTestId('section-slot-overview')).toBeVisible()
  await expect(page.getByTestId('slow-context-panel')).toBeVisible()

  await page.getByRole('button', { name: /ETH-USD/i }).click()

  await expect(page).toHaveURL(/symbol=ETH-USD/)
  await expect(page.getByTestId('focused-symbol-heading')).toContainText('ETH-USD')

  const sectionsNav = page.getByRole('navigation', { name: 'Dashboard sections' })
  await sectionsNav.getByRole('button', { name: 'Health' }).click()

  await expect(page).toHaveURL(/section=health/)
  await expect(sectionsNav.getByRole('button', { name: 'Health' })).toHaveAttribute('aria-current', 'true')
  await expect(page.getByTestId('section-slot-health')).toBeVisible()
})

async function waitForReadableShell(page: Page) {
  const unavailableHeading = page.getByRole('heading', { name: 'Dashboard snapshot is not readable yet.' })
  const retryButton = page.getByRole('button', { name: 'Retry current state' })
  const btcSummary = page.getByTestId('summary-card-BTC-USD')
  const ethSummary = page.getByTestId('summary-card-ETH-USD')

  await expect
    .poll(
      async () => {
        if ((await btcSummary.isVisible()) && (await ethSummary.isVisible())) {
          return 'shell'
        }
        if (await unavailableHeading.isVisible()) {
          await retryButton.click()
          return 'unavailable'
        }
        return 'pending'
      },
      {
        intervals: [250, 500, 1000],
        timeout: 30000,
      },
    )
    .toBe('shell')
}

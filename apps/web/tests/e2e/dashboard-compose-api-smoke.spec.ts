import { expect, test } from '@playwright/test'

test('loads dashboard data from the Go market-state API without frontend route mocks', async ({ page }) => {
  await page.goto('/dashboard')

  await expect(page.getByTestId('summary-card-BTC-USD')).toBeVisible()
  await expect(page.getByTestId('summary-card-ETH-USD')).toBeVisible()
  await expect(page.getByTestId('focused-symbol-heading')).toContainText('BTC-USD')
  await expect(page.getByTestId('section-slot-overview')).toContainText('BTC-USD is TRADEABLE')
  await expect(page.getByTestId('slow-context-panel')).toBeVisible()

  await page.getByRole('button', { name: /ETH-USD/i }).click()

  await expect(page).toHaveURL(/symbol=ETH-USD/)
  await expect(page.getByTestId('focused-symbol-heading')).toContainText('ETH-USD')
  await expect(page.getByTestId('section-slot-overview')).toContainText('ETH-USD is WATCH')
})

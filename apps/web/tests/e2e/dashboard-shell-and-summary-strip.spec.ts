import { expect, test } from '@playwright/test'

test('renders the dashboard shell with both summary cards', async ({ page }) => {
  await page.goto('/dashboard')

  await expect(page.getByTestId('status-rail')).toBeVisible()
  await expect(page.getByTestId('summary-card-BTC-USD')).toBeVisible()
  await expect(page.getByTestId('summary-card-ETH-USD')).toBeVisible()
  await expect(page.getByTestId('focused-symbol-heading')).toContainText('BTC-USD')
})

test('updates symbol and section query params from shell controls', async ({ page }) => {
  await page.goto('/dashboard')

  await page.getByRole('button', { name: /ETH-USD/i }).click()
  await expect(page).toHaveURL(/symbol=ETH-USD/)
  await expect(page.getByTestId('focused-symbol-heading')).toContainText('ETH-USD')

  await page.getByRole('button', { name: /Feed Health And Regime/i }).click()
  await expect(page).toHaveURL(/section=health/)
})

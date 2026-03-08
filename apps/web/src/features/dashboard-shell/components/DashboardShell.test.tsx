import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, test } from 'vitest'
import type { DashboardClient } from '../../../api/dashboard/dashboardClient'
import { DashboardPage } from '../../../pages/dashboard/DashboardPage'
import type { DashboardFixture } from '../model/dashboardShellModel'
import { DASHBOARD_STALE_AFTER_MS, type DashboardClock } from '../../dashboard-state/dashboardQueryState'
import { createDashboardScenarioClient } from '../../../test/dashboardScenarioCatalog'
import {
  degradedDashboardFixture,
  partialDashboardFixture,
  healthyDashboardFixture,
} from '../fixtures/dashboardFixtures'
import { DashboardShell } from './DashboardShell'

type RenderOptions = {
  client?: DashboardClient
  clock?: DashboardClock
  fixture?: DashboardFixture
}

function renderAt(url: string, options?: RenderOptions) {
  window.history.replaceState({}, '', url)
  return render(<DashboardPage client={options?.client} clock={options?.clock} fixture={options?.fixture} />)
}

describe('Dashboard shell', () => {
  test('renders the status rail, both summary cards, and four named slots', () => {
    renderAt('/dashboard', { fixture: healthyDashboardFixture })

    expect(screen.getByTestId('status-rail')).toBeInTheDocument()
    expect(screen.getByTestId('summary-card-BTC-USD')).toBeInTheDocument()
    expect(screen.getByTestId('summary-card-ETH-USD')).toBeInTheDocument()
    expect(screen.getByTestId('focused-symbol-heading')).toHaveTextContent('BTC-USD')
    expect(screen.getByTestId('section-slot-overview')).toBeInTheDocument()
    expect(screen.getByTestId('section-slot-microstructure')).toBeInTheDocument()
    expect(screen.getByTestId('section-slot-derivatives')).toBeInTheDocument()
    expect(screen.getByTestId('section-slot-health')).toBeInTheDocument()
    expect(screen.getByText('Effective state')).toBeInTheDocument()
    expect(screen.getByText('Trusted bucket')).toBeInTheDocument()
  })

  test('switches focus to ETH from the keyboard while keeping both summary cards visible', async () => {
    const user = userEvent.setup()
    renderAt('/dashboard', { fixture: healthyDashboardFixture })

    const ethButton = screen.getByRole('button', { name: /ETH-USD WATCH/i })
    ethButton.focus()
    await user.keyboard('{Enter}')

    expect(screen.getByTestId('focused-symbol-heading')).toHaveTextContent('ETH-USD')
    expect(screen.getByTestId('summary-card-BTC-USD')).toBeInTheDocument()
    expect(screen.getByTestId('summary-card-ETH-USD')).toHaveAttribute('aria-current', 'true')
    expect(window.location.search).toContain('symbol=ETH-USD')
  })

  test('falls back to safe defaults for invalid query params', () => {
    renderAt('/dashboard?symbol=SOL-USD&section=panic', { fixture: healthyDashboardFixture })

    expect(screen.getByTestId('focused-symbol-heading')).toHaveTextContent('BTC-USD')
    expect(window.location.search).toContain('symbol=BTC-USD')
    expect(window.location.search).toContain('section=overview')
  })

  test('shows degraded trust notes from service-shaped fixture data', () => {
    render(
      <DashboardShell
        fixture={degradedDashboardFixture}
        onRetry={() => undefined}
        routeState={{
          symbol: 'ETH-USD',
          section: 'overview',
          setSymbol: () => undefined,
          setSection: () => undefined,
        }}
      />,
    )

    expect(screen.getByTestId('route-warning')).toHaveTextContent('ETH-USD trust reduced')
    expect(screen.getByText(/Coinbase freshness degraded/i)).toBeInTheDocument()
    expect(screen.getByText(/recvTs fallback/i)).toBeInTheDocument()
    expect(
      screen.getAllByText(/Showing the last-known-good overview while degraded trust cues stay visible/i).length,
    ).toBeGreaterThan(0)
    expect(screen.getAllByText('WATCH').length).toBeGreaterThan(0)
  })

  test('keeps the shell usable when one section is unavailable', () => {
    render(
      <DashboardShell
        fixture={partialDashboardFixture}
        onRetry={() => undefined}
        routeState={{
          symbol: 'BTC-USD',
          section: 'health',
          setSymbol: () => undefined,
          setSection: () => undefined,
        }}
      />,
    )

    expect(screen.getByTestId('summary-strip')).toBeInTheDocument()
    expect(screen.getByTestId('section-slot-health')).toHaveTextContent('unavailable')
    expect(screen.getByTestId('route-warning')).toHaveTextContent('Feed Health And Regime unavailable')
    expect(screen.getByTestId('section-slot-overview')).toHaveTextContent('Focused symbol regime')
  })

  test('keeps active section semantics and focused warning copy visible through keyboard navigation', async () => {
    const user = userEvent.setup()

    renderAt('/dashboard?symbol=ETH-USD&section=overview', { fixture: degradedDashboardFixture })

    const healthButton = screen.getByRole('button', { name: 'Feed Health And Regime' })
    healthButton.focus()

    expect(screen.getByTestId('focused-symbol-warning')).toHaveTextContent('ETH-USD trust reduced')
    expect(screen.getByRole('button', { name: 'Overview' })).toHaveAttribute('aria-current', 'true')
    expect(healthButton).toHaveAttribute('aria-describedby')

    await user.keyboard('{Enter}')

    expect(screen.getByRole('button', { name: 'Feed Health And Regime' })).toHaveAttribute('aria-current', 'true')
    expect(window.location.search).toContain('section=health')
  })

  test('loads adapter-backed responses through the page and keeps degraded trust notes visible', async () => {
    renderAt('/dashboard', { client: createDashboardScenarioClient('degraded') })

    expect(await screen.findByTestId('summary-card-BTC-USD')).toBeInTheDocument()
    expect(screen.getByTestId('route-warning')).toBeInTheDocument()
    expect(await screen.findByText(/Timestamp trust is reduced/i)).toBeInTheDocument()
    expect(screen.getByText(/Derivatives Context is unavailable/i)).toBeInTheDocument()
    expect(screen.getAllByText('Global ceiling').length).toBeGreaterThan(0)
  })

  test('switches focused panel content with the symbol while keeping peer summaries visible', async () => {
    const user = userEvent.setup()

    renderAt('/dashboard', { client: createDashboardScenarioClient('healthy') })

    expect(await screen.findByTestId('summary-card-BTC-USD')).toBeInTheDocument()
    expect(screen.getByTestId('section-slot-overview')).toHaveTextContent('BTC-USD is TRADEABLE')
    expect(screen.getByTestId('section-slot-derivatives')).toHaveTextContent('BTC-USD derivatives unavailable')

    await user.click(screen.getByRole('button', { name: /ETH-USD/i }))

    expect(screen.getByTestId('focused-symbol-heading')).toHaveTextContent('ETH-USD')
    expect(screen.getByTestId('summary-card-BTC-USD')).toBeInTheDocument()
    expect(screen.getByTestId('summary-card-ETH-USD')).toBeInTheDocument()
    expect(screen.getByTestId('section-slot-overview')).toHaveTextContent('ETH-USD is WATCH')
    expect(screen.getByTestId('section-slot-derivatives')).toHaveTextContent('ETH-USD derivatives unavailable')
    expect(window.location.search).toContain('symbol=ETH-USD')
  })

  test('keeps unavailable focused-symbol fallback explicit while peer summaries stay visible', async () => {
    renderAt('/dashboard?symbol=ETH-USD&section=overview', { client: createDashboardScenarioClient('unavailable') })

    expect(await screen.findByTestId('summary-card-BTC-USD')).toBeInTheDocument()
    expect(screen.getByTestId('summary-card-ETH-USD')).toHaveTextContent('Unavailable')
    expect(screen.getByTestId('route-warning')).toHaveTextContent('Global trust unavailable')
    expect(screen.getByTestId('focused-symbol-warning')).toHaveTextContent('ETH-USD current state unavailable')
    expect(screen.getByTestId('section-slot-overview')).toHaveTextContent('ETH-USD overview is unavailable')
    expect(screen.getByTestId('section-slot-health')).toHaveTextContent('Global ceiling')
  })

  test('keeps last-known-good state visible after a failed refresh', async () => {
    const user = userEvent.setup()
    let nowMs = Date.parse('2026-01-15T14:32:20Z')
    const clock: DashboardClock = {
      now: () => nowMs,
    }
    const client: DashboardClient = createDashboardScenarioClient('stale', {
      baseMs: Date.parse('2026-01-15T14:32:20Z'),
    })

    renderAt('/dashboard', { client, clock })

    expect(await screen.findByTestId('summary-card-BTC-USD')).toHaveTextContent('TRADEABLE')

    nowMs = Date.parse('2026-01-15T14:31:54Z') + DASHBOARD_STALE_AFTER_MS + 10_000
    await user.click(screen.getByRole('button', { name: /Retry current state/i }))

    expect((await screen.findAllByText(/^stale$/i)).length).toBeGreaterThan(0)
    expect(screen.getByTestId('summary-card-BTC-USD')).toHaveTextContent('TRADEABLE')
    expect(screen.getByTestId('route-warning')).toHaveTextContent('Global trust stale')
    expect((await screen.findAllByText(/last-known-good current state/i)).length).toBeGreaterThan(0)
  })
})

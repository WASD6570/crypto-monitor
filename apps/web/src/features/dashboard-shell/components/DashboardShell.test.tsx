import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import { describe, expect, test } from 'vitest'
import type { DashboardClient } from '../../../api/dashboard/dashboardClient'
import { DashboardPage } from '../../../pages/dashboard/DashboardPage'
import type { DashboardFixture, DashboardSymbol } from '../model/dashboardShellModel'
import { DASHBOARD_STALE_AFTER_MS, type DashboardClock } from '../../dashboard-state/dashboardQueryState'
import {
  createStaticDashboardClient,
  degradedDashboardResponses,
  healthyDashboardResponses,
} from '../../dashboard-state/dashboardStateFixtures'
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

  test('switches focus to ETH while keeping both summary cards visible', async () => {
    const user = userEvent.setup()
    renderAt('/dashboard', { fixture: healthyDashboardFixture })

    await user.click(screen.getByRole('button', { name: /ETH-USD/i }))

    expect(screen.getByTestId('focused-symbol-heading')).toHaveTextContent('ETH-USD')
    expect(screen.getByTestId('summary-card-BTC-USD')).toBeInTheDocument()
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

    expect(screen.getByText(/Coinbase freshness degraded/i)).toBeInTheDocument()
    expect(screen.getByText(/recvTs fallback/i)).toBeInTheDocument()
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
    expect(screen.getByTestId('section-slot-overview')).toHaveTextContent('Focused symbol regime')
  })

  test('loads adapter-backed responses through the page and keeps degraded trust notes visible', async () => {
    renderAt('/dashboard', { client: createStaticDashboardClient(degradedDashboardResponses) })

    expect(await screen.findByTestId('summary-card-BTC-USD')).toBeInTheDocument()
    expect(await screen.findByText(/Timestamp trust is reduced/i)).toBeInTheDocument()
    expect(screen.getByText(/Derivatives Context is unavailable/i)).toBeInTheDocument()
    expect(screen.getAllByText('Global ceiling').length).toBeGreaterThan(0)
  })

  test('switches focused panel content with the symbol while keeping peer summaries visible', async () => {
    const user = userEvent.setup()

    renderAt('/dashboard', { client: createStaticDashboardClient(healthyDashboardResponses) })

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

  test('keeps last-known-good state visible after a failed refresh', async () => {
    const user = userEvent.setup()
    let nowMs = Date.parse('2026-01-15T14:32:20Z')
    const clock: DashboardClock = {
      now: () => nowMs,
    }
    let globalCalls = 0
    let symbolCalls = 0
    const client: DashboardClient = {
      getGlobalState: vi.fn(async () => {
        globalCalls += 1
        if (globalCalls === 1) {
          return structuredClone(healthyDashboardResponses.global)
        }

        throw new Error('current-state request failed: 503 Service Unavailable')
      }),
      getSymbolState: vi.fn(async (symbol) => {
        symbolCalls += 1
        if (symbolCalls <= 2) {
          return structuredClone(healthyDashboardResponses.symbols[symbol as DashboardSymbol])
        }

        throw new Error('current-state request failed: 503 Service Unavailable')
      }),
    }

    renderAt('/dashboard', { client, clock })

    expect(await screen.findByTestId('summary-card-BTC-USD')).toHaveTextContent('TRADEABLE')

    nowMs = Date.parse('2026-01-15T14:31:54Z') + DASHBOARD_STALE_AFTER_MS + 10_000
    await user.click(screen.getByRole('button', { name: /Retry current state/i }))

    expect((await screen.findAllByText(/^stale$/i)).length).toBeGreaterThan(0)
    expect(screen.getByTestId('summary-card-BTC-USD')).toHaveTextContent('TRADEABLE')
    expect((await screen.findAllByText(/last-known-good current state/i)).length).toBeGreaterThan(0)
  })
})

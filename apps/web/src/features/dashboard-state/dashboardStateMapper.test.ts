import { describe, expect, test } from 'vitest'
import { createInitialDashboardDataState, DASHBOARD_STALE_AFTER_MS } from './dashboardQueryState'
import { deriveDashboardViewModel } from './dashboardStateMapper'
import {
  healthyDashboardResponses,
  partialDashboardResponses,
} from './dashboardStateFixtures'
import { createDashboardScenarioState, type DashboardScenarioName } from '../../test/dashboardScenarioCatalog'

describe('dashboard state mapper', () => {
  test.each<{ name: DashboardScenarioName; assert: (viewModel: ReturnType<typeof deriveDashboardViewModel>['viewModel']) => void }>([
    {
      name: 'healthy',
      assert(viewModel: ReturnType<typeof deriveDashboardViewModel>['viewModel']) {
        expect(viewModel.summaries['BTC-USD'].trustState).toBe('ready')
        expect(viewModel.summaries['ETH-USD'].warning).toBeUndefined()
        expect(viewModel.focusedPanels.overview.summary).toContain('BTC-USD is TRADEABLE')
        expect(viewModel.slowContextPanel.badgeLabel).toBe('Context only')
        expect(viewModel.slowContextPanel.rows).toEqual(
          expect.arrayContaining([
            expect.objectContaining({
              label: 'CME volume',
              freshnessLabel: 'Fresh',
            }),
          ]),
        )
        expect(viewModel.primaryWarning).toEqual(
          expect.objectContaining({ label: 'Derivatives Context unavailable' }),
        )
      },
    },
    {
      name: 'degraded',
      assert(viewModel: ReturnType<typeof deriveDashboardViewModel>['viewModel']) {
        expect(viewModel.summaries['ETH-USD'].trustState).toBe('degraded')
        expect(viewModel.summaries['ETH-USD'].warning).toEqual(
          expect.objectContaining({ label: 'ETH-USD trust reduced' }),
        )
        expect(viewModel.focusedPanels.health.warning).toEqual(
          expect.objectContaining({ label: 'Feed Health And Regime degraded' }),
        )
        expect(viewModel.primaryWarning).toEqual(
          expect.objectContaining({ label: 'ETH-USD trust reduced' }),
        )
      },
    },
    {
      name: 'stale',
      assert(viewModel: ReturnType<typeof deriveDashboardViewModel>['viewModel']) {
        expect(viewModel.summaries['BTC-USD'].trustState).toBe('stale')
        expect(viewModel.focusedPanels.overview.warning).toEqual(
          expect.objectContaining({ label: 'Overview stale' }),
        )
        expect(viewModel.primaryWarning).toEqual(
          expect.objectContaining({ label: 'Global trust stale' }),
        )
      },
    },
    {
      name: 'partial',
      assert(viewModel: ReturnType<typeof deriveDashboardViewModel>['viewModel']) {
        expect(viewModel.summaries['ETH-USD'].warning).toEqual(
          expect.objectContaining({ label: 'ETH-USD partial inputs' }),
        )
        expect(viewModel.slowContextPanel.rows).toEqual(
          expect.arrayContaining([
            expect.objectContaining({
              metricFamily: 'etf_daily_flow',
              valueLabel: 'Unavailable',
            }),
          ]),
        )
        expect(viewModel.focusedPanels.microstructure.warning).toEqual(
          expect.objectContaining({ detail: 'Missing Input' }),
        )
        expect(viewModel.sections.derivatives.status).toBe('unavailable')
      },
    },
    {
      name: 'unavailable',
      assert(viewModel: ReturnType<typeof deriveDashboardViewModel>['viewModel']) {
        expect(viewModel.summaries['BTC-USD'].trustState).toBe('ready')
        expect(viewModel.summaries['ETH-USD'].warning).toEqual(
          expect.objectContaining({ label: 'ETH-USD current state unavailable' }),
        )
        expect(viewModel.focusedPanels.overview.trustState).toBe('unavailable')
        expect(viewModel.slowContextPanel.trustState).toBe('unavailable')
        expect(viewModel.focusedPanels.health.trustState).toBe('ready')
      },
    },
  ])('keeps $name scenario trust and fallback states honest', ({ name, assert }) => {
    const scenario = createDashboardScenarioState(name)

    const derived = deriveDashboardViewModel({
      state: scenario.state,
      focusedSymbol: scenario.focusedSymbol,
      nowMs: scenario.nowMs,
    })

    assert(derived.viewModel)
  })

  test('keeps unaffected surfaces visible when one symbol response is partial', () => {
    const state = createInitialDashboardDataState()
    state.global = {
      data: structuredClone(partialDashboardResponses.global),
      pending: false,
      lastSuccessAt: Date.parse('2026-01-15T14:32:10Z'),
    }
    state.symbols['BTC-USD'] = {
      data: structuredClone(partialDashboardResponses.symbols['BTC-USD']),
      pending: false,
      lastSuccessAt: Date.parse('2026-01-15T14:31:54Z'),
    }
    state.symbols['ETH-USD'] = {
      data: structuredClone(partialDashboardResponses.symbols['ETH-USD']),
      pending: false,
      lastSuccessAt: Date.parse('2026-01-15T14:31:50Z'),
    }

    const derived = deriveDashboardViewModel({
      state,
      focusedSymbol: 'ETH-USD',
      nowMs: Date.parse('2026-01-15T14:32:20Z'),
    })

    expect(derived.viewModel.summaries['BTC-USD'].trustState).toBe('ready')
    expect(derived.viewModel.sections.microstructure.status).toBe('degraded')
    expect(derived.viewModel.sections.derivatives.status).toBe('unavailable')
    expect(derived.viewModel.focusedPanels.overview.metrics).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ label: 'Effective state', value: 'WATCH' }),
        expect.objectContaining({ label: 'Global ceiling', value: 'WATCH' }),
      ]),
    )
    expect(derived.viewModel.focusedPanels.health.metrics).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ label: 'Focused availability', value: 'degraded' }),
      ]),
    )
    expect(derived.viewModel.summaries['ETH-USD'].warning).toEqual(
      expect.objectContaining({
        tone: 'degraded',
        label: 'ETH-USD partial inputs',
      }),
    )
    expect(derived.viewModel.focusedPanels.microstructure.warning).toEqual(
      expect.objectContaining({
        tone: 'degraded',
        label: 'Microstructure degraded',
        detail: 'Missing Input',
      }),
    )
    expect(derived.viewModel.primaryWarning).toEqual(
      expect.objectContaining({
        tone: 'degraded',
        label: 'ETH-USD partial inputs',
      }),
    )
    expect(derived.viewModel.focusedPanels.derivatives.summary).toMatch(/not present in the current-state contract yet/i)
    expect(derived.viewModel.trustState).toBe('degraded')
  })

  test('marks last-known-good data stale after a failed refresh', () => {
    const state = createInitialDashboardDataState()
    state.global = {
      data: structuredClone(healthyDashboardResponses.global),
      pending: false,
      error: 'current-state request failed: 503 Service Unavailable',
      lastSuccessAt: Date.parse('2026-01-15T14:32:10Z'),
      lastFailureAt: Date.parse('2026-01-15T14:33:55Z'),
    }
    state.symbols['BTC-USD'] = {
      data: structuredClone(healthyDashboardResponses.symbols['BTC-USD']),
      pending: false,
      error: 'current-state request failed: 503 Service Unavailable',
      lastSuccessAt: Date.parse('2026-01-15T14:31:54Z'),
      lastFailureAt: Date.parse('2026-01-15T14:33:55Z'),
    }
    state.symbols['ETH-USD'] = {
      data: structuredClone(healthyDashboardResponses.symbols['ETH-USD']),
      pending: false,
      lastSuccessAt: Date.parse('2026-01-15T14:31:50Z'),
    }

    const derived = deriveDashboardViewModel({
      state,
      focusedSymbol: 'BTC-USD',
      nowMs: Date.parse('2026-01-15T14:31:54Z') + DASHBOARD_STALE_AFTER_MS + 5_000,
    })

    expect(derived.viewModel.summaries['BTC-USD'].trustState).toBe('stale')
    expect(derived.viewModel.sections.overview.status).toBe('stale')
    expect(derived.viewModel.summaries['BTC-USD'].warning).toEqual(
      expect.objectContaining({
        tone: 'stale',
        label: 'BTC-USD current state stale',
      }),
    )
    expect(derived.viewModel.focusedPanels.overview.note).toMatch(/last-known-good overview/i)
    expect(derived.viewModel.focusedPanels.overview.warning).toEqual(
      expect.objectContaining({
        tone: 'stale',
        label: 'Overview stale',
      }),
    )
    expect(derived.viewModel.primaryWarning).toEqual(
      expect.objectContaining({
        tone: 'stale',
        label: 'Global trust stale',
      }),
    )
    expect(derived.viewModel.focusedPanels.health.note).toMatch(/last successful payload/i)
    expect(derived.viewModel.degradedNotes.join(' ')).toMatch(/stale/i)
  })

  test('builds focused panel content for both symbols without inventing missing data', () => {
    const state = createInitialDashboardDataState()
    state.global = {
      data: structuredClone(healthyDashboardResponses.global),
      pending: false,
      lastSuccessAt: Date.parse('2026-01-15T14:32:10Z'),
    }
    state.symbols['BTC-USD'] = {
      data: structuredClone(healthyDashboardResponses.symbols['BTC-USD']),
      pending: false,
      lastSuccessAt: Date.parse('2026-01-15T14:31:54Z'),
    }
    state.symbols['ETH-USD'] = {
      data: structuredClone(healthyDashboardResponses.symbols['ETH-USD']),
      pending: false,
      lastSuccessAt: Date.parse('2026-01-15T14:31:50Z'),
    }

    const btcDerived = deriveDashboardViewModel({
      state,
      focusedSymbol: 'BTC-USD',
      nowMs: Date.parse('2026-01-15T14:32:20Z'),
    })
    const ethDerived = deriveDashboardViewModel({
      state,
      focusedSymbol: 'ETH-USD',
      nowMs: Date.parse('2026-01-15T14:32:20Z'),
    })

    expect(btcDerived.viewModel.focusedPanels.overview.summary).toContain('BTC-USD is TRADEABLE')
    expect(ethDerived.viewModel.focusedPanels.overview.summary).toContain('ETH-USD is WATCH')
    expect(ethDerived.viewModel.focusedPanels.microstructure.metrics).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ label: 'Trusted bucket', value: '30s' }),
        expect.objectContaining({ label: '30s context', value: 'Complete (0 missing)' }),
      ]),
    )
    expect(ethDerived.viewModel.focusedPanels.derivatives.trustState).toBe('unavailable')
    expect(ethDerived.viewModel.focusedPanels.derivatives.metrics).toContainEqual({
      label: 'Focused symbol',
      value: 'ETH-USD',
    })
    expect(btcDerived.viewModel.slowContextPanel.rows).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          label: 'ETF daily flow',
          valueLabel: '$245,000,000.00',
        }),
      ]),
    )
    expect(ethDerived.viewModel.slowContextPanel.rows).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          label: 'ETF daily flow',
          valueLabel: 'Unavailable',
          note: 'No trusted ETF daily flow is tracked for this focused asset.',
        }),
      ]),
    )
    expect(btcDerived.viewModel.primaryWarning).toEqual(
      expect.objectContaining({
        tone: 'degraded',
        label: 'Derivatives Context unavailable',
      }),
    )
  })
})

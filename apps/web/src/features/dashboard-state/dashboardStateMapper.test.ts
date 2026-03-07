import { describe, expect, test } from 'vitest'
import { createInitialDashboardDataState, DASHBOARD_STALE_AFTER_MS } from './dashboardQueryState'
import { deriveDashboardViewModel } from './dashboardStateMapper'
import {
  healthyDashboardResponses,
  partialDashboardResponses,
} from './dashboardStateFixtures'

describe('dashboard state mapper', () => {
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
    expect(derived.viewModel.focusedPanels.overview.note).toMatch(/last-known-good overview/i)
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
  })
})

import { describe, expect, test } from 'vitest'
import { deriveDashboardViewModel } from '../features/dashboard-state/dashboardStateMapper'
import {
  createDashboardScenarioClient,
  createDashboardScenarioMockPlan,
  createDashboardScenarioState,
} from './dashboardScenarioCatalog'

describe('dashboard scenario catalog', () => {
  test('returns deterministic mock plans for repeated inputs', () => {
    const baseMs = Date.parse('2026-03-08T12:00:00Z')

    const first = createDashboardScenarioMockPlan('degraded', { baseMs })
    const second = createDashboardScenarioMockPlan('degraded', { baseMs })

    expect(first).toEqual(second)
    expect(first).not.toBe(second)
    expect(first.global.responses[0].asOf).toBe('2026-03-08T11:59:50.000Z')
    expect(first.symbols['BTC-USD'].responses[0].asOf).toBe('2026-03-08T11:59:44.000Z')
    expect(first.symbols['ETH-USD'].responses[0].asOf).toBe('2026-03-08T11:59:40.000Z')
  })

  test('preserves stale last-known-good semantics instead of generic unavailable output', () => {
    const scenario = createDashboardScenarioState('stale')

    const derived = deriveDashboardViewModel({
      state: scenario.state,
      focusedSymbol: scenario.focusedSymbol,
      nowMs: scenario.nowMs,
    })

    expect(derived.viewModel.summaries['BTC-USD'].trustState).toBe('stale')
    expect(derived.viewModel.summaries['BTC-USD'].warning).toEqual(
      expect.objectContaining({ label: 'BTC-USD current state stale' }),
    )
    expect(derived.viewModel.primaryWarning).toEqual(
      expect.objectContaining({ label: 'Global trust stale' }),
    )
    expect(derived.viewModel.focusedPanels.overview.warning).toEqual(
      expect.objectContaining({ label: 'Overview stale' }),
    )
  })

  test('keeps partial and unavailable scenarios aligned with service-owned labels', () => {
    const partialScenario = createDashboardScenarioState('partial')
    const unavailableScenario = createDashboardScenarioState('unavailable')

    const partialView = deriveDashboardViewModel({
      state: partialScenario.state,
      focusedSymbol: partialScenario.focusedSymbol,
      nowMs: partialScenario.nowMs,
    }).viewModel
    const unavailableView = deriveDashboardViewModel({
      state: unavailableScenario.state,
      focusedSymbol: unavailableScenario.focusedSymbol,
      nowMs: unavailableScenario.nowMs,
    }).viewModel

    expect(partialView.summaries['ETH-USD'].stateLabel).toBe('WATCH')
    expect(partialView.summaries['ETH-USD'].warning).toEqual(
      expect.objectContaining({ label: 'ETH-USD partial inputs' }),
    )
    expect(partialView.focusedPanels.microstructure.warning).toEqual(
      expect.objectContaining({ detail: 'Missing Input' }),
    )

    expect(unavailableView.summaries['BTC-USD'].trustState).toBe('ready')
    expect(unavailableView.summaries['ETH-USD'].stateLabel).toBe('Unavailable')
    expect(unavailableView.summaries['ETH-USD'].warning).toEqual(
      expect.objectContaining({ label: 'ETH-USD current state unavailable' }),
    )
    expect(unavailableView.focusedPanels.overview.summary).toContain('ETH-USD overview is unavailable')
  })

  test('keeps stale and unavailable client scripts deterministic', async () => {
    const staleClient = createDashboardScenarioClient('stale', { baseMs: Date.parse('2026-03-08T12:00:00Z') })
    const unavailableClient = createDashboardScenarioClient('unavailable', { baseMs: Date.parse('2026-03-08T12:00:00Z') })

    await expect(staleClient.getGlobalState()).resolves.toEqual(
      expect.objectContaining({ asOf: '2026-03-08T11:59:50.000Z' }),
    )
    await expect(staleClient.getGlobalState()).rejects.toThrow('503 Service Unavailable')
    await expect(unavailableClient.getSymbolState('ETH-USD')).rejects.toThrow('503 Service Unavailable')
  })
})

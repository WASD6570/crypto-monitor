import { describe, expect, test } from 'vitest'
import { decodeDashboardSymbolState } from './dashboardDecoders'
import { healthyDashboardResponses } from '../../features/dashboard-state/dashboardStateFixtures'

describe('dashboard decoders', () => {
  test('accepts a healthy service-owned symbol payload', () => {
    const decoded = decodeDashboardSymbolState(healthyDashboardResponses.symbols['BTC-USD'])

    expect(decoded.symbol).toBe('BTC-USD')
    expect(decoded.regime.effectiveState).toBe('TRADEABLE')
    expect(decoded.slowContext.asset).toBe('BTC')
    expect(decoded.slowContext.contexts).toHaveLength(3)
    expect(decoded.slowContext.contexts[0]).toEqual(
      expect.objectContaining({
        metricFamily: 'cme_volume',
        freshness: 'fresh',
      }),
    )
    expect(decoded.provenance.historySeam.reservedSchemaFamily).toBe('market-state-history-and-audit-reads')
  })

  test('rejects malformed critical fields', () => {
    const malformed = {
      ...healthyDashboardResponses.symbols['BTC-USD'],
      asOf: '',
    }

    expect(() => decodeDashboardSymbolState(malformed)).toThrow(/asOf/i)
  })

  test('rejects malformed slow-context freshness values', () => {
    const malformed = structuredClone(healthyDashboardResponses.symbols['BTC-USD'])
    malformed.slowContext.contexts[0].freshness = 'later'

    expect(() => decodeDashboardSymbolState(malformed)).toThrow(/slowContext\.contexts\[0\]\.freshness/i)
  })

  test('falls back to an explicit unavailable slow-context block when the payload omits it', () => {
    const withoutSlowContext = {
      ...healthyDashboardResponses.symbols['ETH-USD'],
    }
    delete (withoutSlowContext as Partial<typeof withoutSlowContext>).slowContext

    const decoded = decodeDashboardSymbolState(withoutSlowContext)

    expect(decoded.slowContext.asset).toBe('ETH')
    expect(decoded.slowContext.contexts).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          metricFamily: 'etf_daily_flow',
          availability: 'unavailable',
          message: 'Slow context is unavailable',
        }),
      ]),
    )
  })
})

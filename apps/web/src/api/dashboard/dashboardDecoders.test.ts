import { describe, expect, test } from 'vitest'
import { decodeDashboardSymbolState } from './dashboardDecoders'
import { healthyDashboardResponses } from '../../features/dashboard-state/dashboardStateFixtures'

describe('dashboard decoders', () => {
  test('accepts a healthy service-owned symbol payload', () => {
    const decoded = decodeDashboardSymbolState(healthyDashboardResponses.symbols['BTC-USD'])

    expect(decoded.symbol).toBe('BTC-USD')
    expect(decoded.regime.effectiveState).toBe('TRADEABLE')
    expect(decoded.provenance.historySeam.reservedSchemaFamily).toBe('market-state-history-and-audit-reads')
  })

  test('rejects malformed critical fields', () => {
    const malformed = {
      ...healthyDashboardResponses.symbols['BTC-USD'],
      asOf: '',
    }

    expect(() => decodeDashboardSymbolState(malformed)).toThrow(/asOf/i)
  })
})

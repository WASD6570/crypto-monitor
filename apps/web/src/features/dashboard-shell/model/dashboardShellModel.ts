export type DashboardTrustState = 'loading' | 'ready' | 'stale' | 'degraded' | 'unavailable'
export type DashboardSymbol = 'BTC-USD' | 'ETH-USD'
export type DashboardSectionKey = 'overview' | 'microstructure' | 'derivatives' | 'health'
export type DashboardWarningTone = Exclude<DashboardTrustState, 'loading' | 'ready'>

export type DashboardWarning = {
  tone: DashboardWarningTone
  label: string
  detail: string
}

export type DashboardSummary = {
  symbol: DashboardSymbol
  stateLabel: string
  trustState: DashboardTrustState
  reasons: string[]
  lastUpdated: string
  freshnessLabel: string
  comparisonLabel: string
  timestampNote?: string
  warning?: DashboardWarning
}

export type DashboardSectionState = {
  key: DashboardSectionKey
  title: string
  status: DashboardTrustState
  note: string
  reasons: string[]
}

export type DashboardPanelMetric = {
  label: string
  value: string
}

export type DashboardFocusedPanel = {
  key: DashboardSectionKey
  title: string
  eyebrow?: string
  trustState: DashboardTrustState
  summary: string
  metrics: DashboardPanelMetric[]
  reasons: string[]
  note?: string
  warning?: DashboardWarning
}

export type DashboardSlowContextRow = {
  metricFamily: string
  label: string
  status: DashboardTrustState
  valueLabel: string
  freshnessLabel: string
  cadenceLabel: string
  asOfLabel: string
  publishedLabel?: string
  ingestLabel?: string
  previousValueLabel?: string
  revisionLabel?: string
  note: string
}

export type DashboardSlowContextPanel = {
  title: string
  eyebrow: string
  badgeLabel: string
  trustState: DashboardTrustState
  summary: string
  note: string
  rows: DashboardSlowContextRow[]
}

export type DashboardViewModel = {
  asOf: string
  oldestAgeLabel: string
  configVersion: string
  algorithmVersion: string
  trustState: DashboardTrustState
  primaryWarning?: DashboardWarning
  degradedNotes: string[]
  isRefreshing: boolean
  lastSuccessLabel?: string
  summaries: Record<DashboardSymbol, DashboardSummary>
  focusedPanels: Record<DashboardSectionKey, DashboardFocusedPanel>
  slowContextPanel: DashboardSlowContextPanel
  sections: Record<DashboardSectionKey, DashboardSectionState>
}

export type DashboardFixture = DashboardViewModel

export const DASHBOARD_SYMBOLS: DashboardSymbol[] = ['BTC-USD', 'ETH-USD']
export const DASHBOARD_SECTIONS: DashboardSectionKey[] = [
  'overview',
  'microstructure',
  'derivatives',
  'health',
]

const validSymbolSet = new Set<DashboardSymbol>(DASHBOARD_SYMBOLS)
const validSectionSet = new Set<DashboardSectionKey>(DASHBOARD_SECTIONS)

export function normalizeDashboardSymbol(value: string | null | undefined): DashboardSymbol {
  if (value && validSymbolSet.has(value as DashboardSymbol)) {
    return value as DashboardSymbol
  }

  return 'BTC-USD'
}

export function normalizeDashboardSection(value: string | null | undefined): DashboardSectionKey {
  if (value && validSectionSet.has(value as DashboardSectionKey)) {
    return value as DashboardSectionKey
  }

  return 'overview'
}

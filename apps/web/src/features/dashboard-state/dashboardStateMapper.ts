import type {
  DashboardAvailability,
  DashboardBucketSectionContract,
  DashboardGlobalStateContract,
  DashboardGlobalStateSummaryContract,
  DashboardSymbolStateContract,
} from '../../api/dashboard/dashboardContracts'
import {
  DASHBOARD_SECTIONS,
  type DashboardFocusedPanel,
  type DashboardPanelMetric,
  DASHBOARD_SYMBOLS,
  type DashboardSectionKey,
  type DashboardSectionState,
  type DashboardSummary,
  type DashboardSymbol,
  type DashboardTrustState,
  type DashboardViewModel,
} from '../dashboard-shell/model/dashboardShellModel'
import {
  DASHBOARD_SEVERE_STALE_AFTER_MS,
  DASHBOARD_STALE_AFTER_MS,
  type DashboardDataState,
  type DashboardSurfaceRecord,
} from './dashboardQueryState'

const SECTION_TITLES: Record<DashboardSectionKey, string> = {
  overview: 'Overview',
  microstructure: 'Microstructure',
  derivatives: 'Derivatives Context',
  health: 'Feed Health And Regime',
}

const TRUST_RANK: Record<DashboardTrustState, number> = {
  ready: 0,
  loading: 1,
  degraded: 2,
  stale: 3,
  unavailable: 4,
}

export type DashboardViewDerivation = {
  viewModel: DashboardViewModel
  hasRenderableShell: boolean
  initialUnavailableReason?: string
}

type DashboardViewInput = {
  state: DashboardDataState
  focusedSymbol: DashboardSymbol
  nowMs: number
}

export function deriveDashboardViewModel({ state, focusedSymbol, nowMs }: DashboardViewInput): DashboardViewDerivation {
  const summaries = DASHBOARD_SYMBOLS.reduce<Record<DashboardSymbol, DashboardSummary>>((accumulator, symbol) => {
    accumulator[symbol] = buildSummary(symbol, state.symbols[symbol], nowMs)
    return accumulator
  }, {} as Record<DashboardSymbol, DashboardSummary>)

  const focusedPanels = buildFocusedPanels(state, focusedSymbol, nowMs)
  const sections = buildSections(focusedPanels)
  const globalTrust = buildGlobalTrust(state.global, focusedSymbol, nowMs)
  const dataTimestamps = collectTimestamps(state)
  const hasRenderableShell = DASHBOARD_SYMBOLS.some((symbol) => Boolean(state.symbols[symbol].data))
  const overallTrust = deriveOverallTrust({ globalTrust, summaries, sections, hasRenderableShell })
  const lastSuccessAt = newestTimestamp([
    state.global.lastSuccessAt,
    ...DASHBOARD_SYMBOLS.map((symbol) => state.symbols[symbol].lastSuccessAt),
  ])
  const initialUnavailableReason = !hasRenderableShell && !hasPendingWork(state)
    ? buildInitialUnavailableReason(state)
    : undefined

  return {
    viewModel: {
      asOf: newestIsoTimestamp(dataTimestamps) ?? '',
      oldestAgeLabel: buildOldestAgeLabel(dataTimestamps, nowMs),
      configVersion:
        state.global.data?.version.configVersion ??
        state.symbols[focusedSymbol].data?.version.configVersion ??
        'awaiting-current-state',
      algorithmVersion:
        state.global.data?.version.algorithmVersion ??
        state.symbols[focusedSymbol].data?.version.algorithmVersion ??
        'awaiting-current-state',
      trustState: overallTrust,
      degradedNotes: buildDashboardNotes(state, summaries, sections, focusedSymbol, globalTrust.trustState),
      isRefreshing: hasRenderableShell && hasPendingWork(state),
      lastSuccessLabel: lastSuccessAt ? `Last successful refresh ${formatAgeLabel(nowMs - lastSuccessAt)} ago` : undefined,
      summaries,
      focusedPanels,
      sections,
    },
    hasRenderableShell,
    initialUnavailableReason,
  }
}

function buildSummary(
  symbol: DashboardSymbol,
  record: DashboardSurfaceRecord<DashboardSymbolStateContract>,
  nowMs: number,
): DashboardSummary {
  if (!record.data) {
    return {
      symbol,
      stateLabel: record.pending ? 'Loading current state' : 'Unavailable',
      trustState: record.pending ? 'loading' : 'unavailable',
      reasons: record.error ? [record.error] : [],
      lastUpdated: '',
      freshnessLabel: record.pending ? 'Awaiting first service read' : 'No current-state payload',
      comparisonLabel: 'WORLD/USA comparison unavailable',
    }
  }

  const reasonCodes = dedupeStrings([
    ...record.data.regime.reasonCodes,
    ...record.data.regime.symbol.reasons,
    ...record.data.composite.reasonCodes,
  ])
  const trust = applyAgeToTrust(
    combineTrusts([
      availabilityToTrust(record.data.regime.availability),
      availabilityToTrust(record.data.composite.availability),
    ]),
    record.data.asOf,
    Boolean(record.error),
    nowMs,
  )
  const reasons = reasonCodes.map(humanizeCode)
  if (record.error) {
    reasons.push('Latest refresh failed; showing last-known-good current state.')
  }

  return {
    symbol,
    stateLabel: record.data.regime.effectiveState,
    trustState: trust,
    reasons: dedupeStrings(reasons),
    lastUpdated: record.data.asOf,
    freshnessLabel: buildFreshnessLabel(record.data.asOf, nowMs),
    comparisonLabel: buildComparisonLabel(record.data),
    timestampNote: hasTimestampTrustReduction(reasonCodes)
      ? 'Timestamp trust is reduced in the service-owned payload.'
      : undefined,
  }
}

function buildFocusedPanels(state: DashboardDataState, focusedSymbol: DashboardSymbol, nowMs: number) {
  const focusedRecord = state.symbols[focusedSymbol]
  const focusedData = focusedRecord.data
  const globalRecord = state.global
  const globalSummary = globalRecord.data?.symbols.find((entry) => entry.symbol === focusedSymbol)

  return {
    overview: buildOverviewPanel(focusedSymbol, focusedRecord, nowMs),
    microstructure: buildMicrostructurePanel(focusedRecord, nowMs),
    derivatives: buildDerivativesPanel(focusedSymbol),
    health: buildHealthPanel(globalRecord, globalSummary, nowMs, focusedData),
  } satisfies Record<DashboardSectionKey, DashboardFocusedPanel>
}

function buildSections(panels: Record<DashboardSectionKey, DashboardFocusedPanel>) {
  return DASHBOARD_SECTIONS.reduce<Record<DashboardSectionKey, DashboardSectionState>>((accumulator, key) => {
    const panel = panels[key]
    accumulator[key] = createSectionState(key, panel.trustState, panel.note ?? panel.summary, panel.reasons)
    return accumulator
  }, {} as Record<DashboardSectionKey, DashboardSectionState>)
}

function buildOverviewPanel(
  focusedSymbol: DashboardSymbol,
  record: DashboardSurfaceRecord<DashboardSymbolStateContract>,
  nowMs: number,
): DashboardFocusedPanel {
  if (!record.data) {
    return createPanelState(
      'overview',
      record.pending ? 'loading' : 'unavailable',
      record.pending ? 'Focused overview pending' : 'Focused overview unavailable',
      record.pending
        ? `${focusedSymbol} overview is waiting for the first current-state payload.`
        : `${focusedSymbol} overview is unavailable until the service-owned symbol response loads.`,
      [],
      record.error ? [record.error] : [],
    )
  }

  const reasons = dedupeStrings([...record.data.regime.reasonCodes, ...record.data.composite.reasonCodes]).map(humanizeCode)
  const metrics: DashboardPanelMetric[] = [
    { label: 'Effective state', value: record.data.regime.effectiveState },
    { label: 'Symbol state', value: record.data.regime.symbol.state },
    { label: 'Global ceiling', value: record.data.regime.global.state },
    { label: 'WORLD composite', value: formatCompositeMetric(record.data.composite.world) },
    { label: 'USA composite', value: formatCompositeMetric(record.data.composite.usa) },
  ]
  let note: string | undefined
  if (record.error) {
    reasons.push('Last-known-good overview is retained after a failed refresh.')
    note = 'Showing the last-known-good overview while the latest refresh retries.'
  }

  return createPanelState(
    'overview',
    applyAgeToTrust(availabilityToTrust(record.data.regime.availability), record.data.asOf, Boolean(record.error), nowMs),
    'Focused symbol regime',
    `${focusedSymbol} is ${record.data.regime.effectiveState}. Symbol state ${record.data.regime.symbol.state}; global ceiling ${record.data.regime.global.state}.`,
    metrics,
    reasons,
    note,
  )
}

function buildMicrostructurePanel(
  record: DashboardSurfaceRecord<DashboardSymbolStateContract>,
  nowMs: number,
): DashboardFocusedPanel {
  if (!record.data) {
    return createPanelState(
      'microstructure',
      record.pending ? 'loading' : 'unavailable',
      record.pending ? 'Microstructure pending' : 'Microstructure unavailable',
      record.pending
        ? 'Microstructure context is waiting on the focused symbol payload.'
        : 'Microstructure context is unavailable because the focused symbol payload is missing.',
      [],
      record.error ? [record.error] : [],
    )
  }

  const bucket = selectMicrostructureBucket(record.data)
  const reasons = bucket.reasonCodes.map(humanizeCode)
  const metrics: DashboardPanelMetric[] = [
    { label: 'Trusted bucket', value: bucket.bucket.window.family },
    { label: 'Bucket close', value: formatReadableTime(bucket.bucket.window.end) },
    {
      label: 'Missing windows',
      value: `${bucket.bucket.window.missingBucketCount}/${bucket.bucket.window.expectedBucketCount}`,
    },
    { label: '30s context', value: formatRecentContextMetric(record.data.recentContext.thirtySeconds) },
    { label: '2m context', value: formatRecentContextMetric(record.data.recentContext.twoMinutes) },
    { label: '5m context', value: formatRecentContextMetric(record.data.recentContext.fiveMinutes) },
  ]
  let note: string | undefined
  if (record.error) {
    reasons.push('Bucket view is using the last-known-good payload after refresh failure.')
    note = 'Microstructure stays visible from cached data while revalidation is still failing.'
  }

  return createPanelState(
    'microstructure',
    applyAgeToTrust(availabilityToTrust(bucket.availability), record.data.asOf, Boolean(record.error), nowMs),
    'Closed-window tape quality',
    `Latest ${bucket.bucket.window.family} bucket closed ${formatReadableTime(bucket.bucket.window.end)} with ${bucket.bucket.window.missingBucketCount}/${bucket.bucket.window.expectedBucketCount} missing windows.`,
    metrics,
    dedupeStrings(reasons),
    note,
  )
}

function buildDerivativesPanel(focusedSymbol: DashboardSymbol): DashboardFocusedPanel {
  return createPanelState(
    'derivatives',
    'unavailable',
    `${focusedSymbol} derivatives unavailable`,
    'Derivatives context is not present in the current-state contract yet, so the shell keeps the slot explicit instead of inventing client logic.',
    [
      { label: 'Focused symbol', value: focusedSymbol },
      { label: 'Contract seam', value: 'Current-state derivatives context not yet exposed' },
    ],
    [],
  )
}

function buildHealthPanel(
  record: DashboardSurfaceRecord<DashboardGlobalStateContract>,
  summary: DashboardGlobalStateSummaryContract | undefined,
  nowMs: number,
  focusedData: DashboardSymbolStateContract | undefined,
): DashboardFocusedPanel {
  if (!record.data) {
    return createPanelState(
      'health',
      record.pending ? 'loading' : 'unavailable',
      record.pending ? 'Health and regime pending' : 'Health and regime unavailable',
      record.pending
        ? 'Feed-health and global ceiling context are waiting on the global current-state surface.'
        : 'Feed-health and global ceiling context are unavailable, so summary trust is reduced.',
      [],
      record.error ? [record.error] : [],
    )
  }

  const reasons = dedupeStrings([
    ...record.data.global.reasons,
    ...(summary?.reasonCodes ?? []),
    ...(focusedData?.composite.reasonCodes ?? []),
  ]).map(humanizeCode)
  const metrics: DashboardPanelMetric[] = [
    { label: 'Global ceiling', value: record.data.global.state },
    { label: 'Focused availability', value: summary?.availability ?? 'available' },
    { label: 'Focused effective state', value: summary?.effectiveState ?? focusedData?.regime.effectiveState ?? 'Awaiting symbol state' },
    { label: 'Global as of', value: formatReadableTime(record.data.asOf) },
  ]
  let note: string | undefined
  if (record.error) {
    reasons.push('Last-known-good health context is retained after a failed refresh.')
    note = 'Feed-health context is using the last successful payload until refresh recovers.'
  }

  return createPanelState(
    'health',
    applyAgeToTrust(
      availabilityToTrust(summary?.availability ?? 'available'),
      record.data.asOf,
      Boolean(record.error),
      nowMs,
    ),
    'Feed health and global ceiling',
    `Global ceiling is ${record.data.global.state}; focused-symbol trust reads from the same service-owned health surface.`,
    metrics,
    reasons,
    note,
  )
}

function buildGlobalTrust(
  record: DashboardSurfaceRecord<DashboardGlobalStateContract>,
  focusedSymbol: DashboardSymbol,
  nowMs: number,
): { trustState: DashboardTrustState } {
  if (!record.data) {
    return { trustState: record.pending ? 'loading' : 'unavailable' }
  }

  const summary = record.data.symbols.find((entry) => entry.symbol === focusedSymbol)
  return {
    trustState: applyAgeToTrust(
      availabilityToTrust(summary?.availability ?? 'available'),
      record.data.asOf,
      Boolean(record.error),
      nowMs,
    ),
  }
}

function deriveOverallTrust(args: {
  globalTrust: { trustState: DashboardTrustState }
  summaries: Record<DashboardSymbol, DashboardSummary>
  sections: Record<DashboardSectionKey, DashboardSectionState>
  hasRenderableShell: boolean
}): DashboardTrustState {
  if (!args.hasRenderableShell) {
    const pending = DASHBOARD_SYMBOLS.some((symbol) => args.summaries[symbol].trustState === 'loading')
    return pending ? 'loading' : 'unavailable'
  }

  const requiredTrust = combineTrusts([
    normalizeForRenderableShell(args.globalTrust.trustState),
    ...DASHBOARD_SYMBOLS.map((symbol) => normalizeForRenderableShell(args.summaries[symbol].trustState)),
    normalizeForRenderableShell(args.sections.overview.status),
    normalizeForRenderableShell(args.sections.microstructure.status),
    normalizeForRenderableShell(args.sections.health.status),
  ])
  const derivativesPenalty = args.sections.derivatives.status === 'ready' ? 'ready' : 'degraded'

  return combineTrusts([requiredTrust, derivativesPenalty])
}

function buildDashboardNotes(
  state: DashboardDataState,
  summaries: Record<DashboardSymbol, DashboardSummary>,
  sections: Record<DashboardSectionKey, DashboardSectionState>,
  focusedSymbol: DashboardSymbol,
  globalTrust: DashboardTrustState,
): string[] {
  const notes = dedupeStrings([
    ...dashboardStatusNotesFromSummary('BTC-USD', summaries['BTC-USD']),
    ...dashboardStatusNotesFromSummary('ETH-USD', summaries['ETH-USD']),
    ...dashboardStatusNotesFromSection('derivatives', sections.derivatives),
    ...dashboardStatusNotesFromSection('health', sections.health),
  ])

  if (globalTrust === 'stale') {
    notes.push('Global ceiling surface is stale; operator confidence is reduced until refresh succeeds.')
  }
  if (globalTrust === 'unavailable') {
    notes.push('Global ceiling surface is unavailable; summary trust stays visible but should not be treated as complete.')
  }
  if (state.global.error && !notes.includes(state.global.error)) {
    notes.push(state.global.error)
  }
  if (state.symbols[focusedSymbol].error && !notes.includes(state.symbols[focusedSymbol].error)) {
    notes.push(state.symbols[focusedSymbol].error)
  }

  return notes
}

function dashboardStatusNotesFromSummary(symbol: DashboardSymbol, summary: DashboardSummary): string[] {
  if (summary.trustState === 'stale') {
    return [`${symbol} is stale; showing last-known-good current state.`]
  }
  if (summary.trustState === 'unavailable') {
    return [`${symbol} current state is unavailable.`]
  }
  return []
}

function dashboardStatusNotesFromSection(section: DashboardSectionKey, state: DashboardSectionState): string[] {
  if (state.status === 'loading' || state.status === 'ready') {
    return []
  }

  return [`${SECTION_TITLES[section]} is ${state.status}. ${state.note}`]
}

function createSectionState(
  key: DashboardSectionKey,
  status: DashboardTrustState,
  note: string,
  reasons: string[],
): DashboardSectionState {
  return {
    key,
    title: SECTION_TITLES[key],
    status,
    note,
    reasons,
  }
}

function createPanelState(
  key: DashboardSectionKey,
  trustState: DashboardTrustState,
  eyebrow: string,
  summary: string,
  metrics: DashboardPanelMetric[],
  reasons: string[],
  note?: string,
): DashboardFocusedPanel {
  return {
    key,
    title: SECTION_TITLES[key],
    eyebrow,
    trustState,
    summary,
    metrics,
    reasons,
    note,
  }
}

function selectMicrostructureBucket(symbol: DashboardSymbolStateContract): DashboardBucketSectionContract {
  if (symbol.buckets.thirtySeconds.availability !== 'unavailable') {
    return symbol.buckets.thirtySeconds
  }
  if (symbol.buckets.twoMinutes.availability !== 'unavailable') {
    return symbol.buckets.twoMinutes
  }
  return symbol.buckets.fiveMinutes
}

function availabilityToTrust(value: DashboardAvailability): DashboardTrustState {
  switch (value) {
    case 'available':
      return 'ready'
    case 'degraded':
    case 'partial':
      return 'degraded'
    case 'unavailable':
      return 'unavailable'
  }
}

function applyAgeToTrust(
  baseTrust: DashboardTrustState,
  asOf: string,
  hasRefreshError: boolean,
  nowMs: number,
): DashboardTrustState {
  if (baseTrust === 'loading' || baseTrust === 'unavailable') {
    return baseTrust
  }

  const ageMs = parseAgeMs(asOf, nowMs)
  if (ageMs === undefined) {
    return baseTrust
  }
  if (ageMs >= DASHBOARD_SEVERE_STALE_AFTER_MS) {
    return 'unavailable'
  }
  if (hasRefreshError || ageMs >= DASHBOARD_STALE_AFTER_MS) {
    return 'stale'
  }
  return baseTrust
}

function buildComparisonLabel(symbol: DashboardSymbolStateContract): string {
  const worldPrice = symbol.composite.world.compositePrice
  const usaPrice = symbol.composite.usa.compositePrice

  if (typeof worldPrice === 'number' && typeof usaPrice === 'number' && worldPrice !== 0) {
    const diffBps = Math.abs(((usaPrice - worldPrice) / worldPrice) * 10_000)
    if (diffBps < 0.5) {
      return `WORLD and USA aligned inside ${formatBps(diffBps)}bp`
    }

    return usaPrice > worldPrice
      ? `USA leads WORLD by ${formatBps(diffBps)}bp`
      : `WORLD stronger than USA by ${formatBps(diffBps)}bp`
  }

  if (symbol.composite.availability === 'unavailable') {
    return 'WORLD/USA comparison unavailable'
  }
  if (symbol.composite.availability !== 'available') {
    return `WORLD/USA comparison carries ${symbol.composite.availability} inputs`
  }
  return 'WORLD/USA comparison pending'
}

function buildFreshnessLabel(asOf: string, nowMs: number): string {
  const ageMs = parseAgeMs(asOf, nowMs)
  if (ageMs === undefined) {
    return 'Awaiting service freshness metadata'
  }
  return `${formatAgeLabel(ageMs)} old`
}

function buildOldestAgeLabel(timestamps: string[], nowMs: number): string {
  const ages = timestamps
    .map((timestamp) => parseAgeMs(timestamp, nowMs))
    .filter((value): value is number => value !== undefined)

  if (ages.length === 0) {
    return 'Awaiting first successful current-state read'
  }

  return `${formatAgeLabel(Math.max(...ages))} oldest critical panel age`
}

function buildInitialUnavailableReason(state: DashboardDataState): string {
  const messages = dedupeStrings([
    state.global.error ?? 'Global current-state read did not complete.',
    ...DASHBOARD_SYMBOLS.map((symbol) => state.symbols[symbol].error ?? `${symbol} current-state read did not complete.`),
  ])

  return messages.join(' ')
}

function collectTimestamps(state: DashboardDataState): string[] {
  return [
    state.global.data?.asOf,
    ...DASHBOARD_SYMBOLS.map((symbol) => state.symbols[symbol].data?.asOf),
  ].filter((value): value is string => Boolean(value))
}

function hasPendingWork(state: DashboardDataState): boolean {
  return state.global.pending || DASHBOARD_SYMBOLS.some((symbol) => state.symbols[symbol].pending)
}

function newestIsoTimestamp(timestamps: string[]): string | undefined {
  return timestamps
    .map((timestamp) => ({ timestamp, ms: Date.parse(timestamp) }))
    .filter((entry) => !Number.isNaN(entry.ms))
    .sort((left, right) => right.ms - left.ms)[0]?.timestamp
}

function newestTimestamp(values: Array<number | undefined>): number | undefined {
  return values.filter((value): value is number => typeof value === 'number').sort((left, right) => right - left)[0]
}

function parseAgeMs(timestamp: string, nowMs: number): number | undefined {
  const parsed = Date.parse(timestamp)
  if (Number.isNaN(parsed)) {
    return undefined
  }
  return Math.max(0, nowMs - parsed)
}

function combineTrusts(states: DashboardTrustState[]): DashboardTrustState {
  return states.reduce((current, next) => {
    return TRUST_RANK[next] > TRUST_RANK[current] ? next : current
  }, 'ready')
}

function normalizeForRenderableShell(state: DashboardTrustState): DashboardTrustState {
  if (state === 'loading') {
    return 'ready'
  }
  if (state === 'unavailable') {
    return 'degraded'
  }
  return state
}

function hasTimestampTrustReduction(reasons: string[]): boolean {
  return reasons.includes('timestamp-fallback') || reasons.includes('timestamp-trust-loss')
}

function formatBps(value: number): string {
  return value >= 10 ? value.toFixed(0) : value.toFixed(1)
}

function formatAgeLabel(ageMs: number): string {
  const totalSeconds = Math.round(ageMs / 1000)
  if (totalSeconds < 60) {
    return `${totalSeconds}s`
  }

  const minutes = Math.floor(totalSeconds / 60)
  const seconds = totalSeconds % 60
  if (minutes < 60) {
    return seconds === 0 ? `${minutes}m` : `${minutes}m ${seconds}s`
  }

  const hours = Math.floor(minutes / 60)
  const remainingMinutes = minutes % 60
  return remainingMinutes === 0 ? `${hours}h` : `${hours}h ${remainingMinutes}m`
}

function formatReadableTime(timestamp: string): string {
  return timestamp.replace('T', ' ').replace('Z', ' UTC')
}

function formatCompositeMetric(side: DashboardSymbolStateContract['composite']['world']): string {
  const parts: string[] = []

  if (typeof side.compositePrice === 'number') {
    parts.push(side.compositePrice.toLocaleString('en-US', { maximumFractionDigits: 2 }))
  } else if (side.unavailable) {
    parts.push('Unavailable')
  } else {
    parts.push('Price pending')
  }

  if (typeof side.coverageRatio === 'number') {
    parts.push(`${Math.round(side.coverageRatio * 100)}% coverage`)
  }
  if (typeof side.healthScore === 'number') {
    parts.push(`${Math.round(side.healthScore * 100)} health`)
  }
  if (side.degraded) {
    parts.push('Degraded')
  }

  return parts.join(' | ')
}

function formatRecentContextMetric(context: DashboardSymbolStateContract['recentContext']['thirtySeconds']): string {
  const lead = context.complete ? 'Complete' : 'Partial'
  return `${lead} (${context.missingBucketCount} missing)`
}

function humanizeCode(code: string): string {
  return code
    .split('-')
    .map((part) => (part.length > 0 ? `${part[0].toUpperCase()}${part.slice(1)}` : part))
    .join(' ')
}

function dedupeStrings(values: string[]): string[] {
  return Array.from(new Set(values.filter((value) => value.length > 0)))
}

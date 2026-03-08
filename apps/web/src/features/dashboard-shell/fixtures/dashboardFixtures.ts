import type { DashboardFixture } from '../model/dashboardShellModel'

const baseSections = {
  overview: {
    key: 'overview',
    title: 'Overview',
    status: 'ready',
    note: 'Shell reserved for summary context, current market state, and upcoming operator notes.',
    reasons: [],
  },
  microstructure: {
    key: 'microstructure',
    title: 'Microstructure',
    status: 'ready',
    note: 'Awaiting order-book texture and short-window imbalance views from later child features.',
    reasons: [],
  },
  derivatives: {
    key: 'derivatives',
    title: 'Derivatives Context',
    status: 'unavailable',
    note: 'Current-state query wiring keeps this slot explicit until a service-owned derivatives surface exists.',
    reasons: [],
  },
  health: {
    key: 'health',
    title: 'Feed Health And Regime',
    status: 'ready',
    note: 'Reserved for venue trust, freshness, and regime detail without moving the shell layout.',
    reasons: [],
  },
} satisfies DashboardFixture['sections']

const baseFocusedPanels = {
  overview: {
    key: 'overview',
    title: 'Overview',
    eyebrow: 'Focused symbol regime',
    trustState: 'ready',
    summary: 'BTC-USD is TRADEABLE. Symbol state TRADEABLE; global ceiling WATCH.',
    metrics: [
      { label: 'Effective state', value: 'TRADEABLE' },
      { label: 'Symbol state', value: 'TRADEABLE' },
      { label: 'Global ceiling', value: 'WATCH' },
      { label: 'WORLD composite', value: '64,000 | 100% coverage | 99 health' },
      { label: 'USA composite', value: '63,996 | 98% coverage | 97 health' },
    ],
    reasons: [],
  },
  microstructure: {
    key: 'microstructure',
    title: 'Microstructure',
    eyebrow: 'Closed-window tape quality',
    trustState: 'ready',
    summary: 'Latest 30s bucket closed 2026-01-15 14:31:54 UTC with 0/1 missing windows.',
    metrics: [
      { label: 'Trusted bucket', value: '30s' },
      { label: 'Bucket close', value: '2026-01-15 14:31:54 UTC' },
      { label: 'Missing windows', value: '0/1' },
      { label: '30s context', value: 'Complete (0 missing)' },
      { label: '2m context', value: 'Complete (0 missing)' },
      { label: '5m context', value: 'Complete (0 missing)' },
    ],
    reasons: [],
  },
  derivatives: {
    key: 'derivatives',
    title: 'Derivatives Context',
    eyebrow: 'BTC-USD derivatives unavailable',
    trustState: 'unavailable',
    summary: 'Derivatives context is not present in the current-state contract yet, so the shell keeps the slot explicit instead of inventing client logic.',
    metrics: [
      { label: 'Focused symbol', value: 'BTC-USD' },
      { label: 'Contract seam', value: 'Current-state derivatives context not yet exposed' },
    ],
    reasons: [],
    warning: {
      tone: 'unavailable',
      label: 'Derivatives Context unavailable',
      detail: 'Derivatives context is not present in the current-state contract yet, so this slot stays explicit.',
    },
  },
  health: {
    key: 'health',
    title: 'Feed Health And Regime',
    eyebrow: 'Feed health and global ceiling',
    trustState: 'ready',
    summary: 'Global ceiling is WATCH; focused-symbol trust reads from the same service-owned health surface.',
    metrics: [
      { label: 'Global ceiling', value: 'WATCH' },
      { label: 'Focused availability', value: 'available' },
      { label: 'Focused effective state', value: 'TRADEABLE' },
      { label: 'Global as of', value: '2026-01-15 14:32:10 UTC' },
    ],
    reasons: [],
  },
} satisfies DashboardFixture['focusedPanels']

export const healthyDashboardFixture: DashboardFixture = {
  asOf: '2026-01-15T14:32:10Z',
  oldestAgeLabel: '22s oldest critical panel age',
  configVersion: 'visibility.v1.3.2',
  algorithmVersion: 'symbol-global-regime.v1',
  trustState: 'degraded',
  primaryWarning: {
    tone: 'degraded',
    label: 'Derivatives Context unavailable',
    detail: 'The dashboard keeps the missing derivatives contract seam visible instead of inventing client-side logic.',
  },
  degradedNotes: ['Derivatives Context is unavailable. Current-state query wiring keeps this slot explicit until a service-owned derivatives surface exists.'],
  isRefreshing: false,
  summaries: {
    'BTC-USD': {
      symbol: 'BTC-USD',
      stateLabel: 'TRADEABLE',
      trustState: 'ready',
      reasons: ['WORLD breadth leads', 'USA confirms tape', 'Feed quality stable'],
      lastUpdated: '2026-01-15T14:31:54Z',
      freshnessLabel: '16s old',
      comparisonLabel: 'WORLD and USA aligned inside 4bp',
    },
    'ETH-USD': {
      symbol: 'ETH-USD',
      stateLabel: 'WATCH',
      trustState: 'ready',
      reasons: ['WORLD remains constructive', 'USA liquidity lags', 'Regime waiting on follow-through'],
      lastUpdated: '2026-01-15T14:31:50Z',
      freshnessLabel: '20s old',
      comparisonLabel: 'WORLD stronger than USA by 11bp',
    },
  },
  focusedPanels: baseFocusedPanels,
  sections: baseSections,
}

export const degradedDashboardFixture: DashboardFixture = {
  ...healthyDashboardFixture,
  trustState: 'degraded',
  primaryWarning: {
    tone: 'degraded',
    label: 'ETH-USD trust reduced',
    detail: 'Timestamp trust is reduced in the service-owned payload, so the latest symbol read needs caution.',
  },
  degradedNotes: ['Coinbase freshness degraded; timestamp fallback is active for one USA confirmation stream.'],
  summaries: {
    ...healthyDashboardFixture.summaries,
    'ETH-USD': {
      ...healthyDashboardFixture.summaries['ETH-USD'],
      trustState: 'degraded',
      freshnessLabel: '47s old',
      reasons: ['USA confirmation degraded', 'Fallback recvTs in use', 'Keep ETH informational until freshness recovers'],
      timestampNote: 'Last state uses recvTs fallback after exchangeTs plausibility failure.',
      warning: {
        tone: 'degraded',
        label: 'ETH-USD trust reduced',
        detail: 'Timestamp trust is reduced in the service-owned payload, so the latest symbol read needs caution.',
      },
    },
  },
  focusedPanels: {
    ...baseFocusedPanels,
    overview: {
      ...baseFocusedPanels.overview,
      trustState: 'degraded',
      summary: 'ETH-USD is WATCH. Symbol state WATCH; global ceiling WATCH.',
      metrics: [
        { label: 'Effective state', value: 'WATCH' },
        { label: 'Symbol state', value: 'WATCH' },
        { label: 'Global ceiling', value: 'WATCH' },
        { label: 'WORLD composite', value: '3,242 | 98% coverage | 94 health' },
        { label: 'USA composite', value: '3,236 | 72% coverage | 64 health | Degraded' },
      ],
      reasons: ['Timestamp Trust Loss', 'Feed Health Degraded'],
      note: 'Showing the last-known-good overview while degraded trust cues stay visible.',
      warning: {
        tone: 'degraded',
        label: 'Overview degraded',
        detail: 'Showing the last-known-good overview while degraded trust cues stay visible.',
      },
    },
    derivatives: {
      ...baseFocusedPanels.derivatives,
      eyebrow: 'ETH-USD derivatives unavailable',
      metrics: [
        { label: 'Focused symbol', value: 'ETH-USD' },
        { label: 'Contract seam', value: 'Current-state derivatives context not yet exposed' },
      ],
    },
    health: {
      ...baseFocusedPanels.health,
      trustState: 'degraded',
      metrics: [
        { label: 'Global ceiling', value: 'WATCH' },
        { label: 'Focused availability', value: 'degraded' },
        { label: 'Focused effective state', value: 'WATCH' },
        { label: 'Global as of', value: '2026-01-15 14:32:10 UTC' },
      ],
      reasons: ['Global Shared Watch', 'Timestamp Trust Loss'],
      warning: {
        tone: 'degraded',
        label: 'Feed Health And Regime degraded',
        detail: 'Global Shared Watch',
      },
    },
  },
}

export const staleDashboardFixture: DashboardFixture = {
  ...healthyDashboardFixture,
  trustState: 'stale',
  oldestAgeLabel: '2m 14s oldest critical panel age',
  primaryWarning: {
    tone: 'stale',
    label: 'BTC-USD current state stale',
    detail: 'Showing the last-known-good symbol state after refresh failure; do not assume live confirmation.',
  },
  degradedNotes: ['Status rail is showing last-known-good shell data while one or more upstream panels are stale.'],
  lastSuccessLabel: 'Last successful refresh 2m 14s ago',
  summaries: {
    ...healthyDashboardFixture.summaries,
    'BTC-USD': {
      ...healthyDashboardFixture.summaries['BTC-USD'],
      trustState: 'stale',
      freshnessLabel: '2m 14s old',
      reasons: ['Last-known-good state preserved', 'Do not assume current tape confirmation'],
      warning: {
        tone: 'stale',
        label: 'BTC-USD current state stale',
        detail: 'Showing the last-known-good symbol state after refresh failure; do not assume live confirmation.',
      },
    },
  },
  focusedPanels: {
    ...baseFocusedPanels,
    overview: {
      ...baseFocusedPanels.overview,
      trustState: 'stale',
      note: 'Showing the last-known-good overview while the latest refresh is stale.',
      warning: {
        tone: 'stale',
        label: 'Overview stale',
        detail: 'Showing the last-known-good overview while the latest refresh is stale.',
      },
    },
    microstructure: {
      ...baseFocusedPanels.microstructure,
      trustState: 'stale',
      note: 'Microstructure stays visible from cached data while the latest refresh is stale.',
      warning: {
        tone: 'stale',
        label: 'Microstructure stale',
        detail: 'Microstructure stays visible from cached data while the latest refresh is stale.',
      },
    },
    health: {
      ...baseFocusedPanels.health,
      trustState: 'stale',
      note: 'Feed-health context is using the last successful payload until refresh recovers.',
      warning: {
        tone: 'stale',
        label: 'Feed Health And Regime stale',
        detail: 'Feed-health context is using the last successful payload until refresh recovers.',
      },
    },
  },
}

export const partialDashboardFixture: DashboardFixture = {
  ...healthyDashboardFixture,
  trustState: 'degraded',
  primaryWarning: {
    tone: 'unavailable',
    label: 'Feed Health And Regime unavailable',
    detail: 'Unavailable in this shell fixture while upstream consumer wiring is still pending.',
  },
  degradedNotes: ['One lower detail region is unavailable; shell remains live so operator focus is preserved.'],
  sections: {
    ...baseSections,
    health: {
      key: 'health',
      title: 'Feed Health And Regime',
      status: 'unavailable',
      note: 'Unavailable in this shell fixture while upstream consumer wiring is still pending.',
      reasons: [],
    },
  },
  focusedPanels: {
    ...baseFocusedPanels,
    health: {
      ...baseFocusedPanels.health,
      trustState: 'unavailable',
      summary: 'Feed-health and global ceiling context are unavailable, so summary trust is reduced.',
      metrics: [],
      note: 'Unavailable in this shell fixture while upstream consumer wiring is still pending.',
      warning: {
        tone: 'unavailable',
        label: 'Feed Health And Regime unavailable',
        detail: 'Unavailable in this shell fixture while upstream consumer wiring is still pending.',
      },
    },
  },
}

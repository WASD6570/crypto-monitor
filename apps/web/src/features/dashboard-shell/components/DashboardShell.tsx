import {
  DASHBOARD_SECTIONS,
  DASHBOARD_SYMBOLS,
  type DashboardFixture,
  type DashboardSectionKey,
} from '../model/dashboardShellModel'
import type { DashboardShellRouteState } from '../hooks/useDashboardShellRouteState'

export type { DashboardFixture } from '../model/dashboardShellModel'

type DashboardShellProps = {
  fixture: DashboardFixture
  onRetry: () => void
  routeState: DashboardShellRouteState
}

function formatReadableTime(value: string) {
  return value.replace('T', ' ').replace('Z', ' UTC')
}

function trustToneClass(trustState: DashboardFixture['trustState']) {
  return `trust-chip trust-chip--${trustState}`
}

function sectionToneClass(status: DashboardFixture['trustState']) {
  return `section-slot section-slot--${status}`
}

export function DashboardShell({ fixture, onRetry, routeState }: DashboardShellProps) {
  const focusedSummary = fixture.summaries[routeState.symbol]

  return (
    <main className="app-shell">
      <div className="dashboard-shell">
        <header className="dashboard-hero">
          <p className="eyebrow">Crypto Market Copilot</p>
          <h1 className="hero-title">Visibility Deck</h1>
          <p className="hero-copy">
            One shell for trust, symbol focus, and future detail panels. The service remains the
            source of truth; the UI only stages operator reading order.
          </p>
        </header>

        <section aria-label="Dashboard status rail" className="status-rail" data-testid="status-rail">
          <div className="status-rail__header">
            <div className="status-rail__primary">
              <p className="status-label">System trust</p>
              <div className="status-headline-row">
                <span className={trustToneClass(fixture.trustState)}>{fixture.trustState}</span>
                {fixture.asOf ? <time dateTime={fixture.asOf}>{formatReadableTime(fixture.asOf)}</time> : null}
              </div>
            </div>
            <div className="status-actions">
              {fixture.lastSuccessLabel ? <p className="inline-note">{fixture.lastSuccessLabel}</p> : null}
              <button className="dashboard-action" onClick={onRetry} type="button">
                {fixture.isRefreshing ? 'Refreshing...' : 'Retry current state'}
              </button>
            </div>
          </div>
          <dl className="status-metrics">
            <div>
              <dt>As of</dt>
              <dd>
                {fixture.asOf ? <time dateTime={fixture.asOf}>{formatReadableTime(fixture.asOf)}</time> : 'Awaiting service time'}
              </dd>
            </div>
            <div>
              <dt>Oldest critical panel</dt>
              <dd>{fixture.oldestAgeLabel}</dd>
            </div>
            <div>
              <dt>Config version</dt>
              <dd>{fixture.configVersion}</dd>
            </div>
            <div>
              <dt>Algorithm version</dt>
              <dd>{fixture.algorithmVersion}</dd>
            </div>
          </dl>
          {fixture.degradedNotes.length > 0 ? (
            <ul className="status-notes" aria-label="Global trust notes">
              {fixture.degradedNotes.map((note) => (
                <li key={note}>{note}</li>
              ))}
            </ul>
          ) : (
            <p className="status-notes status-notes--quiet">No global degraded notes in this fixture.</p>
          )}
        </section>

        <section aria-label="Symbol summary strip" className="summary-strip" data-testid="summary-strip">
          {DASHBOARD_SYMBOLS.map((symbol) => {
            const summary = fixture.summaries[symbol]
            const isActive = routeState.symbol === symbol

            return (
              <button
                key={symbol}
                aria-pressed={isActive}
                className={`summary-card ${isActive ? 'summary-card--active' : ''}`}
                data-testid={`summary-card-${symbol}`}
                onClick={() => routeState.setSymbol(symbol)}
                type="button"
              >
                <div className="summary-card__header">
                  <div>
                    <p className="summary-card__symbol">{summary.symbol}</p>
                    <p className="summary-card__state">{summary.stateLabel}</p>
                  </div>
                  <div className="summary-card__badges">
                    <span className={trustToneClass(summary.trustState)}>{summary.trustState}</span>
                    {isActive ? <span className="focus-pill">Focused</span> : null}
                  </div>
                </div>
                <ul className="reason-stack">
                  {summary.reasons.map((reason) => (
                    <li key={reason}>{reason}</li>
                  ))}
                </ul>
                <div className="summary-card__footer">
                  <p>{summary.comparisonLabel}</p>
                  <p>{summary.freshnessLabel}</p>
                </div>
                <time className="summary-card__time" dateTime={summary.lastUpdated}>
                  {formatReadableTime(summary.lastUpdated)}
                </time>
                {summary.timestampNote ? <p className="inline-note">{summary.timestampNote}</p> : null}
              </button>
            )
          })}
        </section>

        <section aria-label="Focused symbol detail shell" className="detail-shell">
          <div className="detail-shell__header">
            <div>
              <p className="eyebrow">Focused symbol</p>
              <h2 data-testid="focused-symbol-heading">{focusedSummary.symbol} current-state shell</h2>
            </div>
            <p className="detail-shell__trust-copy">
              Trust stays visible at the top of every section. This shell never recomputes market
              state; it only presents service-supplied labels and reasons.
            </p>
            <span className={trustToneClass(focusedSummary.trustState)}>{focusedSummary.trustState}</span>
          </div>

          <nav aria-label="Dashboard sections" className="section-nav">
            {DASHBOARD_SECTIONS.map((section) => (
              <button
                key={section}
                aria-pressed={routeState.section === section}
                className={`section-nav__button ${routeState.section === section ? 'section-nav__button--active' : ''}`}
                onClick={() => routeState.setSection(section)}
                type="button"
              >
                {fixture.focusedPanels[section].title}
              </button>
            ))}
          </nav>

          <div className="detail-grid">
            {DASHBOARD_SECTIONS.map((section) => (
              <DashboardFocusedPanelCard
                active={routeState.section === section}
                key={section}
                panel={fixture.focusedPanels[section]}
              />
            ))}
          </div>
        </section>
      </div>
    </main>
  )
}

type DashboardFocusedPanelCardProps = {
  active: boolean
  panel: DashboardFixture['focusedPanels'][DashboardSectionKey]
}

function DashboardFocusedPanelCard({ active, panel }: DashboardFocusedPanelCardProps) {
  return (
    <section
      aria-label={panel.title}
      className={`${sectionToneClass(panel.trustState)} ${active ? 'section-slot--active' : ''}`}
      data-testid={`section-slot-${panel.key}`}
    >
      <div className="section-slot__header">
        <div>
          <p className="eyebrow">{panel.eyebrow ?? (active ? 'Active section' : 'Focused section')}</p>
          <h3>{panel.title}</h3>
        </div>
        <span className={trustToneClass(panel.trustState)}>{panel.trustState}</span>
      </div>
      <p>{panel.summary}</p>
      {panel.metrics.length > 0 ? (
        <dl className="panel-metrics" aria-label={`${panel.title} metrics`}>
          {panel.metrics.map((metric) => (
            <div className="panel-metric" key={`${metric.label}-${metric.value}`}>
              <dt>{metric.label}</dt>
              <dd>{metric.value}</dd>
            </div>
          ))}
        </dl>
      ) : null}
      {panel.note ? <p className="inline-note">{panel.note}</p> : null}
      {panel.reasons.length > 0 ? (
        <ul className="reason-stack" aria-label={`${panel.title} reasons`}>
          {panel.reasons.map((reason) => (
            <li key={reason}>{reason}</li>
          ))}
        </ul>
      ) : null}
    </section>
  )
}

import { useId } from 'react'
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

function warningToneClass(tone: DashboardFixture['trustState']) {
  return `warning-band warning-band--${tone}`
}

export function DashboardShell({ fixture, onRetry, routeState }: DashboardShellProps) {
  const shellId = useId()
  const focusedSummary = fixture.summaries[routeState.symbol]
  const detailShellId = `${shellId}-detail-shell`
  const focusedWarningId = focusedSummary.warning ? `${shellId}-${routeState.symbol}-focused-warning` : undefined

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
          {fixture.primaryWarning ? (
            <DashboardWarningBand
              ariaLive="polite"
              className="status-rail__warning"
              dataTestId="route-warning"
              warning={fixture.primaryWarning}
            />
          ) : null}
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
            const summaryWarningId = summary.warning ? `${shellId}-${symbol}-summary-warning` : undefined
            const summaryTrustId = `${shellId}-${symbol}-summary-trust`
            const summaryFreshnessId = `${shellId}-${symbol}-summary-freshness`
            const summaryDescription = [summaryTrustId, summaryFreshnessId, summaryWarningId].filter(Boolean).join(' ')

            return (
              <button
                key={symbol}
                aria-controls={detailShellId}
                aria-current={isActive ? 'true' : undefined}
                aria-describedby={summaryDescription}
                aria-label={`${summary.symbol} ${summary.stateLabel}. Trust ${summary.trustState}.${isActive ? ' Focused symbol.' : ' Activate focused symbol.'}`}
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
                    <span className={trustToneClass(summary.trustState)} id={summaryTrustId}>{summary.trustState}</span>
                    {isActive ? <span className="focus-pill">Focused</span> : null}
                  </div>
                </div>
                {summary.warning ? (
                  <DashboardWarningBand className="summary-card__warning" id={summaryWarningId} warning={summary.warning} />
                ) : null}
                <ul className="reason-stack">
                  {summary.reasons.map((reason) => (
                    <li key={reason}>{reason}</li>
                  ))}
                </ul>
                <div className="summary-card__footer">
                  <p>{summary.comparisonLabel}</p>
                  <p id={summaryFreshnessId}>{summary.freshnessLabel}</p>
                </div>
                <time className="summary-card__time" dateTime={summary.lastUpdated}>
                  {formatReadableTime(summary.lastUpdated)}
                </time>
                {summary.timestampNote ? <p className="inline-note">{summary.timestampNote}</p> : null}
              </button>
            )
          })}
        </section>

        <section aria-label="Focused symbol detail shell" className="detail-shell" id={detailShellId}>
          <div className="detail-shell__header">
            <div>
              <p className="eyebrow">Focused symbol</p>
              <h2 aria-live="polite" data-testid="focused-symbol-heading">{focusedSummary.symbol} current-state shell</h2>
            </div>
            <p className="detail-shell__trust-copy">
              Trust stays visible at the top of every section. This shell never recomputes market
              state; it only presents service-supplied labels and reasons.
            </p>
            <span className={trustToneClass(focusedSummary.trustState)}>{focusedSummary.trustState}</span>
          </div>
          {focusedSummary.warning ? (
            <DashboardWarningBand
              ariaLive="polite"
              className="detail-shell__warning"
              dataTestId="focused-symbol-warning"
              id={focusedWarningId}
              warning={focusedSummary.warning}
            />
          ) : null}

          <nav aria-label="Dashboard sections" className="section-nav">
            {DASHBOARD_SECTIONS.map((section) => {
              const isActive = routeState.section === section
              const panel = fixture.focusedPanels[section]
              const panelId = `${shellId}-panel-${section}`
              const panelWarningId = panel.warning ? `${panelId}-warning` : undefined
              const panelTrustId = `${panelId}-trust`
              const buttonDescription = [panelTrustId, panelWarningId, focusedWarningId].filter(Boolean).join(' ')

              return (
                <button
                  key={section}
                  aria-controls={panelId}
                  aria-current={isActive ? 'true' : undefined}
                  aria-describedby={buttonDescription}
                  aria-pressed={isActive}
                  className={`section-nav__button ${isActive ? 'section-nav__button--active' : ''}`}
                  onClick={() => routeState.setSection(section)}
                  type="button"
                >
                  {panel.title}
                </button>
              )
            })}
          </nav>

          <div className="detail-grid">
            {DASHBOARD_SECTIONS.map((section) => (
              <DashboardFocusedPanelCard
                active={routeState.section === section}
                id={`${shellId}-panel-${section}`}
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
  id: string
  panel: DashboardFixture['focusedPanels'][DashboardSectionKey]
}

function DashboardFocusedPanelCard({ active, id, panel }: DashboardFocusedPanelCardProps) {
  const headingId = `${id}-heading`
  const warningId = panel.warning ? `${id}-warning` : undefined
  const noteId = panel.note ? `${id}-note` : undefined

  return (
    <section
      aria-describedby={[warningId, noteId].filter(Boolean).join(' ') || undefined}
      aria-labelledby={headingId}
      className={`${sectionToneClass(panel.trustState)} ${active ? 'section-slot--active' : ''}`}
      data-testid={`section-slot-${panel.key}`}
      id={id}
      role="region"
    >
      <div className="section-slot__header">
        <div>
          <p className="eyebrow">{panel.eyebrow ?? (active ? 'Active section' : 'Focused section')}</p>
          <h3 id={headingId}>{panel.title}</h3>
        </div>
        <span className={trustToneClass(panel.trustState)} id={`${id}-trust`}>{panel.trustState}</span>
      </div>
      {panel.warning ? <DashboardWarningBand className="section-slot__warning" id={warningId} warning={panel.warning} /> : null}
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
      {panel.note ? <p className="inline-note" id={noteId}>{panel.note}</p> : null}
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

type DashboardWarningBandProps = {
  ariaLive?: 'polite' | 'assertive'
  className?: string
  dataTestId?: string
  id?: string
  warning: NonNullable<DashboardFixture['primaryWarning']>
}

function DashboardWarningBand({ ariaLive, className, dataTestId, id, warning }: DashboardWarningBandProps) {
  return (
    <div
      aria-live={ariaLive}
      className={`${warningToneClass(warning.tone)} ${className ?? ''}`.trim()}
      data-testid={dataTestId}
      id={id}
      role="note"
    >
      <p className="warning-band__label">{warning.label}</p>
      <p className="warning-band__detail">{warning.detail}</p>
    </div>
  )
}

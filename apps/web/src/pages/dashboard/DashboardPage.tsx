import {
  DashboardShell,
  type DashboardFixture,
} from '../../features/dashboard-shell/components/DashboardShell'
import { useDashboardShellRouteState } from '../../features/dashboard-shell/hooks/useDashboardShellRouteState'
import { type DashboardClient } from '../../api/dashboard/dashboardClient'
import { type DashboardClock } from '../../features/dashboard-state/dashboardQueryState'
import { useDashboardData } from '../../features/dashboard-state/useDashboardData'

type DashboardPageProps = {
  fixture?: DashboardFixture
  client?: DashboardClient
  clock?: DashboardClock
}

export function DashboardPage({ fixture, client, clock }: DashboardPageProps) {
  const routeState = useDashboardShellRouteState()
  const dashboardData = useDashboardData({
    client,
    clock,
    enabled: fixture === undefined,
    focusedSymbol: routeState.symbol,
  })

  if (fixture) {
    return <DashboardShell fixture={fixture} onRetry={() => undefined} routeState={routeState} />
  }

  if (dashboardData.isInitialUnavailable) {
    return (
      <main className="app-shell app-shell--not-found">
        <div className="empty-state-panel">
          <p className="eyebrow">Current State Unavailable</p>
          <h1>Dashboard snapshot is not readable yet.</h1>
          <p>
            {dashboardData.initialUnavailableReason ??
              'No current-state surface returned a readable payload yet.'}
          </p>
          <button className="dashboard-action dashboard-action--primary" onClick={dashboardData.retry} type="button">
            Retry current state
          </button>
        </div>
      </main>
    )
  }

  return <DashboardShell fixture={dashboardData.viewModel} onRetry={dashboardData.retry} routeState={routeState} />
}

import { useEffect, useMemo, useState } from 'react'
import { DashboardPage } from '../pages/dashboard/DashboardPage'

const DASHBOARD_PATH = '/dashboard'

function normalizePath(pathname: string) {
  if (pathname === '/') {
    return DASHBOARD_PATH
  }

  return pathname
}

export function App() {
  const [pathname, setPathname] = useState(() => normalizePath(window.location.pathname))

  useEffect(() => {
    if (window.location.pathname === '/') {
      window.history.replaceState({}, '', `${DASHBOARD_PATH}${window.location.search}`)
      setPathname(DASHBOARD_PATH)
    }
  }, [])

  const content = useMemo(() => {
    if (pathname === DASHBOARD_PATH) {
      return <DashboardPage />
    }

    return (
      <main className="app-shell app-shell--not-found">
        <div className="empty-state-panel">
          <p className="eyebrow">Unavailable Route</p>
          <h1>Dashboard shell lives at `/dashboard`.</h1>
          <p>
            The web surface is intentionally narrow for now so later state and health panels can
            attach without changing the route shape.
          </p>
          <a className="summary-card summary-card--link" href={DASHBOARD_PATH}>
            Open dashboard shell
          </a>
        </div>
      </main>
    )
  }, [pathname])

  return <>{content}</>
}

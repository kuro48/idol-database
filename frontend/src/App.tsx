import { lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Skeleton } from './components/ui/Skeleton'
import DataShell from './layout/DataShell'
import AdminShell from './layout/AdminShell'

const LoginPage = lazy(() => import('./auth/LoginPage'))
const CallbackPage = lazy(() => import('./pages/auth/CallbackPage'))
const IdolListPage = lazy(() => import('./pages/idols/IdolListPage'))
const GroupListPage = lazy(() => import('./pages/groups/GroupListPage'))
const AgencyListPage = lazy(() => import('./pages/agencies/AgencyListPage'))
const EventListPage = lazy(() => import('./pages/events/EventListPage'))
const TagListPage = lazy(() => import('./pages/tags/TagListPage'))
const ReleaseListPage = lazy(() => import('./pages/releases/ReleaseListPage'))
const DashboardPage = lazy(() => import('./pages/dashboard/DashboardPage'))
const ApiKeysPage = lazy(() => import('./pages/dashboard/ApiKeysPage'))
const AnalyticsPage = lazy(() => import('./pages/dashboard/AnalyticsPage'))
const OshiColorPage = lazy(() => import('./pages/settings/OshiColorPage'))

function PageFallback() {
  return (
    <div style={{ padding: 'var(--space-8)', display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
      <Skeleton height="2rem" width="200px" />
      <Skeleton height="1rem" />
      <Skeleton height="1rem" width="80%" />
      <Skeleton height="1rem" width="60%" />
    </div>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <Suspense fallback={<PageFallback />}>
        <Routes>
          {/* Login — no shell */}
          <Route path="/login" element={<LoginPage />} />

          {/* OIDC callback — no shell, no auth required */}
          <Route path="/callback" element={<CallbackPage />} />

          {/* Public data browser */}
          <Route element={<DataShell />}>
            <Route index element={<Navigate to="/idols" replace />} />
            <Route path="/idols" element={<IdolListPage />} />
            <Route path="/groups" element={<GroupListPage />} />
            <Route path="/agencies" element={<AgencyListPage />} />
            <Route path="/events" element={<EventListPage />} />
            <Route path="/tags" element={<TagListPage />} />
            <Route path="/releases" element={<ReleaseListPage />} />
            <Route path="/settings/oshi-color" element={<OshiColorPage />} />
          </Route>

          {/* Admin dashboard */}
          <Route path="/dashboard" element={<AdminShell />}>
            <Route index element={<DashboardPage />} />
            <Route path="apikeys" element={<ApiKeysPage />} />
            <Route path="analytics" element={<AnalyticsPage />} />
          </Route>

          {/* Fallback */}
          <Route path="*" element={<Navigate to="/idols" replace />} />
        </Routes>
      </Suspense>
    </BrowserRouter>
  )
}

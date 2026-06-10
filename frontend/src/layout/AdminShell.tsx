import { useEffect } from 'react'
import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { LayoutDashboard, KeyRound, BarChart2, ClipboardCheck, LogOut } from 'lucide-react'
import { useAuth } from '../auth/useAuth'
import { userManager } from '../auth/oidcClient'
import { applyAdminTheme } from '../lib/applyTheme'
import styles from './admin-shell.module.css'

const ADMIN_NAV = [
  { to: '/dashboard', label: 'ダッシュボード', icon: LayoutDashboard },
  { to: '/dashboard/apikeys', label: 'APIキー', icon: KeyRound },
  { to: '/dashboard/analytics', label: '分析', icon: BarChart2 },
  { to: '/dashboard/review', label: '審査', icon: ClipboardCheck },
]

export default function AdminShell() {
  const { isAdmin, idToken, logout } = useAuth()
  const navigate = useNavigate()

  useEffect(() => {
    applyAdminTheme()
  }, [])

  if (!isAdmin) {
    return (
      <div className={styles.denied}>
        <p>管理者権限が必要です。</p>
        <button onClick={() => navigate('/login')} type="button">
          サインイン
        </button>
      </div>
    )
  }

  async function handleLogout() {
    logout()
    try {
      await userManager.signoutRedirect(idToken)
    } catch {
      navigate('/login', { replace: true })
    }
  }

  return (
    <div className={styles.shell}>
      <nav className={styles.sidebar} aria-label="Admin navigation">
        <div className={styles.sidebarHeader}>
          <span className={styles.siteName}>idol.db</span>
          <span className={styles.adminBadge}>admin</span>
        </div>

        <ul className={styles.navList}>
          {ADMIN_NAV.map(({ to, label, icon: Icon }) => (
            <li key={to}>
              <NavLink
                to={to}
                end={to === '/dashboard'}
                className={({ isActive }) =>
                  `${styles.navLink} ${isActive ? styles.navLinkActive : ''}`
                }
              >
                <Icon size={16} aria-hidden="true" />
                <span>{label}</span>
              </NavLink>
            </li>
          ))}
        </ul>

        <div className={styles.sidebarFooter}>
          <button
            onClick={handleLogout}
            className={styles.logoutButton}
            type="button"
          >
            <LogOut size={16} aria-hidden="true" />
            <span>サインアウト</span>
          </button>
        </div>
      </nav>

      <div className={styles.main}>
        <header className={styles.topbar}>
          <span className={styles.topbarTitle}>管理コンソール</span>
        </header>
        <main className={styles.content}>
          <Outlet />
        </main>
      </div>
    </div>
  )
}

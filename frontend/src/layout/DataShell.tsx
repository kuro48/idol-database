import { useEffect } from 'react'
import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import {
  Users,
  Music,
  Building2,
  Calendar,
  Tag,
  Disc3,
  Settings,
  LayoutDashboard,
  LogOut,
  LogIn,
} from 'lucide-react'
import { useAuth } from '../auth/useAuth'
import { userManager } from '../auth/oidcClient'
import { applyOshiTheme } from '../lib/applyTheme'
import styles from './data-shell.module.css'

const NAV_ITEMS = [
  { to: '/idols', label: 'アイドル', icon: Users },
  { to: '/groups', label: 'グループ', icon: Music },
  { to: '/agencies', label: '事務所', icon: Building2 },
  { to: '/events', label: 'イベント', icon: Calendar },
  { to: '/tags', label: 'タグ', icon: Tag },
  { to: '/releases', label: 'リリース', icon: Disc3 },
]

export default function DataShell() {
  const { oshiColor, isLoggedIn, isAdmin, displayName, email, logout } = useAuth()
  const navigate = useNavigate()

  useEffect(() => {
    applyOshiTheme(oshiColor)
  }, [oshiColor])

  async function handleLogout() {
    logout()
    try {
      await userManager.signoutRedirect()
    } catch {
      // If the auth server logout fails (network, config), still bring the user
      // back to a safe public page. Local session is already cleared.
      navigate('/', { replace: true })
    }
  }

  return (
    <div className={styles.shell}>
      <nav className={styles.sidebar} aria-label="Main navigation">
        <div className={styles.sidebarHeader}>
          <span className={styles.siteName}>idol.db</span>
        </div>

        <ul className={styles.navList}>
          {NAV_ITEMS.map(({ to, label, icon: Icon }) => (
            <li key={to}>
              <NavLink
                to={to}
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
          {isLoggedIn && (
            <NavLink
              to="/settings/oshi-color"
              className={({ isActive }) =>
                `${styles.navLink} ${isActive ? styles.navLinkActive : ''}`
              }
            >
              <Settings size={16} aria-hidden="true" />
              <span>設定</span>
            </NavLink>
          )}
          {isAdmin && (
            <NavLink
              to="/dashboard"
              className={({ isActive }) =>
                `${styles.navLink} ${isActive ? styles.navLinkActive : ''}`
              }
            >
              <LayoutDashboard size={16} aria-hidden="true" />
              <span>管理者</span>
            </NavLink>
          )}
        </div>
      </nav>

      <div className={styles.main}>
        <header className={styles.topbar}>
          <div className={styles.topbarLeft} />
          <div className={styles.topbarRight}>
            {isLoggedIn ? (
              <>
                {(displayName ?? email) && (
                  <span className={styles.userLabel}>{displayName ?? email}</span>
                )}
                <button
                  onClick={handleLogout}
                  className={styles.authButton}
                  type="button"
                >
                  <LogOut size={15} aria-hidden="true" />
                  <span>サインアウト</span>
                </button>
              </>
            ) : (
              <NavLink to="/login" className={styles.authButton}>
                <LogIn size={15} aria-hidden="true" />
                <span>サインイン</span>
              </NavLink>
            )}
          </div>
        </header>

        <main className={styles.content}>
          <Outlet />
        </main>
      </div>
    </div>
  )
}


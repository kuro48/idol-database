const ADMIN_ACCENT = '#0017C1'
const DEFAULT_ACCENT = '#FF69B4'

export function applyOshiTheme(color: string | null): void {
  const accent = color || DEFAULT_ACCENT
  document.documentElement.style.setProperty('--color-accent', accent)
}

export function applyAdminTheme(): void {
  document.documentElement.style.setProperty('--color-accent', ADMIN_ACCENT)
}

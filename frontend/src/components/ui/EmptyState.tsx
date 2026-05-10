interface EmptyStateProps {
  icon?: React.ReactNode
  title: string
  description?: string
  action?: React.ReactNode
}

export function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 'var(--space-12) var(--space-8)',
        gap: 'var(--space-4)',
        textAlign: 'center',
        color: 'var(--color-text-muted)',
      }}
    >
      {icon && (
        <div
          style={{
            color: 'var(--color-text-faint)',
            marginBottom: 'var(--space-2)',
          }}
        >
          {icon}
        </div>
      )}
      <h3
        style={{
          fontSize: 'var(--text-base)',
          fontWeight: 600,
          color: 'var(--color-text)',
          margin: 0,
        }}
      >
        {title}
      </h3>
      {description && (
        <p
          style={{
            fontSize: 'var(--text-sm)',
            color: 'var(--color-text-muted)',
            margin: 0,
            maxWidth: '320px',
          }}
        >
          {description}
        </p>
      )}
      {action && <div>{action}</div>}
    </div>
  )
}

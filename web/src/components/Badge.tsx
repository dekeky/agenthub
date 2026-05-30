interface BadgeProps {
  children: React.ReactNode
  variant?: 'category' | 'version'
}

export function Badge({ children, variant = 'category' }: BadgeProps) {
  return <span className={`badge badge-${variant}`}>{children}</span>
}

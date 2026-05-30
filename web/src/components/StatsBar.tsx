interface StatsBarProps {
  agents: { category: string }[]
}

export function StatsBar({ agents }: StatsBarProps) {
  const total = agents.length
  const picoclaw = agents.filter(a => a.category === 'picoclaw').length
  const openclaw = agents.filter(a => a.category === 'openclaw').length

  return (
    <div className="stats-bar">
      <div className="stat-item">
        <span className="stat-value">{total}</span>
        <span className="stat-label">全部 Agent</span>
      </div>
      <div className="stat-divider" />
      <div className="stat-item">
        <span className="stat-value">{picoclaw}</span>
        <span className="stat-label">PicoClaw</span>
      </div>
      <div className="stat-divider" />
      <div className="stat-item">
        <span className="stat-value">{openclaw}</span>
        <span className="stat-label">OpenClaw</span>
      </div>
    </div>
  )
}

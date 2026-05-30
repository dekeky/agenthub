import { Link } from 'react-router-dom'
import type { AgentMeta } from '../api/types'
import { Badge } from './Badge'
import { formatDate } from '../utils/format'

interface AgentCardProps {
  agent: AgentMeta
}

export function AgentCard({ agent }: AgentCardProps) {
  return (
    <Link to={`/agents/${encodeURIComponent(agent.agentName)}`} className="agent-card">
      <div className="agent-card-header">
        <h3 className="agent-name">{agent.displayName || agent.agentName}</h3>
        <Badge variant="category">{agent.category}</Badge>
      </div>
      <p className="agent-summary">{agent.summary || '暂无描述'}</p>
      <div className="agent-footer">
        <span className="agent-version">v{agent.latestVersion || '?'}</span>
        <span className="agent-date">{formatDate(agent.updatedAt)}</span>
      </div>
    </Link>
  )
}

import type { AgentMeta } from '../api/types'
import { AgentCard } from './AgentCard'

interface AgentGridProps {
  agents: AgentMeta[]
}

export function AgentGrid({ agents }: AgentGridProps) {
  return (
    <div className="agent-grid">
      {agents.map(agent => (
        <AgentCard key={agent.agentName} agent={agent} />
      ))}
    </div>
  )
}

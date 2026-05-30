import { useState, useMemo } from 'react'
import { useAgents } from '../hooks/useAgents'
import { AgentGrid } from '../components/AgentGrid'
import { SearchBar } from '../components/SearchBar'
import { CategoryFilter } from '../components/CategoryFilter'
import { StatsBar } from '../components/StatsBar'
import { LoadingState, EmptyState, ErrorState } from '../components/States'

const KNOWN_CATEGORIES = ['picoclaw', 'openclaw']

export function BrowsePage() {
  const [search, setSearch] = useState('')
  const [category, setCategory] = useState('')
  const { data: agents, isLoading, isError, error, refetch } = useAgents(category || undefined)

  const filteredAgents = useMemo(() => {
    if (!agents) return []
    if (!search.trim()) return agents
    const q = search.toLowerCase().trim()
    return agents.filter(
      a =>
        a.agentName.toLowerCase().includes(q) ||
        (a.displayName || '').toLowerCase().includes(q) ||
        (a.summary || '').toLowerCase().includes(q)
    )
  }, [agents, search])

  return (
    <div className="view active">
      <header className="view-header">
        <div className="header-left">
          <h1>Agent 仓库</h1>
          <p className="header-subtitle">发现并管理您的 Agent 包</p>
        </div>
        <div className="header-actions">
          <SearchBar value={search} onChange={setSearch} />
          <CategoryFilter
            value={category}
            onChange={setCategory}
            categories={KNOWN_CATEGORIES}
          />
        </div>
      </header>

      <div className="view-content">
        {isLoading && <LoadingState />}

        {isError && (
          <ErrorState message={error?.message || '无法连接到服务器'} onRetry={() => refetch()} />
        )}

        {agents && !isLoading && !isError && (
          <>
            <StatsBar agents={agents} />
            {filteredAgents.length === 0 ? (
              agents.length === 0 ? (
                <EmptyState />
              ) : (
                <div className="state-container">
                  <p style={{ color: 'var(--color-text-tertiary)' }}>未找到匹配的 Agent</p>
                </div>
              )
            ) : (
              <AgentGrid agents={filteredAgents} />
            )}
          </>
        )}
      </div>
    </div>
  )
}

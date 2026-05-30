import { useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import {
  ArrowLeft,
  Download,
  Terminal,
  Pencil,
  Trash2,
  Clock,
} from 'lucide-react'
import { useAgentDetail, useDeleteAgent } from '../hooks/useAgents'
import { Badge } from '../components/Badge'
import { AgentFileExplorer } from '../components/AgentFileExplorer'
import { LoadingState, ErrorState } from '../components/States'
import { Button } from '../components/UI'
import { formatDate } from '../utils/format'
import { getApiUrl } from '../utils/storage'
import { useToast } from '../components/Toast'

type Tab = 'versions' | 'files'

export function AgentDetailPage() {
  const { agentName } = useParams<{ agentName: string }>()
  const navigate = useNavigate()
  const { showToast } = useToast()

  const [viewVersion, setViewVersion] = useState<string | undefined>(undefined)
  const { data: agent, isLoading, isError, error } = useAgentDetail(agentName || '', viewVersion)
  const deleteMutation = useDeleteAgent()
  const [activeTab, setActiveTab] = useState<Tab>('files')
  const activeVersion = viewVersion ?? agent?.latestVersion ?? ''

  const handleDownload = () => {
    if (!agent) return
    const url = `${getApiUrl()}/api/hub/agents/${encodeURIComponent(agent.agentName)}/download`
    const a = document.createElement('a')
    a.href = url
    a.download = `${agent.agentName}-${agent.latestVersion}.zip`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    showToast('success', '开始下载')
  }

  const handleDelete = async () => {
    if (!agent || !agentName) return
    if (!window.confirm(`确定要删除 "${agent.displayName || agent.agentName}" 吗？此操作不可撤销。`)) return
    try {
      await deleteMutation.mutateAsync(agentName)
      showToast('success', 'Agent 已删除')
      navigate('/')
    } catch (e: unknown) {
      showToast('error', `删除失败: ${(e as Error).message}`)
    }
  }

  const handleCopyInstall = () => {
    if (!agent) return
    const cmd = `agenthub-cli install --dest ./agents ${agent.agentName}`
    navigator.clipboard.writeText(cmd).then(() => {
      showToast('success', '安装命令已复制')
    }).catch(() => {
      showToast('error', '复制失败')
    })
  }

  if (isLoading) return <LoadingState />
  if (isError || !agent) {
    return (
      <div className="view active">
        <div className="view-content">
          <ErrorState message={error?.message || 'Agent 未找到'} onRetry={() => navigate('/')} />
        </div>
      </div>
    )
  }

  return (
    <div className="view active">
      <header className="view-header">
        <div className="header-left">
          <div className="header-back">
            <Link to="/" className="back-link">
              <ArrowLeft size={18} />
              返回
            </Link>
          </div>
          <h1>{agent.displayName || agent.agentName}</h1>
          <div className="header-badges">
            <Badge variant="category">{agent.category}</Badge>
            <Badge variant="version">v{agent.latestVersion}</Badge>
          </div>
        </div>
        <div className="header-actions">
          <Button variant="secondary" onClick={handleCopyInstall}>
            <Terminal size={16} />
            复制安装命令
          </Button>
          <Button variant="secondary" onClick={handleDownload}>
            <Download size={16} />
            下载 ZIP
          </Button>
          <Button variant="primary" onClick={() => navigate(`/agents/${encodeURIComponent(agent.agentName)}/edit`)}>
            <Pencil size={16} />
            编辑
          </Button>
          <Button variant="danger" onClick={handleDelete} disabled={deleteMutation.isPending}>
            <Trash2 size={16} />
            {deleteMutation.isPending ? '删除中...' : '删除'}
          </Button>
        </div>
      </header>

      <div className="view-content">
        <div className="detail-summary">
          <p className="detail-description">{agent.summary || '暂无描述'}</p>
          <div className="detail-meta">
            <span><Clock size={14} /> 更新于 {formatDate(agent.updatedAt)}</span>
          </div>
        </div>

        <div className="detail-tabs">
          {(['files', 'versions'] as Tab[]).map(tab => (
            <button
              key={tab}
              className={`tab-btn ${activeTab === tab ? 'active' : ''}`}
              onClick={() => setActiveTab(tab)}
            >
              {tab === 'versions' ? '版本' : '文件'}
            </button>
          ))}
        </div>

        <div className="tab-panels">
          {activeTab === 'versions' && (
            <div className="tab-panel active">
              <div className="version-chips">
                {(agent.versions || []).map(v => (
                  <button
                    key={v}
                    type="button"
                    className={`version-chip ${activeVersion === v ? 'active' : ''}`}
                    onClick={() => setViewVersion(v)}
                  >
                    {v}
                  </button>
                ))}
              </div>
            </div>
          )}

          {activeTab === 'files' && (
            <div className="tab-panel active">
              <AgentFileExplorer
                key={`${agent.agentName}-${activeVersion}`}
                agentName={agent.agentName}
                version={activeVersion}
                files={agent.files}
              />
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

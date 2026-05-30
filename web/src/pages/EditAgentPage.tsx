import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { ArrowLeft, Save } from 'lucide-react'
import { useAgentDetail, useUpdateMeta, useUpdateFile } from '../hooks/useAgents'
import { AgentFileEditor } from '../components/AgentFileEditor'
import { LoadingState, ErrorState } from '../components/States'
import { Button } from '../components/UI'
import { useToast } from '../components/Toast'

const KNOWN_CATEGORIES = ['picoclaw', 'openclaw']

export function EditAgentPage() {
  const { agentName } = useParams<{ agentName: string }>()
  const navigate = useNavigate()
  const { showToast } = useToast()

  const { data: agent, isLoading, isError, error } = useAgentDetail(agentName || '')
  const updateMeta = useUpdateMeta(agentName || '')
  const updateFileMutation = useUpdateFile(agentName || '')

  const [displayName, setDisplayName] = useState('')
  const [summary, setSummary] = useState('')
  const [category, setCategory] = useState('')

  useEffect(() => {
    if (agent) {
      setDisplayName(agent.displayName || '')
      setSummary(agent.summary || '')
      setCategory(agent.category || 'picoclaw')
    }
  }, [agent])

  const handleSaveMeta = async () => {
    try {
      await updateMeta.mutateAsync({
        displayName: displayName || undefined,
        summary: summary || undefined,
        category: category || undefined,
      })
      showToast('success', '元信息已保存')
    } catch (e: unknown) {
      showToast('error', `保存失败: ${(e as Error).message}`)
    }
  }

  const handleSaveFile = async (filePath: string, content: string) => {
    if (!agent) return
    try {
      await updateFileMutation.mutateAsync({
        filePath,
        req: {
          version: agent.latestVersion,
          content,
        },
      })
      showToast('success', '文件已保存')
    } catch (e: unknown) {
      showToast('error', `保存失败: ${(e as Error).message}`)
    }
  }

  const hasMetaChanges =
    agent &&
    (displayName !== (agent.displayName || '') ||
      summary !== (agent.summary || '') ||
      category !== (agent.category || 'picoclaw'))

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
            <Link to={`/agents/${encodeURIComponent(agent.agentName)}`} className="back-link">
              <ArrowLeft size={18} />
              返回详情
            </Link>
          </div>
          <h1>编辑: {agent.displayName || agent.agentName}</h1>
        </div>
      </header>

      <div className="view-content">
        <div className="edit-layout">
          <div className="edit-meta-section">
            <div className="form-section">
              <h3>元信息</h3>
              <div className="form-grid">
                <div className="form-group">
                  <label htmlFor="displayName">显示名称</label>
                  <input
                    id="displayName"
                    type="text"
                    className="form-input"
                    value={displayName}
                    onChange={e => setDisplayName(e.target.value)}
                    placeholder="My Cool Agent"
                  />
                </div>
                <div className="form-group">
                  <label htmlFor="category">运行时类别</label>
                  <div className="select-wrapper">
                    <select
                      id="category"
                      className="form-select"
                      value={category}
                      onChange={e => setCategory(e.target.value)}
                    >
                      {KNOWN_CATEGORIES.map(c => (
                        <option key={c} value={c}>
                          {c.charAt(0).toUpperCase() + c.slice(1)}
                        </option>
                      ))}
                    </select>
                  </div>
                </div>
              </div>
              <div className="form-group" style={{ marginTop: '1rem' }}>
                <label htmlFor="summary">描述</label>
                <textarea
                  id="summary"
                  className="form-input form-textarea"
                  value={summary}
                  onChange={e => setSummary(e.target.value)}
                  placeholder="Agent 的功能描述..."
                  rows={4}
                />
              </div>
              <div className="form-actions">
                <Button
                  variant="primary"
                  onClick={handleSaveMeta}
                  disabled={!hasMetaChanges || updateMeta.isPending}
                >
                  <Save size={16} />
                  {updateMeta.isPending ? '保存中...' : '保存元信息'}
                </Button>
              </div>
            </div>
          </div>

          <div className="edit-file-section">
            <div className="form-section">
              <h3>文件编辑</h3>
              <AgentFileEditor
                key={agent.agentName}
                agentName={agent.agentName}
                version={agent.latestVersion}
                files={agent.files}
                onSave={handleSaveFile}
                saving={updateFileMutation.isPending}
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default EditAgentPage

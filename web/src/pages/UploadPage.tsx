import { useState, useRef, type DragEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { Upload, CloudUpload, FileArchive, X } from 'lucide-react'
import { useUpload } from '../hooks/useAgents'
import { Button } from '../components/UI'
import { useToast } from '../components/Toast'
import { formatSize } from '../utils/format'

const KNOWN_CATEGORIES = ['picoclaw', 'openclaw']

function isZipFile(file: File): boolean {
  return /\.zip$/i.test(file.name)
}

function agentNameFromZip(filename: string): string {
  const base = filename.replace(/\.zip$/i, '').toLowerCase()
  let name = base.replace(/[^a-z0-9-]/g, '-').replace(/-+/g, '-').replace(/^-+|-+$/g, '')
  if (!name || !/^[a-z0-9]/.test(name)) {
    name = `agent-${name || 'upload'}`.replace(/-+/g, '-')
  }
  return name
}

export function UploadPage() {
  const navigate = useNavigate()
  const { showToast } = useToast()
  const uploadMutation = useUpload()

  const [agentName, setAgentName] = useState('')
  const [category, setCategory] = useState('picoclaw')
  const [version, setVersion] = useState('')
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [progress, setProgress] = useState(0)
  const [uploading, setUploading] = useState(false)
  const [dragOver, setDragOver] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleFileSelect = (file: File) => {
    if (!isZipFile(file)) {
      showToast('error', '请选择 ZIP 文件')
      return
    }
    setSelectedFile(file)
    setAgentName(prev => prev.trim() || agentNameFromZip(file.name))
  }

  const handleDrop = (e: DragEvent) => {
    e.preventDefault()
    setDragOver(false)
    const file = e.dataTransfer.files[0]
    if (file) handleFileSelect(file)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!agentName.trim()) {
      showToast('error', '请输入 Agent 名称')
      return
    }
    if (!selectedFile) {
      showToast('error', '请选择要上传的 ZIP 文件')
      return
    }

    const formData = new FormData()
    formData.append('agentName', agentName.trim())
    formData.append('category', category)
    if (version.trim()) formData.append('version', version.trim())
    formData.append('file', selectedFile)

    setUploading(true)
    setProgress(0)

    try {
      await uploadMutation.mutateAsync({
        formData,
        onProgress: (percent) => setProgress(percent),
      })
      showToast('success', `${agentName} 上传成功！`)
      setTimeout(() => {
        navigate('/')
      }, 1000)
    } catch (err: unknown) {
      showToast('error', `上传失败: ${(err as Error).message}`)
    } finally {
      setUploading(false)
    }
  }

  const clearFile = () => {
    setSelectedFile(null)
    if (fileInputRef.current) fileInputRef.current.value = ''
  }

  return (
    <div className="view active">
      <header className="view-header">
        <div className="header-left">
          <h1>上传 Agent</h1>
          <p className="header-subtitle">将新的 Agent 包添加到仓库</p>
        </div>
      </header>

      <div className="view-content">
        <form id="upload-form" className="upload-form" onSubmit={handleSubmit}>
          <div className="form-section">
            <h3>基本信息</h3>
            <div className="form-grid">
              <div className="form-group">
                <label htmlFor="agent-name">Agent 名称</label>
                <input
                  id="agent-name"
                  type="text"
                  className="form-input"
                  required
                  pattern="[a-z0-9][a-z0-9-]*"
                  value={agentName}
                  onChange={e => setAgentName(e.target.value)}
                  placeholder="my-agent"
                  disabled={uploading}
                />
                <span className="form-hint">小写字母、数字和连字符</span>
              </div>
              <div className="form-group">
                <label htmlFor="category-select">运行时类别</label>
                <div className="select-wrapper">
                  <select
                    id="category-select"
                    className="form-select"
                    required
                    value={category}
                    onChange={e => setCategory(e.target.value)}
                    disabled={uploading}
                  >
                    {KNOWN_CATEGORIES.map(c => (
                      <option key={c} value={c}>
                        {c.charAt(0).toUpperCase() + c.slice(1)}
                      </option>
                    ))}
                  </select>
                </div>
              </div>
              <div className="form-group">
                <label htmlFor="version-input">版本号</label>
                <input
                  id="version-input"
                  type="text"
                  className="form-input"
                  value={version}
                  onChange={e => setVersion(e.target.value)}
                  placeholder="1.0.0"
                  disabled={uploading}
                />
                <span className="form-hint">留空自动生成</span>
              </div>
            </div>
          </div>

          <div className="form-section">
            <h3>包文件</h3>
            <div
              className={`upload-zone ${dragOver ? 'dragover' : ''}`}
              onClick={() => !uploading && fileInputRef.current?.click()}
              onDragOver={(e) => { e.preventDefault(); setDragOver(true) }}
              onDragLeave={() => setDragOver(false)}
              onDrop={handleDrop}
            >
              <input
                ref={fileInputRef}
                type="file"
                accept=".zip"
                hidden
                onChange={e => e.target.files?.[0] && handleFileSelect(e.target.files[0])}
              />

              {!selectedFile ? (
                <div className="upload-zone-content">
                  <div className="upload-icon-wrapper">
                    <CloudUpload size={28} />
                  </div>
                  <p className="upload-title">拖拽 ZIP 文件到此处</p>
                  <p className="upload-subtitle">或点击选择文件</p>
                  <div className="upload-requirements">
                    <span>✓ PicoClaw: 无必需文件</span>
                    <span>✓ OpenClaw: AGENT.md</span>
                  </div>
                </div>
              ) : (
                <div className="upload-zone-file">
                  <div className="file-preview">
                    <FileArchive size={24} style={{ color: 'var(--color-accent)' }} />
                    <div className="file-details">
                      <span className="file-name">{selectedFile.name}</span>
                      <span className="file-size">{formatSize(selectedFile.size)}</span>
                    </div>
                    {!uploading && (
                      <button type="button" className="file-remove" onClick={(e) => { e.stopPropagation(); clearFile() }}>
                        <X size={16} />
                      </button>
                    )}
                  </div>
                </div>
              )}
            </div>
          </div>

          {uploading && (
            <div className="upload-progress">
              <div className="progress-header">
                <Upload size={16} className="spin" />
                <span>上传中...</span>
              </div>
              <div className="progress-track">
                <div className="progress-fill" style={{ width: `${progress}%` }} />
              </div>
              <span className="progress-percent">{progress}%</span>
            </div>
          )}

          <div className="form-actions">
            <Button
              variant="primary"
              size="lg"
              type="submit"
              disabled={uploading}
            >
              <Upload size={16} />
              {uploading ? '上传中...' : '上传 Agent'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}

import { AlertCircle, RefreshCw } from 'lucide-react'

export function LoadingState() {
  return (
    <div className="state-container loading">
      <div className="skeleton-grid">
        {[1, 2, 3, 4].map(i => (
          <div key={i} className="skeleton-card" />
        ))}
      </div>
    </div>
  )
}

export function EmptyState({
  title = '暂无 Agent',
  description = '仓库中还没有 Agent 包，点击「上传 Agent」开始添加',
}: {
  title?: string
  description?: string
}) {
  return (
    <div className="state-container empty">
      <div className="empty-illustration">
        <svg width="120" height="120" viewBox="0 0 120 120" fill="none">
          <rect x="20" y="30" width="80" height="60" rx="8" stroke="currentColor" strokeWidth="2" strokeDasharray="4 4" />
          <path d="M45 55h30M45 65h20" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
          <circle cx="60" cy="45" r="8" stroke="currentColor" strokeWidth="2" />
        </svg>
      </div>
      <h3>{title}</h3>
      <p>{description}</p>
    </div>
  )
}

export function ErrorState({
  message = '无法连接到服务器',
  onRetry,
}: {
  message?: string
  onRetry?: () => void
}) {
  return (
    <div className="state-container error">
      <div className="error-illustration">
        <AlertCircle size={28} />
      </div>
      <h3>加载失败</h3>
      <p>{message}</p>
      {onRetry && (
        <button className="btn btn-primary" onClick={onRetry}>
          <RefreshCw size={16} />
          重新加载
        </button>
      )}
    </div>
  )
}

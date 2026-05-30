import { MarkdownContent, isMarkdownFile } from './MarkdownContent'

interface CodePreviewProps {
  fileName: string
  content: string
  loading?: boolean
  error?: string | null
}

export function CodePreview({ fileName, content, loading, error }: CodePreviewProps) {
  const isMd = isMarkdownFile(fileName)

  return (
    <div className="code-preview-wrapper">
      <div className="preview-header">
        <span>{fileName}</span>
        {isMd && <span className="preview-badge">Markdown</span>}
      </div>
      {loading ? (
        <div className="preview-body preview-loading">加载中...</div>
      ) : error ? (
        <div className="preview-body preview-error">{error}</div>
      ) : isMd ? (
        <div className="preview-body preview-markdown">
          <MarkdownContent>{content}</MarkdownContent>
          {!content.trim() && <p className="preview-empty">文件内容为空</p>}
        </div>
      ) : (
        <pre className="code-preview">
          <code>{content || '文件内容为空'}</code>
        </pre>
      )}
    </div>
  )
}

import { useMemo, useState } from 'react'
import { FileText } from 'lucide-react'
import type { FileEntry } from '../api/types'
import { useFileContent } from '../hooks/useAgents'
import { countFiles, findDefaultFile } from '../utils/agentFiles'
import { FileTree } from './FileTree'
import { CodePreview } from './CodePreview'

interface AgentFileExplorerProps {
  agentName: string
  version: string
  files: FileEntry[]
}

export function AgentFileExplorer({ agentName, version, files }: AgentFileExplorerProps) {
  const [selectedFile, setSelectedFile] = useState<string | null>(null)

  const defaultFile = useMemo(() => findDefaultFile(files), [files])
  const activeFile = selectedFile ?? defaultFile

  const fileContentQuery = useFileContent(
    agentName,
    version,
    activeFile,
    !!agentName && !!version && !!activeFile
  )

  const loading =
    fileContentQuery.isLoading ||
    (fileContentQuery.isFetching && fileContentQuery.data === undefined)

  if (files.length === 0) {
    return (
      <div className="state-container file-explorer-empty">
        <FileText size={48} strokeWidth={1} style={{ color: 'var(--color-text-muted)' }} />
        <p style={{ color: 'var(--color-text-tertiary)', marginTop: '1rem' }}>
          该 Agent 暂无文件，请重新上传完整的 Agent 包
        </p>
      </div>
    )
  }

  return (
    <div className="file-explorer">
      <div className="file-explorer-sidebar">
        <div className="file-explorer-sidebar-header">
          <FileText size={14} />
          <span>Agent 文件</span>
          <span className="file-explorer-count">{countFiles(files)}</span>
        </div>
        <FileTree
          files={files}
          selectedPath={activeFile}
          onSelect={setSelectedFile}
          className="file-explorer-tree"
        />
      </div>
      <div className="file-explorer-preview">
        {activeFile ? (
          <CodePreview
            fileName={activeFile}
            content={fileContentQuery.data ?? ''}
            loading={loading}
            error={
              fileContentQuery.isError
                ? (fileContentQuery.error?.message ?? '加载失败')
                : null
            }
          />
        ) : (
          <div className="state-container file-explorer-empty">
            <FileText size={48} strokeWidth={1} style={{ color: 'var(--color-text-muted)' }} />
            <p style={{ color: 'var(--color-text-tertiary)', marginTop: '1rem' }}>
              点击左侧文件进行预览
            </p>
          </div>
        )}
      </div>
    </div>
  )
}

import { useEffect, useMemo, useState } from 'react'
import { FileText, Save } from 'lucide-react'
import type { FileEntry } from '../api/types'
import { useFileContent } from '../hooks/useAgents'
import { findDefaultFile } from '../utils/agentFiles'
import { FileTree } from './FileTree'
import { CodeEditor } from './CodeEditor'
import { Button } from './UI'

interface AgentFileEditorProps {
  agentName: string
  version: string
  files: FileEntry[]
  onSave: (filePath: string, content: string) => Promise<void>
  saving: boolean
}

export function AgentFileEditor({
  agentName,
  version,
  files,
  onSave,
  saving,
}: AgentFileEditorProps) {
  const [selectedFile, setSelectedFile] = useState<string | null>(null)
  const [fileContent, setFileContent] = useState('')
  const [originalContent, setOriginalContent] = useState('')

  const defaultFile = useMemo(() => findDefaultFile(files), [files])
  const activeFile = selectedFile ?? defaultFile

  const fileContentQuery = useFileContent(
    agentName,
    version,
    activeFile,
    !!agentName && !!version && !!activeFile
  )

  useEffect(() => {
    if (fileContentQuery.data !== undefined) {
      setFileContent(fileContentQuery.data)
      setOriginalContent(fileContentQuery.data)
    }
  }, [fileContentQuery.data])

  useEffect(() => {
    setFileContent('')
    setOriginalContent('')
  }, [activeFile])

  const hasFileChanges = fileContent !== originalContent
  const loading = fileContentQuery.isLoading && fileContentQuery.data === undefined

  if (files.length === 0) {
    return (
      <div className="state-container">
        <FileText size={48} strokeWidth={1} style={{ color: 'var(--color-text-muted)' }} />
        <p style={{ color: 'var(--color-text-tertiary)', marginTop: '1rem' }}>
          该 Agent 暂无文件
        </p>
      </div>
    )
  }

  return (
    <div className="edit-file-layout">
      <div className="edit-file-tree">
        <FileTree
          files={files}
          selectedPath={activeFile}
          onSelect={setSelectedFile}
        />
      </div>
      <div className="edit-file-editor">
        {!activeFile ? (
          <div className="state-container">
            <FileText size={48} strokeWidth={1} style={{ color: 'var(--color-text-muted)' }} />
            <p style={{ color: 'var(--color-text-tertiary)', marginTop: '1rem' }}>
              选择左侧文件进行编辑
            </p>
          </div>
        ) : loading ? (
          <div className="state-container">
            <p style={{ color: 'var(--color-text-tertiary)' }}>加载文件内容...</p>
          </div>
        ) : fileContentQuery.isError ? (
          <div className="state-container">
            <p style={{ color: 'var(--color-danger)' }}>
              {fileContentQuery.error?.message || '加载失败'}
            </p>
          </div>
        ) : (
          <>
            <CodeEditor
              fileName={activeFile}
              content={fileContent}
              onChange={setFileContent}
            />
            <div className="form-actions" style={{ marginTop: '1rem' }}>
              <Button
                variant="primary"
                onClick={() => void onSave(activeFile, fileContent)}
                disabled={!hasFileChanges || saving}
              >
                <Save size={16} />
                {saving ? '保存中...' : '保存文件'}
              </Button>
            </div>
          </>
        )}
      </div>
    </div>
  )
}

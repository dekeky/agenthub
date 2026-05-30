import CodeMirror from '@uiw/react-codemirror'
import { markdown, markdownLanguage } from '@codemirror/lang-markdown'
import { json } from '@codemirror/lang-json'
import { createTheme } from '@uiw/codemirror-themes'
import { tags as t } from '@lezer/highlight'
import { useEffect, useState } from 'react'

interface CodeEditorProps {
  fileName: string
  content: string
  onChange: (content: string) => void
  readOnly?: boolean
}

function getLanguageExtension(fileName: string) {
  if (fileName.endsWith('.md')) return [markdown({ base: markdownLanguage })]
  if (fileName.endsWith('.json')) return [json()]
  return []
}

const darkTheme = createTheme({
  theme: 'dark',
  settings: {
    background: '#0a0a0f',
    foreground: '#f4f4f5',
    caret: '#6366f1',
    selection: 'rgba(99, 102, 241, 0.3)',
    gutterBackground: '#13131a',
    gutterForeground: '#71717a',
    lineHighlight: '#1c1c26',
  },
  styles: [
    { tag: t.comment, color: '#71717a' },
    { tag: t.variableName, color: '#f4f4f5' },
    { tag: [t.string, t.special(t.brace)], color: '#22c55e' },
    { tag: t.number, color: '#f59e0b' },
    { tag: t.bool, color: '#f59e0b' },
    { tag: t.null, color: '#f59e0b' },
    { tag: t.keyword, color: '#ef4444' },
    { tag: t.operator, color: '#a1a1aa' },
    { tag: t.className, color: '#818cf8' },
    { tag: t.definition(t.typeName), color: '#818cf8' },
    { tag: t.typeName, color: '#818cf8' },
    { tag: t.angleBracket, color: '#a1a1aa' },
    { tag: t.tagName, color: '#ef4444' },
    { tag: t.attributeName, color: '#f59e0b' },
    { tag: t.heading, color: '#818cf8', fontWeight: 'bold' },
    { tag: t.strong, fontWeight: 'bold' },
    { tag: t.emphasis, fontStyle: 'italic' },
    { tag: t.link, color: '#6366f1', textDecoration: 'underline' },
    { tag: t.meta, color: '#a1a1aa' },
  ],
})

const lightTheme = createTheme({
  theme: 'light',
  settings: {
    background: '#ffffff',
    foreground: '#111827',
    caret: '#4f46e5',
    selection: 'rgba(79, 70, 229, 0.15)',
    gutterBackground: '#f9fafb',
    gutterForeground: '#9ca3af',
    lineHighlight: '#f3f4f6',
  },
  styles: [
    { tag: t.comment, color: '#9ca3af' },
    { tag: t.variableName, color: '#111827' },
    { tag: [t.string, t.special(t.brace)], color: '#16a34a' },
    { tag: t.number, color: '#d97706' },
    { tag: t.bool, color: '#d97706' },
    { tag: t.null, color: '#d97706' },
    { tag: t.keyword, color: '#dc2626' },
    { tag: t.operator, color: '#4b5563' },
    { tag: t.className, color: '#4f46e5' },
    { tag: t.definition(t.typeName), color: '#4f46e5' },
    { tag: t.typeName, color: '#4f46e5' },
    { tag: t.angleBracket, color: '#4b5563' },
    { tag: t.tagName, color: '#dc2626' },
    { tag: t.heading, color: '#4f46e5', fontWeight: 'bold' },
    { tag: t.strong, fontWeight: 'bold' },
    { tag: t.emphasis, fontStyle: 'italic' },
    { tag: t.link, color: '#4f46e5', textDecoration: 'underline' },
    { tag: t.meta, color: '#4b5563' },
  ],
})

export function CodeEditor({ fileName, content, onChange, readOnly }: CodeEditorProps) {
  const [value, setValue] = useState(content)
  const [appTheme, setAppTheme] = useState<'dark' | 'light'>(
    () => (document.documentElement.getAttribute('data-theme') as 'dark' | 'light') || 'dark'
  )

  // Watch for theme changes on the html element
  useEffect(() => {
    const observer = new MutationObserver(() => {
      const theme = document.documentElement.getAttribute('data-theme') as 'dark' | 'light'
      if (theme) setAppTheme(theme)
    })
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] })
    return () => observer.disconnect()
  }, [])

  useEffect(() => {
    setValue(content)
  }, [content])

  const handleChange = (val: string) => {
    setValue(val)
    onChange(val)
  }

  const isLarge = value.length > 1024 * 1024 // 1MB

  if (isLarge) {
    return (
      <div className="code-editor-large">
        <p>文件过大（超过 1MB），无法在线编辑</p>
        <p className="text-sm text-muted">请下载后使用本地编辑器修改</p>
      </div>
    )
  }

  return (
    <div className="code-editor-wrapper">
      <div className="preview-header">
        <span>{fileName}</span>
        {readOnly && <span className="preview-badge">只读</span>}
      </div>
      <CodeMirror
        value={value}
        height="500px"
        extensions={getLanguageExtension(fileName)}
        onChange={handleChange}
        readOnly={readOnly}
        theme={appTheme === 'light' ? lightTheme : darkTheme}
        basicSetup={{
          lineNumbers: true,
          foldGutter: true,
          highlightActiveLine: true,
          bracketMatching: true,
        }}
      />
    </div>
  )
}

export default CodeEditor

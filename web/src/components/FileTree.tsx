import { useEffect, useMemo, useState } from 'react'
import type { FileEntry } from '../api/types'
import { formatSize } from '../utils/format'
import { ChevronRight, FileText, Folder } from 'lucide-react'

interface FileTreeProps {
  files: FileEntry[]
  selectedPath?: string
  onSelect: (path: string) => void
  className?: string
}

interface TreeDir {
  files: FileEntry[]
  dirs: Map<string, TreeDir>
}

function insertFile(root: TreeDir, entry: FileEntry) {
  const parts = entry.path.split('/').filter(Boolean)
  let node = root
  for (let i = 0; i < parts.length; i++) {
    const isLeaf = i === parts.length - 1
    if (isLeaf) {
      node.files.push(entry)
      return
    }
    const dirName = parts[i]
    if (!node.dirs.has(dirName)) {
      node.dirs.set(dirName, { files: [], dirs: new Map() })
    }
    node = node.dirs.get(dirName)!
  }
}

function insertEmptyDir(root: TreeDir, dirPath: string) {
  const parts = dirPath.split('/').filter(Boolean)
  let node = root
  for (const part of parts) {
    if (!node.dirs.has(part)) {
      node.dirs.set(part, { files: [], dirs: new Map() })
    }
    node = node.dirs.get(part)!
  }
}

/** Build hierarchy from version files; parent dirs come from file paths, empty dirs from dir entries. */
function buildTree(entries: FileEntry[]): TreeDir {
  const root: TreeDir = { files: [], dirs: new Map() }
  for (const entry of entries) {
    if (entry.dir === true) {
      insertEmptyDir(root, entry.path)
    } else {
      insertFile(root, entry)
    }
  }
  return root
}

function collectDirPaths(node: TreeDir, prefix = ''): string[] {
  const paths: string[] = []
  for (const [name, child] of node.dirs) {
    const fullPath = prefix ? `${prefix}/${name}` : name
    paths.push(fullPath)
    paths.push(...collectDirPaths(child, fullPath))
  }
  return paths
}

function isMarkdown(name: string): boolean {
  const lower = name.toLowerCase()
  return lower.endsWith('.md') || lower.endsWith('.markdown')
}

function sortDirEntries(node: TreeDir): [string, TreeDir][] {
  return [...node.dirs.entries()].sort(([a], [b]) => a.localeCompare(b))
}

function sortFiles(files: FileEntry[]): FileEntry[] {
  return [...files].sort((a, b) => {
    const nameA = a.path.split('/').pop() ?? a.path
    const nameB = b.path.split('/').pop() ?? b.path
    return nameA.localeCompare(nameB)
  })
}

export function FileTree({ files, selectedPath, onSelect, className }: FileTreeProps) {
  const tree = useMemo(() => buildTree(files), [files])
  const [expandedDirs, setExpandedDirs] = useState<Set<string>>(() => new Set())

  useEffect(() => {
    setExpandedDirs(new Set(collectDirPaths(tree)))
  }, [tree])

  const toggleDir = (dirPath: string) => {
    setExpandedDirs(prev => {
      const next = new Set(prev)
      if (next.has(dirPath)) next.delete(dirPath)
      else next.add(dirPath)
      return next
    })
  }

  if (files.length === 0) {
    return (
      <div className={`file-tree ${className ?? ''}`}>
        <p className="file-tree-empty">暂无文件</p>
      </div>
    )
  }

  return (
    <div className={`file-tree ${className ?? ''}`}>
      <DirectoryNode
        dirPath=""
        node={tree}
        level={0}
        selectedPath={selectedPath}
        expandedDirs={expandedDirs}
        onToggleDir={toggleDir}
        onSelect={onSelect}
      />
    </div>
  )
}

interface DirectoryNodeProps {
  dirPath: string
  node: TreeDir
  level: number
  selectedPath?: string
  expandedDirs: Set<string>
  onToggleDir: (dirPath: string) => void
  onSelect: (path: string) => void
}

function DirectoryNode({
  dirPath,
  node,
  level,
  selectedPath,
  expandedDirs,
  onToggleDir,
  onSelect,
}: DirectoryNodeProps) {
  const paddingLeft = 8 + level * 14

  return (
    <>
      {sortDirEntries(node).map(([name, child]) => {
        const fullPath = dirPath ? `${dirPath}/${name}` : name
        const expanded = expandedDirs.has(fullPath)

        return (
          <div key={fullPath}>
            <button
              type="button"
              className="file-tree-dir"
              style={{ paddingLeft }}
              onClick={() => onToggleDir(fullPath)}
            >
              <ChevronRight
                size={14}
                className={`file-tree-chevron ${expanded ? 'expanded' : ''}`}
              />
              <Folder size={14} className="file-tree-folder-icon" />
              <span className="file-tree-label">{name}</span>
            </button>
            {expanded && (
              <DirectoryNode
                dirPath={fullPath}
                node={child}
                level={level + 1}
                selectedPath={selectedPath}
                expandedDirs={expandedDirs}
                onToggleDir={onToggleDir}
                onSelect={onSelect}
              />
            )}
          </div>
        )
      })}
      {sortFiles(node.files).map(file => {
        const name = file.path.split('/').pop() ?? file.path
        const selected = file.path === selectedPath

        return (
          <button
            key={file.path}
            type="button"
            className={`file-tree-file ${selected ? 'selected' : ''} ${isMarkdown(name) ? 'is-markdown' : ''}`}
            style={{ paddingLeft }}
            onClick={() => onSelect(file.path)}
            title={file.path}
          >
            <FileText size={14} className="file-tree-file-icon" />
            <span className="file-tree-label">{name}</span>
            <span className="file-tree-size">{formatSize(file.size)}</span>
          </button>
        )
      })}
    </>
  )
}

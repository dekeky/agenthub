import type { FileEntry } from '../api/types'

const PREVIEW_FILES = ['AGENT.md', 'SKILL.md', 'README.md']

export function isFileEntry(entry: FileEntry): boolean {
  return entry.dir !== true
}

export function fileEntriesOnly(files: FileEntry[]): FileEntry[] {
  return files.filter(isFileEntry)
}

export function countFiles(files: FileEntry[]): number {
  return fileEntriesOnly(files).length
}

export function findDefaultFile(files: FileEntry[]): string {
  const fileOnly = fileEntriesOnly(files)
  for (const name of PREVIEW_FILES) {
    const match = fileOnly.find(f => f.path === name || f.path.endsWith(`/${name}`))
    if (match) return match.path
  }
  const firstMd = fileOnly.find(f => f.path.toLowerCase().endsWith('.md'))
  if (firstMd) return firstMd.path
  return fileOnly[0]?.path ?? ''
}

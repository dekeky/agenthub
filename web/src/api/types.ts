export interface AgentMeta {
  agentName: string
  category: string
  displayName: string
  summary: string
  latestVersion: string
  versions: string[]
  updatedAt: string
}

export interface AgentDetail extends AgentMeta {
  files: FileEntry[]
}

export interface FileEntry {
  path: string
  size: number
  /** true when this entry is a directory (size is always 0). */
  dir?: boolean
}

export interface FileContent {
  agentName: string
  version: string
  path: string
  content: string
}

export interface UpdateMetaRequest {
  displayName?: string
  summary?: string
  category?: string
}

export interface UpdateFileRequest {
  version: string
  content: string
}

export interface ApiResponse<T> {
  code: number
  errMsg: string
  body: T
}

export interface ListAgentsResponse {
  agents: AgentMeta[]
  total: number
}

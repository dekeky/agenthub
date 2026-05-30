import { api, uploadFile, testConnection } from './client'
import { encodeFilePath } from '../utils/path'
import type {
  AgentMeta,
  AgentDetail,
  FileContent,
  UpdateMetaRequest,
  UpdateFileRequest,
  ListAgentsResponse,
} from './types'

export async function listAgents(category?: string): Promise<AgentMeta[]> {
  const path = category
    ? `/api/hub/agents?category=${encodeURIComponent(category)}`
    : '/api/hub/agents'
  const data = await api.get<ListAgentsResponse>(path)
  return data.agents || []
}

export async function getAgent(agentName: string, version?: string): Promise<AgentDetail> {
  const base = `/api/hub/agents/${encodeURIComponent(agentName)}`
  const path =
    version && version.trim()
      ? `${base}?version=${encodeURIComponent(version)}`
      : base
  const data = await api.get<AgentDetail>(path)
  return { ...data, files: data.files ?? [] }
}

export async function getFile(
  agentName: string,
  version: string,
  filePath: string
): Promise<string> {
  const path = `/api/hub/agents/${encodeURIComponent(agentName)}/files/${encodeFilePath(filePath)}?version=${encodeURIComponent(version)}`
  const data = await api.get<FileContent>(path)
  return data?.content ?? ''
}

export async function updateAgentMeta(
  agentName: string,
  req: UpdateMetaRequest
): Promise<AgentMeta> {
  return api.put<AgentMeta>(`/api/hub/agents/${encodeURIComponent(agentName)}`, req)
}

export async function updateFile(
  agentName: string,
  filePath: string,
  req: UpdateFileRequest
): Promise<unknown> {
  const path = `/api/hub/agents/${encodeURIComponent(agentName)}/files/${encodeFilePath(filePath)}`
  return api.put(path, req)
}

export async function getCategories(): Promise<string[]> {
  const data = await api.get<{ categories: string[] }>('/api/hub/categories')
  return data.categories || []
}

export async function deleteAgent(agentName: string): Promise<void> {
  await api.delete(`/api/hub/agents/${encodeURIComponent(agentName)}`)
}

export { uploadFile, testConnection }

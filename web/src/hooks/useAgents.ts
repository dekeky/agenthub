import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  listAgents,
  getAgent,
  getFile,
  updateAgentMeta,
  updateFile,
  uploadFile,
  testConnection,
  deleteAgent,
} from '../api/agents'
import type { UpdateMetaRequest, UpdateFileRequest } from '../api/types'

export function useAgents(category?: string) {
  return useQuery({
    queryKey: ['agents', category],
    queryFn: () => listAgents(category),
  })
}

export function useAgentDetail(agentName: string, version?: string) {
  return useQuery({
    queryKey: ['agent', agentName, version ?? 'latest'],
    queryFn: () => getAgent(agentName, version),
    enabled: !!agentName,
  })
}

export function useFileContent(
  agentName: string,
  version: string,
  filePath: string,
  enabled = true
) {
  return useQuery({
    queryKey: ['file', agentName, version, filePath],
    queryFn: () => getFile(agentName, version, filePath),
    enabled: enabled && !!agentName && !!version && !!filePath,
    retry: 1,
  })
}

export function useUpdateMeta(agentName: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (req: UpdateMetaRequest) => updateAgentMeta(agentName, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['agent', agentName] })
      queryClient.invalidateQueries({ queryKey: ['agents'] })
    },
  })
}

export function useUpdateFile(agentName: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ filePath, req }: { filePath: string; req: UpdateFileRequest }) =>
      updateFile(agentName, filePath, req),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: ['file', agentName, variables.req.version, variables.filePath],
      })
      queryClient.invalidateQueries({ queryKey: ['agent', agentName] })
    },
  })
}

export function useUpload() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      formData,
      onProgress,
    }: {
      formData: FormData
      onProgress?: (percent: number) => void
    }) => uploadFile(formData, onProgress),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['agents'] })
    },
  })
}

export function useDeleteAgent() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (agentName: string) => deleteAgent(agentName),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['agents'] })
    },
  })
}

export function useConnection() {
  return useQuery({
    queryKey: ['connection'],
    queryFn: testConnection,
    refetchInterval: 30000,
    retry: false,
  })
}

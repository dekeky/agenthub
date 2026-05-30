import { getApiUrl, getUploadToken } from '../utils/storage'
import type { ApiResponse } from './types'

class ApiError extends Error {
  code: number
  constructor(code: number, message: string) {
    super(message)
    this.code = code
    this.name = 'ApiError'
  }
}

async function request<T>(
  method: string,
  path: string,
  options: {
    body?: unknown
    headers?: Record<string, string>
  } = {}
): Promise<T> {
  const url = `${getApiUrl()}${path}`
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...options.headers,
  }

  const token = getUploadToken()
  if (token) {
    headers['X-Upload-Token'] = token
  }

  const config: RequestInit = {
    method,
    headers,
  }

  if (options.body !== undefined && method !== 'GET') {
    config.body = JSON.stringify(options.body)
  }

  const response = await fetch(url, config)
  const data: ApiResponse<T> = await response.json()

  // Backend returns HTTP status code as "code" (e.g. 200, 404).
  // Accept 2xx as success; treat anything else as error.
  if (data.code !== 0 && (data.code < 200 || data.code >= 300)) {
    throw new ApiError(data.code, data.errMsg || 'Request failed')
  }

  return data.body
}

export async function uploadFile(
  formData: FormData,
  onProgress?: (percent: number) => void
): Promise<unknown> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest()
    const url = `${getApiUrl()}/api/hub/agents`

    xhr.open('POST', url)

    const token = getUploadToken()
    if (token) {
      xhr.setRequestHeader('X-Upload-Token', token)
    }

    xhr.upload.addEventListener('progress', (e) => {
      if (e.lengthComputable && onProgress) {
        const percent = Math.round((e.loaded / e.total) * 100)
        onProgress(percent)
      }
    })

    xhr.addEventListener('load', () => {
      try {
        const data = JSON.parse(xhr.responseText)
        if (data.code !== 0 && (data.code < 200 || data.code >= 300)) {
          reject(new Error(data.errMsg || 'Upload failed'))
        } else {
          resolve(data.body?.agent || data.body)
        }
      } catch {
        reject(new Error('Invalid server response'))
      }
    })

    xhr.addEventListener('error', () => reject(new Error('Network error')))
    xhr.addEventListener('abort', () => reject(new Error('Upload cancelled')))

    xhr.send(formData)
  })
}

export async function testConnection(): Promise<boolean> {
  try {
    const response = await fetch(`${getApiUrl()}/health`)
    return response.ok
  } catch {
    return false
  }
}

export const api = {
  get: <T>(path: string) => request<T>('GET', path),
  post: <T>(path: string, body?: unknown) => request<T>('POST', path, { body }),
  put: <T>(path: string, body?: unknown) => request<T>('PUT', path, { body }),
  delete: <T>(path: string) => request<T>('DELETE', path),
}

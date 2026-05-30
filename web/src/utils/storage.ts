const STORAGE_KEYS = {
  UPLOAD_TOKEN: 'agenthub_upload_token',
  THEME: 'agenthub_theme',
} as const

// Always use relative paths so the frontend works on any origin.
// The Go server serves both the API and the static files from the same address.
export function getApiUrl(): string {
  return ''
}

export function getUploadToken(): string {
  return localStorage.getItem(STORAGE_KEYS.UPLOAD_TOKEN) || ''
}

export function setUploadToken(token: string): void {
  localStorage.setItem(STORAGE_KEYS.UPLOAD_TOKEN, token)
}

// Read upload token from server config at startup.
// The Go server embeds it in a <meta> tag in index.html.
export function initUploadToken(): void {
  const meta = document.querySelector('meta[name="upload-token"]')
  if (meta instanceof HTMLMetaElement && meta.content) {
    const existing = localStorage.getItem(STORAGE_KEYS.UPLOAD_TOKEN)
    if (!existing) {
      localStorage.setItem(STORAGE_KEYS.UPLOAD_TOKEN, meta.content)
    }
  }
}

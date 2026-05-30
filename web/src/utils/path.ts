/** Encode each path segment separately (matches Go client escapePath). */
export function encodeFilePath(filePath: string): string {
  return filePath
    .replace(/^\/+/, '')
    .split('/')
    .map(encodeURIComponent)
    .join('/')
}

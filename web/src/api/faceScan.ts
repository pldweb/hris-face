import { apiClient } from './client'

export interface FaceScanResult {
  found: boolean
  full_name?: string
  similarity?: number
}

export async function scanFace(imageBlob: Blob): Promise<FaceScanResult> {
  const { data } = await apiClient.post<FaceScanResult>('/admin/face-scan', imageBlob, {
    headers: { 'Content-Type': 'image/jpeg' },
  })
  return data
}

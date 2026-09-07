import { apiClient } from './client'

export interface Correction {
  id: string
  employee_id: string
  full_name: string
  requested_type: 'check_in' | 'check_out'
  requested_time: string
  reason: string
  status: 'pending' | 'approved' | 'rejected'
  reviewed_at: string | null
  created_at: string
}

export async function createCorrection(input: {
  requested_type: string
  requested_time: string
  reason: string
}): Promise<Correction> {
  const { data } = await apiClient.post('/corrections', input)
  return data
}

export async function listMyCorrections(): Promise<Correction[]> {
  const { data } = await apiClient.get('/corrections/me')
  return data.data ?? []
}

export async function listCorrections(status?: string): Promise<Correction[]> {
  const { data } = await apiClient.get('/admin/corrections', { params: { status } })
  return data.data ?? []
}

export async function reviewCorrection(id: string, decision: 'approve' | 'reject', note = '') {
  await apiClient.patch(`/admin/corrections/${id}`, { decision, note })
}

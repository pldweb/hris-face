import { apiClient } from './client'

export type LeaveType = 'annual' | 'sick' | 'permit'
export type LeaveStatus = 'pending' | 'approved' | 'rejected'

export interface LeaveRequest {
  id: string
  employee_id: string
  full_name?: string
  type: LeaveType
  start_date: string
  end_date: string
  days_count: number
  reason: string
  status: LeaveStatus
  reviewed_at: string | null
  created_at: string
}

export interface LeaveBalance {
  year: number
  quota: number
  used: number
  remaining: number
}

export async function createLeaveRequest(input: {
  type: LeaveType
  start_date: string
  end_date: string
  reason: string
}): Promise<LeaveRequest> {
  const { data } = await apiClient.post('/leave-requests', input)
  return data
}

export async function listMyLeaveRequests(status?: string): Promise<LeaveRequest[]> {
  const { data } = await apiClient.get('/leave-requests/me', { params: { status } })
  return data.data ?? []
}

export async function fetchMyLeaveBalance(year?: number): Promise<LeaveBalance> {
  const { data } = await apiClient.get('/leave-requests/me/balance', { params: { year } })
  return data
}

export async function listAdminLeaveRequests(filters?: {
  status?: string
  type?: string
  employee_id?: string
  from?: string
  to?: string
}): Promise<LeaveRequest[]> {
  const { data } = await apiClient.get('/admin/leave-requests', { params: filters })
  return data.data ?? []
}

export async function decideLeaveRequest(id: string, decision: 'approve' | 'reject', note = '') {
  const { data } = await apiClient.patch(`/admin/leave-requests/${id}`, { decision, note })
  return data as LeaveRequest
}

export async function updateLeaveRequest(
  id: string,
  input: { type: LeaveType; start_date: string; end_date: string; reason: string },
): Promise<LeaveRequest> {
  const { data } = await apiClient.put(`/admin/leave-requests/${id}`, input)
  return data
}

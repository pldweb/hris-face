import { apiClient } from './client'

export interface AttendanceRow {
  id: string
  employee_id: string
  nik: string
  full_name: string
  department: string
  type: 'check_in' | 'check_out'
  occurred_at: string
  status: 'on_time' | 'late' | 'early_leave' | 'absent'
  source: string
  low_confidence: boolean
}

export interface AttendanceFilter {
  from?: string
  to?: string
  department_id?: string
  status?: string
  limit?: number
  offset?: number
}

export interface DailySummary {
  date: string
  total_active: number
  on_time: number
  late: number
  not_yet: number
  checked_out: number
}

export async function listAttendances(filter: AttendanceFilter, team = false) {
  const { data } = await apiClient.get(team ? '/team/attendances' : '/admin/attendances', {
    params: filter,
  })
  return { rows: (data.data ?? []) as AttendanceRow[], total: (data.total ?? 0) as number }
}

export async function fetchToday(team = false): Promise<DailySummary> {
  const { data } = await apiClient.get(team ? '/team/today' : '/admin/attendances/today')
  return data
}

export async function deleteAttendance(id: string): Promise<void> {
  await apiClient.delete(`/admin/attendances/${id}`)
}

export async function updateAttendance(id: string, occurredAtISO: string): Promise<AttendanceRow> {
  const { data } = await apiClient.put(`/admin/attendances/${id}`, { occurred_at: occurredAtISO })
  return data
}

// Downloads through the same axios instance so the Authorization header and the
// active filter both apply -- a plain <a href> would send neither.
export async function exportReport(filter: AttendanceFilter, format: 'csv' | 'xlsx'): Promise<void> {
  const path = format === 'xlsx' ? '/admin/reports/export.xlsx' : '/admin/reports/export'
  const response = await apiClient.get(path, { params: filter, responseType: 'blob' })
  const url = URL.createObjectURL(response.data as Blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `absensi-${new Date().toISOString().slice(0, 10)}.${format}`
  link.click()
  URL.revokeObjectURL(url)
}

export interface DayRecord {
  date: string
  check_in_at: string | null
  check_out_at: string | null
  status: string
  minutes: number
}

export async function fetchMyHistory(month: string): Promise<DayRecord[]> {
  const { data } = await apiClient.get('/attendance/me', { params: { month } })
  return data.data ?? []
}

import { apiClient } from './client'

export interface Schedule {
  id: string
  name: string
  start_time: string
  end_time: string
  late_tolerance_minutes: number
  work_days: number[]
}

export interface Location {
  id: string
  name: string
}

export interface Device {
  id: string
  employee_id: string
  full_name: string
  user_agent: string
  approved: boolean
  created_at: string
}

export async function listSchedules(): Promise<Schedule[]> {
  const { data } = await apiClient.get('/admin/schedules')
  return data.data ?? []
}

export async function createSchedule(input: Omit<Schedule, 'id'>): Promise<Schedule> {
  const { data } = await apiClient.post('/admin/schedules', input)
  return data
}

export async function updateSchedule(id: string, input: Omit<Schedule, 'id'>): Promise<void> {
  await apiClient.put(`/admin/schedules/${id}`, input)
}

export async function listLocations(): Promise<Location[]> {
  const { data } = await apiClient.get('/admin/locations')
  return data.data ?? []
}

export async function createLocation(name: string): Promise<Location> {
  const { data } = await apiClient.post('/admin/locations', { name })
  return data
}

export async function updateLocation(id: string, name: string): Promise<void> {
  await apiClient.put(`/admin/locations/${id}`, { name })
}

export async function deleteLocation(id: string): Promise<void> {
  await apiClient.delete(`/admin/locations/${id}`)
}

export async function listDevices(): Promise<Device[]> {
  const { data } = await apiClient.get('/admin/devices')
  return data.data ?? []
}

export async function approveDevice(id: string): Promise<void> {
  await apiClient.post(`/admin/devices/${id}/approve`)
}

export async function assignSchedule(employeeId: string, scheduleId: string): Promise<void> {
  await apiClient.put(`/admin/employees/${employeeId}/schedule`, { schedule_id: scheduleId })
}

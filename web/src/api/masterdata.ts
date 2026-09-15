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
  lat?: number
  lng?: number
  radius_meters?: number
}

export interface LocationInput {
  name: string
  // Present only when the caller means to write lat/lng/radius (even as
  // null, to clear them) -- see Service.UpdateLocation on the API side. A
  // plain rename must never touch a geofence HR already configured elsewhere.
  geofence?: { lat: number | null; lng: number | null; radius_meters: number | null }
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

function toLocationPayload(input: LocationInput) {
  return {
    name: input.name,
    ...(input.geofence
      ? {
          set_geofence: true,
          lat: input.geofence.lat,
          lng: input.geofence.lng,
          radius_meters: input.geofence.radius_meters,
        }
      : {}),
  }
}

export async function createLocation(input: LocationInput): Promise<Location> {
  const { data } = await apiClient.post('/admin/locations', toLocationPayload(input))
  return data
}

export async function updateLocation(id: string, input: LocationInput): Promise<void> {
  await apiClient.put(`/admin/locations/${id}`, toLocationPayload(input))
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

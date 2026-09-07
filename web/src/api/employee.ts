import { apiClient } from './client'

export interface Employee {
  id: string
  nik: string
  full_name: string
  email: string
  department_id?: string
  location_id?: string
  status: 'pending_enrollment' | 'active' | 'inactive'
}

export interface Department {
  id: string
  name: string
}

export async function listEmployees(): Promise<Employee[]> {
  const { data } = await apiClient.get('/admin/employees')
  return Array.isArray(data.data) ? data.data : []
}

export interface CreateEmployeeInput {
  nik: string
  full_name: string
  email: string
  department_id?: string
  location_id?: string
}

export interface UpdateEmployeeInput {
  full_name: string
  email: string
  department_id?: string
  location_id?: string
}

export interface CreateEmployeeResult {
  employee: Employee
  temp_password: string
}

export async function createEmployee(input: CreateEmployeeInput): Promise<CreateEmployeeResult> {
  const { data } = await apiClient.post('/admin/employees', input)
  return data
}

export async function updateEmployee(id: string, input: UpdateEmployeeInput): Promise<void> {
  await apiClient.put(`/admin/employees/${id}`, input)
}

export async function deactivateEmployee(id: string): Promise<void> {
  await apiClient.delete(`/admin/employees/${id}`)
}

export async function listDepartments(): Promise<Department[]> {
  const { data } = await apiClient.get('/admin/departments')
  return Array.isArray(data.data) ? data.data : []
}

export async function createDepartment(name: string): Promise<Department> {
  const { data } = await apiClient.post('/admin/departments', { name })
  return data
}

export interface Me {
  employee_id: string
  full_name: string
  status: string
  check_in_at: string | null
  check_out_at: string | null
}

export async function fetchMe(): Promise<Me> {
  const { data } = await apiClient.get<Me>('/me')
  return data
}

export interface ImportRow {
  line: number
  nik: string
  full_name: string
  email: string
  temp_password?: string
  error?: string
}

export interface ImportResult {
  created: ImportRow[]
  failed: ImportRow[]
}

export async function importEmployeesCsv(file: File): Promise<ImportResult> {
  const form = new FormData()
  form.append('file', file)
  const { data } = await apiClient.post('/admin/employees/import', form)
  return data
}

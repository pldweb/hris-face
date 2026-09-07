<script setup lang="ts">
import { DeleteOutlined, EditOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import { computed, onMounted, ref } from 'vue'
import {
  createEmployee,
  createDepartment,
  deactivateEmployee,
  importEmployeesCsv,
  listDepartments,
  listEmployees,
  updateEmployee,
  type Department,
  type Employee,
  type ImportResult,
} from '../../api/employee'
import { listLocations, type Location } from '../../api/masterdata'

const employees = ref<Employee[]>([])
const departments = ref<Department[]>([])
const locations = ref<Location[]>([])
const loading = ref(true)
const modalOpen = ref(false)
const submitting = ref(false)
const deactivatingId = ref<string | null>(null)

// null: create flow (NIK editable). set: editing that employee (NIK locked, PUT instead of POST).
const editingId = ref<string | null>(null)

const form = ref({
  nik: '',
  full_name: '',
  email: '',
  department_id: undefined as string | undefined,
  location_id: undefined as string | undefined,
})
const newDeptName = ref('')
const creatingDept = ref(false)

const tempPasswordResult = ref<{ name: string; password: string } | null>(null)
const importResult = ref<ImportResult | null>(null)
const importing = ref(false)

// Returning false stops AntD from uploading on its own: the file goes through
// the shared axios client so it carries the Authorization header.
function onPickCsv(file: File) {
  importing.value = true
  importEmployeesCsv(file)
    .then((result) => {
      importResult.value = result
      loadAll()
      if (result.failed.length) {
        message.warning(`${result.created.length} berhasil, ${result.failed.length} gagal`)
      } else {
        message.success(`${result.created.length} karyawan diimpor`)
      }
    })
    .catch((err) => message.error(err?.response?.data?.error ?? 'Gagal mengimpor CSV'))
    .finally(() => {
      importing.value = false
    })
  return false
}

// Stable reference: rebuilding this array inline in the template on every
// render tripped an internal crash in rc-select's option memoization.
const departmentOptions = computed(() => departments.value.map((d) => ({ label: d.name, value: d.id })))
const locationOptions = computed(() => locations.value.map((l) => ({ label: l.name, value: l.id })))

const modalTitle = computed(() => (editingId.value ? 'Ubah Karyawan' : 'Tambah Karyawan'))

const columns = [
  { title: 'NIK', dataIndex: 'nik' },
  { title: 'Nama', dataIndex: 'full_name' },
  { title: 'Email', dataIndex: 'email' },
  { title: 'Status', dataIndex: 'status', key: 'status' },
  { title: 'Aksi', key: 'actions' },
]

const statusMeta: Record<Employee['status'], { color: string; label: string }> = {
  pending_enrollment: { color: 'warning', label: 'Belum enrollment' },
  active: { color: 'success', label: 'Aktif' },
  inactive: { color: 'default', label: 'Nonaktif' },
}

async function loadAll() {
  loading.value = true
  try {
    ;[employees.value, departments.value, locations.value] = await Promise.all([
      listEmployees(),
      listDepartments(),
      listLocations(),
    ])
  } catch {
    message.error('Gagal memuat data karyawan')
  } finally {
    loading.value = false
  }
}

function openModal() {
  editingId.value = null
  form.value = { nik: '', full_name: '', email: '', department_id: undefined, location_id: undefined }
  modalOpen.value = true
}

function openEditModal(record: Employee) {
  editingId.value = record.id
  form.value = {
    nik: record.nik,
    full_name: record.full_name,
    email: record.email,
    department_id: record.department_id,
    location_id: record.location_id,
  }
  modalOpen.value = true
}

async function onDeactivate(record: Employee) {
  deactivatingId.value = record.id
  try {
    await deactivateEmployee(record.id)
    message.success(`${record.full_name} dinonaktifkan`)
    await loadAll()
  } catch (err: any) {
    message.error(err?.response?.data?.error ?? 'Gagal menonaktifkan karyawan')
  } finally {
    deactivatingId.value = null
  }
}

async function onCreateDepartment() {
  if (!newDeptName.value.trim()) return
  creatingDept.value = true
  try {
    const dept = await createDepartment(newDeptName.value.trim())
    departments.value.push(dept)
    form.value.department_id = dept.id
    newDeptName.value = ''
  } catch {
    message.error('Gagal membuat departemen')
  } finally {
    creatingDept.value = false
  }
}

async function onSubmit() {
  submitting.value = true
  try {
    if (editingId.value) {
      const { full_name, email, department_id, location_id } = form.value
      await updateEmployee(editingId.value, { full_name, email, department_id, location_id })
      modalOpen.value = false
      message.success('Karyawan diperbarui')
      await loadAll()
    } else {
      const result = await createEmployee(form.value)
      employees.value.push(result.employee)
      modalOpen.value = false
      tempPasswordResult.value = { name: result.employee.full_name, password: result.temp_password }
    }
  } catch (err: any) {
    message.error(err?.response?.data?.error ?? (editingId.value ? 'Gagal memperbarui karyawan' : 'Gagal membuat karyawan'))
  } finally {
    submitting.value = false
  }
}

onMounted(loadAll)
</script>

<template>
  <div>
    <div class="toolbar">
      <h2>Karyawan</h2>
      <a-space>
        <a-upload :show-upload-list="false" accept=".csv,text/csv" :before-upload="onPickCsv">
          <a-button :loading="importing">Import CSV</a-button>
        </a-upload>
        <a-button type="primary" @click="openModal">Tambah Karyawan</a-button>
      </a-space>
    </div>

    <a-table
      :columns="columns"
      :data-source="employees"
      :loading="loading"
      row-key="id"
      size="middle"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'status'">
          <a-tag :color="statusMeta[record.status as Employee['status']].color">
            {{ statusMeta[record.status as Employee['status']].label }}
          </a-tag>
        </template>
        <template v-else-if="column.key === 'actions'">
          <a-space>
            <a-button type="text" size="small" @click="openEditModal(record)">
              <template #icon><EditOutlined /></template>
              Ubah
            </a-button>
            <a-popconfirm
              title="Nonaktifkan karyawan ini?"
              description="Karyawan tidak akan bisa login atau absen. Riwayat kehadirannya tetap tersimpan."
              ok-text="Nonaktifkan"
              cancel-text="Batal"
              :ok-button-props="{ danger: true }"
              @confirm="onDeactivate(record)"
            >
              <a-button
                type="text"
                danger
                size="small"
                :disabled="record.status === 'inactive'"
                :loading="deactivatingId === record.id"
              >
                <template #icon><DeleteOutlined /></template>
                Hapus
              </a-button>
            </a-popconfirm>
          </a-space>
        </template>
      </template>
    </a-table>

    <a-modal v-model:open="modalOpen" :title="modalTitle" :confirm-loading="submitting" @ok="onSubmit">
      <a-form layout="vertical">
        <a-form-item label="NIK">
          <a-input v-model:value="form.nik" :disabled="!!editingId" />
        </a-form-item>
        <a-form-item label="Nama Lengkap">
          <a-input v-model:value="form.full_name" />
        </a-form-item>
        <a-form-item label="Email">
          <a-input v-model:value="form.email" type="email" />
        </a-form-item>
        <a-form-item label="Departemen">
          <a-select
            v-model:value="form.department_id"
            :options="departmentOptions"
            allow-clear
            placeholder="Pilih departemen"
          />
        </a-form-item>
        <a-form-item label="Departemen baru (opsional)">
          <a-space>
            <a-input v-model:value="newDeptName" placeholder="Nama departemen" />
            <a-button :loading="creatingDept" @click="onCreateDepartment">Tambah</a-button>
          </a-space>
        </a-form-item>
        <a-form-item label="Lokasi">
          <a-select
            v-model:value="form.location_id"
            :options="locationOptions"
            allow-clear
            placeholder="Pilih lokasi"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal
      :open="!!importResult"
      title="Hasil Import"
      :footer="null"
      width="720px"
      @cancel="importResult = null"
    >
      <p>
        {{ importResult?.created.length ?? 0 }} karyawan dibuat,
        {{ importResult?.failed.length ?? 0 }} baris gagal.
      </p>
      <a-alert
        v-if="importResult?.created.length"
        type="warning"
        show-icon
        class="import-warning"
        message="Password sementara di bawah ini hanya ditampilkan sekali. Salin sekarang."
      />
      <a-table
        v-if="importResult?.created.length"
        :data-source="importResult.created"
        row-key="line"
        size="small"
        :pagination="false"
        :scroll="{ y: 220 }"
        :columns="[
          { title: 'Nama', dataIndex: 'full_name' },
          { title: 'Email', dataIndex: 'email' },
          { title: 'Password sementara', dataIndex: 'temp_password' },
        ]"
      />
      <a-table
        v-if="importResult?.failed.length"
        :data-source="importResult.failed"
        row-key="line"
        size="small"
        :pagination="false"
        class="import-failed"
        :columns="[
          { title: 'Baris', dataIndex: 'line', width: 70 },
          { title: 'NIK', dataIndex: 'nik' },
          { title: 'Nama', dataIndex: 'full_name' },
          { title: 'Masalah', dataIndex: 'error' },
        ]"
      />
    </a-modal>

    <a-modal
      :open="!!tempPasswordResult"
      title="Karyawan dibuat"
      :footer="null"
      @cancel="tempPasswordResult = null"
    >
      <p>
        Akun untuk <strong>{{ tempPasswordResult?.name }}</strong> berhasil dibuat. Password
        sementara di bawah ini hanya ditampilkan sekali — sampaikan ke karyawan secara langsung,
        bukan lewat email tanpa enkripsi.
      </p>
      <a-typography-paragraph copyable class="temp-password">
        {{ tempPasswordResult?.password }}
      </a-typography-paragraph>
    </a-modal>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.import-warning {
  margin-bottom: 12px;
}

.import-failed {
  margin-top: 16px;
}

.temp-password {
  font-family: monospace;
  font-size: 16px;
  background: #fafafa;
  padding: 8px 12px;
  border-radius: 6px;
}
</style>

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
  updateDepartment,
  updateEmployee,
  type Department,
  type Employee,
  type EmployeeStatus,
  type ImportResult,
} from '../../api/employee'
import {
  createLocation,
  listLocations,
  listSchedules,
  updateLocation,
  type Location,
  type Schedule,
} from '../../api/masterdata'
import InlineMasterSelect from './InlineMasterSelect.vue'

// InlineMasterSelect only ever offers a name field, so these adapt the richer
// masterdata.ts calls (which also carry geofence data for the dedicated
// editor in MasterData.vue) down to the plain rename InlineMasterSelect can
// actually do. Omitting `geofence` here is what keeps a quick rename from
// this form from wiping a radius HR already set up elsewhere.
const createLocationByName = (name: string) => createLocation({ name })
const updateLocationByName = (id: string, name: string) => updateLocation(id, { name })

const employees = ref<Employee[]>([])
const departments = ref<Department[]>([])
const locations = ref<Location[]>([])
const schedules = ref<Schedule[]>([])
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
  schedule_id: undefined as string | undefined,
  manager_id: undefined as string | undefined,
  annual_leave_quota: 12,
  status: 'pending_enrollment' as EmployeeStatus,
  allow_remote: false,
})

// Resetting someone's password is a separate, deliberate act, so it stays
// collapsed: an ordinary edit must not silently reissue their credentials.
const resetPassword = ref(false)
const newPassword = ref('')

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
// Only append the hours when the schedule's own name does not already carry
// them, otherwise a name like "Reguler 08:00-17:00" renders them twice.
const scheduleOptions = computed(() =>
  schedules.value.map((s) => ({
    label: s.name.includes(s.start_time) ? s.name : `${s.name} (${s.start_time}-${s.end_time})`,
    value: s.id,
  })),
)
// An employee cannot manage themselves, so the one being edited is excluded.
const managerOptions = computed(() =>
  employees.value
    .filter((e) => e.id !== editingId.value && e.status !== 'inactive')
    .map((e) => ({ label: `${e.full_name} (${e.nik})`, value: e.id })),
)
const statusOptions = [
  { label: 'Belum enrollment', value: 'pending_enrollment' },
  { label: 'Aktif', value: 'active' },
  { label: 'Nonaktif', value: 'inactive' },
]

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
    ;[employees.value, departments.value, locations.value, schedules.value] = await Promise.all([
      listEmployees(),
      listDepartments(),
      listLocations(),
      listSchedules(),
    ])
  } catch {
    message.error('Gagal memuat data karyawan')
  } finally {
    loading.value = false
  }
}

function openModal() {
  editingId.value = null
  form.value = {
    nik: '', full_name: '', email: '',
    department_id: undefined, location_id: undefined,
    schedule_id: undefined, manager_id: undefined,
    annual_leave_quota: 12, status: 'pending_enrollment',
    allow_remote: false,
  }
  resetPassword.value = false
  newPassword.value = ''
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
    schedule_id: record.schedule_id,
    manager_id: record.manager_id,
    annual_leave_quota: record.annual_leave_quota ?? 12,
    status: record.status,
    allow_remote: record.allow_remote ?? false,
  }
  resetPassword.value = false
  newPassword.value = ''
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

// The inline department/location editor changes master data behind the form,
// so the option lists have to be refetched without disturbing what is typed.
async function reloadMasterLists() {
  ;[departments.value, locations.value] = await Promise.all([listDepartments(), listLocations()])
}

async function onSubmit() {
  submitting.value = true
  try {
    if (editingId.value) {
      const f = form.value
      if (resetPassword.value && newPassword.value.length < 8) {
        message.error('Password baru minimal 8 karakter')
        return
      }
      await updateEmployee(editingId.value, {
        nik: f.nik,
        full_name: f.full_name,
        email: f.email,
        // '' rather than undefined: the API reads undefined as "leave alone"
        // and '' as "clear", which is how a department can be removed at all.
        department_id: f.department_id ?? '',
        location_id: f.location_id ?? '',
        schedule_id: f.schedule_id ?? '',
        manager_id: f.manager_id ?? '',
        annual_leave_quota: f.annual_leave_quota,
        status: f.status,
        allow_remote: f.allow_remote,
        ...(resetPassword.value ? { password: newPassword.value } : {}),
      })
      modalOpen.value = false
      message.success(resetPassword.value ? 'Karyawan diperbarui, password direset' : 'Karyawan diperbarui')
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

    <a-modal v-model:open="modalOpen" :title="modalTitle" width="620px" :confirm-loading="submitting" @ok="onSubmit">
      <a-form layout="vertical">
        <a-row :gutter="12">
          <a-col :span="12">
            <a-form-item label="NIK">
              <a-input v-model:value="form.nik" placeholder="Nomor induk karyawan" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Nama Lengkap">
              <a-input v-model:value="form.full_name" placeholder="Nama lengkap" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="Email">
          <a-input v-model:value="form.email" type="email" placeholder="nama@perusahaan.com" />
        </a-form-item>

        <a-form-item label="Departemen">
          <InlineMasterSelect
            v-model="form.department_id"
            :options="departmentOptions"
            placeholder="Pilih departemen"
            entity-label="departemen"
            :create="createDepartment"
            :update="updateDepartment"
            @changed="reloadMasterLists"
          />
        </a-form-item>
        <a-form-item label="Lokasi kerja">
          <InlineMasterSelect
            v-model="form.location_id"
            :options="locationOptions"
            placeholder="Pilih lokasi"
            entity-label="lokasi"
            :create="createLocationByName"
            :update="updateLocationByName"
            @changed="reloadMasterLists"
          />
        </a-form-item>

        <template v-if="editingId">
          <a-form-item label="Jadwal kerja">
            <a-select
              v-model:value="form.schedule_id"
              :options="scheduleOptions"
              allow-clear
              placeholder="Pilih jadwal"
            />
          </a-form-item>
          <a-row :gutter="12">
            <a-col :span="12">
              <a-form-item label="Atasan">
                <a-select
                  v-model:value="form.manager_id"
                  :options="managerOptions"
                  allow-clear
                  show-search
                  option-filter-prop="label"
                  placeholder="Tanpa atasan"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Status">
                <a-select v-model:value="form.status" :options="statusOptions" />
              </a-form-item>
            </a-col>
          </a-row>
          <a-form-item label="Kuota cuti tahunan (hari)">
            <a-input-number v-model:value="form.annual_leave_quota" :min="0" :max="60" style="width: 100%" />
          </a-form-item>

          <a-form-item>
            <a-checkbox v-model:checked="form.allow_remote">Izinkan absen dari luar jaringan kantor</a-checkbox>
            <p class="pw-hint">
              Wajib dicentang kalau karyawan absen dari HP/lokasi pribadi, bukan dari jaringan kantor.
              Tanpa ini, absen selalu ditolak dengan pesan "di luar jaringan kantor".
            </p>
          </a-form-item>

          <a-checkbox v-model:checked="resetPassword">Reset password karyawan</a-checkbox>
          <a-form-item v-if="resetPassword" label="Password baru" class="pw-field">
            <a-input-password v-model:value="newPassword" placeholder="Minimal 8 karakter" />
            <p class="pw-hint">
              Karyawan akan keluar dari semua perangkat dan harus login dengan password ini.
              Sampaikan langsung, bukan lewat email tanpa enkripsi.
            </p>
          </a-form-item>
        </template>
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

.pw-field {
  margin-top: 12px;
}

.pw-hint {
  margin: 6px 0 0;
  color: #94a3b8;
  font-size: 12px;
  line-height: 1.5;
}

.temp-password {
  font-family: monospace;
  font-size: 16px;
  background: #fafafa;
  padding: 8px 12px;
  border-radius: 6px;
}
</style>

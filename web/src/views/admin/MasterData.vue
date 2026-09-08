<script setup lang="ts">
// Schedules, locations and pending devices. All three are small lists HR edits
// rarely, so they share one screen rather than three near-empty ones.
import { DeleteOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import dayjs from '../../lib/dayjs'
import { onMounted, ref } from 'vue'
import {
  createDepartment,
  deleteDepartment,
  listDepartments,
  updateDepartment,
  type Department,
} from '../../api/employee'
import {
  approveDevice,
  createLocation,
  createSchedule,
  deleteLocation,
  listDevices,
  listLocations,
  listSchedules,
  updateLocation,
  updateSchedule,
  type Device,
  type Location,
  type Schedule,
} from '../../api/masterdata'

const schedules = ref<Schedule[]>([])
const locations = ref<Location[]>([])
const devices = ref<Device[]>([])
const departments = ref<Department[]>([])
const loading = ref(true)

const scheduleModal = ref(false)
const editingId = ref<string | null>(null)
const submitting = ref(false)
const form = ref({ name: '', start_time: '08:00', end_time: '17:00', late_tolerance_minutes: 15 })
const newLocation = ref('')

const newDepartment = ref('')
const departmentModal = ref(false)
const editingDepartmentId = ref<string | null>(null)
const departmentSubmitting = ref(false)
const departmentForm = ref({ name: '' })
const deletingDepartmentId = ref<string | null>(null)

const locationModal = ref(false)
const editingLocationId = ref<string | null>(null)
const locationSubmitting = ref(false)
const locationForm = ref({ name: '' })
const deletingLocationId = ref<string | null>(null)

const DAY_LABELS = ['Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab', 'Min']

async function load() {
  loading.value = true
  try {
    ;[schedules.value, locations.value, devices.value, departments.value] = await Promise.all([
      listSchedules(),
      listLocations(),
      listDevices(),
      listDepartments(),
    ])
  } catch {
    message.error('Gagal memuat master data')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  form.value = { name: '', start_time: '08:00', end_time: '17:00', late_tolerance_minutes: 15 }
  scheduleModal.value = true
}

function openEdit(s: Schedule) {
  editingId.value = s.id
  form.value = {
    name: s.name,
    start_time: s.start_time,
    end_time: s.end_time,
    late_tolerance_minutes: s.late_tolerance_minutes,
  }
  scheduleModal.value = true
}

async function onSubmitSchedule() {
  if (!form.value.name.trim()) {
    message.error('Nama jadwal wajib diisi')
    return
  }
  submitting.value = true
  try {
    const payload = { ...form.value, work_days: [1, 2, 3, 4, 5] }
    if (editingId.value) {
      await updateSchedule(editingId.value, payload)
    } else {
      await createSchedule(payload)
    }
    scheduleModal.value = false
    message.success('Jadwal disimpan')
    load()
  } catch {
    message.error('Gagal menyimpan jadwal')
  } finally {
    submitting.value = false
  }
}

async function onAddLocation() {
  if (!newLocation.value.trim()) return
  try {
    await createLocation(newLocation.value.trim())
    newLocation.value = ''
    load()
  } catch {
    message.error('Gagal menambah lokasi')
  }
}

async function onAddDepartment() {
  if (!newDepartment.value.trim()) return
  try {
    await createDepartment(newDepartment.value.trim())
    newDepartment.value = ''
    load()
  } catch {
    message.error('Gagal menambah departemen')
  }
}

function openEditDepartment(dept: Department) {
  editingDepartmentId.value = dept.id
  departmentForm.value = { name: dept.name }
  departmentModal.value = true
}

async function onSubmitDepartment() {
  if (!departmentForm.value.name.trim()) {
    message.error('Nama departemen wajib diisi')
    return
  }
  if (!editingDepartmentId.value) return
  departmentSubmitting.value = true
  try {
    await updateDepartment(editingDepartmentId.value, departmentForm.value.name.trim())
    departmentModal.value = false
    message.success('Departemen disimpan')
    load()
  } catch {
    message.error('Gagal menyimpan departemen')
  } finally {
    departmentSubmitting.value = false
  }
}

async function onDeleteDepartment(dept: Department) {
  deletingDepartmentId.value = dept.id
  try {
    await deleteDepartment(dept.id)
    message.success(`Departemen ${dept.name} dihapus`)
    await load()
  } catch (err: any) {
    message.error(err?.response?.data?.error ?? 'Gagal menghapus departemen')
  } finally {
    deletingDepartmentId.value = null
  }
}

function openEditLocation(loc: Location) {
  editingLocationId.value = loc.id
  locationForm.value = { name: loc.name }
  locationModal.value = true
}

async function onSubmitLocation() {
  if (!locationForm.value.name.trim()) {
    message.error('Nama lokasi wajib diisi')
    return
  }
  if (!editingLocationId.value) return
  locationSubmitting.value = true
  try {
    await updateLocation(editingLocationId.value, locationForm.value.name.trim())
    locationModal.value = false
    message.success('Lokasi disimpan')
    load()
  } catch {
    message.error('Gagal menyimpan lokasi')
  } finally {
    locationSubmitting.value = false
  }
}

async function onDeleteLocation(loc: Location) {
  deletingLocationId.value = loc.id
  try {
    await deleteLocation(loc.id)
    message.success(`Lokasi ${loc.name} dihapus`)
    await load()
  } catch (err: any) {
    message.error(err?.response?.data?.error ?? 'Gagal menghapus lokasi')
  } finally {
    deletingLocationId.value = null
  }
}

onMounted(load)

async function onApprove(device: Device) {
  try {
    await approveDevice(device.id)
    message.success(`Perangkat ${device.full_name} disetujui`)
    load()
  } catch {
    message.error('Gagal menyetujui perangkat')
  }
}
</script>

<template>
  <div>
    <h2>Master Data</h2>

    <a-card title="Jadwal Kerja" size="small" class="section">
      <template #extra>
        <a-button type="primary" size="small" @click="openCreate">Tambah Jadwal</a-button>
      </template>
      <a-table
        :data-source="schedules"
        :loading="loading"
        row-key="id"
        size="small"
        :pagination="false"
        :columns="[
          { title: 'Nama', dataIndex: 'name' },
          { title: 'Masuk', key: 'start', width: 90 },
          { title: 'Pulang', key: 'end', width: 90 },
          { title: 'Toleransi', key: 'tolerance', width: 110 },
          { title: 'Hari kerja', key: 'days', width: 200 },
          { title: '', key: 'actions', width: 80 },
        ]"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'start'"><span class="tabular-nums">{{ record.start_time }}</span></template>
          <template v-else-if="column.key === 'end'"><span class="tabular-nums">{{ record.end_time }}</span></template>
          <template v-else-if="column.key === 'tolerance'">{{ record.late_tolerance_minutes }} menit</template>
          <template v-else-if="column.key === 'days'">
            <a-tag v-for="d in record.work_days" :key="d">{{ DAY_LABELS[d - 1] }}</a-tag>
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-button type="link" size="small" @click="openEdit(record)">Ubah</a-button>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-card title="Lokasi Kerja" size="small" class="section">
      <a-space class="add-row">
        <a-input v-model:value="newLocation" placeholder="Nama lokasi" @press-enter="onAddLocation" />
        <a-button @click="onAddLocation">Tambah</a-button>
      </a-space>
      <a-list size="small" :data-source="locations" :loading="loading">
        <template #renderItem="{ item }">
          <a-list-item>
            {{ item.name }}
            <template #actions>
              <a-button type="link" size="small" @click="openEditLocation(item)">Ubah</a-button>
              <a-popconfirm
                title="Hapus lokasi ini?"
                description="Lokasi tidak bisa dihapus jika masih dipakai karyawan aktif."
                ok-text="Hapus"
                cancel-text="Batal"
                :ok-button-props="{ danger: true }"
                @confirm="onDeleteLocation(item)"
              >
                <a-button type="text" danger size="small" :loading="deletingLocationId === item.id">
                  <template #icon><DeleteOutlined /></template>
                  Hapus
                </a-button>
              </a-popconfirm>
            </template>
          </a-list-item>
        </template>
        <template #emptyText><a-empty description="Belum ada lokasi" /></template>
      </a-list>
    </a-card>

    <a-card title="Departemen" size="small" class="section">
      <a-space class="add-row">
        <a-input v-model:value="newDepartment" placeholder="Nama departemen" @press-enter="onAddDepartment" />
        <a-button @click="onAddDepartment">Tambah</a-button>
      </a-space>
      <a-list size="small" :data-source="departments" :loading="loading">
        <template #renderItem="{ item }">
          <a-list-item>
            {{ item.name }}
            <template #actions>
              <a-button type="link" size="small" @click="openEditDepartment(item)">Ubah</a-button>
              <a-popconfirm
                title="Hapus departemen ini?"
                description="Departemen tidak bisa dihapus jika masih dipakai karyawan."
                ok-text="Hapus"
                cancel-text="Batal"
                :ok-button-props="{ danger: true }"
                @confirm="onDeleteDepartment(item)"
              >
                <a-button type="text" danger size="small" :loading="deletingDepartmentId === item.id">
                  <template #icon><DeleteOutlined /></template>
                  Hapus
                </a-button>
              </a-popconfirm>
            </template>
          </a-list-item>
        </template>
        <template #emptyText><a-empty description="Belum ada departemen" /></template>
      </a-list>
    </a-card>

    <a-card title="Perangkat" size="small" class="section">
      <p class="hint">
        Maksimal 2 perangkat disetujui per karyawan. Perangkat berikutnya menunggu persetujuan di sini.
      </p>
      <a-table
        :data-source="devices"
        :loading="loading"
        row-key="id"
        size="small"
        :pagination="false"
        :columns="[
          { title: 'Karyawan', dataIndex: 'full_name' },
          { title: 'Browser', dataIndex: 'user_agent', ellipsis: true },
          { title: 'Terdaftar', key: 'created', width: 140 },
          { title: 'Status', key: 'status', width: 160 },
        ]"
      >
        <template #emptyText><a-empty description="Belum ada perangkat terdaftar" /></template>
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'created'">
            <span class="tabular-nums">{{ dayjs(record.created_at).format('DD MMM HH:mm') }}</span>
          </template>
          <template v-else-if="column.key === 'status'">
            <a-tag v-if="record.approved" color="success">Disetujui</a-tag>
            <a-button v-else type="primary" size="small" @click="onApprove(record)">Setujui</a-button>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:open="scheduleModal"
      :title="editingId ? 'Ubah Jadwal' : 'Tambah Jadwal'"
      :confirm-loading="submitting"
      @ok="onSubmitSchedule"
    >
      <a-form layout="vertical">
        <a-form-item label="Nama jadwal">
          <a-input v-model:value="form.name" placeholder="Reguler 08:00-17:00" />
        </a-form-item>
        <a-form-item label="Jam masuk">
          <a-input v-model:value="form.start_time" placeholder="08:00" />
        </a-form-item>
        <a-form-item label="Jam pulang">
          <a-input v-model:value="form.end_time" placeholder="17:00" />
        </a-form-item>
        <a-form-item label="Toleransi terlambat (menit)">
          <a-input-number v-model:value="form.late_tolerance_minutes" :min="0" :max="120" style="width: 100%" />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal
      v-model:open="departmentModal"
      title="Ubah Departemen"
      :confirm-loading="departmentSubmitting"
      @ok="onSubmitDepartment"
    >
      <a-form layout="vertical">
        <a-form-item label="Nama departemen">
          <a-input v-model:value="departmentForm.name" placeholder="Nama departemen" @press-enter="onSubmitDepartment" />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal
      v-model:open="locationModal"
      title="Ubah Lokasi"
      :confirm-loading="locationSubmitting"
      @ok="onSubmitLocation"
    >
      <a-form layout="vertical">
        <a-form-item label="Nama lokasi">
          <a-input v-model:value="locationForm.name" placeholder="Nama lokasi" @press-enter="onSubmitLocation" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped>
.section {
  margin-top: 16px;
}

.add-row {
  margin-bottom: 12px;
}

.hint {
  color: rgba(0, 0, 0, 0.45);
  margin: 0 0 12px;
}
</style>

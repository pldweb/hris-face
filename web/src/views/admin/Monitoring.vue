<script setup lang="ts">
// docs/PRD.md 10.4: four stat tiles then the table. The tiles are not decoration
// -- "belum absen" at 09:00 is HR's actual to-do list for the day.
import {
  BellOutlined,
  CalendarOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  DeleteOutlined,
  EditOutlined,
  FileExcelOutlined,
  FileTextOutlined,
  LogoutOutlined,
  ReloadOutlined,
  SearchOutlined,
} from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import dayjs, { type Dayjs } from '../../lib/dayjs'
import { computed, onMounted, reactive, ref } from 'vue'
import { listDepartments, type Department } from '../../api/employee'
import {
  exportReport,
  deleteAttendance,
  fetchToday,
  listAttendances,
  updateAttendance,
  type AttendanceRow,
  type DailySummary,
} from '../../api/report'

const props = defineProps<{ team?: boolean }>()

const summary = ref<DailySummary | null>(null)
const rows = ref<AttendanceRow[]>([])
const departments = ref<Department[]>([])
const total = ref(0)
const loading = ref(true)
const exporting = ref(false)
const deletingID = ref<string | null>(null)
const editingID = ref<string | null>(null)
const editModalOpen = ref(false)
const editRecord = ref<AttendanceRow | null>(null)
const editValue = ref<Dayjs | null>(null)

const page = ref(1)
const pageSize = 20

function defaultRange(): [Dayjs, Dayjs] {
  return [dayjs().startOf('month'), dayjs()]
}

// draft holds what the fields show; applied holds what was last submitted via
// "Terapkan Filter". Reloading on every keystroke/select made the table flicker
// and refetch mid-adjustment -- the explicit apply step matches how HR actually
// works: pick a period + department + status together, then run it once.
const draft = reactive({
  range: defaultRange() as [Dayjs, Dayjs],
  departmentId: undefined as string | undefined,
  status: undefined as string | undefined,
})
const applied = reactive({
  range: defaultRange() as [Dayjs, Dayjs],
  departmentId: undefined as string | undefined,
  status: undefined as string | undefined,
})

const filter = computed(() => ({
  from: applied.range?.[0]?.format('YYYY-MM-DD'),
  to: applied.range?.[1]?.format('YYYY-MM-DD'),
  department_id: applied.departmentId,
  status: applied.status,
  limit: pageSize,
  offset: (page.value - 1) * pageSize,
}))

const columns = computed(() => [
  { title: 'Tanggal', key: 'date', width: 110 },
  { title: 'Jam', key: 'time', width: 90 },
  { title: 'NIK', dataIndex: 'nik', width: 100 },
  { title: 'Nama', dataIndex: 'full_name' },
  { title: 'Departemen', dataIndex: 'department', width: 140 },
  { title: 'Tipe', key: 'type', width: 90 },
  { title: 'Status', key: 'status', width: 130 },
  ...(!props.team ? [{ title: 'Aksi', key: 'actions', width: 160, align: 'center' }] : []),
])

const statusMeta: Record<string, { color: string; label: string }> = {
  on_time: { color: 'success', label: 'Tepat waktu' },
  late: { color: 'warning', label: 'Terlambat' },
  early_leave: { color: 'warning', label: 'Pulang cepat' },
  absent: { color: 'error', label: 'Tidak hadir' },
}

async function load() {
  loading.value = true
  try {
    const [attendance, today] = await Promise.all([
      listAttendances(filter.value, props.team),
      fetchToday(props.team),
    ])
    rows.value = attendance.rows
    total.value = attendance.total
    summary.value = today
  } catch {
    message.error('Gagal memuat data absensi')
  } finally {
    loading.value = false
  }
}

function onApplyFilter() {
  applied.range = draft.range
  applied.departmentId = draft.departmentId
  applied.status = draft.status
  page.value = 1
  load()
}

function onResetFilter() {
  draft.range = defaultRange()
  draft.departmentId = undefined
  draft.status = undefined
  onApplyFilter()
}

// pct-of-total for the stat card captions; guarded against a 0/0 flash before
// the day's first summary loads.
function pct(value: number | undefined): number {
  const total = summary.value?.total_active ?? 0
  if (!total || !value) return 0
  return Math.round((value / total) * 100)
}

async function onExport(format: 'csv' | 'xlsx') {
  exporting.value = true
  try {
    await exportReport(filter.value, format)
  } catch {
    message.error('Gagal mengekspor data')
  } finally {
    exporting.value = false
  }
}

async function onDelete(record: AttendanceRow) {
  deletingID.value = record.id
  try {
    await deleteAttendance(record.id)
    message.success(`Absensi ${record.full_name} dihapus`)
    await load()
  } catch (err: any) {
    message.error(err?.response?.data?.error ?? 'Gagal menghapus absensi')
  } finally {
    deletingID.value = null
  }
}

function onEditOpen(record: AttendanceRow) {
  editRecord.value = record
  editValue.value = dayjs(record.occurred_at)
  editModalOpen.value = true
}

async function onEditSubmit() {
  if (!editRecord.value || !editValue.value) return
  editingID.value = editRecord.value.id
  try {
    await updateAttendance(editRecord.value.id, editValue.value.toISOString())
    message.success('Waktu absensi diperbarui')
    editModalOpen.value = false
    await load()
  } catch (err: any) {
    message.error(err?.response?.data?.error ?? 'Gagal memperbarui absensi')
  } finally {
    editingID.value = null
  }
}

// One-click shortcut to VIEW this month's recap on screen -- sets the period
// to [1st of this month, today] and applies it, same as manually picking that
// range. It does NOT export a file: that stays a separate, deliberate action
// via the Export buttons right next to it, once HR has looked at the table.
function onRekapBulanIni() {
  draft.range = defaultRange()
  draft.departmentId = undefined
  draft.status = undefined
  onApplyFilter()
}

onMounted(async () => {
  if (!props.team) {
    try {
      departments.value = await listDepartments()
    } catch {
      // Department filter is a convenience; the table still works without it.
    }
  }
  load()
})
</script>

<template>
  <div>
    <div class="toolbar">
      <div>
        <p class="eyebrow">{{ props.team ? 'Tim' : 'Monitoring' }}</p>
        <h2>{{ props.team ? 'Kehadiran Tim' : 'Monitoring Kehadiran' }}</h2>
        <p class="page-desc">Pantau dan kelola data kehadiran karyawan secara real-time.</p>
      </div>
      <a-space v-if="!props.team">
        <a-button @click="onRekapBulanIni">
          <template #icon><CalendarOutlined style="color: #21409a" /></template>
          Rekap Bulan Ini
        </a-button>
        <a-button :loading="exporting" @click="onExport('xlsx')">
          <template #icon><FileExcelOutlined style="color: #16a34a" /></template>
          Export Excel
        </a-button>
        <a-button :loading="exporting" @click="onExport('csv')">
          <template #icon><FileTextOutlined style="color: #2563eb" /></template>
          Export CSV
        </a-button>
      </a-space>
    </div>

    <a-row :gutter="16" class="stats">
      <a-col :xs="12" :md="6">
        <a-card size="small" class="stat-card">
          <span class="stat-icon stat-icon--success"><CheckCircleOutlined /></span>
          <p class="stat-title">Tepat waktu</p>
          <p class="stat-value">{{ summary?.on_time ?? 0 }}</p>
          <p class="stat-caption">{{ pct(summary?.on_time) }}% dari total</p>
        </a-card>
      </a-col>
      <a-col :xs="12" :md="6">
        <a-card size="small" class="stat-card">
          <span class="stat-icon stat-icon--warning"><ClockCircleOutlined /></span>
          <p class="stat-title">Terlambat</p>
          <p class="stat-value">{{ summary?.late ?? 0 }}</p>
          <p class="stat-caption">{{ pct(summary?.late) }}% dari total</p>
        </a-card>
      </a-col>
      <a-col :xs="12" :md="6">
        <a-card size="small" class="stat-card">
          <span class="stat-icon stat-icon--danger"><BellOutlined /></span>
          <p class="stat-title">Belum absen</p>
          <p class="stat-value">{{ summary?.not_yet ?? 0 }}</p>
          <p class="stat-caption">{{ pct(summary?.not_yet) }}% dari total</p>
        </a-card>
      </a-col>
      <a-col :xs="12" :md="6">
        <a-card size="small" class="stat-card">
          <span class="stat-icon stat-icon--info"><LogoutOutlined /></span>
          <p class="stat-title">Sudah pulang</p>
          <p class="stat-value">{{ summary?.checked_out ?? 0 }}</p>
          <p class="stat-caption">{{ pct(summary?.checked_out) }}% dari total</p>
        </a-card>
      </a-col>
    </a-row>

    <a-card size="small" class="filter-card">
      <a-form layout="inline" class="filters">
        <a-form-item label="Periode">
          <a-range-picker v-model:value="draft.range" format="DD MMM YYYY" />
        </a-form-item>
        <a-form-item v-if="!props.team" label="Departemen">
          <a-select
            v-model:value="draft.departmentId"
            :options="departments.map((d) => ({ label: d.name, value: d.id }))"
            style="width: 180px"
            allow-clear
            placeholder="Semua"
          />
        </a-form-item>
        <a-form-item label="Status">
          <a-select
            v-model:value="draft.status"
            :options="[
              { label: 'Tepat waktu', value: 'on_time' },
              { label: 'Terlambat', value: 'late' },
              { label: 'Pulang cepat', value: 'early_leave' },
            ]"
            style="width: 160px"
            allow-clear
            placeholder="Semua"
          />
        </a-form-item>
        <a-form-item>
          <a-space>
            <a-button type="primary" @click="onApplyFilter">
              <template #icon><SearchOutlined /></template>
              Terapkan Filter
            </a-button>
            <a-button @click="onResetFilter">
              <template #icon><ReloadOutlined /></template>
              Reset
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-card>

    <a-card size="small" class="table-card">
    <a-table
      :columns="columns"
      :data-source="rows"
      :loading="loading"
      row-key="id"
      size="small"
      :pagination="{ current: page, pageSize, total, showSizeChanger: false }"
      @change="(p: any) => { page = p.current; load() }"
    >
      <template #emptyText>
        <a-empty class="table-empty">
          <template #description>
            <p class="empty-title">Tidak ada data</p>
            <p class="empty-desc">Data kehadiran tidak ditemukan pada periode yang dipilih.</p>
          </template>
          <a-button @click="onResetFilter">Ubah Filter</a-button>
        </a-empty>
      </template>
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'date'">
          <span class="tabular-nums">{{ dayjs(record.occurred_at).format('DD MMM YYYY') }}</span>
        </template>
        <template v-else-if="column.key === 'time'">
          <span class="tabular-nums">{{ dayjs(record.occurred_at).format('HH:mm') }}</span>
        </template>
        <template v-else-if="column.key === 'type'">
          {{ record.type === 'check_out' ? 'Pulang' : 'Masuk' }}
        </template>
        <template v-else-if="column.key === 'status'">
          <a-tag :color="statusMeta[record.status]?.color">{{ statusMeta[record.status]?.label ?? record.status }}</a-tag>
          <a-tooltip v-if="record.source === 'manual_correction'" title="Hasil koreksi manual, bukan verifikasi wajah">
            <a-tag>manual</a-tag>
          </a-tooltip>
          <a-tooltip v-if="record.low_confidence" title="Skor kemiripan tipis, perlu ditinjau">
            <a-tag color="orange">review</a-tag>
          </a-tooltip>
        </template>
        <template v-else-if="column.key === 'actions'">
          <a-button type="text" size="small" @click="onEditOpen(record)">
            <template #icon><EditOutlined /></template>
            Ubah
          </a-button>
          <a-popconfirm
            title="Hapus record absensi ini?"
            description="Aksi ini menghapus absensi dari daftar dan dicatat di audit log."
            ok-text="Hapus"
            cancel-text="Batal"
            :ok-button-props="{ danger: true }"
            @confirm="onDelete(record)"
          >
            <a-button type="text" danger size="small" :loading="deletingID === record.id">
              <template #icon><DeleteOutlined /></template>
              Hapus
            </a-button>
          </a-popconfirm>
        </template>
      </template>
    </a-table>
    </a-card>

    <a-modal
      v-model:open="editModalOpen"
      title="Ubah Waktu Absensi"
      ok-text="Simpan"
      cancel-text="Batal"
      :confirm-loading="editingID !== null"
      @ok="onEditSubmit"
    >
      <p v-if="editRecord">{{ editRecord.full_name }} &middot; {{ editRecord.type === 'check_out' ? 'Pulang' : 'Masuk' }}</p>
      <a-date-picker v-model:value="editValue" show-time format="DD MMM YYYY HH:mm" style="width: 100%" />
    </a-modal>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
  gap: 16px;
}

.eyebrow {
  margin: 0 0 4px;
  color: #21409a;
  font-size: 12px;
  font-weight: 700;
}

.toolbar h2 {
  margin: 0;
  letter-spacing: -0.02em;
}

.page-desc {
  margin: 6px 0 0;
  color: #64748b;
  font-size: 13px;
}

.stats {
  margin-bottom: 16px;
}

.stat-card {
  position: relative;
  overflow: hidden;
}
.stat-card :deep(.ant-card-body) {
  padding: 18px;
}

.stat-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  font-size: 18px;
  margin-bottom: 12px;
}
.stat-icon--success { color: #16a34a; background: #dcfce7; }
.stat-icon--warning { color: #d48806; background: #fef3c7; }
.stat-icon--danger { color: #dc2626; background: #fee2e2; }
.stat-icon--info { color: #2563eb; background: #dbeafe; }

.stat-title {
  margin: 0;
  color: #64748b;
  font-size: 13px;
}
.stat-value {
  margin: 4px 0 0;
  color: #172554;
  font-size: 28px;
  font-weight: 700;
  line-height: 1.2;
  font-variant-numeric: tabular-nums;
}
.stat-caption {
  margin: 4px 0 0;
  color: #94a3b8;
  font-size: 12px;
}

.filter-card,
.table-card {
  margin-bottom: 16px;
}
.filter-card :deep(.ant-card-body) {
  padding: 16px;
}

.filters {
  row-gap: 12px;
}

.table-empty {
  padding: 32px 0;
}
.empty-title {
  margin: 8px 0 4px;
  font-weight: 600;
  color: #172554;
}
.empty-desc {
  margin: 0 0 12px;
  color: #64748b;
  font-size: 13px;
}
</style>

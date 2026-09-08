<script setup lang="ts">
import { message, Modal } from 'ant-design-vue'
import dayjs from '../../lib/dayjs'
import { onMounted, ref } from 'vue'
import type { Dayjs } from '../../lib/dayjs'
import { listAdminLeaveRequests, decideLeaveRequest, updateLeaveRequest, type LeaveRequest, type LeaveType } from '../../api/leave'

const items = ref<LeaveRequest[]>([])
const loading = ref(true)
const statusFilter = ref<string | undefined>('pending')
const typeFilter = ref<string | undefined>(undefined)
const note = ref('')

const editOpen = ref(false)
const editing = ref<LeaveRequest | null>(null)
const savingEdit = ref(false)
const editForm = ref({
  type: 'annual' as LeaveType,
  range: null as [Dayjs, Dayjs] | null,
  reason: '',
})

const typeOptions = [
  { label: 'Cuti Tahunan', value: 'annual' },
  { label: 'Sakit', value: 'sick' },
  { label: 'Izin', value: 'permit' },
]

function openEdit(item: LeaveRequest) {
  editing.value = item
  editForm.value = {
    type: item.type,
    range: [dayjs(item.start_date), dayjs(item.end_date)],
    reason: item.reason,
  }
  editOpen.value = true
}

async function onSaveEdit() {
  const f = editForm.value
  if (!f.range || !f.reason.trim()) {
    message.error('Tanggal dan alasan wajib diisi')
    return
  }
  if (!editing.value) return
  savingEdit.value = true
  try {
    await updateLeaveRequest(editing.value.id, {
      type: f.type,
      start_date: f.range[0].format('YYYY-MM-DD'),
      end_date: f.range[1].format('YYYY-MM-DD'),
      reason: f.reason.trim(),
    })
    editOpen.value = false
    message.success('Pengajuan cuti diperbarui')
    load()
  } catch (err: any) {
    message.error(err?.response?.data?.error ?? 'Gagal memperbarui pengajuan')
  } finally {
    savingEdit.value = false
  }
}

const columns = [
  { title: 'Diajukan', key: 'created', width: 150 },
  { title: 'Karyawan', dataIndex: 'full_name' },
  { title: 'Jenis', key: 'type', width: 130 },
  { title: 'Tanggal', key: 'dates', width: 190 },
  { title: 'Hari kerja', dataIndex: 'days_count', width: 100 },
  { title: 'Alasan', dataIndex: 'reason' },
  { title: 'Status', key: 'status', width: 110 },
  { title: '', key: 'actions', width: 230 },
]

const statusMeta: Record<string, { color: string; label: string }> = {
  pending: { color: 'processing', label: 'Menunggu' },
  approved: { color: 'success', label: 'Disetujui' },
  rejected: { color: 'default', label: 'Ditolak' },
}

const typeLabels: Record<string, string> = {
  annual: 'Cuti Tahunan',
  sick: 'Sakit',
  permit: 'Izin',
}

function labelForType(type: string) {
  return typeLabels[type] ?? type
}

async function load() {
  loading.value = true
  try {
    items.value = await listAdminLeaveRequests({ status: statusFilter.value, type: typeFilter.value })
  } catch {
    message.error('Gagal memuat pengajuan')
  } finally {
    loading.value = false
  }
}

function confirmReview(item: LeaveRequest, decision: 'approve' | 'reject') {
  note.value = ''
  Modal.confirm({
    title: decision === 'approve' ? 'Setujui pengajuan?' : 'Tolak pengajuan?',
    content:
      decision === 'approve'
        ? `Pengajuan ${labelForType(item.type)} ${item.full_name} (${item.days_count} hari kerja, ` +
          `${dayjs(item.start_date).format('DD MMM')}–${dayjs(item.end_date).format('DD MMM YYYY')}) akan disetujui.`
        : `Pengajuan ${item.full_name} akan ditolak. Karyawan menerima pemberitahuan.`,
    okText: decision === 'approve' ? 'Setujui' : 'Tolak',
    okType: decision === 'approve' ? 'primary' : 'danger',
    cancelText: 'Batal',
    async onOk() {
      try {
        await decideLeaveRequest(item.id, decision, note.value)
        message.success(decision === 'approve' ? 'Pengajuan disetujui' : 'Pengajuan ditolak')
        load()
      } catch (err: any) {
        message.error(err?.response?.data?.error ?? 'Gagal memproses pengajuan')
      }
    },
  })
}

onMounted(load)
</script>

<template>
  <div>
    <div class="toolbar">
      <h2>Cuti & Izin</h2>
      <a-space>
        <a-select
          v-model:value="statusFilter"
          :options="[
            { label: 'Menunggu', value: 'pending' },
            { label: 'Disetujui', value: 'approved' },
            { label: 'Ditolak', value: 'rejected' },
          ]"
          style="width: 160px"
          allow-clear
          placeholder="Semua status"
          @change="load"
        />
        <a-select
          v-model:value="typeFilter"
          :options="[
            { label: 'Cuti Tahunan', value: 'annual' },
            { label: 'Sakit', value: 'sick' },
            { label: 'Izin', value: 'permit' },
          ]"
          style="width: 160px"
          allow-clear
          placeholder="Semua jenis"
          @change="load"
        />
      </a-space>
    </div>

    <a-table :columns="columns" :data-source="items" :loading="loading" row-key="id" size="small">
      <template #emptyText>
        <a-empty description="Tidak ada pengajuan cuti" />
      </template>
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'created'">
          <span class="tabular-nums">{{ dayjs(record.created_at).format('DD MMM HH:mm') }}</span>
        </template>
        <template v-else-if="column.key === 'type'">
          <a-tag>{{ labelForType(record.type) }}</a-tag>
        </template>
        <template v-else-if="column.key === 'dates'">
          <span class="tabular-nums">
            {{ dayjs(record.start_date).format('DD MMM YYYY') }}–{{ dayjs(record.end_date).format('DD MMM YYYY') }}
          </span>
        </template>
        <template v-else-if="column.key === 'status'">
          <a-tag :color="statusMeta[record.status]?.color">{{ statusMeta[record.status]?.label }}</a-tag>
        </template>
        <template v-else-if="column.key === 'actions'">
          <a-space>
            <template v-if="record.status === 'pending'">
              <a-button type="primary" size="small" @click="confirmReview(record, 'approve')">Setujui</a-button>
              <a-button danger size="small" @click="confirmReview(record, 'reject')">Tolak</a-button>
            </template>
            <a-button type="text" size="small" @click="openEdit(record)">Ubah</a-button>
          </a-space>
        </template>
      </template>
    </a-table>

    <a-modal
      v-model:open="editOpen"
      title="Ubah Pengajuan Cuti"
      :confirm-loading="savingEdit"
      ok-text="Simpan"
      cancel-text="Batal"
      @ok="onSaveEdit"
    >
      <p v-if="editing" class="edit-subject">
        Pengajuan atas nama <strong>{{ editing.full_name }}</strong>
      </p>
      <a-form layout="vertical">
        <a-form-item label="Jenis">
          <a-select v-model:value="editForm.type" :options="typeOptions" />
        </a-form-item>
        <a-form-item label="Tanggal">
          <a-range-picker v-model:value="editForm.range" format="DD MMMM YYYY" :allow-clear="false" style="width: 100%" />
        </a-form-item>
        <a-form-item label="Alasan">
          <a-textarea v-model:value="editForm.reason" :rows="3" />
        </a-form-item>
      </a-form>
      <p class="edit-hint">
        Jumlah hari kerja dihitung ulang otomatis dari jadwal kerja karyawan, dan kuota cuti
        tahunan diperiksa ulang saat disimpan.
      </p>
    </a-modal>
  </div>
</template>

<style scoped>
.edit-subject {
  margin: 0 0 12px;
  color: #475569;
}

.edit-hint {
  margin: 4px 0 0;
  color: #94a3b8;
  font-size: 12px;
  line-height: 1.5;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
</style>

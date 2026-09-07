<script setup lang="ts">
import { message, Modal } from 'ant-design-vue'
import dayjs from '../../lib/dayjs'
import { onMounted, ref } from 'vue'
import { listAdminLeaveRequests, decideLeaveRequest, type LeaveRequest } from '../../api/leave'

const items = ref<LeaveRequest[]>([])
const loading = ref(true)
const statusFilter = ref<string | undefined>('pending')
const typeFilter = ref<string | undefined>(undefined)
const note = ref('')

const columns = [
  { title: 'Diajukan', key: 'created', width: 150 },
  { title: 'Karyawan', dataIndex: 'full_name' },
  { title: 'Jenis', key: 'type', width: 130 },
  { title: 'Tanggal', key: 'dates', width: 190 },
  { title: 'Hari kerja', dataIndex: 'days_count', width: 100 },
  { title: 'Alasan', dataIndex: 'reason' },
  { title: 'Status', key: 'status', width: 110 },
  { title: '', key: 'actions', width: 170 },
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
          <a-space v-if="record.status === 'pending'">
            <a-button type="primary" size="small" @click="confirmReview(record, 'approve')">Setujui</a-button>
            <a-button danger size="small" @click="confirmReview(record, 'reject')">Tolak</a-button>
          </a-space>
        </template>
      </template>
    </a-table>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
</style>

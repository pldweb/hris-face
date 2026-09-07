<script setup lang="ts">
import { message, Modal } from 'ant-design-vue'
import dayjs from '../../lib/dayjs'
import { onMounted, ref } from 'vue'
import { listCorrections, reviewCorrection, type Correction } from '../../api/correction'

const items = ref<Correction[]>([])
const loading = ref(true)
const statusFilter = ref<string | undefined>('pending')
const note = ref('')

const columns = [
  { title: 'Diajukan', key: 'created', width: 150 },
  { title: 'Karyawan', dataIndex: 'full_name' },
  { title: 'Untuk', key: 'requested', width: 190 },
  { title: 'Alasan', dataIndex: 'reason' },
  { title: 'Status', key: 'status', width: 110 },
  { title: '', key: 'actions', width: 170 },
]

const statusMeta: Record<string, { color: string; label: string }> = {
  pending: { color: 'processing', label: 'Menunggu' },
  approved: { color: 'success', label: 'Disetujui' },
  rejected: { color: 'default', label: 'Ditolak' },
}

async function load() {
  loading.value = true
  try {
    items.value = await listCorrections(statusFilter.value)
  } catch {
    message.error('Gagal memuat pengajuan')
  } finally {
    loading.value = false
  }
}

function confirmReview(item: Correction, decision: 'approve' | 'reject') {
  note.value = ''
  Modal.confirm({
    title: decision === 'approve' ? 'Setujui koreksi?' : 'Tolak koreksi?',
    content:
      decision === 'approve'
        ? `Absen ${item.requested_type === 'check_out' ? 'pulang' : 'masuk'} ${item.full_name} ` +
          `akan dicatat pada ${dayjs(item.requested_time).format('DD MMM YYYY HH:mm')} dan ditandai sebagai koreksi manual.`
        : `Pengajuan ${item.full_name} akan ditolak. Karyawan menerima pemberitahuan.`,
    okText: decision === 'approve' ? 'Setujui' : 'Tolak',
    okType: decision === 'approve' ? 'primary' : 'danger',
    cancelText: 'Batal',
    async onOk() {
      try {
        await reviewCorrection(item.id, decision, note.value)
        message.success(decision === 'approve' ? 'Koreksi disetujui' : 'Koreksi ditolak')
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
      <h2>Koreksi Absen</h2>
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
    </div>

    <a-table :columns="columns" :data-source="items" :loading="loading" row-key="id" size="small">
      <template #emptyText>
        <a-empty description="Tidak ada pengajuan koreksi" />
      </template>
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'created'">
          <span class="tabular-nums">{{ dayjs(record.created_at).format('DD MMM HH:mm') }}</span>
        </template>
        <template v-else-if="column.key === 'requested'">
          <span class="tabular-nums">
            {{ record.requested_type === 'check_out' ? 'Pulang' : 'Masuk' }}
            {{ dayjs(record.requested_time).format('DD MMM YYYY HH:mm') }}
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

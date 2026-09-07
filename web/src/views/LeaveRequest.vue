<script setup lang="ts">
import { message } from 'ant-design-vue'
import dayjs, { type Dayjs } from '../lib/dayjs'
import { onMounted, ref } from 'vue'
import {
  createLeaveRequest,
  fetchMyLeaveBalance,
  listMyLeaveRequests,
  type LeaveBalance,
  type LeaveRequest,
  type LeaveType,
} from '../api/leave'
import EmployeeNav from './EmployeeNav.vue'

const requests = ref<LeaveRequest[]>([])
const balance = ref<LeaveBalance | null>(null)
const loading = ref(true)

const modalOpen = ref(false)
const submitting = ref(false)
const form = ref({
  type: 'annual' as LeaveType,
  range: [dayjs(), dayjs()] as [Dayjs, Dayjs],
  reason: '',
})

const typeLabel: Record<LeaveType, string> = {
  annual: 'Cuti Tahunan',
  sick: 'Sakit',
  permit: 'Izin',
}

const statusMeta: Record<string, { color: string; label: string }> = {
  pending: { color: 'processing', label: 'Menunggu' },
  approved: { color: 'success', label: 'Disetujui' },
  rejected: { color: 'default', label: 'Ditolak' },
}

async function load() {
  loading.value = true
  try {
    ;[requests.value, balance.value] = await Promise.all([
      listMyLeaveRequests(),
      fetchMyLeaveBalance(),
    ])
  } catch {
    message.error('Gagal memuat data cuti')
  } finally {
    loading.value = false
  }
}

function openModal() {
  form.value = { type: 'annual', range: [dayjs(), dayjs()], reason: '' }
  modalOpen.value = true
}

async function onSubmit() {
  if (form.value.reason.trim().length < 5) {
    message.error('Alasan minimal 5 karakter')
    return
  }
  submitting.value = true
  try {
    await createLeaveRequest({
      type: form.value.type,
      start_date: form.value.range[0].format('YYYY-MM-DD'),
      end_date: form.value.range[1].format('YYYY-MM-DD'),
      reason: form.value.reason.trim(),
    })
    modalOpen.value = false
    message.success('Pengajuan cuti dikirim')
    load()
  } catch (err: any) {
    message.error(err?.response?.data?.error ?? 'Gagal mengajukan cuti')
  } finally {
    submitting.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="leave-shell">
    <EmployeeNav />
    <div class="leave-page">
      <div class="leave-column">
        <header class="page-header">
          <h2>Cuti & Izin</h2>
          <a-button type="primary" @click="openModal">Ajukan Cuti/Izin</a-button>
        </header>

        <a-row :gutter="16" class="stats">
          <a-col :span="8">
            <a-card size="small"><a-statistic title="Kuota tahunan" :value="balance?.quota ?? 0" /></a-card>
          </a-col>
          <a-col :span="8">
            <a-card size="small"><a-statistic title="Terpakai" :value="balance?.used ?? 0" /></a-card>
          </a-col>
          <a-col :span="8">
            <a-card size="small"><a-statistic title="Sisa" :value="balance?.remaining ?? 0" /></a-card>
          </a-col>
        </a-row>
        <p class="hint">Kuota berlaku untuk cuti tahunan. Sakit dan izin tidak memotong kuota.</p>

        <a-table
          :data-source="requests"
          :loading="loading"
          row-key="id"
          size="small"
          :columns="[
            { title: 'Jenis', key: 'type' },
            { title: 'Tanggal', key: 'dates' },
            { title: 'Hari kerja', dataIndex: 'days_count', key: 'days_count' },
            { title: 'Alasan', dataIndex: 'reason', key: 'reason' },
            { title: 'Status', key: 'status' },
            { title: 'Diajukan', key: 'created_at' },
          ]"
        >
          <template #emptyText>
            <a-empty description="Belum ada pengajuan cuti" />
          </template>
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'type'">
              {{ typeLabel[record.type as LeaveType] ?? record.type }}
            </template>
            <template v-else-if="column.key === 'dates'">
              {{ dayjs(record.start_date).format('DD MMM YYYY') }} – {{ dayjs(record.end_date).format('DD MMM YYYY') }}
            </template>
            <template v-else-if="column.key === 'status'">
              <a-tag :color="statusMeta[record.status]?.color">{{ statusMeta[record.status]?.label ?? record.status }}</a-tag>
            </template>
            <template v-else-if="column.key === 'created_at'">
              {{ dayjs(record.created_at).format('DD MMM HH:mm') }}
            </template>
          </template>
        </a-table>
      </div>

      <a-modal v-model:open="modalOpen" title="Ajukan Cuti/Izin" :confirm-loading="submitting" @ok="onSubmit">
        <a-form layout="vertical">
          <a-form-item label="Jenis">
            <a-radio-group v-model:value="form.type">
              <a-radio-button value="annual">Cuti Tahunan</a-radio-button>
              <a-radio-button value="sick">Sakit</a-radio-button>
              <a-radio-button value="permit">Izin</a-radio-button>
            </a-radio-group>
          </a-form-item>
          <a-form-item label="Tanggal mulai - selesai">
            <a-range-picker
              v-model:value="form.range"
              format="DD MMMM YYYY"
              :allow-clear="false"
              style="width: 100%"
            />
          </a-form-item>
          <a-form-item label="Alasan">
            <a-textarea v-model:value="form.reason" :rows="3" placeholder="Contoh: acara keluarga" />
          </a-form-item>
        </a-form>
      </a-modal>
    </div>
  </div>
</template>

<style scoped>
.leave-shell {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.leave-page {
  flex: 1;
  background: var(--ant-color-bg-layout, #f5f6f8);
  padding: 24px;
}

.leave-column {
  max-width: 900px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.page-header h2 {
  margin: 0;
}

.stats {
  margin-bottom: 8px;
}

.hint {
  color: #64748b;
  font-size: 13px;
  margin: 0 0 16px;
}
</style>

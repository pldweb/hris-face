<script setup lang="ts">
// The employee's own month, plus the correction escape hatch. This screen is
// what makes a failed face match recoverable instead of a lost day.
import { message } from 'ant-design-vue'
import dayjs, { type Dayjs } from '../lib/dayjs'
import { computed, onMounted, ref } from 'vue'
import { createCorrection, listMyCorrections, type Correction } from '../api/correction'
import { fetchMyHistory, type DayRecord } from '../api/report'
import EmployeeNav from './EmployeeNav.vue'

const month = ref<Dayjs>(dayjs())
const days = ref<DayRecord[]>([])
const corrections = ref<Correction[]>([])
const loading = ref(true)

const modalOpen = ref(false)
const submitting = ref(false)
const form = ref({
  requested_type: 'check_in' as 'check_in' | 'check_out',
  date: dayjs() as Dayjs,
  time: dayjs().hour(8).minute(0) as Dayjs,
  reason: '',
})

const statusMeta: Record<string, { color: string; label: string }> = {
  on_time: { color: 'success', label: 'Tepat waktu' },
  late: { color: 'warning', label: 'Terlambat' },
  early_leave: { color: 'warning', label: 'Pulang cepat' },
  absent: { color: 'default', label: '—' },
}

const totalMinutes = computed(() => days.value.reduce((sum, d) => sum + d.minutes, 0))
const presentDays = computed(() => days.value.filter((d) => d.check_in_at).length)

async function load() {
  loading.value = true
  try {
    ;[days.value, corrections.value] = await Promise.all([
      fetchMyHistory(month.value.format('YYYY-MM')),
      listMyCorrections(),
    ])
  } catch {
    message.error('Gagal memuat riwayat')
  } finally {
    loading.value = false
  }
}

function openCorrection(date?: string) {
  form.value = {
    requested_type: 'check_in',
    date: date ? dayjs(date) : dayjs(),
    time: dayjs().hour(8).minute(0),
    reason: '',
  }
  modalOpen.value = true
}

async function onSubmit() {
  if (form.value.reason.trim().length < 5) {
    message.error('Alasan minimal 5 karakter')
    return
  }
  submitting.value = true
  try {
    const when = form.value.date
      .hour(form.value.time.hour())
      .minute(form.value.time.minute())
      .second(0)
    await createCorrection({
      requested_type: form.value.requested_type,
      requested_time: when.toISOString(),
      reason: form.value.reason.trim(),
    })
    modalOpen.value = false
    message.success('Pengajuan terkirim, menunggu persetujuan HR')
    load()
  } catch (err: any) {
    message.error(err?.response?.data?.error ?? 'Gagal mengajukan koreksi')
  } finally {
    submitting.value = false
  }
}

function formatDuration(minutes: number) {
  if (!minutes) return '—'
  return `${Math.floor(minutes / 60)}j ${minutes % 60}m`
}

onMounted(load)
</script>

<template>
  <div class="history-shell">
    <EmployeeNav />
    <div class="history-page">
    <div class="history-column">
      <header class="page-header">
        <h2>Riwayat Absensi</h2>
      </header>

      <a-row :gutter="16" class="stats">
        <a-col :span="8">
          <a-card size="small"><a-statistic title="Hari hadir" :value="presentDays" /></a-card>
        </a-col>
        <a-col :span="8">
          <a-card size="small"><a-statistic title="Total jam" :value="Math.floor(totalMinutes / 60)" suffix="jam" /></a-card>
        </a-col>
        <a-col :span="8">
          <a-card size="small">
            <a-statistic title="Koreksi menunggu" :value="corrections.filter((c) => c.status === 'pending').length" />
          </a-card>
        </a-col>
      </a-row>

      <div class="controls">
        <a-date-picker v-model:value="month" picker="month" format="MMMM YYYY" :allow-clear="false" @change="load" />
        <a-button type="primary" @click="openCorrection()">Ajukan Koreksi</a-button>
      </div>

      <a-table
        :data-source="days"
        :loading="loading"
        row-key="date"
        size="small"
        :pagination="false"
        :columns="[
          { title: 'Tanggal', key: 'date', width: 140 },
          { title: 'Masuk', key: 'in', width: 90 },
          { title: 'Pulang', key: 'out', width: 90 },
          { title: 'Durasi', key: 'duration', width: 100 },
          { title: 'Status', key: 'status' },
        ]"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'date'">
            <span class="tabular-nums">{{ dayjs(record.date).format('ddd, DD MMM') }}</span>
          </template>
          <template v-else-if="column.key === 'in'">
            <span class="tabular-nums">{{ record.check_in_at ? dayjs(record.check_in_at).format('HH:mm') : '—' }}</span>
          </template>
          <template v-else-if="column.key === 'out'">
            <span class="tabular-nums">{{ record.check_out_at ? dayjs(record.check_out_at).format('HH:mm') : '—' }}</span>
          </template>
          <template v-else-if="column.key === 'duration'">
            <span class="tabular-nums">{{ formatDuration(record.minutes) }}</span>
          </template>
          <template v-else-if="column.key === 'status'">
            <a-tag :color="statusMeta[record.status]?.color">{{ statusMeta[record.status]?.label ?? record.status }}</a-tag>
            <a-button
              v-if="!record.check_in_at || !record.check_out_at"
              type="link"
              size="small"
              @click="openCorrection(record.date)"
            >
              Ajukan koreksi
            </a-button>
          </template>
        </template>
      </a-table>

      <a-card v-if="corrections.length" title="Pengajuan Koreksi" size="small" class="corrections">
        <a-list size="small" :data-source="corrections">
          <template #renderItem="{ item }">
            <a-list-item>
              <a-list-item-meta
                :title="`${item.requested_type === 'check_out' ? 'Pulang' : 'Masuk'} ${dayjs(item.requested_time).format('DD MMM YYYY HH:mm')}`"
                :description="item.reason"
              />
              <a-tag
                :color="item.status === 'approved' ? 'success' : item.status === 'rejected' ? 'default' : 'processing'"
              >
                {{ item.status === 'approved' ? 'Disetujui' : item.status === 'rejected' ? 'Ditolak' : 'Menunggu' }}
              </a-tag>
            </a-list-item>
          </template>
        </a-list>
      </a-card>
    </div>

    <a-modal v-model:open="modalOpen" title="Ajukan Koreksi Absen" :confirm-loading="submitting" @ok="onSubmit">
      <a-form layout="vertical">
        <a-form-item label="Jenis">
          <a-radio-group v-model:value="form.requested_type">
            <a-radio-button value="check_in">Absen masuk</a-radio-button>
            <a-radio-button value="check_out">Absen pulang</a-radio-button>
          </a-radio-group>
        </a-form-item>
        <a-form-item label="Tanggal">
          <a-date-picker v-model:value="form.date" format="DD MMMM YYYY" :allow-clear="false" style="width: 100%" />
        </a-form-item>
        <a-form-item label="Jam">
          <a-time-picker v-model:value="form.time" format="HH:mm" :allow-clear="false" style="width: 100%" />
        </a-form-item>
        <a-form-item label="Alasan">
          <a-textarea
            v-model:value="form.reason"
            :rows="3"
            placeholder="Contoh: kamera tidak mengenali wajah setelah 3 percobaan"
          />
        </a-form-item>
      </a-form>
    </a-modal>
    </div>
  </div>
</template>

<style scoped>
.history-shell {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.history-page {
  flex: 1;
  background: var(--ant-color-bg-layout, #f5f6f8);
  padding: 24px;
}

.history-column {
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
  margin-bottom: 16px;
}

.controls {
  display: flex;
  justify-content: space-between;
  margin-bottom: 16px;
}

.corrections {
  margin-top: 16px;
}
</style>

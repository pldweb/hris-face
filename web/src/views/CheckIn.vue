<script setup lang="ts">
// Direction contract: .impeccable/surfaces/web-src-views-checkin-vue.md
// One full viewport, camera + clock, no dashboard-card treatment.
// The ring around the camera IS the state indicator (10.3) -- text below it
// is a second channel, never the only one (screen readers get aria-live, 10.7).
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { submitAttendance, submitChallenge, CheckInError, type CheckInResult } from '../api/attendance'
import { fetchMe, type Me } from '../api/employee'
import dayjs from '../lib/dayjs'
import EmployeeNav from './EmployeeNav.vue'
import { useFaceFrame } from './useFaceFrame'

const { t } = useI18n()

type CameraError = 'permission' | 'no_device' | null
type SubmitPhase = 'idle' | 'submitting' | 'challenge' | 'success' | 'error'

const videoEl = ref<HTMLVideoElement | null>(null)
const canvasEl = ref<HTMLCanvasElement | null>(null)
const cameraError = ref<CameraError>(null)
const submitPhase = ref<SubmitPhase>('idle')
const errorMessage = ref('')
const result = ref<CheckInResult | null>(null)
const me = ref<Me | null>(null)
const now = ref(dayjs())
let clockTimer: ReturnType<typeof setInterval>

const { state: frameState, start: startFraming, reset: resetFraming } = useFaceFrame(() => {
  if (submitPhase.value === 'idle' && mode.value !== 'done') submitFrame()
})

// Which action this screen offers now, derived from today's own record rather
// than a button the employee picks -- there is only ever one valid next step.
const mode = computed<'check_in' | 'check_out' | 'done'>(() => {
  if (!me.value?.check_in_at) return 'check_in'
  if (!me.value?.check_out_at) return 'check_out'
  return 'done'
})

const actionLabel = computed(() =>
  mode.value === 'check_out' ? t('checkin.actionCheckOut') : t('checkin.actionCheckIn'),
)

const ring = computed(() => {
  if (submitPhase.value === 'submitting') return 'submitting'
  if (submitPhase.value === 'success') return 'success'
  if (submitPhase.value === 'error') return 'error'
  return frameState.value === 'detected' || frameState.value === 'stable' ? 'detected' : 'idle'
})

const statusText = computed(() => {
  if (submitPhase.value === 'submitting') return t('checkin.stateSubmitting')
  if (submitPhase.value === 'challenge') return t('checkin.stateChallenge')
  if (submitPhase.value === 'success' && result.value) {
    const statusLabel =
      result.value.status === 'late'
        ? t('checkin.statusLate')
        : result.value.status === 'early_leave'
          ? t('checkin.statusEarlyLeave')
          : t('checkin.statusOnTime')
    return t('checkin.stateSuccess', {
      name: result.value.full_name,
      time: dayjs(result.value.occurred_at).format('HH:mm'),
      status: statusLabel,
    })
  }
  if (submitPhase.value === 'error') return errorMessage.value
  return ring.value === 'detected' ? t('checkin.stateDetected') : t('checkin.stateIdle')
})

const greeting = computed(() => {
  const hour = now.value.hour()
  const period = hour < 11 ? 'morning' : hour < 15 ? 'afternoon' : 'evening'
  return t('checkin.greeting', {
    period: t(`checkin.${period}`),
    name: me.value?.full_name ?? '',
  }).trim()
})

// Today's own record, refreshed after a successful check-in so the summary
// row reflects reality instead of a permanent em dash.
// /me is the source of truth; the just-submitted result only fills the gap
// until that refetch lands, and only for its own type.
const checkInAt = computed(() =>
  me.value?.check_in_at
    ? dayjs(me.value.check_in_at)
    : result.value?.type === 'check_in'
      ? dayjs(result.value.occurred_at)
      : null,
)
const checkOutAt = computed(() =>
  me.value?.check_out_at
    ? dayjs(me.value.check_out_at)
    : result.value?.type === 'check_out'
      ? dayjs(result.value.occurred_at)
      : null,
)
const workedDuration = computed(() => {
  if (!checkInAt.value || !checkOutAt.value) return null
  const mins = checkOutAt.value.diff(checkInAt.value, 'minute')
  return `${Math.floor(mins / 60)}j ${mins % 60}m`
})

async function initCamera() {
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ video: { width: 640, height: 480 } })
    if (videoEl.value) {
      videoEl.value.srcObject = stream
      await videoEl.value.play()
      await startFraming(videoEl.value)
    }
  } catch (err: any) {
    cameraError.value = err?.name === 'NotFoundError' ? 'no_device' : 'permission'
  }
}

function captureJpeg(): Promise<Blob | null> {
  const video = videoEl.value
  const canvas = canvasEl.value
  if (!video || !canvas) return Promise.resolve(null)
  canvas.width = video.videoWidth
  canvas.height = video.videoHeight
  canvas.getContext('2d')?.drawImage(video, 0, 0)
  return new Promise((resolve) => canvas.toBlob(resolve, 'image/jpeg', 0.9))
}

// Captures a short burst so the server can see movement. The frames are spread
// over ~1.2s because two frames 30ms apart look like the same still.
async function captureBurst(count = 4): Promise<Blob[]> {
  const frames: Blob[] = []
  for (let i = 0; i < count; i++) {
    const blob = await captureJpeg()
    if (blob) frames.push(blob)
    if (i < count - 1) await new Promise((r) => setTimeout(r, 400))
  }
  return frames
}

async function runChallenge(kind: 'check_in' | 'check_out') {
  submitPhase.value = 'challenge'
  const frames = await captureBurst()
  result.value = await submitChallenge(kind, frames)
}

// Normal flow passes no kind (derived from mode); the "ulangi absen" buttons
// on the done screen pass one explicitly, since mode is 'done' at that point
// and can no longer say which session to resubmit.
async function submitFrame(explicitKind?: 'check_in' | 'check_out') {
  const kind = explicitKind ?? (mode.value === 'check_out' ? 'check_out' : 'check_in')
  const blob = await captureJpeg()
  if (!blob) return

  submitPhase.value = 'submitting'
  try {
    result.value = await submitAttendance(kind, blob)
    submitPhase.value = 'success'
    await loadMe()
    // Return to idle so the next action (check-out) is reachable without a
    // reload; the summary row keeps the confirmation, so nothing is lost.
    setTimeout(() => {
      if (submitPhase.value === 'success') {
        submitPhase.value = 'idle'
        resetFraming()
      }
    }, 5000)
  } catch (err) {
    // 428 means the passive model was unsure, not that the person failed.
    // Escalate to the movement challenge rather than turning them away.
    if (err instanceof CheckInError && err.code === 'challenge_required') {
      try {
        await runChallenge(kind)
        submitPhase.value = 'success'
        await loadMe()
        setTimeout(() => {
          if (submitPhase.value === 'success') {
            submitPhase.value = 'idle'
            resetFraming()
          }
        }, 5000)
        return
      } catch (challengeErr) {
        submitPhase.value = 'error'
        errorMessage.value =
          challengeErr instanceof CheckInError
            ? mapErrorMessage(challengeErr)
            : 'Verifikasi gerakan gagal.'
        setTimeout(() => {
          submitPhase.value = 'idle'
          resetFraming()
        }, 3000)
        return
      }
    }

    submitPhase.value = 'error'
    if (err instanceof CheckInError) {
      errorMessage.value = mapErrorMessage(err)
    } else {
      errorMessage.value = 'Terjadi kesalahan. Coba lagi.'
    }
    setTimeout(() => {
      submitPhase.value = 'idle'
      resetFraming()
    }, 3000)
  }
}

function mapErrorMessage(err: CheckInError): string {
  switch (err.code) {
    case 'liveness':
      return t('checkin.errorLiveness')
    case 'no_match':
      return t('checkin.errorNoMatch')
    case 'wrong_person':
      return t('checkin.errorWrongPerson', { name: err.matchedName ?? '?' })
    case 'no_check_in_yet':
      return t('checkin.errorNoCheckInYet')
    case 'device_not_approved':
      return t('checkin.errorDeviceNotApproved')
    case 'still_image':
      return t('checkin.errorStillImage')
    case 'challenge_failed':
      return t('checkin.errorChallengeFailed')
    case 'outside_network':
      return t('checkin.errorOutsideNetwork')
    default:
      return 'Terjadi kesalahan. Coba lagi.'
  }
}

async function loadMe() {
  try {
    me.value = await fetchMe()
  } catch {
    // A missing profile must not block attendance; the greeting just stays generic.
  }
}

onMounted(() => {
  loadMe()
  initCamera()
  clockTimer = setInterval(() => {
    now.value = dayjs()
  }, 1000)
})

onUnmounted(() => clearInterval(clockTimer))
</script>

<template>
  <div class="checkin-shell">
    <EmployeeNav />
    <div class="checkin-page">
    <div class="checkin-column">
      <header class="checkin-header">
        <p class="greeting">{{ greeting }}</p>
        <p class="date">{{ now.format('dddd, D MMMM YYYY') }}</p>
      </header>

      <p class="clock tabular-nums">{{ now.format('HH:mm:ss') }}</p>

      <div v-if="cameraError === 'permission'" class="camera-fallback">
        <a-result status="warning" :title="t('checkin.errorPermissionDenied')" />
      </div>
      <div v-else-if="cameraError === 'no_device'" class="camera-fallback">
        <a-result status="info" :title="t('checkin.errorNoCamera')">
          <template #extra>
            <a-button type="primary">{{ t('checkin.errorNoCameraAction') }}</a-button>
          </template>
        </a-result>
      </div>
      <template v-else>
        <div class="camera-frame" :class="`ring-${ring}`">
          <video ref="videoEl" class="camera-video" muted playsinline />
          <canvas ref="canvasEl" hidden />
        </div>

        <p class="status-text" aria-live="polite">{{ statusText }}</p>

        <a-button
          v-if="mode !== 'done'"
          type="primary"
          size="large"
          block
          :loading="submitPhase === 'submitting' || submitPhase === 'challenge'"
          :disabled="ring !== 'detected'"
          @click="submitFrame()"
        >
          {{ actionLabel }}
        </a-button>
        <template v-else>
          <p class="all-done">{{ t('checkin.allDone') }}</p>
          <div class="rescan-row">
            <a-button
              :loading="submitPhase === 'submitting' || submitPhase === 'challenge'"
              :disabled="ring !== 'detected'"
              @click="submitFrame('check_in')"
            >
              {{ t('checkin.rescanIn') }}
            </a-button>
            <a-button
              :loading="submitPhase === 'submitting' || submitPhase === 'challenge'"
              :disabled="ring !== 'detected'"
              @click="submitFrame('check_out')"
            >
              {{ t('checkin.rescanOut') }}
            </a-button>
          </div>
        </template>
      </template>

      <div class="page-links">
        <router-link to="/enroll?ulang=1">Daftar ulang wajah</router-link>
      </div>

      <div class="summary-row">
        <span>{{ t('checkin.summaryIn') }} {{ checkInAt ? checkInAt.format('HH:mm') : '—' }}</span>
        <span class="dot">·</span>
        <span>{{ t('checkin.summaryOut') }} {{ checkOutAt ? checkOutAt.format('HH:mm') : '—' }}</span>
        <span class="dot">·</span>
        <span>{{ t('checkin.summaryDuration') }} {{ workedDuration ?? '—' }}</span>
      </div>
    </div>
    </div>
  </div>
</template>

<style scoped>
.checkin-shell {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.checkin-page {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--ant-color-bg-layout, #f5f6f8);
  padding: 24px;
}

.checkin-column {
  width: 100%;
  max-width: 720px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 18px;
}

.checkin-header {
  text-align: center;
}

.greeting {
  font-size: 20px;
  font-weight: 600;
  margin: 0;
}

.date {
  margin: 4px 0 0;
  color: rgba(0, 0, 0, 0.45);
}

.clock {
  font-size: 64px;
  font-weight: 600;
  line-height: 1;
  margin: 8px 0;
}

.camera-frame {
  width: 100%;
  max-width: 480px;
  aspect-ratio: 4 / 3;
  border-radius: 12px;
  overflow: hidden;
  border: 5px solid #cbd5e1;
  box-shadow: 0 18px 40px rgba(23, 37, 84, .12);
  transition: border-color 200ms ease;
  background: #000;
}

.camera-frame.ring-idle {
  border-color: #d9d9d9;
}
.camera-frame.ring-detected {
  border-color: #1d3a8f;
}
.camera-frame.ring-submitting {
  border-color: #1d3a8f;
  animation: pulse-border 900ms ease-in-out infinite;
}
.camera-frame.ring-success {
  border-color: #52c41a;
}
.camera-frame.ring-error {
  border-color: #ff4d4f;
}

@keyframes pulse-border {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.camera-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transform: scaleX(-1); /* mirror, matching how users expect to see themselves */
}

.camera-fallback {
  width: 100%;
  max-width: 480px;
}

.status-text {
  min-height: 22px;
  color: #475569;
  font-weight: 600;
  margin: 0;
}

.page-links {
  font-size: 13px;
  padding-top: 2px;
}

.all-done {
  color: rgba(0, 0, 0, 0.45);
  margin: 0;
}

.rescan-row {
  display: flex;
  gap: 8px;
  width: 100%;
}

.rescan-row .ant-btn {
  flex: 1;
}

.summary-row {
  color: #64748b;
  padding: 10px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: rgba(255, 255, 255, .8);
  font-variant-numeric: tabular-nums;
  display: flex;
  gap: 8px;
  margin-top: 8px;
}
</style>

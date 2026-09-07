<script setup lang="ts">
// Extends the CheckIn surface's visual system (ring-frame camera, centered
// column) rather than inventing a new one -- same world, different task.
// docs/PRD.md 6.1: 3-5 photos, explicit pose prompts, server re-validates all.
import { message } from 'ant-design-vue'
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { submitEnrollment } from '../api/enrollment'
import { useFaceFrame } from './useFaceFrame'

const MIN_PHOTOS = 3
const MAX_PHOTOS = 5

const { t } = useI18n()
const POSES = [t('enroll.poseFront'), t('enroll.poseLeft'), t('enroll.poseRight'), t('enroll.poseExtra'), t('enroll.poseExtra')]
const router = useRouter()
const route = useRoute()
// ?ulang=1 replaces an existing face instead of registering a first one; the
// server archives the old embeddings rather than deleting them.
const isReEnroll = route.query.ulang === '1' 
const videoEl = ref<HTMLVideoElement | null>(null)
const canvasEl = ref<HTMLCanvasElement | null>(null)
const cameraReady = ref(false)
const cameraError = ref(false)

const captures = ref<{ blob: Blob; url: string }[]>([])
const submitting = ref(false)
const rejectedIndexes = ref<Set<number>>(new Set())

const { state: frameState, start: startFraming } = useFaceFrame(() => {})

const canCapture = computed(
  () => captures.value.length < MAX_PHOTOS && (frameState.value === 'detected' || frameState.value === 'stable'),
)
const canSubmit = computed(() => captures.value.length >= MIN_PHOTOS && !submitting.value)
const currentPoseLabel = computed(() => POSES[captures.value.length] ?? t('enroll.poseExtra'))

async function initCamera() {
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ video: { width: 640, height: 480 } })
    if (videoEl.value) {
      videoEl.value.srcObject = stream
      await videoEl.value.play()
      await startFraming(videoEl.value)
      cameraReady.value = true
    }
  } catch {
    cameraError.value = true
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

async function onCapture() {
  const blob = await captureJpeg()
  if (!blob) return
  captures.value.push({ blob, url: URL.createObjectURL(blob) })
}

function onRemove(index: number) {
  URL.revokeObjectURL(captures.value[index].url)
  captures.value.splice(index, 1)
  rejectedIndexes.value.clear()
}

async function onSubmit() {
  submitting.value = true
  rejectedIndexes.value.clear()
  try {
    const outcome = await submitEnrollment(captures.value.map((c) => c.blob), isReEnroll)
    if (outcome.ok) {
      message.success(t('enroll.success'))
      router.push('/checkin')
      return
    }
    if (outcome.kind === 'rejections') {
      rejectedIndexes.value = new Set(outcome.rejections.map((r) => r.index))
      const reasons = outcome.rejections.map((r) => `Foto ${r.index + 1}: ${r.reason}`).join('; ')
      message.error(t('enroll.rejectedSummary', { reasons }))
    } else {
      message.error(outcome.message)
    }
  } finally {
    submitting.value = false
  }
}

onMounted(initCamera)
onUnmounted(() => captures.value.forEach((c) => URL.revokeObjectURL(c.url)))
</script>

<template>
  <div class="enroll-page">
    <div class="enroll-column">
      <header class="enroll-header">
        <h2>{{ isReEnroll ? t('enroll.titleRe') : t('enroll.title') }}</h2>
        <p>{{ isReEnroll ? t('enroll.subtitleRe') : t('enroll.subtitle') }}</p>
      </header>

      <a-result v-if="cameraError" status="warning" :title="t('enroll.cameraError')" />

      <template v-else>
        <div class="camera-frame" :class="frameState === 'idle' ? 'ring-idle' : 'ring-detected'">
          <video ref="videoEl" class="camera-video" muted playsinline />
          <canvas ref="canvasEl" hidden />
        </div>

        <p class="pose-hint" aria-live="polite">
          {{ captures.length < MAX_PHOTOS ? currentPoseLabel : t('enroll.ready') }}
        </p>

        <a-button type="primary" :disabled="!canCapture" @click="onCapture">{{ t('enroll.capture') }}</a-button>

        <div class="thumbnails">
          <div v-for="(c, i) in captures" :key="c.url" class="thumb" :class="{ rejected: rejectedIndexes.has(i) }">
            <img :src="c.url" alt="" />
            <a-button size="small" danger shape="circle" class="thumb-remove" @click="onRemove(i)">✕</a-button>
          </div>
        </div>

        <a-button type="primary" size="large" block :disabled="!canSubmit" :loading="submitting" @click="onSubmit">
          {{ t('enroll.submit', { count: captures.length, min: MIN_PHOTOS }) }}
        </a-button>
      </template>
    </div>
  </div>
</template>

<style scoped>
.enroll-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--ant-color-bg-layout, #f5f6f8);
  padding: 24px;
}

.enroll-column {
  width: 100%;
  max-width: 480px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.enroll-header {
  text-align: center;
}

.enroll-header p {
  color: rgba(0, 0, 0, 0.45);
  margin: 4px 0 0;
}

.camera-frame {
  width: 100%;
  aspect-ratio: 4 / 3;
  border-radius: 12px;
  overflow: hidden;
  border: 4px solid #d9d9d9;
  transition: border-color 200ms ease;
  background: #000;
}

.camera-frame.ring-detected {
  border-color: #1d3a8f;
}

.camera-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transform: scaleX(-1);
}

.pose-hint {
  margin: 0;
  color: rgba(0, 0, 0, 0.65);
}

.thumbnails {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: center;
}

.thumb {
  position: relative;
  width: 64px;
  height: 64px;
  border-radius: 6px;
  overflow: hidden;
  border: 2px solid transparent;
}

.thumb.rejected {
  border-color: #ff4d4f;
}

.thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.thumb-remove {
  position: absolute;
  top: -6px;
  right: -6px;
}
</style>

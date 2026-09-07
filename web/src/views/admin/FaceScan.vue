<script setup lang="ts">
import { CameraOutlined, CheckCircleFilled, CloseCircleFilled, ScanOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { scanFace, type FaceScanResult } from '../../api/faceScan'
import { useFaceFrame } from '../useFaceFrame'

const videoEl = ref<HTMLVideoElement | null>(null)
const canvasEl = ref<HTMLCanvasElement | null>(null)
const stream = ref<MediaStream | null>(null)
const cameraError = ref('')
const scanning = ref(false)
const result = ref<FaceScanResult | null>(null)
const { state: frameState, start: startFraming } = useFaceFrame(() => {})

const canScan = computed(() => !scanning.value && (frameState.value === 'detected' || frameState.value === 'stable'))

async function initCamera() {
  try {
    stream.value = await navigator.mediaDevices.getUserMedia({ video: { width: 640, height: 480 } })
    if (!videoEl.value) return
    videoEl.value.srcObject = stream.value
    await videoEl.value.play()
    await startFraming(videoEl.value)
  } catch {
    cameraError.value = 'Kamera tidak tersedia atau izinnya ditolak.'
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

async function onScan() {
  const blob = await captureJpeg()
  if (!blob) return
  scanning.value = true
  result.value = null
  try {
    result.value = await scanFace(blob)
  } catch (err: any) {
    message.error(err?.response?.data?.error ?? 'Gagal memeriksa wajah.')
  } finally {
    scanning.value = false
  }
}

onMounted(initCamera)
onUnmounted(() => stream.value?.getTracks().forEach((track) => track.stop()))
</script>

<template>
  <div class="scan-page">
    <div class="scan-heading">
      <div>
        <h2>Scan Wajah</h2>
        <p>Periksa apakah wajah sudah terdaftar di sistem.</p>
      </div>
      <a-tag color="blue"><ScanOutlined /> Pencarian data</a-tag>
    </div>

    <a-card class="scan-card">
      <div class="camera-frame" :class="`frame-${frameState}`">
        <video ref="videoEl" muted playsinline />
        <div class="camera-guide"><span /></div>
        <div class="camera-state">
          <CameraOutlined />
          {{ frameState === 'idle' ? 'Arahkan wajah ke kamera' : 'Wajah terdeteksi' }}
        </div>
        <canvas ref="canvasEl" hidden />
      </div>

      <a-alert v-if="cameraError" type="warning" show-icon :message="cameraError" />

      <div v-if="result" class="scan-result" :class="result.found ? 'is-found' : 'is-not-found'">
        <CheckCircleFilled v-if="result.found" class="result-icon" />
        <CloseCircleFilled v-else class="result-icon" />
        <div>
          <strong>{{ result.found ? 'Wajah terdaftar' : 'Wajah belum ditemukan' }}</strong>
          <p v-if="result.found">Data karyawan: {{ result.full_name }}</p>
          <p v-else>Pastikan wajah terlihat jelas atau lakukan enrollment terlebih dahulu.</p>
        </div>
      </div>

      <a-button class="scan-action" type="primary" size="large" :disabled="!canScan" :loading="scanning" @click="onScan">
        <template #icon><ScanOutlined /></template>
        Scan Wajah
      </a-button>
      <p class="scan-note">Scan ini tidak mencatat absensi dan hanya membandingkan dengan data wajah aktif.</p>
    </a-card>
  </div>
</template>

<style scoped>
.scan-page { max-width: 760px; margin: 0 auto; }
.scan-heading { display: flex; justify-content: space-between; align-items: flex-start; gap: 16px; margin-bottom: 20px; }
.scan-heading h2 { margin: 0; color: #172554; font-size: 26px; letter-spacing: -0.02em; }
.scan-heading p { margin: 6px 0 0; color: #64748b; }
.scan-card { border-color: #e2e8f0; }
.camera-frame { position: relative; overflow: hidden; aspect-ratio: 4 / 3; background: #0f172a; border: 4px solid #cbd5e1; border-radius: 12px; transition: border-color 180ms ease; }
.camera-frame.frame-detected, .camera-frame.frame-stable { border-color: #1d3a8f; }
.camera-frame video { display: block; width: 100%; height: 100%; object-fit: cover; transform: scaleX(-1); }
.camera-guide { position: absolute; inset: 13% 27%; border: 2px solid rgba(255,255,255,.75); border-radius: 48%; box-shadow: 0 0 0 999px rgba(15,23,42,.2); pointer-events: none; }
.camera-state { position: absolute; left: 16px; bottom: 16px; display: flex; align-items: center; gap: 8px; padding: 7px 11px; color: #fff; background: rgba(15,23,42,.78); border-radius: 6px; font-size: 13px; }
.scan-card :deep(.ant-alert) { margin-top: 16px; }
.scan-result { display: flex; align-items: flex-start; gap: 12px; margin-top: 20px; padding: 14px 16px; border: 1px solid; border-radius: 8px; }
.scan-result.is-found { color: #166534; background: #f0fdf4; border-color: #bbf7d0; }
.scan-result.is-not-found { color: #991b1b; background: #fef2f2; border-color: #fecaca; }
.result-icon { margin-top: 2px; font-size: 20px; }
.scan-result strong { display: block; font-size: 16px; }
.scan-result p { margin: 3px 0 0; color: inherit; opacity: .86; }
.scan-action { width: 100%; height: 48px; margin-top: 20px; font-weight: 600; }
.scan-note { margin: 12px 0 0; color: #64748b; font-size: 13px; text-align: center; }
@media (max-width: 640px) { .scan-heading { flex-direction: column; } .scan-page { max-width: none; } }
</style>

// Client-side face framing ONLY -- decides when to auto-capture a stable shot.
// It never decides identity or liveness; the server re-verifies everything
// from the captured frame (docs/PRD.md section 6.2, 10.1).
import { FaceDetector, FilesetResolver } from '@mediapipe/tasks-vision'
import { onUnmounted, ref } from 'vue'

const STABLE_MS = 700
const WASM_BASE = 'https://cdn.jsdelivr.net/npm/@mediapipe/tasks-vision@0.10.14/wasm'
const MODEL_URL =
  'https://storage.googleapis.com/mediapipe-models/face_detector/blaze_face_short_range/float16/1/blaze_face_short_range.tflite'

export type FrameState = 'idle' | 'detected' | 'stable'

export function useFaceFrame(onStable: () => void) {
  const state = ref<FrameState>('idle')
  let detector: FaceDetector | null = null
  let rafId = 0
  let stableSince = 0
  let fired = false

  async function start(video: HTMLVideoElement) {
    const fileset = await FilesetResolver.forVisionTasks(WASM_BASE)
    detector = await FaceDetector.createFromOptions(fileset, {
      baseOptions: { modelAssetPath: MODEL_URL },
      runningMode: 'VIDEO',
    })
    loop(video)
  }

  function loop(video: HTMLVideoElement) {
    rafId = requestAnimationFrame(() => loop(video))
    if (!detector || video.readyState < 2) return

    const result = detector.detectForVideo(video, performance.now())
    const hasFace = result.detections.length === 1 // reject 0 or >1 in-frame

    if (!hasFace) {
      state.value = 'idle'
      stableSince = 0
      fired = false
      return
    }

    state.value = 'detected'
    if (stableSince === 0) stableSince = performance.now()

    if (!fired && performance.now() - stableSince > STABLE_MS) {
      state.value = 'stable'
      fired = true
      onStable()
    }
  }

  function reset() {
    stableSince = 0
    fired = false
    state.value = 'idle'
  }

  function stop() {
    cancelAnimationFrame(rafId)
    detector?.close()
  }

  onUnmounted(stop)

  return { state, start, stop, reset }
}

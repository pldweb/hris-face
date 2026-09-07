import { apiClient } from './client'

export interface PhotoRejection {
  index: number
  reason: string
}

export type EnrollOutcome =
  | { ok: true }
  | { ok: false; kind: 'rejections'; rejections: PhotoRejection[] }
  | { ok: false; kind: 'duplicate' | 'error'; message: string }

// Submits the captured batch; the server re-validates every photo and is the
// sole authority on pass/fail (docs/PRD.md 6.1). A partial batch failure comes
// back as per-photo rejections so the UI can ask for a retake of just those.
export async function submitEnrollment(photos: Blob[], replace = false): Promise<EnrollOutcome> {
  const form = new FormData()
  photos.forEach((blob, i) => form.append('photos', blob, `photo-${i}.jpg`))

  try {
    await apiClient.post(replace ? '/enrollment/re-enroll' : '/enrollment/photos', form)
    return { ok: true }
  } catch (err: any) {
    const status = err?.response?.status
    if (status === 422 && err.response.data?.rejections) {
      return { ok: false, kind: 'rejections', rejections: err.response.data.rejections }
    }
    if (status === 409) {
      return { ok: false, kind: 'duplicate', message: err.response.data?.error ?? 'Duplikat terdeteksi' }
    }
    return { ok: false, kind: 'error', message: err?.response?.data?.error ?? 'Terjadi kesalahan' }
  }
}

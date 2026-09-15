import { apiClient } from './client'

export interface CheckInResult {
  employee_id: string
  full_name: string
  occurred_at: string
  type: 'check_in' | 'check_out'
  status: 'on_time' | 'late' | 'early_leave' | 'absent'
  low_confidence: boolean
}

export type CheckInErrorCode =
  | 'no_match'
  | 'wrong_person'
  | 'liveness'
  | 'no_check_in_yet'
  | 'outside_network'
  | 'outside_radius'
  | 'location_required'
  | 'device_not_approved'
  | 'challenge_required'
  | 'still_image'
  | 'challenge_failed'
  | 'unknown'

export interface Coords {
  lat: number
  lng: number
}

export class CheckInError extends Error {
  code: CheckInErrorCode
  matchedName?: string
  constructor(code: CheckInErrorCode, message: string, matchedName?: string) {
    super(message)
    this.code = code
    this.matchedName = matchedName
  }
}

// Sends one JPEG frame; the server is the sole authority on match/liveness/timestamp.
// docs/PRD.md section 6.2 -- the client never decides pass/fail on its own.
export async function submitAttendance(
  kind: 'check_in' | 'check_out',
  imageBlob: Blob,
  coords?: Coords | null,
): Promise<CheckInResult> {
  const path = kind === 'check_in' ? '/attendance/check-in' : '/attendance/check-out'
  try {
    const { data } = await apiClient.post<CheckInResult>(path, imageBlob, {
      headers: { 'Content-Type': 'image/jpeg' },
      params: coords ? { lat: coords.lat, lng: coords.lng } : undefined,
    })
    return data
  } catch (err: any) {
    const status = err?.response?.status
    const message = err?.response?.data?.error ?? 'Terjadi kesalahan'
    const matchedName = err?.response?.data?.matched_name as string | undefined
    if (matchedName) throw new CheckInError('wrong_person', message, matchedName)
    if (status === 422 && message.includes('Gunakan wajah asli')) {
      throw new CheckInError('liveness', message)
    }
    if (status === 428) throw new CheckInError('challenge_required', message)
    if (status === 422 && message.includes('Tidak terdeteksi gerakan')) {
      throw new CheckInError('still_image', message)
    }
    if (status === 422 && message.includes('Verifikasi gerakan gagal')) {
      throw new CheckInError('challenge_failed', message)
    }
    if (status === 422) throw new CheckInError('no_match', message)
    if (status === 409) throw new CheckInError('no_check_in_yet', message)
    if (status === 403 && message.includes('Perangkat')) {
      throw new CheckInError('device_not_approved', message)
    }
    if (status === 403 && message.includes('radius')) {
      throw new CheckInError('outside_radius', message)
    }
    if (status === 403 && message.includes('lokasi')) {
      throw new CheckInError('location_required', message)
    }
    if (status === 403) throw new CheckInError('outside_network', message)
    throw new CheckInError('unknown', message)
  }
}

// Multi-frame fallback used when the server answers 428: the passive anti-spoof
// model was unsure about a single frame, so it wants evidence of movement.
export async function submitChallenge(
  kind: 'check_in' | 'check_out',
  frames: Blob[],
  coords?: Coords | null,
): Promise<CheckInResult> {
  const form = new FormData()
  frames.forEach((blob, i) => form.append('frames', blob, `frame-${i}.jpg`))
  if (coords) {
    form.append('lat', String(coords.lat))
    form.append('lng', String(coords.lng))
  }
  const path = kind === 'check_in' ? '/attendance/challenge/check-in' : '/attendance/challenge/check-out'
  try {
    const { data } = await apiClient.post(path, form)
    return data.result as CheckInResult
  } catch (err: any) {
    const status = err?.response?.status
    const message = err?.response?.data?.error ?? 'Terjadi kesalahan'
    const matchedName = err?.response?.data?.matched_name as string | undefined
    if (matchedName) throw new CheckInError('wrong_person', message, matchedName)
    if (status === 422 && message.includes('Tidak terdeteksi gerakan')) {
      throw new CheckInError('still_image', message)
    }
    if (status === 422) throw new CheckInError('challenge_failed', message)
    if (status === 409) throw new CheckInError('no_check_in_yet', message)
    if (status === 403 && message.includes('radius')) {
      throw new CheckInError('outside_radius', message)
    }
    if (status === 403 && message.includes('lokasi')) {
      throw new CheckInError('location_required', message)
    }
    if (status === 403) throw new CheckInError('outside_network', message)
    throw new CheckInError('unknown', message)
  }
}

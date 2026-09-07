import axios from 'axios'
import { useAuthStore } from '../stores/auth'

export const apiClient = axios.create({
  baseURL: '/api/v1',
  withCredentials: true, // sends the httpOnly refresh-token cookie
})

apiClient.interceptors.request.use((config) => {
  const auth = useAuthStore()
  if (auth.accessToken) {
    config.headers.Authorization = `Bearer ${auth.accessToken}`
  }
  return config
})

// On a 401, try one silent refresh via the httpOnly cookie and replay the
// original request; a second 401 means the session is really over.
let refreshInFlight: Promise<string | null> | null = null

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const auth = useAuthStore()
    const original = error.config
    if (error.response?.status !== 401 || original._retried || original.url?.includes('/auth/')) {
      throw error
    }
    original._retried = true

    refreshInFlight ??= refreshAccessToken()
    const newToken = await refreshInFlight
    refreshInFlight = null

    if (!newToken) {
      auth.logout()
      throw error
    }

    original.headers.Authorization = `Bearer ${newToken}`
    return apiClient.request(original)
  },
)

async function refreshAccessToken(): Promise<string | null> {
  const auth = useAuthStore()
  try {
    const { data } = await axios.post(
      '/api/v1/auth/refresh',
      {},
      { withCredentials: true },
    )
    auth.setToken(data.access_token)
    return data.access_token as string
  } catch {
    return null
  }
}

import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref<string | null>(sessionStorage.getItem('access_token'))
  const role = ref<string | null>(sessionStorage.getItem('user_role') ?? (accessToken.value ? readRole(accessToken.value) : null))
  const fullName = ref<string | null>(null)

  function setToken(token: string) {
    accessToken.value = token
    role.value = readRole(token)
    sessionStorage.setItem('access_token', token)
    if (role.value) sessionStorage.setItem('user_role', role.value)
  }

  function logout() {
    accessToken.value = null
    role.value = null
    sessionStorage.removeItem('access_token')
    sessionStorage.removeItem('user_role')
  }

  async function logoutFully() {
    const { apiClient } = await import('../api/client')
    try {
      await apiClient.post('/auth/logout')
    } finally {
      logout()
    }
  }

  return { accessToken, role, fullName, setToken, logout, logoutFully }
})

function readRole(token: string): string | null {
  try {
    const payload = token.split('.')[1]
    const json = atob(payload.replace(/-/g, '+').replace(/_/g, '/') + '==')
    return JSON.parse(json).role ?? null
  } catch {
    return null
  }
}

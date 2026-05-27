import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api, setAuthToken, getAuthToken } from '@/api/client'
import type { UserProfile } from '@shared/types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<UserProfile | null>(null)
  const token = ref<string | null>(getAuthToken())
  const loading = ref(false)

  const isLoggedIn = computed(() => !!user.value && !!token.value)
  const userId = computed(() => user.value?.id ?? null)
  const userName = computed(() => user.value?.name ?? '')

  async function login(email: string, password: string) {
    loading.value = true
    try {
      const resp = await api<{ data: { token: string; user: UserProfile } }>('/auth/login', {
        method: 'POST',
        body: { email, password },
      })
      token.value = resp.data.token
      user.value = resp.data.user
      setAuthToken(resp.data.token)
    } finally {
      loading.value = false
    }
  }

  async function fetchProfile() {
    if (!token.value) return
    try {
      const resp = await api<{ data: UserProfile }>('/users/me')
      user.value = resp.data
    } catch {
      logout()
    }
  }

  function logout() {
    user.value = null
    token.value = null
    setAuthToken(null)
  }

  return { user, token, loading, isLoggedIn, userId, userName, login, fetchProfile, logout }
})

import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { login as apiLogin, register as apiRegister, getProfile } from '@/api/user'
import { clearToken, getToken, setToken } from '@/utils/storage'
import type { UserInfo } from '@/constants/user'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(getToken())
  const user = ref<UserInfo | null>(null)

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  async function login(username: string, password: string) {
    const res = await apiLogin({ username, password })
    token.value = res.token
    user.value = res.user
    setToken(res.token)
  }

  async function register(payload: { username: string; email: string; password: string; bio?: string }) {
    const res = await apiRegister(payload)
    token.value = res.token
    user.value = res.user
    setToken(res.token)
  }

  async function fetchProfile() {
    if (!token.value) return
    user.value = await getProfile()
  }

  function logout() {
    token.value = ''
    user.value = null
    clearToken()
  }

  return { token, user, isLoggedIn, isAdmin, login, register, fetchProfile, logout }
})

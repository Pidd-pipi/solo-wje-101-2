import { computed } from 'vue'
import { useUserStore } from '@/stores/useUserStore'

export function useAuth() {
  const store = useUserStore()
  const isLoggedIn = computed(() => store.isLoggedIn)
  const isAdmin = computed(() => store.isAdmin)
  const user = computed(() => store.user)
  return { store, isLoggedIn, isAdmin, user }
}

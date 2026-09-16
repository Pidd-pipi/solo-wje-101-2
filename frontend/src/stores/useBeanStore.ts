import { defineStore } from 'pinia'
import { ref } from 'vue'
import { listBeans } from '@/api/bean'
import { favoriteBean, unfavoriteBean } from '@/api/favorite'
import type { CoffeeBean } from '@/constants/bean'

export const useBeanStore = defineStore('bean', () => {
  const beans = ref<CoffeeBean[]>([])
  const total = ref(0)
  // Monotonic token: only the most recent load() may commit, so fast consecutive
  // filters/searches/resets cannot let a slow stale response overwrite fresh data.
  let loadSeq = 0

  async function load(params: { page?: number; page_size?: number; origin?: string; process?: string; keyword?: string } = {}) {
    const seq = ++loadSeq
    const res = await listBeans(params)
    // Drop responses from a superseded request entirely (no half/stale page).
    if (seq !== loadSeq) return
    beans.value = res.list
    total.value = res.total
  }

  // Toggle favorite and take the server-returned state as truth, so a duplicate
  // favorite or a cancel always reflects back on the card.
  async function toggleFavorite(bean: CoffeeBean) {
    const favored = !!bean.is_favored
    const res = favored
      ? await unfavoriteBean(bean.id)
      : await favoriteBean(bean.id)
    const target = beans.value.find((b) => b.id === bean.id)
    if (target) target.is_favored = res.favored
    return res.favored
  }

  return { beans, total, load, toggleFavorite }
})

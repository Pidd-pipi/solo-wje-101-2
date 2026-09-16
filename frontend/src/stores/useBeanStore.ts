import { defineStore } from 'pinia'
import { ref } from 'vue'
import { listBeans } from '@/api/bean'
import type { CoffeeBean } from '@/constants/bean'

export const useBeanStore = defineStore('bean', () => {
  const beans = ref<CoffeeBean[]>([])
  const total = ref(0)

  async function load(params: { page?: number; page_size?: number; origin?: string; process?: string; keyword?: string } = {}) {
    const res = await listBeans(params)
    beans.value = res.list
    total.value = res.total
  }

  return { beans, total, load }
})

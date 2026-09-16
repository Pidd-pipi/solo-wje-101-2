import { computed, ref } from 'vue'

export function usePagination(pageSize = 10) {
  const page = ref(1)
  const total = ref(0)

  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))

  function setPage(p: number) {
    page.value = p
  }

  return { page, total, totalPages, pageSize, setPage }
}

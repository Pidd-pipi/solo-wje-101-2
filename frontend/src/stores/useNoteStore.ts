import { defineStore } from 'pinia'
import { ref } from 'vue'
import { listNotes } from '@/api/note'
import type { NoteItem } from '@/constants/note'

export const useNoteStore = defineStore('note', () => {
  const items = ref<NoteItem[]>([])
  const total = ref(0)

  async function load(params: { page?: number; page_size?: number; roast?: string; sort?: string } = {}) {
    const res = await listNotes(params)
    items.value = res.list
    total.value = res.total
  }

  return { items, total, load }
})

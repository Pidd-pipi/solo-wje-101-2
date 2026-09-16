<template>
  <div class="flavor-tags">
    <el-tag v-for="t in tags" :key="t" size="small" effect="plain" class="tag">{{ t }}</el-tag>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { parseTags } from '@/constants/note'

const props = defineProps<{ tags: string }>()
const tags = computed(() => parseTags(props.tags))
</script>

<style scoped>
.flavor-tags { display: flex; flex-wrap: wrap; gap: 6px; }
.tag { background: #fdf3e3; border-color: #e8c8a0; color: #7b4b2a; }
</style>

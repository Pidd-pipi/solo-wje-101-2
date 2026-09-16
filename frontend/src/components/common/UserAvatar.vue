<template>
  <el-avatar :size="size" :src="src" class="avatar">{{ initial }}</el-avatar>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ name?: string; src?: string; size?: number }>()
const initial = computed(() => (props.name || '咖').slice(0, 1))
</script>

<style scoped>
.avatar { background: #7b4b2a; color: #fff; font-weight: 700; }
</style>

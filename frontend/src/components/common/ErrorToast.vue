<template>
  <div v-if="visible" class="error-toast">{{ message }}</div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{ message: string; visible: boolean }>()
const shown = ref(props.visible)
watch(() => props.visible, (v) => {
  shown.value = v
})
</script>

<style scoped>
.error-toast { position: fixed; top: 20px; left: 50%; transform: translateX(-50%); background: #f56c6c; color: #fff; padding: 10px 20px; border-radius: 6px; z-index: 9999; }
</style>

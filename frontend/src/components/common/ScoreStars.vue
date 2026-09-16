<template>
  <div class="score-stars">
    <el-rate :model-value="modelValue" :max="10" disabled allow-half class="rate" />
    <span class="score">{{ modelValue?.toFixed(1) }}</span>
  </div>
</template>

<script setup lang="ts">
defineProps<{ modelValue?: number }>()
</script>

<style scoped>
.score-stars { display: inline-flex; align-items: center; gap: 8px; }
.score { font-weight: 700; color: #7b4b2a; }
</style>

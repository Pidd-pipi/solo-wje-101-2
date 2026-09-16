<template>
  <el-empty :description="description">
    <el-button v-if="actionText" type="primary" @click="$emit('action')">{{ actionText }}</el-button>
  </el-empty>
</template>

<script setup lang="ts">
defineProps<{ description?: string; actionText?: string }>()
defineEmits<{ (e: 'action'): void }>()
</script>

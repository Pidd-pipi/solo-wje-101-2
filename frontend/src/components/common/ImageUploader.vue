<template>
  <div class="uploader">
    <el-upload :show-file-list="false" :http-request="doUpload" accept="image/*">
      <el-button :loading="loading" type="primary" plain>上传图片</el-button>
    </el-upload>
    <el-image v-if="modelValue" :src="modelValue" fit="cover" class="preview" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import axios from 'axios'
import { getToken } from '@/utils/storage'

defineProps<{ modelValue: string }>()
const emit = defineEmits<{ (e: 'update:modelValue', url: string): void }>()
const loading = ref(false)

async function doUpload(option: { file: File }) {
  const form = new FormData()
  form.append('file', option.file)
  loading.value = true
  try {
    const res = await axios.post('/api/v1/uploads', form, {
      headers: { Authorization: `Bearer ${getToken()}`, 'Content-Type': 'multipart/form-data' },
    })
    emit('update:modelValue', res.data.data.url)
    ElMessage.success('上传成功')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.uploader { display: flex; gap: 12px; align-items: center; }
.preview { width: 120px; height: 90px; border-radius: 6px; }
</style>

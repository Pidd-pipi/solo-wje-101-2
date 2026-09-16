<template>
  <div class="page">
    <h1>豆种库</h1>
    <SearchFilter @search="onSearch" @reset="onReset">
      <template #filters>
        <el-form-item label="产地">
          <el-select v-model="origin" clearable placeholder="全部产地" style="width: 150px" @change="load">
            <el-option v-for="o in ORIGINS" :key="o" :label="o" :value="o" />
          </el-select>
        </el-form-item>
        <el-form-item label="处理法">
          <el-select v-model="process" clearable placeholder="全部处理法" style="width: 150px" @change="load">
            <el-option v-for="(label, value) in ProcessMethodMap" :key="value" :label="label" :value="value" />
          </el-select>
        </el-form-item>
      </template>
    </SearchFilter>
    <el-row :gutter="16">
      <el-col v-for="b in beans" :key="b.id" :xs="24" :sm="12" :md="8">
        <el-card class="bean-card" shadow="hover">
          <div class="card-head">
            <h3>{{ b.name }} <el-tag size="small" type="warning">{{ ProcessMethodMap[b.process_method] }}</el-tag></h3>
            <el-tooltip v-if="isLoggedIn" :content="b.is_favored ? '取消收藏' : '收藏豆种'" placement="top">
              <el-button
                circle
                size="small"
                :type="b.is_favored ? 'warning' : 'default'"
                :loading="pendingIds.has(b.id)"
                @click="toggleFav(b)"
              >
                <el-icon><StarFilled v-if="b.is_favored" /><Star v-else /></el-icon>
              </el-button>
            </el-tooltip>
          </div>
          <div class="meta">{{ b.origin || '-' }}</div>
          <FlavorTags :tags="b.flavor_tags" />
          <p class="desc">{{ b.description }}</p>
          <el-button v-if="isAdmin" size="small" type="danger" plain @click="removeBean(b.id)">删除</el-button>
        </el-card>
      </el-col>
    </el-row>
    <EmptyState v-if="!beans.length" description="暂无豆种" />
    <el-button v-if="isAdmin" type="primary" style="margin-top: 16px" @click="showAdd = true">新增豆种</el-button>
    <el-dialog v-model="showAdd" title="新增豆种" width="480px">
      <el-form label-width="80px">
        <el-form-item label="名称"><el-input v-model="addForm.name" /></el-form-item>
        <el-form-item label="产地"><el-input v-model="addForm.origin" /></el-form-item>
        <el-form-item label="处理法">
          <el-select v-model="addForm.process_method" style="width: 200px">
            <el-option v-for="(label, value) in ProcessMethodMap" :key="value" :label="label" :value="value" />
          </el-select>
        </el-form-item>
        <el-form-item label="风味标签"><el-input v-model="addForm.flavor_tags" placeholder='如 ["坚果","焦糖"]' /></el-form-item>
        <el-form-item label="描述"><el-input v-model="addForm.description" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAdd = false">取消</el-button>
        <el-button type="primary" @click="addBean">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Star, StarFilled } from '@element-plus/icons-vue'
import SearchFilter from '@/components/common/SearchFilter.vue'
import FlavorTags from '@/components/common/FlavorTags.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { useBeanStore } from '@/stores/useBeanStore'
import { useAuth } from '@/hooks/useAuth'
import { createBean, deleteBean } from '@/api/bean'
import { ProcessMethodMap, type ProcessMethod, type CoffeeBean } from '@/constants/bean'

const store = useBeanStore()
const router = useRouter()
const { isLoggedIn, isAdmin } = useAuth()
const beans = computed(() => store.beans)
const origin = ref('')
const process = ref('')
const keyword = ref('')
const showAdd = ref(false)
const addForm = reactive({ name: '', origin: '', process_method: 'washed', flavor_tags: '[]', description: '' })
const pendingIds = ref(new Set<number>())

const ORIGINS = ['埃塞俄比亚', '哥伦比亚', '哥斯达黎加', '印度尼西亚']

onMounted(() => load())

async function load() {
  await store.load({ page: 1, page_size: 20, origin: origin.value, process: process.value, keyword: keyword.value })
}
function onSearch(kw: string) {
  keyword.value = kw
  load()
}
function onReset() {
  origin.value = ''
  process.value = ''
  keyword.value = ''
  load()
}
async function addBean() {
  if (!addForm.name) {
    ElMessage.warning('请填写名称')
    return
  }
  await createBean({ ...addForm, process_method: addForm.process_method as ProcessMethod })
  ElMessage.success('豆种已添加')
  showAdd.value = false
  await load()
}
async function removeBean(id: number) {
  await deleteBean(id)
  ElMessage.success('已下架豆种')
  await load()
}
async function toggleFav(bean: CoffeeBean) {
  if (!isLoggedIn.value) {
    ElMessage.warning('请先登录后再收藏豆种')
    router.push('/login')
    return
  }
  if (pendingIds.value.has(bean.id)) return
  pendingIds.value.add(bean.id)
  try {
    const favored = await store.toggleFavorite(bean)
    ElMessage.success(favored ? '已收藏豆种' : '已取消收藏')
  } finally {
    pendingIds.value.delete(bean.id)
  }
}
</script>

<style scoped>
.page { max-width: 1200px; margin: 0 auto; }
.bean-card { margin-bottom: 16px; }
.card-head { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.card-head h3 { margin: 0; }
.meta { color: #999; font-size: 12px; margin: 6px 0; }
.desc { color: #666; margin-top: 8px; }
</style>

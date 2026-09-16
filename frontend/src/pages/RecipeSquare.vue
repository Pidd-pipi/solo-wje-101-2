<template>
  <div class="page">
    <h1>配方广场</h1>
    <SearchFilter @search="onSearch" @reset="onReset">
      <template #filters>
        <el-form-item label="器具">
          <el-select v-model="device" clearable placeholder="全部器具" style="width: 160px" @change="load">
            <el-option v-for="d in DEVICES" :key="d" :label="d" :value="d" />
          </el-select>
        </el-form-item>
      </template>
    </SearchFilter>
    <el-row :gutter="16">
      <el-col v-for="r in recipes" :key="r.id" :xs="24" :sm="12" :md="8">
        <el-card class="recipe-card" shadow="hover">
          <h3>{{ r.name }} <el-tag size="small">{{ r.device }}</el-tag></h3>
          <div class="meta">{{ r.water_temp }}°C · {{ r.grind_size }} · 粉水比 {{ r.ratio }}</div>
          <ol>
            <li v-for="s in stepsOf(r)" :key="s.step_number">
              第{{ s.step_number }}步：{{ s.description }}（{{ s.duration_seconds }}s）
            </li>
          </ol>
        </el-card>
      </el-col>
    </el-row>
    <EmptyState v-if="!recipes.length" description="暂无配方" />
    <el-button v-if="isLoggedIn" type="primary" style="margin-top: 16px" @click="showAdd = true">分享我的配方</el-button>
    <el-dialog v-model="showAdd" title="分享冲煮配方" width="520px">
      <el-form label-width="80px">
        <el-form-item label="名称"><el-input v-model="addForm.name" /></el-form-item>
        <el-form-item label="器具">
          <el-select v-model="addForm.device" style="width: 200px">
            <el-option v-for="d in DEVICES" :key="d" :label="d" :value="d" />
          </el-select>
        </el-form-item>
        <el-form-item label="水温"><el-input-number v-model="addForm.water_temp" :min="80" :max="100" /></el-form-item>
        <el-form-item label="研磨度"><el-input v-model="addForm.grind_size" /></el-form-item>
        <el-form-item label="粉水比"><el-input v-model="addForm.ratio" placeholder="1:15" /></el-form-item>
        <el-form-item label="步骤(JSON)"><el-input v-model="addForm.steps" type="textarea" :rows="4" placeholder='[{"step_number":1,"description":"闷蒸","duration_seconds":30}]' /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAdd = false">取消</el-button>
        <el-button type="primary" @click="addRecipe">发布</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import SearchFilter from '@/components/common/SearchFilter.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { listRecipes, createRecipe } from '@/api/recipe'
import { useAuth } from '@/hooks/useAuth'
import type { BrewRecipe, RecipeStep } from '@/types/api'

const { isLoggedIn } = useAuth()
const recipes = ref<BrewRecipe[]>([])
const device = ref('')
const keyword = ref('')
const showAdd = ref(false)
const addForm = reactive({ name: '', device: '手冲壶', water_temp: 92, grind_size: '中细', ratio: '1:15', steps: '[]' })

const DEVICES = ['手冲壶', '法压壶', '意式机', '爱乐压', '冷萃壶']

onMounted(() => load())

async function load() {
  const res = await listRecipes({ page: 1, page_size: 20, device: device.value, keyword: keyword.value })
  recipes.value = res.list
}
function onSearch(kw: string) {
  keyword.value = kw
  load()
}
function onReset() {
  device.value = ''
  keyword.value = ''
  load()
}
function stepsOf(r: BrewRecipe): RecipeStep[] {
  try {
    return JSON.parse(r.steps || '[]')
  } catch {
    return []
  }
}
async function addRecipe() {
  if (!addForm.name) {
    ElMessage.warning('请填写配方名称')
    return
  }
  await createRecipe({ ...addForm })
  ElMessage.success('配方已分享')
  showAdd.value = false
  await load()
}
</script>

<style scoped>
.page { max-width: 1200px; margin: 0 auto; }
.recipe-card { margin-bottom: 16px; }
.meta { color: #999; font-size: 12px; margin: 6px 0; }
</style>

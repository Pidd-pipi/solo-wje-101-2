<template>
  <div class="home">
    <section class="hero">
      <h1>记录每一杯咖啡的独特风味</h1>
      <p class="sub">品鉴笔记 · 豆种库 · 冲煮配方 · 咖啡爱好者社区</p>
      <el-input v-model="keyword" size="large" placeholder="搜索咖啡名称 / 产地" class="search" @keyup.enter="doSearch">
        <template #append><el-button @click="doSearch">搜索</el-button></template>
      </el-input>
      <div class="filters">
        <el-radio-group v-model="roast" @change="load">
          <el-radio-button value="">全部烘焙度</el-radio-button>
          <el-radio-button v-for="(label, value) in RoastLevelMap" :key="value" :value="value">{{ label }}</el-radio-button>
        </el-radio-group>
        <el-radio-group v-model="sort" @change="load" class="sort">
          <el-radio-button value="new">最新</el-radio-button>
          <el-radio-button value="hot">热门</el-radio-button>
        </el-radio-group>
      </div>
    </section>

    <el-row :gutter="16">
      <el-col v-for="item in items" :key="item.note.id" :xs="24" :sm="12" :md="8">
        <el-card class="note-card" shadow="hover" @click="$router.push(`/note/${item.note.id}`)">
          <el-image v-if="item.note.image_url" :src="item.note.image_url" fit="cover" class="cover" lazy />
          <div class="body">
            <h3>{{ item.note.coffee_name }}</h3>
            <div class="meta">{{ item.note.origin || '-' }} · {{ RoastLevelMap[item.note.roast_level] }} · {{ item.note.brew_method || '-' }}</div>
            <ScoreStars :model-value="item.note.overall_score" />
            <FlavorTags :tags="item.note.flavor_tags" />
            <div class="foot">👍 {{ item.like_count }} · {{ formatDate(item.note.created_at) }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    <EmptyState v-if="!items.length && !loading" description="暂无品鉴笔记" action-text="写第一篇笔记" @action="$router.push('/note/create')" />
    <el-pagination v-if="total > 0" layout="prev, pager, next" :total="total" :page-size="pageSize" :current-page="page" @current-change="onPage" class="pager" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import ScoreStars from '@/components/common/ScoreStars.vue'
import FlavorTags from '@/components/common/FlavorTags.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { useNoteStore } from '@/stores/useNoteStore'
import { RoastLevelMap } from '@/constants/note'
import { formatDate } from '@/utils/dateFormat'

const store = useNoteStore()
const router = useRouter()
const items = computed(() => store.items)
const total = computed(() => store.total)
const page = ref(1)
const pageSize = 9
const roast = ref('')
const sort = ref('new')
const keyword = ref('')
const loading = ref(false)

onMounted(() => load())

async function load() {
  loading.value = true
  try {
    await store.load({ page: page.value, page_size: pageSize, roast: roast.value, sort: sort.value })
  } finally {
    loading.value = false
  }
}
function doSearch() {
  if (keyword.value) {
    router.push({ path: '/beans', query: { keyword: keyword.value } })
  }
}
function onPage(p: number) {
  page.value = p
  load()
}
</script>

<style scoped>
.home { max-width: 1200px; margin: 0 auto; }
.hero { text-align: center; padding: 32px 0 16px; }
.hero h1 { color: #7b4b2a; }
.sub { color: #999; }
.search { max-width: 560px; margin: 16px auto; }
.filters { margin: 12px auto; display: flex; justify-content: center; gap: 16px; }
.sort { margin-left: 8px; }
.note-card { margin-bottom: 16px; cursor: pointer; }
.cover { width: 100%; height: 160px; border-radius: 6px; }
.body { padding-top: 8px; }
.meta { color: #999; font-size: 12px; margin: 6px 0; }
.foot { color: #aaa; font-size: 12px; margin-top: 8px; }
.pager { margin-top: 16px; justify-content: center; }
</style>

<template>
  <div class="page" v-if="note">
    <el-page-header @back="$router.back()" content="品鉴详情" />
    <el-row :gutter="16">
      <el-col :xs="24" :md="14">
        <el-card>
          <el-image v-if="note.image_url" :src="note.image_url" fit="cover" class="cover" />
          <h1>{{ note.coffee_name }}</h1>
          <div class="meta">{{ note.origin || '-' }} · {{ RoastLevelMap[note.roast_level] }} · {{ note.brew_method || '-' }}</div>
          <ScoreStars :model-value="note.overall_score" />
          <FlavorTags :tags="note.flavor_tags" />
          <el-descriptions :column="2" border class="scores">
            <el-descriptions-item label="香气">{{ note.aroma_score }}</el-descriptions-item>
            <el-descriptions-item label="酸质">{{ note.acidity_score }}</el-descriptions-item>
            <el-descriptions-item label="醇厚">{{ note.body_score }}</el-descriptions-item>
            <el-descriptions-item label="综合">{{ note.overall_score }}</el-descriptions-item>
          </el-descriptions>
          <p class="notes">{{ note.notes_text }}</p>
          <div class="actions">
            <el-button :type="liked ? 'warning' : 'default'" :loading="liking" @click="toggleLike">
              👍 {{ likeCount }}
            </el-button>
            <el-button v-if="isOwner" type="danger" plain @click="remove">删除</el-button>
          </div>
        </el-card>
        <el-card v-if="recipe" class="block">
          <template #header>关联配方：{{ recipe.name }}</template>
          <p>{{ recipe.device }} · {{ recipe.water_temp }}°C · {{ recipe.grind_size }} · 粉水比 {{ recipe.ratio }}</p>
          <ol>
            <li v-for="s in steps" :key="s.step_number">
              第{{ s.step_number }}步：{{ s.description }}（{{ s.duration_seconds }}s）
            </li>
          </ol>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="10">
        <el-card>
          <template #header>评论（{{ comments.length }}）</template>
          <div v-for="c in comments" :key="c.id" class="comment">
            <div class="c-head">用户 #{{ c.user_id }} · {{ formatDateTime(c.created_at) }}</div>
            <div>{{ c.content }}</div>
          </div>
          <el-empty v-if="!comments.length" description="暂无评论" />
          <div class="reply">
            <el-input v-model="reply" type="textarea" :rows="3" placeholder="写下你的评论…" />
            <el-button type="primary" :loading="replying" @click="submitReply">发表评论</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import ScoreStars from '@/components/common/ScoreStars.vue'
import FlavorTags from '@/components/common/FlavorTags.vue'
import { getNote, listComments, createComment, likeNote, unlikeNote, deleteNote } from '@/api/note'
import { getRecipe } from '@/api/recipe'
import { useAuth } from '@/hooks/useAuth'
import { RoastLevelMap, type TastingNote } from '@/constants/note'
import type { Comment, BrewRecipe, RecipeStep } from '@/types/api'
import { formatDateTime } from '@/utils/dateFormat'

const route = useRoute()
const router = useRouter()
const { isLoggedIn, user } = useAuth()
const note = ref<TastingNote | null>(null)
const likeCount = ref(0)
const liked = ref(false)
const liking = ref(false)
const comments = ref<Comment[]>([])
const reply = ref('')
const replying = ref(false)
const recipe = ref<BrewRecipe | null>(null)

const steps = computed<RecipeStep[]>(() => {
  try {
    return JSON.parse(recipe.value?.steps || '[]')
  } catch {
    return []
  }
})
const isOwner = computed(() => !!user.value && note.value?.user_id === user.value.id)

onMounted(async () => {
  const id = route.params.id as string
  const res = await getNote(id)
  note.value = res.note
  likeCount.value = res.like_count
  comments.value = await listComments(res.note.id)
  if (res.note.brew_recipe_id) {
    try {
      recipe.value = await getRecipe(res.note.brew_recipe_id)
    } catch {
      recipe.value = null
    }
  }
})

async function toggleLike() {
  if (!isLoggedIn.value) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }
  liking.value = true
  try {
    if (liked.value) {
      await unlikeNote(note.value!.id)
      liked.value = false
      likeCount.value = Math.max(0, likeCount.value - 1)
    } else {
      await likeNote(note.value!.id)
      liked.value = true
      likeCount.value += 1
    }
  } finally {
    liking.value = false
  }
}

async function submitReply() {
  if (!isLoggedIn.value) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }
  if (!reply.value.trim()) return
  replying.value = true
  try {
    await createComment(note.value!.id, reply.value)
    comments.value = await listComments(note.value!.id)
    reply.value = ''
  } finally {
    replying.value = false
  }
}

async function remove() {
  await deleteNote(note.value!.id)
  ElMessage.success('笔记已删除')
  router.push('/')
}
</script>

<style scoped>
.page { max-width: 1000px; margin: 0 auto; }
.cover { width: 100%; max-height: 360px; border-radius: 8px; }
.meta { color: #999; margin: 8px 0; }
.scores { margin-top: 12px; }
.notes { line-height: 1.8; margin-top: 12px; }
.actions { margin-top: 16px; display: flex; gap: 12px; }
.block { margin-top: 16px; }
.comment { border-bottom: 1px solid #f0f0f0; padding: 10px 0; }
.c-head { color: #999; font-size: 12px; }
.reply { margin-top: 12px; display: flex; flex-direction: column; gap: 10px; }
</style>

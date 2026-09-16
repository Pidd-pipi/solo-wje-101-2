<template>
  <div class="page" v-if="data">
    <el-card class="head">
      <div class="user-line">
        <UserAvatar :name="data.user.username" :size="64" />
        <div>
          <h2>{{ data.user.username }} <el-tag size="small">{{ data.user.role === 'admin' ? '管理员' : '咖啡爱好者' }}</el-tag></h2>
          <p class="bio">{{ data.user.bio || '这个人很懒，什么都没写' }}</p>
        </div>
      </div>
      <div class="stats">
        <div class="stat"><b>{{ data.note_count }}</b><span>品鉴次数</span></div>
        <div class="stat"><b>{{ data.avg_score.toFixed(1) }}</b><span>平均分</span></div>
        <div class="stat"><b>{{ data.favorite_count }}</b><span>收藏豆种</span></div>
        <div class="stat"><b>{{ data.likes_received }}</b><span>收到点赞</span></div>
        <div class="stat"><b>{{ data.followers }}</b><span>粉丝</span></div>
        <div class="stat"><b>{{ data.following }}</b><span>关注</span></div>
      </div>
      <div class="origins" v-if="data.top_origins.length">
        最爱产地 TOP3：<el-tag v-for="o in data.top_origins" :key="o" size="small" class="origin-tag">{{ o }}</el-tag>
      </div>
      <el-button v-if="isLoggedIn && user?.id !== Number($route.params.id)" :type="following ? 'default' : 'primary'" @click="toggleFollow">
        {{ following ? '已关注' : '关注' }}
      </el-button>
    </el-card>

    <el-card class="pref-card">
      <h3>偏好画像</h3>
      <el-row :gutter="16">
        <el-col :xs="24" :sm="8">
          <h4>烘焙偏好（品鉴）</h4>
          <div v-if="data.preference.roast.length" class="bars">
            <div v-for="r in data.preference.roast" :key="r.key" class="bar-row">
              <span class="bar-label">{{ r.label }}</span>
              <el-progress :percentage="barPercent(data.preference.roast, r.count)" :stroke-width="10" :show-text="true" />
            </div>
          </div>
          <EmptyState v-else description="暂无数据" />
        </el-col>
        <el-col :xs="24" :sm="8">
          <h4>处理法偏好（收藏）</h4>
          <div v-if="data.preference.process.length" class="bars">
            <div v-for="p in data.preference.process" :key="p.key" class="bar-row">
              <span class="bar-label">{{ p.label }}</span>
              <el-progress :percentage="barPercent(data.preference.process, p.count)" :stroke-width="10" status="warning" :show-text="true" />
            </div>
          </div>
          <EmptyState v-else description="暂无数据" />
        </el-col>
        <el-col :xs="24" :sm="8">
          <h4>风味偏好（收藏+品鉴）</h4>
          <div v-if="data.preference.flavor.length" class="flavor-cloud">
            <el-tag
              v-for="f in data.preference.flavor"
              :key="f.key"
              :type="f.count >= flavorMax ? 'danger' : 'success'"
              class="flavor-tag"
            >{{ f.label }} · {{ f.count }}</el-tag>
          </div>
          <EmptyState v-else description="暂无数据" />
        </el-col>
      </el-row>
    </el-card>

    <el-card class="fav-card" v-if="data.recent_favorites.length">
      <div class="section-head">
        <h3>最近收藏</h3>
        <el-tag type="warning" effect="plain">共 {{ data.favorite_count }} 个豆种</el-tag>
      </div>
      <el-row :gutter="16">
        <el-col v-for="f in data.recent_favorites" :key="f.id" :xs="24" :sm="12" :md="8">
          <div class="fav-item" @click="$router.push('/beans')">
            <el-icon class="fav-star"><StarFilled /></el-icon>
            <div class="fav-info">
              <div class="fav-name">{{ f.name }}</div>
              <div class="meta">{{ f.origin || '-' }} · {{ ProcessMethodMap[f.process_method] || f.process_method }}</div>
              <div class="meta">收藏于 {{ formatDateTime(f.favored_at) }}</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <h3>品鉴历史</h3>
    <el-row :gutter="16">
      <el-col v-for="n in data.notes" :key="n.id" :xs="24" :sm="12" :md="8">
        <el-card class="note-card" shadow="hover" @click="$router.push(`/note/${n.id}`)">
          <h4>{{ n.coffee_name }}</h4>
          <div class="meta">{{ n.origin }} · {{ RoastLevelMap[n.roast_level] }}</div>
          <ScoreStars :model-value="n.overall_score" />
        </el-card>
      </el-col>
    </el-row>
    <EmptyState v-if="!data.notes.length" description="暂无品鉴记录" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { StarFilled } from '@element-plus/icons-vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import ScoreStars from '@/components/common/ScoreStars.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { getUserProfile, followUser, unfollowUser } from '@/api/user'
import { useAuth } from '@/hooks/useAuth'
import { RoastLevelMap } from '@/constants/note'
import { ProcessMethodMap, type PreferenceItem } from '@/constants/bean'
import { formatDateTime } from '@/utils/dateFormat'
import type { ProfileData } from '@/api/user'

const route = useRoute()
const { isLoggedIn, user } = useAuth()
const data = ref<ProfileData | null>(null)
const following = ref(false)

const flavorMax = computed(() => Math.max(1, ...(data.value?.preference.flavor.map((f) => f.count) ?? [1])))

function barPercent(items: PreferenceItem[], count: number): number {
  const max = Math.max(1, ...items.map((i) => i.count))
  return Math.round((count / max) * 100)
}

onMounted(async () => {
  data.value = await getUserProfile(route.params.id as string)
})

async function toggleFollow() {
  if (!isLoggedIn.value) {
    ElMessage.warning('请先登录')
    return
  }
  if (following.value) {
    await unfollowUser(data.value!.user.id)
    following.value = false
    ElMessage.success('已取消关注')
  } else {
    await followUser(data.value!.user.id)
    following.value = true
    ElMessage.success('关注成功')
  }
}
</script>

<style scoped>
.page { max-width: 1000px; margin: 0 auto; }
.head { margin-bottom: 20px; }
.user-line { display: flex; gap: 16px; align-items: center; }
.bio { color: #999; }
.stats { display: flex; gap: 32px; margin: 16px 0; flex-wrap: wrap; }
.stat { display: flex; flex-direction: column; }
.stat b { font-size: 22px; color: #7b4b2a; }
.stat span { color: #999; font-size: 12px; }
.origins { margin: 12px 0; }
.origin-tag { margin-right: 6px; }
.pref-card, .fav-card { margin-bottom: 20px; }
.bars { display: flex; flex-direction: column; gap: 10px; }
.bar-row { display: flex; align-items: center; gap: 8px; }
.bar-label { width: 44px; color: #666; font-size: 13px; flex-shrink: 0; }
.flavor-cloud { display: flex; flex-wrap: wrap; gap: 8px; }
.flavor-tag { margin: 0; }
.section-head { display: flex; align-items: center; justify-content: space-between; }
.fav-item { display: flex; gap: 10px; align-items: flex-start; padding: 10px; border: 1px solid #f0e6dd; border-radius: 8px; margin-bottom: 12px; cursor: pointer; }
.fav-item:hover { background: #fdf8f3; }
.fav-star { color: #e6a23c; margin-top: 3px; }
.fav-name { font-weight: 600; }
.note-card { margin-bottom: 16px; cursor: pointer; }
.meta { color: #999; font-size: 12px; }
</style>

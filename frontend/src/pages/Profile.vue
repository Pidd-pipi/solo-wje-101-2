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
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import UserAvatar from '@/components/common/UserAvatar.vue'
import ScoreStars from '@/components/common/ScoreStars.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { getUserProfile, followUser, unfollowUser } from '@/api/user'
import { useAuth } from '@/hooks/useAuth'
import { RoastLevelMap } from '@/constants/note'
import type { ProfileData } from '@/api/user'

const route = useRoute()
const { isLoggedIn, user } = useAuth()
const data = ref<ProfileData | null>(null)
const following = ref(false)

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
.stats { display: flex; gap: 32px; margin: 16px 0; }
.stat { display: flex; flex-direction: column; }
.stat b { font-size: 22px; color: #7b4b2a; }
.stat span { color: #999; font-size: 12px; }
.origins { margin: 12px 0; }
.origin-tag { margin-right: 6px; }
.note-card { margin-bottom: 16px; cursor: pointer; }
.meta { color: #999; font-size: 12px; }
</style>

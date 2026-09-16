<template>
  <el-container class="shell">
    <el-header class="header">
      <div class="brand" @click="$router.push('/')">☕ 咖啡品鉴社区</div>
      <el-menu mode="horizontal" :ellipsis="false" router :default-active="$route.path">
        <el-menu-item index="/">首页</el-menu-item>
        <el-menu-item index="/note/create">写品鉴笔记</el-menu-item>
        <el-menu-item index="/beans">豆种库</el-menu-item>
        <el-menu-item index="/recipes">配方广场</el-menu-item>
      </el-menu>
      <div class="user-area">
        <template v-if="isLoggedIn">
          <el-dropdown @command="onCommand">
            <span class="user-name">
              <UserAvatar :name="user?.username" :size="24" />
              {{ user?.username }}
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">我的主页</el-dropdown-item>
                <el-dropdown-item command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
        <el-button v-else type="primary" size="small" @click="$router.push('/login')">登录/注册</el-button>
      </div>
    </el-header>
    <el-main><router-view /></el-main>
  </el-container>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/useUserStore'
import { useAuth } from '@/hooks/useAuth'
import UserAvatar from '@/components/common/UserAvatar.vue'

const store = useUserStore()
const { isLoggedIn, user } = useAuth()
const router = useRouter()

function onCommand(cmd: string) {
  if (cmd === 'profile') {
    router.push(`/profile/${user.value?.id}`)
  } else if (cmd === 'logout') {
    store.logout()
    router.push('/')
  }
}
</script>

<style>
html, body, #app { margin: 0; height: 100%; background: #f7f1e8; font-family: "PingFang SC", "Microsoft YaHei", sans-serif; }
.shell { min-height: 100vh; }
.header { display: flex; align-items: center; gap: 16px; background: #fff; box-shadow: 0 2px 8px rgba(0,0,0,.06); }
.brand { font-size: 20px; font-weight: 700; color: #7b4b2a; cursor: pointer; white-space: nowrap; }
.user-area { margin-left: auto; }
.user-name { cursor: pointer; display: inline-flex; align-items: center; gap: 6px; font-weight: 600; color: #7b4b2a; }
</style>

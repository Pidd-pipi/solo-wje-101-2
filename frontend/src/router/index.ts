import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/useUserStore'

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'home', component: () => import('@/pages/Home.vue'), meta: { title: '首页' } },
  { path: '/note/create', name: 'noteCreate', component: () => import('@/pages/NoteCreate.vue'), meta: { title: '创建品鉴笔记', requiresAuth: true } },
  { path: '/note/:id', name: 'noteDetail', component: () => import('@/pages/NoteDetail.vue'), meta: { title: '品鉴详情' } },
  { path: '/beans', name: 'beans', component: () => import('@/pages/BeanLibrary.vue'), meta: { title: '豆种库' } },
  { path: '/profile/:id', name: 'profile', component: () => import('@/pages/Profile.vue'), meta: { title: '个人主页' } },
  { path: '/recipes', name: 'recipes', component: () => import('@/pages/RecipeSquare.vue'), meta: { title: '配方广场' } },
  { path: '/login', name: 'login', component: () => import('@/pages/Login.vue'), meta: { title: '登录' } },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  document.title = `${to.meta.title || ''} - 咖啡品鉴社区`
  if (to.meta.requiresAuth) {
    const store = useUserStore()
    if (!store.token) {
      return { path: '/login', query: { redirect: to.fullPath } }
    }
  }
  return true
})

export default router

import { createRouter, createWebHashHistory } from 'vue-router'
import { getAuthToken } from '@/api/client'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      redirect: '/ask',
    },
    {
      path: '/ask',
      name: 'ask',
      component: () => import('@/views/AskView.vue'),
      meta: { title: { en: 'Reports', 'zh-CN': '报告' } },
    },
    {
      path: '/overview',
      name: 'overview',
      component: () => import('@/views/OverviewView.vue'),
      meta: { title: { en: 'My Chart', 'zh-CN': '我的命盘' } },
    },
    {
      path: '/bonds',
      name: 'bonds',
      component: () => import('@/views/BondsView.vue'),
      meta: { title: { en: 'Bonds', 'zh-CN': '关系' } },
    },
    {
      path: '/naming',
      name: 'naming',
      component: () => import('@/views/NamingView.vue'),
      meta: { title: { en: 'Naming', 'zh-CN': '起名' } },
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('@/views/SettingsView.vue'),
      meta: { title: { en: 'Settings', 'zh-CN': '设置' } },
    },
  ],
})

router.beforeEach((_to, _from) => {
  if (!getAuthToken()) {
    window.location.href = '/login'
    return false
  }
  return true
})

export default router

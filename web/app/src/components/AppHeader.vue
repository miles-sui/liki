<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useLocaleStore } from '@/stores/locale'
import { useThemeStore } from '@/stores/theme'
import { Sunny, Moon } from '@element-plus/icons-vue'

defineProps<{ drawerOpen: boolean }>()
const emit = defineEmits<{ (e: 'update:drawerOpen', v: boolean): void }>()

const route = useRoute()
const auth = useAuthStore()
const locale = useLocaleStore()
const theme = useThemeStore()

const pageTitle = computed(() => {
  const meta = route.meta?.title as { en: string; 'zh-CN': string } | undefined
  if (!meta) return ''
  return locale.current === 'zh-CN' ? meta['zh-CN'] : meta.en
})

const avatarLabel = computed(() => (auth.userName || '?').charAt(0).toUpperCase())

function handleCommand(cmd: string) {
  if (cmd === 'locale') {
    locale.toggle()
  } else if (cmd === 'logout') {
    auth.logout()
    window.location.hash = '#/ask'
  }
}
</script>

<template>
  <div class="header-bar">
    <div class="header-left">
      <el-button
        class="mobile-only"
        :icon="drawerOpen ? 'Close' : 'Menu'"
        text
        @click="emit('update:drawerOpen', !drawerOpen)"
      />
      <h2 class="header-title">{{ pageTitle }}</h2>
    </div>
    <div class="header-right">
      <el-switch
        :model-value="theme.dark"
        :active-icon="Moon"
        :inactive-icon="Sunny"
        @change="theme.toggle()"
      />
      <el-dropdown trigger="click" @command="handleCommand">
        <span class="user-trigger">
          <el-avatar :size="28">{{ avatarLabel }}</el-avatar>
          <span class="user-name">{{ auth.userName }}</span>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="locale">
              {{ locale.current === 'zh-CN' ? 'EN' : '中文' }}
            </el-dropdown-item>
            <el-dropdown-item command="logout" divided>
              {{ locale.current === 'zh-CN' ? '退出登录' : 'Logout' }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<style scoped>
.header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.header-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
  color: var(--el-text-color-primary);
}
.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}
.user-trigger {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}
.user-name {
  font-size: 13px;
  color: var(--el-text-color-regular);
}
.mobile-only {
  display: none;
}
@media (max-width: 767px) {
  .mobile-only {
    display: inline-flex;
  }
  .user-name {
    display: none;
  }
}
</style>

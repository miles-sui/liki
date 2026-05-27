<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useLocaleStore } from '@/stores/locale'
import {
  Document,
  DataAnalysis,
  Connection,
  Edit,
  Setting,
} from '@element-plus/icons-vue'

const props = defineProps<{ collapsed: boolean; mobileOpen: boolean }>()
const emit = defineEmits<{ (e: 'update:mobileOpen', v: boolean): void; (e: 'navigate'): void }>()

const router = useRouter()
const route = useRoute()
const { t } = useLocaleStore()


const navItems = [
  { path: '/ask', icon: Document, label: t('Reports', '报告') },
  { path: '/overview', icon: DataAnalysis, label: t('My Chart', '我的命盘') },
  { path: '/bonds', icon: Connection, label: t('Bonds', '关系') },
  { path: '/naming', icon: Edit, label: t('Naming', '起名') },
]

const activeIndex = computed(() => route.path)

function onSelect(path: string) {
  router.push(path)
  emit('update:mobileOpen', false)
  emit('navigate')
}

function onDrawerClose() {
  emit('update:mobileOpen', false)
}
</script>

<template>
  <!-- Mobile drawer -->
  <el-drawer
    :model-value="mobileOpen"
    direction="ltr"
    size="220px"
    :with-header="false"
    @close="onDrawerClose"
  >
    <div class="sidebar-logo">
      <span class="text-red-500 font-brand text-xl">25</span><span class="text-green-600 font-brand text-xl">Types</span>
    </div>
    <el-menu
      :default-active="activeIndex"
      class="sidebar-menu"
      @select="onSelect"
    >
      <el-menu-item v-for="item in navItems" :key="item.path" :index="item.path">
        <el-icon><component :is="item.icon" /></el-icon>
        <span>{{ item.label }}</span>
      </el-menu-item>
      <el-divider style="margin:8px 0" />
      <el-menu-item index="/settings">
        <el-icon><Setting /></el-icon>
        <span>{{ t('Settings', '设置') }}</span>
      </el-menu-item>
    </el-menu>
    <div class="sidebar-footer">
      <el-tag size="small" type="info">Free</el-tag>
    </div>
  </el-drawer>

  <!-- Desktop sidebar -->
  <div class="sidebar-desktop" :class="{ collapsed: collapsed }">
    <div class="sidebar-logo">
      <span class="text-red-500 font-brand text-xl">25</span><span class="text-green-600 font-brand text-xl">{{ collapsed ? '' : 'Types' }}</span>
    </div>
    <el-menu
      :default-active="activeIndex"
      :collapse="collapsed"
      class="sidebar-menu"
      @select="onSelect"
    >
      <el-menu-item v-for="item in navItems" :key="item.path" :index="item.path">
        <el-icon><component :is="item.icon" /></el-icon>
        <template #title>{{ item.label }}</template>
      </el-menu-item>
      <el-divider style="margin:8px 0" />
      <el-menu-item index="/settings">
        <el-icon><Setting /></el-icon>
        <template #title>{{ t('Settings', '设置') }}</template>
      </el-menu-item>
    </el-menu>
    <div class="sidebar-footer">
      <el-tag v-if="!collapsed" size="small" type="info">Free</el-tag>
    </div>
  </div>
</template>

<style scoped>
.sidebar-desktop {
  display: flex;
  flex-direction: column;
  height: 100%;
  border-right: 1px solid var(--el-border-color-light);
  background: var(--el-bg-color);
}
.sidebar-desktop.collapsed {
  align-items: center;
}
.sidebar-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 48px;
  flex-shrink: 0;
  font-family: 'Inter', sans-serif;
  letter-spacing: -0.5px;
  user-select: none;
}
.sidebar-menu {
  flex: 1;
  border-right: none !important;
}
.sidebar-footer {
  padding: 12px;
  display: flex;
  justify-content: center;
}
</style>

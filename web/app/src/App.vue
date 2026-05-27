<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Loading } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { useLocaleStore } from '@/stores/locale'
import AppSidebar from '@/components/AppSidebar.vue'
import AppHeader from '@/components/AppHeader.vue'

const auth = useAuthStore()
const locale = useLocaleStore()

const windowWidth = ref(window.innerWidth)
const mobileDrawerOpen = ref(false)
const appReady = ref(false)

function onResize() {
  windowWidth.value = window.innerWidth
}

onMounted(async () => {
  window.addEventListener('resize', onResize)
  if (auth.token) {
    await auth.fetchProfile()
    if (!auth.isLoggedIn) {
      window.location.href = '/login'
      return
    }
  }
  appReady.value = true
})

onUnmounted(() => {
  window.removeEventListener('resize', onResize)
})

const isMobile = computed(() => windowWidth.value < 768)
const isTablet = computed(() => windowWidth.value >= 768 && windowWidth.value < 1024)
const sidebarCollapsed = computed(() => isTablet.value)
const sidebarWidth = computed(() => {
  if (isMobile.value) return '0px'
  if (isTablet.value) return '64px'
  return '220px'
})
</script>

<template>
  <el-container style="height:100vh">
    <el-aside
      v-if="!isMobile"
      :width="sidebarWidth"
      style="transition:width .3s;overflow:hidden"
    >
      <AppSidebar :collapsed="sidebarCollapsed" :mobile-open="false" />
    </el-aside>
    <el-container>
      <el-header style="height:48px;padding:0 16px;border-bottom:1px solid var(--el-border-color-light);display:flex;align-items:center">
        <AppHeader
          :drawer-open="mobileDrawerOpen"
          @update:drawer-open="(v: boolean) => mobileDrawerOpen = v"
        />
      </el-header>
      <el-main>
        <router-view v-if="appReady" />
        <div v-else style="display:flex;align-items:center;justify-content:center;height:100%">
          <el-icon class="is-loading" :size="24"><Loading /></el-icon>
        </div>
      </el-main>
    </el-container>
  </el-container>

  <!-- Mobile drawer rendered at App level so it overlays everything -->
  <AppSidebar
    v-if="isMobile"
    :collapsed="false"
    :mobile-open="mobileDrawerOpen"
    @update:mobile-open="(v: boolean) => mobileDrawerOpen = v"
  />
</template>

<style>
html, body, #app {
  height: 100%;
  margin: 0;
}
</style>

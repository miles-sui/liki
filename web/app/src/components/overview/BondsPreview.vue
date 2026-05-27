<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useLocaleStore } from '@/stores/locale'
import { api } from '@/api/client'
import ConcordBadge from '@/components/bonds/ConcordBadge.vue'
import type { BondItem } from '@shared/types'

const router = useRouter()
const auth = useAuthStore()
const { t } = useLocaleStore()

const bonds = ref<BondItem[]>([])
const loading = ref(false)

onMounted(async () => {
  if (!auth.isLoggedIn) return
  loading.value = true
  try {
    const resp = await api<{ data: { items: BondItem[]; total: number } }>(`/profiles/${auth.userName}/bonds`)
    bonds.value = resp.data?.items?.slice(0, 3) || []
  } catch { /* no bonds */ }
  finally { loading.value = false }
})
</script>

<template>
  <el-card shadow="never" class="premium-card">
    <template #header>
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="w-2 h-2 rounded-full" style="background:var(--wuxing-wood)"></span>
          <span class="font-semibold">{{ t('Bonds', '关系') }}</span>
        </div>
        <span class="text-xs text-gray-400">{{ bonds.length }} {{ t('active', '个活跃') }}</span>
      </div>
    </template>
    <el-skeleton v-if="loading" :rows="2" animated />
    <div v-else-if="bonds.length === 0" class="premium-empty">
      <div class="icon-wrap">🤝</div>
      <div class="text-sm text-gray-500">{{ t('Compare with others to build bonds.', '与他人对比即可生成关系图谱。') }}</div>
    </div>
    <div v-else class="space-y-1">
      <div v-for="b in bonds" :key="b.other_user.name" class="flex items-center justify-between py-2 px-3 rounded-lg hover:bg-gray-50 transition-colors">
        <span class="text-sm font-medium">{{ b.other_user.name }}</span>
        <ConcordBadge :concord="b.bond.concord" />
      </div>
      <div class="mt-3 pt-2 border-t border-gray-100">
        <el-button size="small" text type="primary" @click="router.push('/bonds')">
          {{ t('View All Bonds', '查看全部关系') }} →
        </el-button>
      </div>
    </div>
  </el-card>
</template>

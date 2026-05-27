<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useLocaleStore } from '@/stores/locale'
import type { MinggeData } from '@shared/types'
import ElementRadar from './ElementRadar.vue'

defineProps<{ mingge: MinggeData | null; loading?: boolean }>()

const auth = useAuthStore()
const { t } = useLocaleStore()
</script>

<template>
  <el-card shadow="never" class="premium-card">
    <template #header>
      <div class="flex items-center gap-2">
        <span class="w-2 h-2 rounded-full" style="background:var(--wuxing-fire)"></span>
        <span class="font-semibold">{{ t('25types Identity', '25types 类型') }}</span>
      </div>
    </template>
    <el-skeleton v-if="loading" :rows="3" animated />
    <div v-else-if="!mingge" class="premium-empty">
      <div class="icon-wrap">🧬</div>
      <div class="text-sm text-gray-500 mb-4">{{ t('Complete an assessment to discover your type.', '完成评估发现你的类型。') }}</div>
      <a href="/assess"><el-button type="primary">{{ t('Start Assessment', '开始评估') }}</el-button></a>
    </div>
    <div v-else>
      <div class="text-center mb-3">
        <div class="text-lg font-bold">{{ auth.userName }}</div>
      </div>
      <ElementRadar :element-count="mingge.element_count" />
    </div>
  </el-card>
</template>

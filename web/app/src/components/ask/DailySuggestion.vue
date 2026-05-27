<script setup lang="ts">
import { useLocaleStore } from '@/stores/locale'
import type { DailySuggestionData } from '@shared/types'

defineProps<{ suggestion: DailySuggestionData | null; loading: boolean }>()
const emit = defineEmits<{ (e: 'askToday'): void }>()

const { t } = useLocaleStore()
</script>

<template>
  <el-card shadow="never" class="premium-card daily-card">
    <el-skeleton v-if="loading" :rows="2" animated />
    <div v-else-if="suggestion" class="flex items-start gap-4">
      <div
        class="daily-dot"
        :style="{ backgroundColor: suggestion.color || '#667eea' }"
      ></div>
      <div class="flex-1 min-w-0">
        <div class="text-sm text-gray-700 leading-relaxed line-clamp-2">{{ suggestion.suggestion }}</div>
        <div class="flex items-center gap-2 mt-2">
          <el-tag size="small" effect="plain" round>{{ suggestion.element }}</el-tag>
          <el-tag size="small" effect="plain" round type="info">{{ suggestion.direction }}</el-tag>
        </div>
      </div>
      <el-button type="primary" plain size="small" round @click="emit('askToday')">
        {{ t('Ask today', '问今天') }} →
      </el-button>
    </div>
    <div v-else class="text-sm text-gray-400 text-center py-4">
      {{ t('Complete an assessment for daily insights.', '完成评估解锁每日指引。') }}
    </div>
  </el-card>
</template>

<style scoped>
.daily-dot {
  width: 12px; height: 12px;
  border-radius: 50%;
  margin-top: 4px;
  flex-shrink: 0;
  box-shadow: 0 0 12px currentColor;
}
.daily-card { background: linear-gradient(135deg, #fafbff, #f5f3ff); }
</style>

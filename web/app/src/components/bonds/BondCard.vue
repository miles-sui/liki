<script setup lang="ts">
import { computed } from 'vue'
import { useLocaleStore } from '@/stores/locale'
import { elementColor, elementIndex } from '@shared/elements'
import ConcordBadge from './ConcordBadge.vue'

const props = defineProps<{
  self: Record<string, number>
  other: Record<string, number>
  deltaA: Record<string, number>
  deltaB: Record<string, number>
  concord: string
}>()

const { t } = useLocaleStore()

const elementKeys = computed(() => Object.keys(props.self))
</script>

<template>
  <div class="bond-cards">
    <el-card shadow="never" class="bond-card">
      <div class="text-sm text-gray-400 mb-2">{{ t('Your Energy', '你的能量') }}</div>
      <div v-for="k in elementKeys" :key="k" class="flex items-center gap-1 text-xs mb-1">
        <span class="w-8 text-gray-500">{{ k }}</span>
        <div class="flex-1 h-2 bg-gray-100 rounded-full overflow-hidden">
          <div class="h-full rounded-full" :style="{ width: Math.abs(self[k]) * 100 + '%', backgroundColor: elementColor(elementIndex(k)) }"></div>
        </div>
      </div>
    </el-card>
    <div class="flex justify-center py-2">
      <ConcordBadge :concord="concord" />
    </div>
    <el-card shadow="never" class="bond-card">
      <div class="text-sm text-gray-400 mb-2">{{ t('Their Energy', 'TA的能量') }}</div>
      <div v-for="k in elementKeys" :key="k" class="flex items-center gap-1 text-xs mb-1">
        <span class="w-8 text-gray-500">{{ k }}</span>
        <div class="flex-1 h-2 bg-gray-100 rounded-full overflow-hidden">
          <div class="h-full rounded-full" :style="{ width: Math.abs(other[k]) * 100 + '%', backgroundColor: elementColor(elementIndex(k)) }"></div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.bond-cards {
  display: flex;
  align-items: center;
  gap: 8px;
}
.bond-card {
  flex: 1;
  min-width: 0;
}
@media (max-width: 640px) {
  .bond-cards { flex-direction: column; }
}
</style>

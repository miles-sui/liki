<script setup lang="ts">
import { useLocaleStore } from '@/stores/locale'
import { formatPillar, elementColor, elementIndex } from '@shared/elements'
import type { DayunData } from '@shared/types'

defineProps<{ dayun: DayunData | null; loading?: boolean }>()

const { t } = useLocaleStore()
</script>

<template>
  <el-card shadow="never" class="premium-card">
    <template #header>
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="w-2 h-2 rounded-full" style="background:var(--wuxing-earth)"></span>
          <span class="font-semibold">{{ t('Big Fortune', '大运') }}</span>
        </div>
        <a href="/app#/ask"><el-tag size="small" type="primary" class="cursor-pointer">{{ t('解读', '解读') }}</el-tag></a>
      </div>
    </template>
    <el-skeleton v-if="loading" :rows="2" animated />
    <div v-else-if="!dayun" class="premium-empty">
      <div class="icon-wrap">🔄</div>
      <div class="text-sm text-gray-500">{{ t('Save birth info to see big fortune.', '保存出生信息后查看大运。') }}</div>
    </div>
    <div v-else>
      <template v-if="dayun.pillars?.[dayun.current_pillar_index || 0]">
        <div class="text-xs text-gray-400 uppercase tracking-wide">{{ t('Current Pillar', '当前大运') }}</div>
        <div class="flex items-center gap-2 mt-1">
          <span class="text-xl font-bold">{{ formatPillar(dayun.pillars[dayun.current_pillar_index].stem, dayun.pillars[dayun.current_pillar_index].branch) }}</span>
          <span class="text-sm text-gray-500">
            {{ dayun.pillars[dayun.current_pillar_index].age_start }}-{{ dayun.pillars[dayun.current_pillar_index].age_end }}
          </span>
        </div>
        <div class="flex gap-1 mt-2">
          <el-tag
            size="small"
            :color="elementColor(elementIndex(dayun.pillars[dayun.current_pillar_index].element || ''))"
            effect="dark"
            style="border:none"
          >{{ dayun.pillars[dayun.current_pillar_index].element }}</el-tag>
          <el-tag size="small" type="info" round>{{ dayun.pillars[dayun.current_pillar_index].ten_god }}</el-tag>
        </div>
      </template>
    </div>
  </el-card>
</template>

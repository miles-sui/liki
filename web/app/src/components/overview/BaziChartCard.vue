<script setup lang="ts">
import { computed } from 'vue'
import { useLocaleStore } from '@/stores/locale'
import { formatPillar, elementName, elementColor } from '@shared/elements'
import type { MinggeData } from '@shared/types'

const props = defineProps<{ mingge: MinggeData | null; loading?: boolean }>()
const emit = defineEmits<{ (e: 'addBirthInfo'): void }>()

const locale = useLocaleStore()
const { t } = locale

const chart = computed(() => props.mingge?.chart)
const pillars = computed(() => {
  const c = chart.value
  if (!c) return []
  return [
    { label: t('Year', '年柱'), stem: c.year_pillar?.stem, branch: c.year_pillar?.branch, nayin: c.na_yin?.[0] },
    { label: t('Month', '月柱'), stem: c.month_pillar?.stem, branch: c.month_pillar?.branch, nayin: c.na_yin?.[1] },
    { label: t('Day', '日柱'), stem: c.day_pillar?.stem, branch: c.day_pillar?.branch, nayin: c.na_yin?.[2] },
    { label: t('Hour', '时柱'), stem: c.hour_pillar?.stem, branch: c.hour_pillar?.branch, nayin: c.na_yin?.[3] },
  ]
})

const elementList = computed(() => {
  if (!props.mingge?.element_count) return []
  return Object.entries(props.mingge.element_count)
    .map(([k, v]) => ({ code: Number(k), count: v }))
    .sort((a, b) => b.count - a.count)
})

const maxCount = computed(() => Math.max(1, ...elementList.value.map(e => e.count)))

const statGrid = computed(() => [
  { label: t('Day Master', '日主'), value: props.mingge?.day_master_name || '-' },
  { label: t('Strength', '强弱'), value: props.mingge?.strength || '-' },
  { label: t('Yong Shen', '用神'), value: props.mingge?.yong_shen || '-' },
  { label: t('Pattern', '格局'), value: props.mingge?.pattern || '-' },
])
</script>

<template>
  <el-card shadow="never" class="premium-card">
    <template #header>
      <div class="flex items-center gap-2">
        <span class="w-2 h-2 rounded-full" style="background:var(--wuxing-water)"></span>
        <span class="font-semibold">{{ t('BaZi Chart', '八字排盘') }}</span>
      </div>
    </template>
    <el-skeleton v-if="loading" :rows="6" animated />
    <div v-else-if="!mingge" class="premium-empty">
      <div class="icon-wrap">📅</div>
      <div class="text-sm text-gray-500 mb-4">{{ t('Add your birth info to generate the chart.', '录入出生信息生成八字排盘。') }}</div>
      <el-button type="primary" @click="emit('addBirthInfo')">{{ t('Add Birth Info', '录入出生信息') }}</el-button>
    </div>
    <div v-else>
      <!-- 4 pillars -->
      <div class="pillar-grid">
        <div v-for="p in pillars" :key="p.label" class="pillar-cell">
          <div class="label">{{ p.label }}</div>
          <div class="value">{{ formatPillar(p.stem, p.branch) }}</div>
          <div class="nayin">{{ p.nayin || '-' }}</div>
        </div>
      </div>

      <!-- Day master info grid -->
      <div class="grid grid-cols-4 gap-2 text-center mt-4">
        <div v-for="s in statGrid" :key="s.label">
          <div class="text-xs text-gray-400 mb-0.5">{{ s.label }}</div>
          <el-tag size="small" round>{{ s.value }}</el-tag>
        </div>
      </div>

      <!-- Element bars -->
      <div class="text-xs text-gray-400 mt-4 mb-2">{{ t('Elements', '五行分布') }}</div>
      <div v-for="el in elementList" :key="el.code" class="element-bar-wrap">
        <span class="w-8 text-xs text-right">{{ elementName(el.code, locale.current) }}</span>
        <div class="element-bar-track">
          <div
            class="element-bar-fill"
            :style="{ width: (el.count / maxCount * 100) + '%', backgroundColor: elementColor(el.code) }"
          ></div>
        </div>
        <span class="text-xs w-4 text-gray-400">{{ el.count }}</span>
      </div>
    </div>
  </el-card>
</template>

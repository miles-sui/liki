<script setup lang="ts">
import { computed } from 'vue'
import { useLocaleStore } from '@/stores/locale'
import type { Report } from '@shared/types'

const props = defineProps<{
  report: Report | null
  streaming: boolean
  streamText: string
}>()
const emit = defineEmits<{
  (e: 'reAsk'): void
  (e: 'share'): void
}>()

const { t } = useLocaleStore()

const displayText = computed(() => {
  if (props.streaming) return props.streamText
  return props.report?.content || ''
})

const hasContent = computed(() => !!props.report || props.streaming)
</script>

<template>
  <el-card shadow="never" class="premium-card report-reader">
    <!-- Empty -->
    <div v-if="!hasContent" class="premium-empty">
      <div class="icon-wrap">📋</div>
      <div class="text-sm text-gray-500">{{ t('Select a report or ask a question to get started.', '选择一个报告或输入问题开始解读。') }}</div>
    </div>

    <!-- Active report -->
    <div v-else>
      <!-- Header -->
      <div v-if="report" class="flex items-center justify-between mb-4 pb-3" style="border-bottom:1px solid var(--el-border-color-lighter)">
        <div class="flex items-center gap-2 min-w-0">
          <h2 class="text-lg font-bold m-0 truncate">{{ report.title || report.question || t('Report', '报告') }}</h2>
          <el-tag size="small" effect="plain" round>{{ report.scene }}</el-tag>
        </div>
        <div class="flex gap-1 flex-shrink-0">
          <el-button size="small" text @click="emit('reAsk')">{{ t('Edit', '修改') }}</el-button>
          <el-button size="small" text @click="emit('share')">{{ t('Share', '分享') }}</el-button>
        </div>
      </div>

      <!-- Content -->
      <div class="streaming-text" :class="{ streaming: streaming }">
        <template v-if="displayText">{{ displayText }}</template>
        <span v-if="streaming" class="streaming-cursor"></span>
      </div>

      <!-- Traceability -->
      <div v-if="report?.engine_data" class="mt-4">
        <el-collapse>
          <el-collapse-item :title="t('Engine trace', '数据溯源')">
            <pre class="text-xs text-gray-500 overflow-x-auto p-3 rounded-lg" style="background:var(--el-fill-color-light)">{{ JSON.stringify(report.engine_data, null, 2) }}</pre>
          </el-collapse-item>
        </el-collapse>
      </div>
    </div>
  </el-card>
</template>

<style scoped>
.report-reader { min-height: 200px; }
</style>

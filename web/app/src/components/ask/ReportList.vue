<script setup lang="ts">
import { useLocaleStore } from '@/stores/locale'
import type { Report } from '@shared/types'

defineProps<{ items: Report[]; activeId: number | null; loading: boolean }>()
const emit = defineEmits<{
  (e: 'select', report: Report): void
  (e: 'delete', report: Report): void
}>()

const { t } = useLocaleStore()
</script>

<template>
  <div class="report-list">
    <div class="flex items-center justify-between px-1 mb-3">
      <span class="text-xs font-semibold uppercase tracking-wider text-gray-400">{{ t('History', '报告历史') }}</span>
      <span v-if="items.length" class="text-xs text-gray-400">{{ items.length }}</span>
    </div>
    <el-skeleton v-if="loading" :rows="4" animated />
    <div
      v-else-if="items.length === 0"
      class="text-center py-8 text-xs text-gray-400"
    >
      {{ t('No reports yet', '暂无报告') }}
    </div>
    <div v-else class="space-y-1">
      <div
        v-for="r in items" :key="r.id"
        class="report-item"
        :class="{ active: r.id === activeId }"
        @click="emit('select', r)"
      >
        <div class="flex-1 min-w-0">
          <div class="text-sm truncate font-medium">{{ r.title || r.question || t('Report', '报告') }}</div>
          <div class="flex items-center gap-1.5 mt-1">
            <span class="w-1.5 h-1.5 rounded-full" :class="r.id === activeId ? 'bg-blue-500' : 'bg-gray-300'"></span>
            <span class="text-xs text-gray-400">{{ new Date(r.created_at).toLocaleDateString() }}</span>
          </div>
        </div>
        <el-popconfirm
          :title="t('Delete?', '确认删除？')"
          @confirm="emit('delete', r)"
        >
          <template #reference>
            <el-button size="small" text class="delete-btn" @click.stop>
              <span class="text-gray-300 hover:text-red-400">×</span>
            </el-button>
          </template>
        </el-popconfirm>
      </div>
    </div>
  </div>
</template>

<style scoped>
.report-list { width: 100%; }
.report-item {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  border-radius: 10px;
  cursor: pointer;
  transition: all .15s ease;
  border: 1px solid transparent;
}
.report-item:hover { background: var(--el-fill-color-light); }
.report-item.active {
  background: var(--el-color-primary-light-9);
  border-color: var(--el-color-primary-light-5);
}
.delete-btn { visibility: hidden; }
.report-item:hover .delete-btn { visibility: visible; }
</style>

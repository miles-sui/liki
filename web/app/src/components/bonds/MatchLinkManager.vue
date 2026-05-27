<script setup lang="ts">
import { useLocaleStore } from '@/stores/locale'
import { ElMessage } from 'element-plus'
import type { MatchLink } from '@shared/types'

defineProps<{ links: MatchLink[]; loading: boolean }>()
const emit = defineEmits<{
  (e: 'create'): void
  (e: 'delete', id: number): void
}>()

const { t } = useLocaleStore()

async function copyLink(link: MatchLink) {
  const url = `${window.location.origin}/m/${link.token}`
  if (navigator.clipboard) {
    try {
      await navigator.clipboard.writeText(url)
      ElMessage.success(t('Link copied', '链接已复制'))
      return
    } catch { /* fall through to fallback */ }
  }
  const input = document.createElement('input')
  input.value = url; document.body.appendChild(input); input.select()
  const ok = document.execCommand('copy'); document.body.removeChild(input)
  if (ok) ElMessage.success(t('Link copied', '链接已复制'))
  else ElMessage.info(url)
}
</script>

<template>
  <el-card shadow="never">
    <template #header>
      <div class="flex items-center justify-between">
        <span class="font-medium">{{ t('Match Links', '匹配链接') }}</span>
        <el-button size="small" type="primary" @click="emit('create')">
          {{ t('Create Link', '创建链接') }}
        </el-button>
      </div>
    </template>
    <el-skeleton v-if="loading" :rows="2" animated />
    <div v-else-if="links.length === 0" class="text-sm text-gray-400 text-center py-4">
      {{ t('No match links yet', '暂无匹配链接') }}
    </div>
    <div v-else v-for="l in links" :key="l.id" class="flex items-center justify-between py-2 border-b last:border-0">
      <div>
        <div class="text-sm font-mono text-gray-500">{{ l.token.slice(0, 12) }}...</div>
        <div class="text-xs text-gray-400">{{ l.match_count }} {{ t('matches', '人已匹配') }}</div>
      </div>
      <div class="flex gap-2">
        <el-button size="small" text @click="copyLink(l)">{{ t('Copy', '复制') }}</el-button>
        <el-popconfirm
          :title="t('Delete this link?', '确认删除此链接？')"
          @confirm="emit('delete', l.id)"
        >
          <template #reference>
            <el-button size="small" text type="danger">{{ t('Delete', '删除') }}</el-button>
          </template>
        </el-popconfirm>
      </div>
    </div>
  </el-card>
</template>

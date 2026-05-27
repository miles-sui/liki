<script setup lang="ts">
import { ref } from 'vue'
import { useLocaleStore } from '@/stores/locale'

defineProps<{ generating: boolean }>()
const emit = defineEmits<{ (e: 'submit', text: string): void }>()

const { t } = useLocaleStore()

const text = ref('')

function handleSubmit() {
  const val = text.value.trim()
  if (!val) return
  emit('submit', val)
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSubmit()
  }
}
</script>

<template>
  <div class="report-input-wrap">
    <div class="report-input-inner">
      <el-input
        v-model="text"
        type="textarea"
        :rows="2"
        :placeholder="t('Ask anything based on your chart...', '输入你想了解的问题...')"
        :disabled="generating"
        class="text-area"
        @keydown="onKeydown"
      />
      <el-button
        class="submit-btn"
        type="primary"
        :loading="generating"
        :disabled="!text.trim() || generating"
        @click="handleSubmit"
        round
      >
        <span v-if="!generating" class="flex items-center gap-1">
          <span>{{ t('Ask', '提问') }}</span>
          <span class="text-lg leading-none">→</span>
        </span>
        <span v-else>{{ t('Generating...', '生成中...') }}</span>
      </el-button>
    </div>
    <div class="text-xs text-gray-400 mt-2 text-center">
      {{ t('Engine data is free. Full report from ¥9.9.', '引擎数据免费，完整报告 ¥9.9。按 Enter 发送，Shift+Enter 换行。') }}
    </div>
  </div>
</template>

<style scoped>
.report-input-wrap {
  position: sticky;
  bottom: 0;
  background: linear-gradient(to top, #fff 80%, transparent);
  padding-top: 12px;
}
.report-input-inner {
  display: flex;
  gap: 10px;
  align-items: flex-end;
  padding: 12px 16px;
  border-radius: 14px;
  border: 1px solid var(--el-border-color);
  background: #fff;
  box-shadow: 0 2px 16px rgba(0,0,0,.04);
  transition: border-color .2s, box-shadow .2s;
}
.report-input-inner:focus-within {
  border-color: var(--el-color-primary);
  box-shadow: 0 0 0 3px var(--el-color-primary-light-8);
}
.text-area { flex: 1; }
.text-area :deep(.el-textarea__inner) {
  border: none !important;
  box-shadow: none !important;
  padding: 0;
  font-size: .95rem;
}
.submit-btn { flex-shrink: 0; }
</style>

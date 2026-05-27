<script setup lang="ts">
import { useLocaleStore } from '@/stores/locale'

defineProps<{ activeScene: string | null }>()
const emit = defineEmits<{ (e: 'select', scene: string): void }>()

const { t } = useLocaleStore()

const scenes = [
  { id: 'liunian', en: 'Yearly Fortune', zh: '流年运势', icon: '📅' },
  { id: 'relationship', en: 'Love Match', zh: '恋爱匹配', icon: '💕' },
  { id: 'naming', en: 'Baby Name', zh: '宝宝起名', icon: '👶' },
  { id: 'dates', en: 'Wedding Date', zh: '结婚择日', icon: '💍' },
  { id: 'career', en: 'Career Path', zh: '高考选专业', icon: '🎓' },
  { id: 'dayun', en: 'Big Fortune', zh: '大运解读', icon: '🔮' },
  { id: 'bazi', en: 'Partner Match', zh: '合伙人匹配', icon: '🤝' },
  { id: 'general', en: 'More', zh: '更多', icon: '✨' },
]
</script>

<template>
  <div class="scene-grid">
    <div
      v-for="s in scenes" :key="s.id"
      class="scene-chip"
      :class="{ active: activeScene === s.id }"
      @click="emit('select', s.id)"
    >
      <span class="text-lg">{{ s.icon }}</span>
      <span class="text-sm font-medium">{{ t(s.en, s.zh) }}</span>
    </div>
  </div>
</template>

<style scoped>
.scene-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 14px;
  border-radius: 10px;
  border: 1px solid var(--el-border-color-lighter);
  cursor: pointer;
  transition: all .2s ease;
  user-select: none;
  background: #fff;
}
.scene-chip:hover {
  border-color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
  transform: translateY(-1px);
}
.scene-chip.active {
  border-color: var(--el-color-primary);
  background: linear-gradient(135deg, var(--el-color-primary-light-9), var(--el-color-primary-light-8));
  box-shadow: 0 2px 8px rgba(102, 126, 234, .15);
}
</style>

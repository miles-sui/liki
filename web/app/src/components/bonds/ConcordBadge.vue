<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ concord: string }>()

const CONCORD_MAP: Record<string, string> = { '顺': 'success', 'harmonious': 'success', '逆': 'danger', 'challenging': 'danger' }
const tagType = computed(() => CONCORD_MAP[props.concord] || 'info')
</script>

<template>
  <el-tag :type="tagType" size="small">{{ concord }}</el-tag>
</template>

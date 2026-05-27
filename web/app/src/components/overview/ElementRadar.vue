<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import echarts, { type ECharts } from '@/echarts-setup'
import { elementName, elementColor } from '@shared/elements'

const props = defineProps<{ elementCount: Record<number, number> }>()

const chartRef = ref<HTMLElement>()
let chart: ECharts | null = null

function render() {
  if (!chartRef.value) return
  if (!chart) {
    chart = echarts.init(chartRef.value)
  }
  const elements = [1, 2, 3, 4, 5]
  const indic = elements.map(e => ({ name: elementName(e, 'zh-CN'), max: 10 }))
  const values = elements.map(e => props.elementCount[e] || 0)

  chart.setOption({
    radar: {
      indicator: indic,
      center: ['50%', '50%'],
      radius: '65%',
      axisName: { fontSize: 11, color: '#999' },
    },
    series: [{
      type: 'radar',
      data: [{ value: values, name: '', areaStyle: { opacity: 0.15 } }],
      lineStyle: { color: '#409eff', width: 2 },
      itemStyle: { color: '#409eff' },
      symbol: 'circle', symbolSize: 5,
    }],
  }, true)
}

function onResize() {
  chart?.resize()
}

onMounted(() => {
  render()
  window.addEventListener('resize', onResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', onResize)
  chart?.dispose()
})

watch(() => props.elementCount, () => render(), { deep: true })
</script>

<template>
  <div ref="chartRef" style="width:100%;height:200px"></div>
</template>

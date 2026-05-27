<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import echarts, { type ECharts } from '@/echarts-setup'
import type { FlowYearlyData } from '@shared/types'

const props = defineProps<{ data: FlowYearlyData | null; loading?: boolean }>()

const chartRef = ref<HTMLElement>()
let chart: ECharts | null = null

function onResize() { chart?.resize() }

function render() {
  if (!chartRef.value || !props.data?.months?.length) return
  if (!chart) {
    chart = echarts.init(chartRef.value)
  }
  chart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: 40, right: 20, top: 20, bottom: 30 },
    xAxis: {
      type: 'category',
      data: props.data.months.map(m => m.name_en || m.id),
      axisLabel: { fontSize: 10, rotate: 30 },
    },
    yAxis: {
      type: 'value',
      name: 'Energy',
      axisLabel: { fontSize: 10 },
    },
    series: [{
      type: 'bar',
      data: props.data.months.map(m => m.generates),
      itemStyle: {
        color: (params: { dataIndex: number }) => {
          const colors = ['#1A7A6F', '#CB3B2D', '#9A7410', '#8E8880', '#1C2238']
          return colors[params.dataIndex % 5]
        },
        borderRadius: [3, 3, 0, 0],
      },
      barMaxWidth: 30,
    }],
  }, true)
}

onMounted(() => { render(); window.addEventListener('resize', onResize) })
onUnmounted(() => { window.removeEventListener('resize', onResize); chart?.dispose() })
watch(() => props.data, () => render())
</script>

<template>
  <el-card shadow="never">
    <template #header>
      <span class="font-medium">Flow River</span>
    </template>
    <el-skeleton v-if="loading" :rows="4" animated />
    <el-empty v-else-if="!data" description="No flow data" />
    <div v-else ref="chartRef" style="width:100%;height:200px"></div>
  </el-card>
</template>

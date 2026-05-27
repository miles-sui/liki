import * as echarts from 'echarts/core'
import { RadarChart, BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([RadarChart, BarChart, GridComponent, TooltipComponent, CanvasRenderer])

export default echarts
export type ECharts = echarts.ECharts

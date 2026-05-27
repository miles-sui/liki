// Custom ECharts build — only the chart types used by 25types
import * as echarts from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import {
  RadarChart, BarChart, LineChart, GraphChart, CustomChart,
} from 'echarts/charts';
import {
  GridComponent, LegendComponent, TooltipComponent, TitleComponent,
  DatasetComponent, GraphicComponent, ToolboxComponent,
  DataZoomComponent, MarkLineComponent, MarkPointComponent,
} from 'echarts/components';

echarts.use([
  CanvasRenderer,
  RadarChart, BarChart, LineChart, GraphChart, CustomChart,
  GridComponent, LegendComponent, TooltipComponent, TitleComponent,
  DatasetComponent, GraphicComponent, ToolboxComponent,
  DataZoomComponent, MarkLineComponent, MarkPointComponent,
]);

// Expose global for existing page scripts
window.echarts = echarts;

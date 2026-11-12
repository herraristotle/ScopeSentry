import * as echarts from 'echarts/core'

import {
  BarChart,
  LineChart,
  PieChart,
  PictorialBarChart,
  RadarChart
} from 'echarts/charts'

import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  PolarComponent,
  AriaComponent,
  ParallelComponent,
  LegendComponent
} from 'echarts/components'

import { CanvasRenderer } from 'echarts/renderers'

echarts.use([
  LegendComponent,
  TitleComponent,
  TooltipComponent,
  GridComponent,
  PolarComponent,
  AriaComponent,
  ParallelComponent,
  BarChart,
  LineChart,
  PieChart,
  CanvasRenderer,
  PictorialBarChart,
  RadarChart
])

export default echarts

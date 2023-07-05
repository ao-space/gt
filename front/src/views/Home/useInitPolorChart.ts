import * as echarts from 'echarts/core'
import { PolarComponent, PolarComponentOption, LegendComponent, LegendComponentOption } from 'echarts/components'
import { BarChart, BarSeriesOption } from 'echarts/charts'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([PolarComponent, LegendComponent, BarChart, CanvasRenderer])

type EChartsOption = echarts.ComposeOption<PolarComponentOption | LegendComponentOption | BarSeriesOption>

let option: EChartsOption

export const useInitPolorChart = (chartDom: HTMLElement | undefined) => {
  // let chartDom = document.getElementById('main')!;

  const myChart = echarts.init(chartDom as HTMLElement)
  option = {
    angleAxis: {
      type: 'category',
      data: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
    },
    radiusAxis: {},
    polar: {},
    series: [
      {
        type: 'bar',
        data: [1, 2, 3, 4, 3, 5, 1],
        coordinateSystem: 'polar',
        name: 'Total',
        stack: 'a',
        emphasis: {
          focus: 'series'
        }
      },
      {
        type: 'bar',
        data: [2, 4, 6, 1, 3, 2, 1],
        coordinateSystem: 'polar',
        name: 'Reached',
        stack: 'a',
        emphasis: {
          focus: 'series'
        }
      },
      {
        type: 'bar',
        data: [1, 2, 3, 4, 1, 2, 5],
        coordinateSystem: 'polar',
        name: 'Opened',
        stack: 'a',
        emphasis: {
          focus: 'series'
        }
      }
    ],
    legend: {
      show: true,
      data: ['Total', 'Reached', 'Opened']
    }
  }

  // eslint-disable-next-line no-unused-expressions
  option && myChart.setOption(option)
}

export default { useInitPolorChart }

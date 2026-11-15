<template>
  <section v-loading="loading" class="contact-stats" aria-label="联系人统计">
    <div v-if="failed" class="stats-error" role="status">
      联系人统计加载失败
      <el-button link type="primary" @click="refresh">重新加载</el-button>
    </div>
    <template v-else>
      <article class="stat-panel trend-panel">
        <header>
          <h3>好友年份趋势</h3>
          <span>添加时间未知 {{ stats.unknownYear }} 人</span>
        </header>
        <v-chart
          v-if="stats.years.length"
          class="stat-chart"
          :option="yearOption"
          autoresize
          aria-label="每年添加好友人数折线图"
        />
        <el-empty
          v-else
          :description="stats.total ? '暂无已知添加时间' : '暂无联系人数据'"
          :image-size="72"
        />
      </article>
      <article class="stat-panel">
        <header>
          <h3>好友性别比例</h3>
          <span>共 {{ stats.total }} 人</span>
        </header>
        <v-chart
          v-if="stats.total"
          class="stat-chart"
          :option="genderOption"
          autoresize
          aria-label="好友性别比例环形图"
        />
        <el-empty v-else description="暂无联系人数据" :image-size="72" />
      </article>
      <article class="stat-panel">
        <header>
          <h3>好友分类排行</h3>
          <span>按人数降序</span>
        </header>
        <v-chart
          v-if="stats.categories.length"
          class="stat-chart"
          :option="categoryOption"
          autoresize
          aria-label="好友分类人数条形图"
        />
        <el-empty v-else description="暂无联系人数据" :image-size="72" />
      </article>
    </template>
  </section>
</template>

<script setup>
  import { computed, onMounted, ref } from 'vue'
  import VChart from 'vue-echarts'
  import { use } from 'echarts/core'
  import { LineChart, PieChart, BarChart } from 'echarts/charts'
  import {
    GridComponent,
    TooltipComponent,
    LegendComponent,
    DataZoomComponent
  } from 'echarts/components'
  import { CanvasRenderer } from 'echarts/renderers'
  import { getRidgeContactStat } from '@/api/ridge/ridge_contact'

  use([
    LineChart,
    PieChart,
    BarChart,
    GridComponent,
    TooltipComponent,
    LegendComponent,
    DataZoomComponent,
    CanvasRenderer
  ])

  const loading = ref(false)
  const failed = ref(false)
  const stats = ref({
    total: 0,
    unknownYear: 0,
    years: [],
    genders: [],
    categories: []
  })
  const tooltip = { confine: true, renderMode: 'richText' }
  const axisLabel = { color: '#606266' }
  const countAxis = {
    type: 'value',
    minInterval: 1,
    axisLabel,
    splitLine: { lineStyle: { color: '#ebeef5' } }
  }

  const yearOption = computed(() => ({
    tooltip: {
      ...tooltip,
      trigger: 'axis',
      valueFormatter: (value) => `${value} 人`
    },
    grid: {
      top: 24,
      left: 40,
      right: 18,
      bottom: stats.value.years.length > 12 ? 65 : 35
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: stats.value.years.map((item) => item.name),
      axisLabel
    },
    yAxis: countAxis,
    dataZoom:
      stats.value.years.length > 12
        ? [
            {
              type: 'slider',
              height: 16,
              bottom: 8,
              startValue: Math.max(0, stats.value.years.length - 12),
              endValue: stats.value.years.length - 1
            }
          ]
        : [],
    series: [
      {
        name: '添加好友',
        type: 'line',
        smooth: true,
        data: stats.value.years.map((item) => item.value),
        symbolSize: 7,
        itemStyle: { color: '#00ccc6' },
        lineStyle: { width: 3, color: '#00ccc6' },
        areaStyle: { color: '#00ccc6', opacity: 0.12 }
      }
    ]
  }))

  const genderOption = computed(() => ({
    color: ['#409eff', '#e88ca5', '#a8abb2'],
    tooltip: { ...tooltip, trigger: 'item', formatter: '{b}：{c} 人（{d}%）' },
    legend: { bottom: 4, icon: 'circle', textStyle: axisLabel },
    series: [
      {
        name: '性别比例',
        type: 'pie',
        radius: ['42%', '65%'],
        center: ['50%', '44%'],
        label: { show: false },
        emphasis: { label: { show: true, formatter: '{b}\n{d}%' } },
        data: stats.value.genders
      }
    ]
  }))

  const categoryOption = computed(() => ({
    tooltip: {
      ...tooltip,
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      valueFormatter: (value) => `${value} 人`
    },
    grid: {
      top: 15,
      left: 12,
      right: stats.value.categories.length > 6 ? 45 : 30,
      bottom: 20,
      containLabel: true
    },
    xAxis: countAxis,
    yAxis: {
      type: 'category',
      inverse: true,
      data: stats.value.categories.map((item) => item.name),
      axisLabel: { ...axisLabel, width: 88, overflow: 'truncate' },
      axisLine: { show: false },
      axisTick: { show: false }
    },
    dataZoom:
      stats.value.categories.length > 6
        ? [
            {
              type: 'slider',
              yAxisIndex: 0,
              right: 0,
              width: 12,
              startValue: 0,
              endValue: 5,
              filterMode: 'empty',
              showDetail: false
            }
          ]
        : [],
    series: [
      {
        name: '联系人',
        type: 'bar',
        barMaxWidth: 20,
        data: stats.value.categories.map((item) => item.value),
        itemStyle: { color: '#67c23a', borderRadius: [0, 3, 3, 0] },
        label: { show: true, position: 'right', color: '#606266' }
      }
    ]
  }))

  const refresh = async () => {
    loading.value = true
    failed.value = false
    try {
      const res = await getRidgeContactStat()
      if (res.code !== 0 || !res.data) throw new Error('获取统计失败')
      stats.value = res.data
    } catch {
      failed.value = true
    } finally {
      loading.value = false
    }
  }

  onMounted(refresh)
  defineExpose({ refresh })
</script>

<style scoped>
  .contact-stats {
    display: grid;
    grid-template-columns: minmax(0, 1.4fr) minmax(0, 1fr) minmax(0, 1.1fr);
    gap: 8px;
    margin: 8px 0;
    min-height: 310px;
  }
  .stat-panel {
    min-width: 0;
    padding: 18px 16px 10px;
    border-radius: 4px;
    background: var(--el-bg-color);
  }
  header {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    justify-content: space-between;
    gap: 8px;
  }
  h3 {
    margin: 0;
    color: var(--el-text-color-primary);
    font-size: 15px;
    font-weight: 600;
  }
  header span {
    color: var(--el-text-color-secondary);
    font-size: 12px;
  }
  .stat-chart {
    width: 100%;
    height: 260px;
  }
  .stats-error {
    grid-column: 1 / -1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    background: var(--el-bg-color);
    color: var(--el-text-color-secondary);
  }
  @media (max-width: 1100px) {
    .contact-stats {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
    .trend-panel {
      grid-column: 1 / -1;
    }
  }
  @media (max-width: 640px) {
    .contact-stats {
      grid-template-columns: minmax(0, 1fr);
    }
  }
</style>

<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Reports &amp; Analytics</h2>
      <p class="text-muted text-sm">Call volume, agent performance, and system usage.</p>
    </div>
    <div class="date-filter">
      <button
        class="filter-btn"
        :class="{ active: activeFilter === 'last24h' }"
        @click="setFilter('last24h')"
      >Last 24h</button>
      <button
        class="filter-btn"
        :class="{ active: activeFilter === 'last7days' }"
        @click="setFilter('last7days')"
      >Last 7 Days</button>
      <button
        class="filter-btn"
        :class="{ active: activeFilter === 'last30days' }"
        @click="setFilter('last30days')"
      >Last 30 Days</button>
      <button
        class="filter-btn"
        :class="{ active: activeFilter === 'custom' }"
        @click="setFilter('custom')"
      >Custom</button>
      <button class="btn-primary" @click="exportCsv">Export CSV</button>
    </div>
  </div>

  <div v-if="activeFilter === 'custom'" class="custom-range">
    <label>
      From
      <input type="date" v-model="customStart" @change="fetchData" />
    </label>
    <label>
      To
      <input type="date" v-model="customEnd" @change="fetchData" />
    </label>
  </div>

  <div class="kpi-grid">
    <div class="kpi-card">
      <div class="kpi-label">Total Calls</div>
      <div class="kpi-value">{{ kpis.totalCalls }}</div>
      <div class="kpi-trend up" v-if="kpis.totalCallsTrend">{{ kpis.totalCallsTrend }}</div>
    </div>
    <div class="kpi-card">
      <div class="kpi-label">Avg Handle Time</div>
      <div class="kpi-value">{{ kpis.avgHandle }}</div>
      <div class="kpi-trend down" v-if="kpis.avgHandleTrend">{{ kpis.avgHandleTrend }}</div>
    </div>
    <div class="kpi-card">
      <div class="kpi-label">Missed Calls</div>
      <div class="kpi-value">{{ kpis.missed }}</div>
      <div class="kpi-trend bad" v-if="kpis.missedPct">{{ kpis.missedPct }}</div>
    </div>
    <div class="kpi-card">
      <div class="kpi-label">SLA Breached</div>
      <div class="kpi-value">{{ kpis.slaBreach }}</div>
      <div class="kpi-trend good" v-if="kpis.slaBreachPct">{{ kpis.slaBreachPct }}</div>
    </div>
  </div>

  <div class="charts-section">
    <div class="chart-container main">
      <h3>Call Volume ({{ volumeMeta.interval === 'hour' ? 'Hourly' : 'Daily' }})</h3>
      <div class="chart-wrapper">
        <Bar v-if="barChartData.labels.length" :data="barChartData" :options="barChartOptions" />
        <div v-else class="chart-empty">No data available</div>
      </div>
    </div>

    <div class="chart-container side">
      <h3>Disposition</h3>
      <div class="chart-wrapper donut-wrapper">
        <Doughnut v-if="donutTotal > 0" :data="donutChartData" :options="donutChartOptions" />
        <div v-else class="chart-empty">No data available</div>
        <div v-if="donutTotal > 0" class="donut-center">
          <span class="total">{{ donutTotal }}</span>
          <span class="label">Calls</span>
        </div>
      </div>
      <div class="legend">
        <div class="item">
          <span class="dot color-1"></span>Answered ({{ dispositions.answered }})
        </div>
        <div class="item">
          <span class="dot color-2"></span>Missed ({{ dispositions.missed }})
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Bar, Doughnut } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js'
import { reportsAPI } from '../../services/api'

ChartJS.register(CategoryScale, LinearScale, BarElement, ArcElement, Title, Tooltip, Legend)

const kpis = reactive({
  totalCalls: '—',
  totalCallsTrend: '',
  avgHandle: '—',
  avgHandleTrend: '',
  missed: '—',
  missedPct: '',
  slaBreach: '—',
  slaBreachPct: ''
})

const volumeRows = ref([])
const volumeMeta = reactive({ interval: 'hour', start: '', end: '' })
const activeFilter = ref('last7days')
const customStart = ref('')
const customEnd = ref('')

const dispositions = reactive({ answered: 0, missed: 0 })

function formatDateInput(d) {
  const iso = d.toISOString().split('T')[0]
  return iso
}

function getFilterConfig(filter) {
  const today = new Date()
  const end = formatDateInput(today)
  let start, interval

  if (filter === 'last24h') {
    const yesterday = new Date(today)
    yesterday.setDate(yesterday.getDate() - 1)
    start = formatDateInput(yesterday)
    interval = 'hour'
  } else if (filter === 'last7days') {
    const weekAgo = new Date(today)
    weekAgo.setDate(weekAgo.getDate() - 6)
    start = formatDateInput(weekAgo)
    interval = 'hour'
  } else if (filter === 'last30days') {
    const monthAgo = new Date(today)
    monthAgo.setDate(monthAgo.getDate() - 29)
    start = formatDateInput(monthAgo)
    interval = 'day'
  } else if (filter === 'custom') {
    const s = customStart.value
    const e = customEnd.value
    if (s && e) {
      const startDate = new Date(s)
      const endDate = new Date(e)
      const diffDays = (endDate - startDate) / (1000 * 60 * 60 * 24)
      interval = diffDays <= 3 ? 'hour' : 'day'
      return { start: s, end: e, interval }
    }
    // Fallback when custom dates not yet set
    const weekAgo = new Date(today)
    weekAgo.setDate(weekAgo.getDate() - 6)
    start = formatDateInput(weekAgo)
    interval = 'day'
  }

  return { start, end, interval }
}

function formatPeriod(periodStr, interval) {
  if (!periodStr) return ''
  const d = new Date(periodStr)
  if (isNaN(d)) return periodStr
  if (interval === 'hour') {
    return d.toLocaleTimeString(undefined, { hour: '2-digit', hour12: true })
  }
  return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
}

const barChartData = computed(() => {
  const labels = volumeRows.value.map((r) => formatPeriod(r.period, volumeMeta.interval))
  const data = volumeRows.value.map((r) => r.total_calls ?? 0)
  return {
    labels,
    datasets: [
      {
        label: 'Total Calls',
        data,
        backgroundColor: '#3b82f6',
        borderRadius: 4,
        barPercentage: 0.7,
        categoryPercentage: 0.8,
      },
    ],
  }
})

const barChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: '#0f172a',
      titleColor: '#f8fafc',
      bodyColor: '#f8fafc',
      padding: 10,
      cornerRadius: 6,
      callbacks: {
        title: (items) => items[0]?.label || '',
        label: (item) => `${item.raw} calls`,
      },
    },
  },
  scales: {
    y: {
      beginAtZero: true,
      grid: { color: '#f1f5f9' },
      ticks: { color: '#94a3b8', font: { size: 10 } },
      border: { display: false },
    },
    x: {
      grid: { display: false },
      ticks: {
        color: '#94a3b8',
        font: { size: 10 },
        maxTicksLimit: 12,
      },
      border: { display: false },
    },
  },
}

const donutTotal = computed(() => {
  return volumeRows.value.reduce((sum, r) => sum + (r.total_calls ?? 0), 0)
})

const donutChartData = computed(() => {
  const answered = volumeRows.value.reduce((sum, r) => sum + (r.answered ?? 0), 0)
  const missed = volumeRows.value.reduce((sum, r) => sum + (r.missed ?? 0), 0)
  return {
    labels: ['Answered', 'Missed'],
    datasets: [
      {
        data: [answered, missed],
        backgroundColor: ['#3b82f6', '#ef4444'],
        borderWidth: 0,
        hoverOffset: 4,
      },
    ],
  }
})

const donutChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  cutout: '70%',
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: '#0f172a',
      titleColor: '#f8fafc',
      bodyColor: '#f8fafc',
      padding: 10,
      cornerRadius: 6,
      callbacks: {
        label: (item) => `${item.label}: ${item.raw}`,
      },
    },
  },
}

const setFilter = (filter) => {
  activeFilter.value = filter
  fetchData()
}

const fetchData = async () => {
  const { start, end, interval } = getFilterConfig(activeFilter.value)
  try {
    const [volRes, kpiRes] = await Promise.all([
      reportsAPI.callVolume({ start, end, interval }),
      reportsAPI.kpi({ start, end }),
    ])

    const rows = volRes.data || []
    volumeRows.value = rows
    if (volRes._meta) {
      volumeMeta.interval = volRes._meta.interval || interval
      volumeMeta.start = volRes._meta.start
      volumeMeta.end = volRes._meta.end
    } else {
      volumeMeta.interval = interval
      volumeMeta.start = start
      volumeMeta.end = end
    }

    const kpi = kpiRes.data || {}
    const total = kpi.total_calls || 0
    kpis.totalCalls = total.toLocaleString()
    kpis.totalCallsTrend = ''

    const acdSec = Math.round(kpi.acd_seconds || 0)
    kpis.avgHandle = acdSec > 60 ? `${Math.floor(acdSec / 60)}m ${acdSec % 60}s` : `${acdSec}s`
    kpis.avgHandleTrend = ''

    const missed = kpi.missed_calls || 0
    kpis.missed = String(missed)
    kpis.missedPct = total > 0 ? `${Math.round((missed / total) * 100)}% of total` : ''

    kpis.slaBreach = '—'
    kpis.slaBreachPct = ''

    const totalAnswered = rows.reduce((s, r) => s + (r.answered || 0), 0)
    const totalMissed = rows.reduce((s, r) => s + (r.missed || 0), 0)
    dispositions.answered = totalAnswered
    dispositions.missed = totalMissed
  } catch (err) {
    console.error('Failed to load report data:', err)
  }
}

onMounted(() => {
  const today = formatDateInput(new Date())
  const weekAgo = formatDateInput(new Date(Date.now() - 6 * 24 * 60 * 60 * 1000))
  customStart.value = weekAgo
  customEnd.value = today
  fetchData()
})

const exportCsv = async () => {
  try {
    const { start, end } = getFilterConfig(activeFilter.value)
    const res = await reportsAPI.export({ type: 'call-volume', start, end })
    const blob = new Blob([res.data], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `report-${start}-to-${end}.csv`
    link.click()
    URL.revokeObjectURL(url)
  } catch (err) {
    console.error('Failed to export report:', err)
  }
}
</script>

<style scoped>
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
}

.date-filter {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.filter-btn {
  background: white;
  border: 1px solid var(--border-color);
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--text-main);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.filter-btn:hover {
  border-color: var(--border-hover);
  background: var(--bg-hover);
}

.filter-btn.active {
  background: var(--primary-subtle);
  border-color: var(--primary-color);
  color: var(--primary-text);
}

.btn-primary {
  background-color: var(--primary-color);
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-weight: 500;
  font-size: var(--text-sm);
  cursor: pointer;
  transition: background var(--transition-fast);
}

.btn-primary:hover {
  background-color: var(--primary-hover);
}

.custom-range {
  display: flex;
  gap: 16px;
  margin-bottom: var(--spacing-lg);
  align-items: center;
}

.custom-range label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: var(--text-sm);
  color: var(--text-main);
  font-weight: 500;
}

.custom-range input[type='date'] {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  padding: 6px 10px;
  font-size: var(--text-sm);
  color: var(--text-main);
  font-family: inherit;
}

/* KPIs */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--spacing-lg);
  margin-bottom: var(--spacing-xl);
}

.kpi-card {
  background: white;
  padding: var(--spacing-lg);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
  box-shadow: var(--shadow-sm);
}

.kpi-label {
  font-size: 11px;
  text-transform: uppercase;
  color: var(--text-muted);
  font-weight: 700;
  margin-bottom: 4px;
}
.kpi-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 4px;
}
.kpi-trend {
  font-size: 11px;
  font-weight: 600;
}
.kpi-trend.up {
  color: var(--status-good);
}
.kpi-trend.down {
  color: var(--status-good);
}
.kpi-trend.bad {
  color: var(--status-bad);
}
.kpi-trend.good {
  color: var(--status-good);
}

/* CHARTS */
.charts-section {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: var(--spacing-lg);
}

.chart-container {
  background: white;
  padding: var(--spacing-lg);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
  box-shadow: var(--shadow-sm);
}

.chart-container h3 {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 24px;
}

.chart-wrapper {
  height: 240px;
  position: relative;
}

.chart-empty {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  font-size: var(--text-sm);
}

.donut-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
}

.donut-center {
  position: absolute;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  pointer-events: none;
}

.total {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
}
.label {
  font-size: 10px;
  color: var(--text-muted);
  text-transform: uppercase;
}

.legend .item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: var(--text-sm);
  margin-bottom: 8px;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}
.color-1 {
  background: var(--primary-color);
}
.color-2 {
  background: #ef4444;
}

@media (max-width: 1024px) {
  .charts-section {
    grid-template-columns: 1fr;
  }
  .kpi-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .kpi-grid {
    grid-template-columns: 1fr;
  }
  .view-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-md);
  }
}
</style>

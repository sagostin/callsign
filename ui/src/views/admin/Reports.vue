<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Reports &amp; Analytics</h2>
      <p class="text-muted text-sm">Call volume, agent performance, and system usage.</p>
    </div>
    <div class="date-filter">
      <button 
        class="btn-secondary"
        :class="{ active: activeFilter === 'last7days' }"
        @click="filterLast7Days"
      >Last 7 Days</button>
      <button class="btn-primary" @click="exportCsv">Export CSV</button>
    </div>
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
      <h3>Call Volume (Hourly)</h3>
      <div class="chart-placeholder">
        <div class="bar" v-for="(h, i) in hourlyData" :key="i" :style="{ height: h + '%' }"></div>
      </div>
      <div class="chart-axis">
        <span>8am</span><span>10am</span><span>12pm</span><span>2pm</span><span>4pm</span>
      </div>
    </div>

    <div class="chart-container side">
      <h3>Disposition</h3>
      <div class="donut-chart" :style="donutStyle">
        <div class="donut-segment"></div>
        <div class="donut-center">
          <span class="total">{{ kpis.totalCalls }}</span>
          <span class="label">Calls</span>
        </div>
      </div>
      <div class="legend">
        <div class="item"><span class="dot color-1"></span>Answered ({{ dispositions.answered }}%)</div>
        <div class="item"><span class="dot color-2"></span>Voicemail ({{ dispositions.voicemail }}%)</div>
        <div class="item"><span class="dot color-3"></span>Abandoned ({{ dispositions.abandoned }}%)</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { cdrAPI, reportsAPI } from '../../services/api'

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

const hourlyData = ref([])
const dispositions = reactive({ answered: 0, voicemail: 0, abandoned: 0 })
const activeFilter = ref('last7days')

const filterLast7Days = () => {
  activeFilter.value = 'last7days'
  fetchReportData({ days: 7 })
}

const donutStyle = computed(() => {
  const a = dispositions.answered
  const v = dispositions.voicemail
  return {
    background: `conic-gradient(var(--primary-color) 0% ${a}%, #F59E0B ${a}% ${a + v}%, #EF4444 ${a + v}% 100%)`
  }
})

const fetchReportData = async (params = {}) => {
  try {
    const res = await reportsAPI.callVolume(params)
    const data = res.data
    
    if (data?.summary) {
      const s = data.summary
      kpis.totalCalls = s.total_calls?.toLocaleString() || '0'
      kpis.totalCallsTrend = s.trend ? `${s.trend > 0 ? '+' : ''}${s.trend}% vs last week` : ''
      const avgSec = s.avg_duration || 0
      kpis.avgHandle = avgSec > 60 ? `${Math.floor(avgSec / 60)}m ${avgSec % 60}s` : `${avgSec}s`
      kpis.missed = String(s.missed_calls || 0)
      const total = s.total_calls || 1
      kpis.missedPct = s.missed_calls ? `${Math.round(s.missed_calls / total * 100)}% of total` : ''
      kpis.slaBreach = String(s.sla_breached || 0)
      kpis.slaBreachPct = s.sla_breached ? `${(s.sla_breached / total * 100).toFixed(1)}%` : ''
      
      if (s.hourly && Array.isArray(s.hourly) && s.hourly.length > 0) {
        const max = Math.max(...s.hourly, 1)
        hourlyData.value = s.hourly.map(v => Math.round(v / max * 100))
      }
      
      if (s.dispositions) {
        dispositions.answered = s.dispositions.answered || 85
        dispositions.voicemail = s.dispositions.voicemail || 10
        dispositions.abandoned = s.dispositions.abandoned || 5
      }
    }
  } catch (err) {
    console.error('Failed to load report data:', err)
  }
}

onMounted(fetchReportData)

const exportCsv = async () => {
  try {
    const res = await reportsAPI.callVolume()
    const data = res.data?.summary

    // Early exit if no data available
    if (!data) {
      console.error('No report data to export')
      return
    }

    const headers = ['Metric', 'Value']
    const rows = [
      ['Total Calls', data.total_calls ?? 0],
      ['Avg Handle Time (seconds)', data.avg_duration ?? 0],
      ['Missed Calls', data.missed_calls ?? 0],
      ['SLA Breached', data.sla_breached ?? 0],
      ['Trend (%)', data.trend ?? 0],
      ['Hourly Data', (data.hourly ?? []).join('; ')],
      ['Disposition Answered (%)', data.dispositions?.answered ?? 0],
      ['Disposition Voicemail (%)', data.dispositions?.voicemail ?? 0],
      ['Disposition Abandoned (%)', data.dispositions?.abandoned ?? 0],
    ]

    const csvContent = [
      headers.join(','),
      ...rows.map(row => row.map(cell => String(cell)).join(','))
    ].join('\n')

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `report-${new Date().toISOString().slice(0, 10)}.csv`
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
}

.btn-secondary {
  background: white;
  border: 1px solid var(--border-color);
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--text-main);
  cursor: pointer;
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

.kpi-label { font-size: 11px; text-transform: uppercase; color: var(--text-muted); font-weight: 700; margin-bottom: 4px; }
.kpi-value { font-size: 28px; font-weight: 700; color: var(--text-primary); margin-bottom: 4px; }
.kpi-trend { font-size: 11px; font-weight: 600; }
.kpi-trend.up { color: var(--status-good); }
.kpi-trend.down { color: var(--status-good); }
.kpi-trend.bad { color: var(--status-bad); }

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

.chart-container h3 { font-size: 16px; font-weight: 600; margin-bottom: 24px; }

.chart-placeholder {
  height: 200px;
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  padding: 0 10px;
  border-bottom: 1px solid var(--border-color);
}

.bar {
  width: 8%;
  background: var(--primary-light);
  border-radius: 4px 4px 0 0;
  transition: height 0.5s ease;
}
.bar:hover { background: var(--primary-color); }

.chart-axis {
  display: flex;
  justify-content: space-between;
  margin-top: 8px;
  font-size: 10px;
  color: var(--text-muted);
}

.donut-chart {
  width: 140px;
  height: 140px;
  border-radius: 50%;
  margin: 0 auto 24px;
  position: relative;
}

.donut-center {
  position: absolute;
  top: 20px;
  left: 20px;
  width: 100px;
  height: 100px;
  background: white;
  border-radius: 50%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.total { font-size: 20px; font-weight: 700; color: var(--text-primary); }
.label { font-size: 10px; color: var(--text-muted); text-transform: uppercase; }

.legend .item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: var(--text-sm);
  margin-bottom: 8px;
}

.dot { width: 8px; height: 8px; border-radius: 50%; }
.color-1 { background: var(--primary-color); }
.color-2 { background: #F59E0B; }
.color-3 { background: #EF4444; }
</style>

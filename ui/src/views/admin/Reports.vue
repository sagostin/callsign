<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Reports &amp; Analytics</h2>
      <p class="text-muted text-sm">Call volume, agent performance, and system usage.</p>
    </div>
    <div class="date-filter">
      <button class="btn-secondary">Last 7 Days</button>
      <button class="btn-primary">Export CSV</button>
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
import { cdrAPI } from '../../services/api'

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

const hourlyData = ref([40, 60, 30, 80, 50, 90, 70, 45, 65, 35])
const dispositions = reactive({ answered: 85, voicemail: 10, abandoned: 5 })

const donutStyle = computed(() => {
  const a = dispositions.answered
  const v = dispositions.voicemail
  return {
    background: `conic-gradient(var(--primary-color) 0% ${a}%, #F59E0B ${a}% ${a + v}%, #EF4444 ${a + v}% 100%)`
  }
})

const fetchReportData = async () => {
  try {
    const res = await cdrAPI.getSummary ? await cdrAPI.getSummary() : await cdrAPI.list({ limit: 1000 })
    const data = res.data

    if (data?.summary) {
      const s = data.summary
      kpis.totalCalls = s.total_calls?.toLocaleString() || '0'
      kpis.totalCallsTrend = s.trend ? `${s.trend > 0 ? '+' : ''}${s.trend}% vs last week` : ''
      const avgSec = s.avg_duration || 0
      kpis.avgHandle = avgSec > 60 ? `${Math.floor(avgSec / 60)}m ${avgSec % 60}s` : `${avgSec}s`
      kpis.avgHandleTrend = s.avg_trend ? `${s.avg_trend}%` : ''
      kpis.missed = String(s.missed_calls || 0)
      const total = s.total_calls || 1
      kpis.missedPct = s.missed_calls ? `${Math.round(s.missed_calls / total * 100)}% of total` : ''
      kpis.slaBreach = String(s.sla_breached || 0)
      kpis.slaBreachPct = s.sla_breached ? `${(s.sla_breached / total * 100).toFixed(1)}%` : ''

      if (s.dispositions) {
        dispositions.answered = s.dispositions.answered || 85
        dispositions.voicemail = s.dispositions.voicemail || 10
        dispositions.abandoned = s.dispositions.abandoned || 5
      }
      if (s.hourly) {
        const max = Math.max(...s.hourly, 1)
        hourlyData.value = s.hourly.map(v => Math.round(v / max * 100))
      }
    } else {
      const records = data?.data || data || []
      kpis.totalCalls = records.length.toLocaleString()
      const missed = records.filter(r => r.status === 'Missed' || r.status === 'No Answer').length
      kpis.missed = String(missed)
      kpis.missedPct = records.length > 0 ? `${Math.round(missed / records.length * 100)}% of total` : ''
      if (records.length > 0) {
        const avgDur = Math.round(records.reduce((s, r) => s + (r.duration || 0), 0) / records.length)
        kpis.avgHandle = avgDur > 60 ? `${Math.floor(avgDur / 60)}m ${avgDur % 60}s` : `${avgDur}s`
      }
      kpis.slaBreach = '0'
    }
  } catch (err) {
    console.error('Failed to load report data:', err)
  }
}

onMounted(fetchReportData)
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

<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Reports & Analytics</h2>
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
      <div class="kpi-value">1,248</div>
      <div class="kpi-trend up">+12% vs last week</div>
    </div>
    <div class="kpi-card">
      <div class="kpi-label">Avg Handle Time</div>
      <div class="kpi-value">4m 12s</div>
      <div class="kpi-trend down">-5% (Good)</div>
    </div>
    <div class="kpi-card">
      <div class="kpi-label">Missed Calls</div>
      <div class="kpi-value">24</div>
      <div class="kpi-trend bad">3% of total</div>
    </div>
    <div class="kpi-card">
      <div class="kpi-label">SLA Breached</div>
      <div class="kpi-value">2</div>
      <div class="kpi-trend good">0.1%</div>
    </div>
  </div>

  <div class="charts-section">
    <div class="chart-container main">
      <h3>Call Volume (Hourly)</h3>
      <div class="chart-placeholder">
        <!-- Visual Mockup of a Bar Chart using CSS -->
        <div class="bar" style="height: 40%"></div>
        <div class="bar" style="height: 60%"></div>
        <div class="bar" style="height: 30%"></div>
        <div class="bar" style="height: 80%"></div>
        <div class="bar" style="height: 50%"></div>
        <div class="bar" style="height: 90%"></div>
        <div class="bar" style="height: 70%"></div>
        <div class="bar" style="height: 45%"></div>
        <div class="bar" style="height: 65%"></div>
        <div class="bar" style="height: 35%"></div>
      </div>
      <div class="chart-axis">
        <span>8am</span><span>10am</span><span>12pm</span><span>2pm</span><span>4pm</span>
      </div>
    </div>

    <div class="chart-container side">
      <h3>Disposition</h3>
      <div class="donut-chart">
        <div class="donut-segment"></div>
        <div class="donut-center">
          <span class="total">1.2k</span>
          <span class="label">Calls</span>
        </div>
      </div>
      <div class="legend">
        <div class="item"><span class="dot color-1"></span>Answered (85%)</div>
        <div class="item"><span class="dot color-2"></span>Voicemail (10%)</div>
        <div class="item"><span class="dot color-3"></span>Abandoned (5%)</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { FileDown } from 'lucide-vue-next'
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
.kpi-trend.down { color: var(--status-good); /* Sometimes down is good like handling time */ }
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
  background: conic-gradient(
    var(--primary-color) 0% 85%,
    #F59E0B 85% 95%,
    #EF4444 95% 100%
  );
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

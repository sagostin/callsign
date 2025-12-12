<template>
  <div class="view-header">
    <h2>System Overview</h2>
    <p class="text-muted text-sm">Real-time system status and metrics.</p>
  </div>

  <div class="dashboard-grid">
    <StatCard 
      label="Active Calls" 
      value="24" 
      subtext="Peak: 42 today"
      :trend="12"
      iconName="calls"
    />
    <StatCard 
      label="Registrations" 
      value="856" 
      subtext="98% Online" 
      trend="1"
      iconName="server"
    />
    <StatCard 
      label="Failed Calls (1h)" 
      value="3" 
      subtext="SIP 503 from Gateway B"
      iconName="alert"
    />
    <StatCard 
      label="System Health" 
      value="Operational" 
      subtext="Uptime: 45d 12h"
      iconName="activity"
    />
  </div>

  <div class="layout-split">
    <!-- Recent Alerts -->
    <div class="panel">
      <div class="panel-header">
        <h3>Recent Alerts</h3>
        <button class="btn-link">View All</button>
      </div>
      <div class="alert-list">
        <div class="alert-item">
          <span class="badge bad">CRITICAL</span>
          <div class="alert-content">
            <div class="alert-msg">SIP Trunk Group "East" requires attention</div>
            <div class="alert-time">2 mins ago</div>
          </div>
        </div>
        <div class="alert-item">
          <span class="badge warn">WARNING</span>
          <div class="alert-content">
            <div class="alert-msg">High latency on Node 4 (150ms)</div>
            <div class="alert-time">15 mins ago</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Active Calls Table (Mini) -->
    <div class="panel">
      <div class="panel-header">
        <h3>Live Calls</h3>
      </div>
      <table class="simple-table">
        <thead>
          <tr>
            <th>From</th>
            <th>To</th>
            <th>Duration</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>Ext 101</td>
            <td>(555) 123-4567</td>
            <td>04:23</td>
          </tr>
          <tr>
            <td>(415) 555-9090</td>
            <td>Queue: Sales</td>
            <td>01:12</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import StatCard from '../../components/features/StatCard.vue'
</script>

<style scoped>
.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: var(--spacing-lg);
  margin: var(--spacing-lg) 0 var(--spacing-xl);
}

.layout-split {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-lg);
}

.panel {
  background: white;
  border-radius: var(--radius-md);
  padding: var(--spacing-lg);
  box-shadow: var(--shadow-sm);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-md);
  border-bottom: 1px solid var(--border-color);
  padding-bottom: var(--spacing-sm);
}

.panel-header h3 {
  font-size: var(--text-sm);
  text-transform: uppercase;
  color: var(--text-muted);
  letter-spacing: 0.05em;
  font-weight: 600;
}

.alert-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.alert-item {
  display: flex;
  gap: var(--spacing-md);
  align-items: flex-start;
  padding: 12px;
  background-color: var(--bg-app);
  border-radius: var(--radius-md);
  transition: transform var(--transition-fast);
}

.alert-item:hover {
  transform: translateX(2px);
}

.alert-content {
  flex: 1;
}

.alert-msg {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--text-primary);
}

.alert-time {
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 2px;
}

.simple-table {
  width: 100%;
}
.simple-table th {
  text-align: left;
  font-size: 11px;
  color: var(--text-muted);
  font-weight: 700;
  text-transform: uppercase;
  padding-bottom: 12px;
}
.simple-table td {
  font-size: var(--text-sm);
  padding: 8px 0;
  border-bottom: 1px solid var(--border-color);
  color: var(--text-main);
}
.simple-table tr:last-child td {
  border-bottom: none;
}
.btn-link {
  background: none;
  border: none;
  color: var(--primary-color);
  font-size: var(--text-xs);
  cursor: pointer;
  font-weight: 600;
}
</style>

<template>
  <div class="view-header">
    <h2>Infrastructure</h2>
    <p class="text-muted text-sm">Cluster health, node status, and capacity planning.</p>
  </div>

  <div class="node-grid">
    <div class="node-card active">
      <div class="node-header">
        <h3>Node A (Primary)</h3>
        <StatusBadge status="good" />
      </div>
      <div class="node-stats">
        <div class="stat-row">
          <span>Registrations</span>
          <span>4,231</span>
        </div>
        <div class="stat-row">
          <span>Active Calls</span>
          <span>128</span>
        </div>
        <div class="stat-row">
          <span>CPU Load</span>
          <span class="text-good">12%</span>
        </div>
        <div class="stat-row">
          <span>RTP Health</span>
          <span class="text-good">99.9%</span>
        </div>
      </div>
    </div>

    <div class="node-card active">
      <div class="node-header">
        <h3>Node B (Failover)</h3>
        <StatusBadge status="good" />
      </div>
      <div class="node-stats">
        <div class="stat-row">
          <span>Registrations</span>
          <span>0</span>
        </div>
        <div class="stat-row">
          <span>Active Calls</span>
          <span>0</span>
        </div>
        <div class="stat-row">
          <span>CPU Load</span>
          <span class="text-good">2%</span>
        </div>
        <div class="stat-row">
          <span>Sync Status</span>
          <span class="text-good">Synced</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import StatusBadge from '../components/common/StatusBadge.vue'
</script>

<style scoped>
.node-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--spacing-lg);
  margin-top: var(--spacing-lg);
}

.node-card {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-lg);
  box-shadow: var(--shadow-sm);
}

.node-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-md);
  padding-bottom: var(--spacing-sm);
  border-bottom: 1px solid var(--border-color);
}

.node-header h3 {
  font-size: var(--text-lg);
  margin: 0;
}

.stat-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px dashed var(--border-color);
  font-size: var(--text-sm);
}

.stat-row:last-child {
  border-bottom: none;
}

.text-good { color: var(--status-good); font-weight: 600; }
</style>

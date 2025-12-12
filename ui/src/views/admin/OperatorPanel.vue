<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Operator Panel</h2>
      <p class="text-muted text-sm">Real-time view of active extensions and calls.</p>
    </div>
    <div class="header-actions">
      <div class="status-indicator">
        <span class="dot pulse"></span>
        Live
      </div>
    </div>
  </div>

  <div class="panel-grid">
    <div v-for="ext in extensions" :key="ext.id" class="ext-card" :class="ext.status.toLowerCase()">
      <div class="ext-header">
        <span class="ext-number">{{ ext.number }}</span>
        <span class="ext-status">{{ ext.status }}</span>
      </div>
      <div class="ext-name">{{ ext.name }}</div>
      <div class="ext-detail" v-if="ext.currentCmd">{{ ext.currentCmd }}</div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const extensions = ref([
  { id: 1, number: '101', name: 'Alice Smith', status: 'Available', currentCmd: '' },
  { id: 2, number: '102', name: 'Bob Jones', status: 'Busy', currentCmd: 'Talking 03:42' },
  { id: 3, number: '103', name: 'Charlie Day', status: 'Ringing', currentCmd: 'Incoming...' },
  { id: 4, number: '104', name: 'Dana White', status: 'Offline', currentCmd: '' },
  { id: 5, number: '105', name: 'Evan Gold', status: 'Available', currentCmd: '' },
])
</script>

<style scoped>
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #f0fdf4;
  color: #166534;
  padding: 4px 12px;
  border-radius: 99px;
  font-size: 12px;
  font-weight: 600;
}

.dot {
  width: 8px;
  height: 8px;
  background-color: #22c55e;
  border-radius: 50%;
}

.pulse {
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(34, 197, 94, 0.7); }
  70% { transform: scale(1); box-shadow: 0 0 0 6px rgba(34, 197, 94, 0); }
  100% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(34, 197, 94, 0); }
}

.panel-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: var(--spacing-md);
}

.ext-card {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  background: white;
  transition: all 0.2s;
}

.ext-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-sm);
}

.ext-card.busy { border-color: #fca5a5; background: #fef2f2; }
.ext-card.ringing { border-color: #fdba74; background: #fff7ed; animation: border-pulse 1s infinite alternate; }
.ext-card.offline { opacity: 0.6; background: #f9fafb; }

.ext-header {
  display: flex;
  justify-content: space-between;
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 4px;
}

.ext-status { font-size: 11px; font-weight: 600; text-transform: uppercase; align-self: center; }

.ext-name { font-size: 13px; color: var(--text-muted); margin-bottom: 8px; }

.ext-detail {
  font-size: 11px;
  color: var(--text-main);
  background: rgba(255,255,255,0.5);
  padding: 2px 6px;
  border-radius: 4px;
}

@keyframes border-pulse {
  from { border-color: #fdba74; }
  to { border-color: #f97316; }
}
</style>

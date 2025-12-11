<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Gateways (Trunks)</h2>
      <p class="text-muted text-sm">Manage external SIP trunks and PSTN gateways.</p>
    </div>
    <div class="header-actions">
      <button class="btn-primary" @click="$router.push('/admin/gateways/new')">+ Add Gateway</button>
    </div>
  </div>

  <DataTable :columns="columns" :data="gateways" actions>
    <template #status="{ value }">
      <StatusBadge :status="value" />
    </template>
    
    <template #type="{ value }">
        <span class="badge" :class="value.toLowerCase()">{{ value }}</span>
    </template>

    <template #actions>
      <button class="btn-link" @click="$router.push('/admin/gateways/1')">Edit</button>
      <button class="btn-link text-bad" @click="deleteGateway">Delete</button>
    </template>
  </DataTable>
</template>

<script setup>
import { ref } from 'vue'
import DataTable from '../components/common/DataTable.vue'
import StatusBadge from '../components/common/StatusBadge.vue'

const columns = [
  { key: 'name', label: 'Gateway Name' },
  { key: 'hostname', label: 'Hostname / IP' },
  { key: 'type', label: 'Type' },
  { key: 'status', label: 'Status' }
]

const gateways = ref([
  { name: 'Flowroute Primary', hostname: 'sip.flowroute.com', type: 'Public', status: 'Registered' },
  { name: 'Twilio Elastic SIP', hostname: 'callsign.pstn.twilio.com', type: 'Public', status: 'Registered' },
  { name: 'Local PRI Gateway', hostname: '192.168.1.200', type: 'Local', status: 'Error' },
])

const deleteGateway = () => confirm('Delete this gateway?') && alert('Gateway deleted.')
</script>

<style scoped>
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
}

.btn-primary {
  background-color: var(--primary-color);
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-weight: 500;
  font-size: var(--text-sm);
}

.btn-link {
  background: none;
  border: none;
  color: var(--primary-color);
  font-size: var(--text-xs);
  margin-left: 8px;
  cursor: pointer;
  font-weight: 500;
}

.text-bad { color: var(--status-bad); }

.badge { padding: 2px 6px; border-radius: 4px; font-size: 11px; font-weight: 600; background: var(--bg-secondary); }
.badge.public { background: #e0f2fe; color: #0369a1; }
.badge.local { background: #f3e8ff; color: #7e22ce; }
</style>

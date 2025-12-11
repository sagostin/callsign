<template>
  <div class="view-header flex justify-between items-center" style="margin-bottom: var(--spacing-lg)">
    <div>
      <h2>Phone Numbers</h2>
      <p class="text-muted text-sm">Manage incoming numbers (DIDs) and their routing.</p>
    </div>
    <button class="btn-primary" @click="$router.push('/numbers/new')">
      + Add Number
    </button>
  </div>
  
  <div class="messaging-config-bar" style="margin-bottom: 20px; background: #f8fafc; padding: 12px; border-radius: 6px; border: 1px solid #e2e8f0; display: flex; align-items: center; justify-content: space-between;">
    <div class="msg-info">
       <span style="font-weight: 600; font-size: 13px;">Tenant Messaging Configuration</span>
       <span style="font-size: 12px; color: #64748b; margin-left: 8px;">Configure API keys for SMS/MMS</span>
    </div>
    <button class="btn-secondary" @click="showMsgModal = true">Manage Settings</button>
  </div>

  <div class="modal-overlay" v-if="showMsgModal">
     <div class="modal">
        <h3>Messaging API Settings</h3>
        <p class="text-xs text-muted">Override global defaults with tenant-specific credentials.</p>
        
        <div class="form-group">
           <label>Provider</label>
           <select class="input-field">
              <option>Use System Default (Twilio)</option>
              <option>Twilio (Custom)</option>
              <option>Bandwidth (Custom)</option>
           </select>
        </div>
         <div class="form-group">
           <label>Account SID / User ID</label>
           <input type="text" class="input-field" placeholder="AC...">
        </div>
         <div class="form-group">
           <label>Auth Token / Secret</label>
           <input type="password" class="input-field" value="">
        </div>
        
        <div class="modal-actions">
           <button class="btn-secondary" @click="showMsgModal = false">Cancel</button>
           <button class="btn-primary" @click="showMsgModal = false">Save Configuration</button>
        </div>
     </div>
  </div>

  <DataTable :columns="columns" :data="numbers" actions>
    <template #status="{ value }">
      <StatusBadge :status="value" />
    </template>
    
    <template #routing="{ value }">
      <span class="route-target">{{ value }}</span>
    </template>

    <template #actions="{ row }">
      <button class="btn-link" @click="$router.push(`/numbers/${row.number.replace(/\D/g, '')}`)">Edit</button>
      <button class="btn-link">Stats</button>
    </template>
  </DataTable>
</template>

<script setup>
import { ref } from 'vue'
import DataTable from '../components/common/DataTable.vue'
import StatusBadge from '../components/common/StatusBadge.vue'

const columns = [
  { key: 'number', label: 'Number', width: '160px' },
  { key: 'usage', label: 'Usage' },
  { key: 'routing', label: 'Routing Target' },
  { key: 'carrier', label: 'Carrier' },
  { key: 'status', label: 'Status', width: '100px' }
]

const numbers = [
  { number: '(415) 555-0100', usage: 'Main Line', routing: 'IVR: Main Menu', carrier: 'Bandwidth', status: 'Active' },
  { number: '(415) 555-0101', usage: 'Fax', routing: 'Fax-to-Email', carrier: 'Bandwidth', status: 'Active' },
  { number: '(310) 555-9988', usage: 'Sales Direct', routing: 'Ring Group: Sales', carrier: 'Twilio', status: 'Active' },
  { number: '(212) 555-1234', usage: 'Reserved', routing: '-', carrier: 'Twilio', status: 'Idle' },
]

const showMsgModal = ref(false)
</script>

<style scoped>
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
}
.route-target {
  font-weight: 500;
  color: var(--text-primary);
}

.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 6px 12px; border-radius: var(--radius-sm); font-size: 12px; cursor: pointer; color: var(--text-main); }

.modal-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 100; }
.modal { background: white; padding: 24px; border-radius: var(--radius-md); width: 400px; display: flex; flex-direction: column; gap: 12px; box-shadow: var(--shadow-lg); }
.modal h3 { font-size: 16px; font-weight: 700; margin: 0; }
.form-group { display: flex; flex-direction: column; gap: 4px; }
.form-group label { font-size: 11px; font-weight: 700; color: var(--text-muted); text-transform: uppercase; }
.input-field { padding: 8px; border: 1px solid var(--border-color); border-radius: 4px; font-size: 13px; }
.modal-actions { display: flex; justify-content: flex-end; gap: 8px; margin-top: 8px; }
.text-xs { font-size: 11px; }
</style>

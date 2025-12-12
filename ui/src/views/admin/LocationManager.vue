<template>
  <div class="location-manager">
    <div class="header-actions">
      <h3>Physical Locations</h3>
      <button class="btn-secondary small">+ Add Location</button>
    </div>
    <p class="text-muted text-sm mb-lg">Manage site-specific settings for E911 and Caller ID.</p>
    
    <div v-for="(loc, index) in locations" :key="index" class="location-card">
       <div class="card-header">
         <span class="loc-name">{{ loc.name }}</span>
         <div class="actions">
            <button class="btn-link">Edit</button>
            <button class="btn-link text-bad">Remove</button>
         </div>
       </div>
       <div class="card-body">
          <div class="info-row">
             <span class="label">Main Number:</span>
             <span class="value">{{ loc.mainNumber }}</span>
          </div>
          <div class="info-row">
             <span class="label">Caller ID Name:</span>
             <span class="value">{{ loc.cidName }}</span>
          </div>
          <div class="info-row">
             <span class="label">Emergency Fallback:</span>
             <span class="value text-warn">{{ loc.fallback }}</span>
          </div>
          <div class="info-row">
             <span class="label">Address:</span>
             <span class="value">{{ loc.address }}</span>
          </div>
       </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const locations = ref([
  { 
    name: 'HQ - San Francisco', 
    mainNumber: '(415) 555-0100', 
    cidName: 'CallSign HQ',
    fallback: '(415) 555-9999',
    address: '123 Market St, San Francisco, CA 94105'
  },
  { 
    name: 'Warehouse - Nevada', 
    mainNumber: '(775) 555-0200', 
    cidName: 'CallSign WH',
    fallback: '(775) 555-9988',
    address: '456 Industrial Pkwy, Reno, NV 89502'
  }
])
</script>

<style scoped>
.location-manager { padding-top: 8px; }

.header-actions { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.header-actions h3 { margin: 0; font-size: 16px; font-weight: 600; }

.mb-lg { margin-bottom: 24px; }

.location-card {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  padding: 16px;
  margin-bottom: 16px;
  background: white;
}

.card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; border-bottom: 1px dashed var(--border-color); padding-bottom: 8px; }
.loc-name { font-weight: 600; font-size: 14px; color: var(--text-primary); }

.card-body { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }

.info-row { display: flex; flex-direction: column; gap: 2px; }
.label { font-size: 10px; text-transform: uppercase; color: var(--text-muted); font-weight: 700; }
.value { font-size: 13px; color: var(--text-main); }
.text-warn { color: #d97706; }

.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 4px 12px; border-radius: 4px; font-size: 12px; cursor: pointer; }
.btn-link { background: none; border: none; color: var(--primary-color); font-size: 12px; font-weight: 600; cursor: pointer; margin-left: 8px; }
.text-bad { color: var(--status-bad); }

</style>

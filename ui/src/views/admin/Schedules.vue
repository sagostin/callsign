<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Schedules</h2>
      <p class="text-muted text-sm">Manage time-based routing schedules and holiday lists.</p>
    </div>
  </div>

  <div class="tabs">
    <button class="tab" :class="{ active: activeTab === 'schedules' }" @click="activeTab = 'schedules'">Schedules</button>
    <button class="tab" :class="{ active: activeTab === 'holidays' }" @click="activeTab = 'holidays'">Holidays</button>
  </div>

  <!-- SCHEDULES TAB -->
  <div class="tab-content" v-if="activeTab === 'schedules'">
    <div class="panel-header">
       <h3>Time Schedules</h3>
       <button class="btn-primary" @click="$router.push('/admin/schedules/new')">+ Add Schedule</button>
    </div>

    <DataTable :columns="scheduleColumns" :data="schedules" actions>
      <template #status="{ value }">
         <span class="tag-active" v-if="value === 'Active'">Active</span>
         <span class="tag-inactive" v-else>Inactive</span>
      </template>
      <template #actions>
        <button class="btn-link" @click="$router.push('/admin/schedules/1')">Edit</button>
        <button class="btn-link text-bad">Delete</button>
      </template>
    </DataTable>
  </div>

  <!-- HOLIDAYS TAB -->
  <div class="tab-content" v-else-if="activeTab === 'holidays'">
     <div v-if="!editingHoliday">
      <div class="panel-header">
        <h3>Holiday Lists</h3>
        <button class="btn-primary" @click="editingHoliday = true">+ New Holiday List</button>
      </div>
      <DataTable :columns="holidayColumns" :data="holidayLists" actions>
        <template #source="{ value }">
           <span class="badge-source">{{ value }}</span>
        </template>
        <template #actions>
          <button class="btn-link" @click="editingHoliday = true">Edit</button>
          <button class="btn-link text-bad">Delete</button>
        </template>
      </DataTable>
    </div>
    <div v-else>
       <div class="builder-header">
         <button class="back-link" @click="editingHoliday = false">← Back to Holiday Lists</button>
      </div>
      <HolidayEditor />
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import DataTable from '../../components/common/DataTable.vue'
import HolidayEditor from '../../components/ivr/HolidayEditor.vue'

const activeTab = ref('schedules')
const editingHoliday = ref(false)

const scheduleColumns = [
  { key: 'name', label: 'Name' },
  { key: 'extension', label: 'Extension' },
  { key: 'status', label: 'Status' },
  { key: 'desc', label: 'Description' }
]

const schedules = ref([
  { name: 'Business Hours', extension: '5001', status: 'Active', desc: 'M-F 9am-5pm' },
  { name: 'Lunch Break', extension: '*901', status: 'Active', desc: 'Daily 12pm-1pm' },
])

const holidayColumns = [
  { key: 'name', label: 'List Name' },
  { key: 'count', label: 'Dates' },
  { key: 'source', label: 'Source' }
]

const holidayLists = ref([
  { name: 'US Federal 2024', count: '11 Dates', source: 'External URL' },
  { name: 'Office Closures', count: '3 Dates', source: 'Manual' },
])
</script>

<style scoped>
.view-header { margin-bottom: 24px; }
.tabs { display: flex; gap: 2px; border-bottom: 1px solid var(--border-color); }
.tab { padding: 8px 16px; background: transparent; border: 1px solid transparent; border-bottom: none; cursor: pointer; font-size: var(--text-sm); font-weight: 500; color: var(--text-muted); border-radius: var(--radius-sm) var(--radius-sm) 0 0; transition: all var(--transition-fast); }
.tab.active { background: white; border-color: var(--border-color); color: var(--primary-color); margin-bottom: -1px; }

.tab-content { background: white; border: 1px solid var(--border-color); border-top: none; padding: var(--spacing-xl); border-radius: 0 0 var(--radius-md) var(--radius-md); box-shadow: var(--shadow-sm); }
.panel-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }

.btn-primary { background: var(--primary-color); color: white; border: none; padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-link { background: none; border: none; color: var(--primary-color); font-size: 12px; font-weight: 600; cursor: pointer; }
.text-bad { color: var(--status-bad); }

.tag-active { background: #dcfce7; color: #166534; padding: 2px 8px; border-radius: 99px; font-size: 11px; font-weight: 600; }
.tag-inactive { background: #f1f5f9; color: #64748b; padding: 2px 8px; border-radius: 99px; font-size: 11px; font-weight: 600; }
.badge-source { font-size: 11px; background: #f3f4f6; color: #4b5563; padding: 2px 6px; border-radius: 4px; border: 1px solid #e5e7eb; }

.builder-header { margin-bottom: 16px; }
.back-link { background: none; border: none; color: var(--text-muted); cursor: pointer; padding: 0; font-size: 12px; }
.back-link:hover { text-decoration: underline; color: var(--primary-color); }
</style>

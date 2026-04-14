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
      <template #actions="{ row }">
        <button class="btn-link" @click="$router.push(`/admin/schedules/${row.id}`)">Edit</button>
        <button class="btn-link text-bad" @click="deleteSchedule(row)">Delete</button>
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
        <template #actions="{ row }">
          <button class="btn-link" @click="editingHoliday = true">Edit</button>
          <button class="btn-link text-bad" @click="deleteHolidayList(row)">Delete</button>
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
import { ref, onMounted, inject } from 'vue'
import DataTable from '../../components/common/DataTable.vue'
import HolidayEditor from '../../components/ivr/HolidayEditor.vue'
import { timeConditionsAPI, holidaysAPI } from '../../services/api'

const toast = inject('toast')
const activeTab = ref('schedules')
const editingHoliday = ref(false)
const isLoading = ref(false)

const scheduleColumns = [
  { key: 'name', label: 'Name' },
  { key: 'extension', label: 'Extension' },
  { key: 'status', label: 'Status' },
  { key: 'desc', label: 'Description' }
]

const schedules = ref([])

const holidayColumns = [
  { key: 'name', label: 'List Name' },
  { key: 'count', label: 'Dates' },
  { key: 'source', label: 'Source' }
]

const holidayLists = ref([])

async function fetchSchedules() {
  isLoading.value = true
  try {
    const response = await timeConditionsAPI.list()
    schedules.value = (response.data || []).map(tc => ({
      id: tc.id,
      name: tc.name,
      extension: tc.extension || '',
      status: tc.enabled !== false ? 'Active' : 'Inactive',
      desc: buildDescription(tc)
    }))
  } catch (error) {
    console.error('Failed to load schedules', error)
    toast?.error(error.message, 'Failed to load schedules')
  } finally {
    isLoading.value = false
  }
}

function buildDescription(tc) {
  if (!tc.start_time || !tc.end_time) return 'No time set'
  const days = (tc.weekdays || []).map(d => ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'][d]).join(', ')
  return days ? `${days} ${tc.start_time}-${tc.end_time}` : 'No schedule set'
}

async function deleteSchedule(schedule) {
  if (!confirm(`Delete "${schedule.name}"?`)) return
  try {
    await timeConditionsAPI.delete(schedule.id)
    schedules.value = schedules.value.filter(s => s.id !== schedule.id)
    toast?.success('Schedule deleted')
  } catch (error) {
    toast?.error(error.message, 'Failed to delete schedule')
  }
}

async function fetchHolidayLists() {
  try {
    const response = await holidaysAPI.list()
    holidayLists.value = (response.data || []).map(list => ({
      id: list.id,
      name: list.name,
      count: list.dates?.length || list.count || 0,
      source: list.external_url ? 'External URL' : 'Manual'
    }))
  } catch (error) {
    console.error('Failed to load holiday lists', error)
    toast?.error(error.message, 'Failed to load holiday lists')
  }
}

async function deleteHolidayList(list) {
  if (!confirm(`Delete "${list.name}"?`)) return
  try {
    await holidaysAPI.delete(list.id)
    holidayLists.value = holidayLists.value.filter(l => l.id !== list.id)
    toast?.success('Holiday list deleted')
  } catch (error) {
    toast?.error(error.message, 'Failed to delete holiday list')
  }
}

onMounted(() => {
  fetchSchedules()
  fetchHolidayLists()
})
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

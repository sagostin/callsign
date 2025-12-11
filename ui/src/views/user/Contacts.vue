<template>
  <div class="view-container">
    <div class="page-header">
      <h2>Contacts</h2>
      <button class="btn-primary">+ Add Contact</button>
    </div>

    <div class="tabs">
      <button class="tab" :class="{ active: activeTab === 'added' }" @click="activeTab = 'added'">My Contacts</button>
      <button class="tab" :class="{ active: activeTab === 'local' }" @click="activeTab = 'local'">Company Directory</button>
      <button class="tab" :class="{ active: activeTab === 'speed-dials' }" @click="activeTab = 'speed-dials'">Speed Dials</button>
      <button class="tab" :class="{ active: activeTab === 'system' }" @click="activeTab = 'system'">System Resources</button>
    </div>

    <div class="contact-list" v-if="activeTab === 'added'">
       <div v-if="contacts.added.length === 0" class="empty-state">No personal contacts added.</div>
       <div class="contact-card" v-for="c in contacts.added" :key="c.id">
        <div class="avatar bg-indigo-100 text-indigo-600">{{ c.initials }}</div>
        <div class="info">
           <div class="name">{{ c.name }}</div>
           <div class="ext">{{ c.number }}</div>
        </div>
        <div class="actions">
           <button class="btn-icon"><Phone class="icon-sm" /></button>
           <button class="btn-icon"><MessageSquare class="icon-sm" /></button>
        </div>
      </div>
    </div>

    <div class="contact-list" v-if="activeTab === 'local'">
      <div class="contact-card" v-for="c in contacts.local" :key="c.id">
        <div class="avatar">{{ c.initials }}</div>
        <div class="info">
           <div class="name">{{ c.name }}</div>
           <div class="ext">Ext. {{ c.ext }}</div>
        </div>
        <div class="actions">
           <button class="btn-icon"><Phone class="icon-sm" /></button>
           <button class="btn-icon"><MessageSquare class="icon-sm" /></button>
        </div>
      </div>
    </div>

    <!-- Speed Dials Tab -->
    <div class="speed-dials-content" v-if="activeTab === 'speed-dials'">
      <div class="speed-dial-group" v-for="group in speedDialGroups" :key="group.id">
        <div class="group-header">
          <h3>{{ group.name }}</h3>
          <span class="prefix-badge">Prefix: {{ group.prefix }}</span>
        </div>
        <div class="contact-list">
          <div class="contact-card speed-dial" v-for="entry in group.entries" :key="entry.code">
            <div class="avatar speed-code">
              <Zap class="icon-sm" />
            </div>
            <div class="info">
               <div class="name">{{ entry.label }}</div>
               <div class="ext">Dial: {{ group.prefix }}{{ entry.code }} → {{ entry.destination }}</div>
            </div>
            <div class="actions">
               <button class="btn-icon" @click="dialSpeedDial(group.prefix, entry.code)"><Phone class="icon-sm" /></button>
            </div>
          </div>
        </div>
      </div>
      <div v-if="speedDialGroups.length === 0" class="empty-state">
        No speed dials configured. Contact your administrator.
      </div>
    </div>

    <div class="contact-list" v-if="activeTab === 'system'">
      <div class="contact-card" v-for="c in contacts.system" :key="c.id">
        <div class="avatar bg-purple-100 text-purple-600"><Bot class="icon-sm" /></div>
        <div class="info">
           <div class="name">{{ c.name }}</div>
           <div class="ext">Ext. {{ c.ext }}</div>
        </div>
        <div class="actions">
           <button class="btn-icon"><Phone class="icon-sm" /></button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Phone, MessageSquare, Bot, Zap } from 'lucide-vue-next'

const activeTab = ref('local')

const contacts = ref({
  added: [],
  local: [
    { id: 1, name: 'Alice Smith', ext: '101', initials: 'AS' },
    { id: 2, name: 'Bob Jones', ext: '102', initials: 'BJ' },
    { id: 3, name: 'Charlie Day', ext: '103', initials: 'CD' },
    { id: 4, name: 'Dave Miller', ext: '104', initials: 'DM' },
    { id: 5, name: 'Eve Polastri', ext: '105', initials: 'EP' },
  ],
  system: [
    { id: 10, name: 'Main IVR', ext: '8000' },
    { id: 11, name: 'Sales Queue', ext: '8001' },
    { id: 12, name: 'Support Queue', ext: '8002' },
  ]
})

const speedDialGroups = ref([
  { 
    id: 1, 
    name: 'Executive Directory', 
    prefix: '*0', 
    entries: [
      { code: '1', label: 'CEO Mobile', destination: '+1 (555) 000-1001' },
      { code: '2', label: 'CTO Mobile', destination: '+1 (555) 000-1002' },
    ]
  },
  { 
    id: 2, 
    name: 'Vendor Support', 
    prefix: '*9', 
    entries: [
      { code: '1', label: 'IT Helpdesk', destination: '1-800-555-1234' },
      { code: '5', label: 'Building Security', destination: '+1 (555) 999-0000' },
    ]
  },
])

const dialSpeedDial = (prefix, code) => {
  alert(`Dialing ${prefix}${code}...`)
}
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h2 { font-size: 20px; font-weight: 700; }
.page-header h2 { font-size: 20px; font-weight: 700; }
.btn-primary { background: var(--primary-color); color: white; border: none; padding: 8px 16px; border-radius: 6px; font-weight: 600; cursor: pointer; }

.tabs { display: flex; gap: 16px; border-bottom: 1px solid var(--border-color); margin-bottom: 24px; }
.tab { background: none; border: none; padding-bottom: 12px; cursor: pointer; color: var(--text-muted); font-weight: 500; font-size: 14px; border-bottom: 2px solid transparent; }
.tab.active { color: var(--primary-color); border-bottom-color: var(--primary-color); }

.contact-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(250px, 1fr)); gap: 16px; }
.empty-state { grid-column: 1 / -1; padding: 40px; text-align: center; color: var(--text-muted); font-style: italic; background: var(--bg-app); border-radius: 8px; }

.contact-card {
  background: white; border: 1px solid var(--border-color); padding: 16px; border-radius: var(--radius-md);
  display: flex; align-items: center; gap: 12px;
}
.avatar { width: 40px; height: 40px; background: #e2e8f0; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-weight: 700; color: #475569; }
.avatar.speed-code { background: #fef3c7; color: #b45309; }
.info { flex: 1; }
.name { font-weight: 600; font-size: 14px; }
.ext { font-size: 12px; color: var(--text-muted); }

.actions { display: flex; gap: 4px; }
.btn-icon { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 6px; border-radius: 4px; }
.btn-icon:hover { background: #f1f5f9; color: var(--primary-color); }
.icon-sm { width: 16px; height: 16px; }

/* Speed Dials Section */
.speed-dials-content { display: flex; flex-direction: column; gap: 24px; }
.group-header { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.group-header h3 { font-size: 14px; font-weight: 700; color: var(--text-primary); margin: 0; }
.prefix-badge { font-size: 11px; font-family: monospace; font-weight: 600; background: #fef3c7; color: #b45309; padding: 3px 8px; border-radius: 4px; }
.contact-card.speed-dial { border-left: 3px solid #f59e0b; }
</style>


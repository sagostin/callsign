<template>
  <div class="view-container">
    <div class="page-header">
      <h2>Contacts</h2>
      <button class="btn-primary" @click="addContact">+ Add Contact</button>
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
      <div v-if="contacts.system.length === 0" class="empty-state">No system resources available.</div>
      <div class="contact-card" v-for="c in contacts.system" :key="c.id">
        <div class="avatar bg-purple-100 text-purple-600"><Bot class="icon-sm" /></div>
        <div class="info">
           <div class="name">{{ c.name }}</div>
           <div class="ext">{{ c.type }} • Ext. {{ c.ext }}</div>
        </div>
        <div class="actions">
           <button class="btn-icon" @click="dialContact(c.ext)"><Phone class="icon-sm" /></button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, inject } from 'vue'
import { useRouter } from 'vue-router'
import { Phone, MessageSquare, Bot, Zap } from 'lucide-vue-next'
import { extensionsAPI, extensionPortalAPI, speedDialsAPI, queuesAPI, ringGroupsAPI } from '../../services/api'

const router = useRouter()
const toast = inject('toast')
const activeTab = ref('local')
const loading = ref(false)

const contacts = ref({
  added: [],
  local: [],
  system: []
})

const speedDialGroups = ref([])

const fetchContacts = async () => {
  loading.value = true
  try {
    // Company directory from extensions
    const extRes = await extensionsAPI.list()
    const exts = extRes.data?.extensions || extRes.data || []
    contacts.value.local = exts.map((e, i) => ({
      id: e.id || i,
      name: e.effective_caller_id_name || e.description || `Ext ${e.extension}`,
      ext: e.extension,
      initials: (e.effective_caller_id_name || e.description || 'X').split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase()
    }))
  } catch (err) {
    console.error('Failed to load directory:', err)
  }

  try {
    // Personal contacts
    const contactRes = await extensionPortalAPI.getContacts()
    const list = contactRes.data?.contacts || contactRes.data || []
    contacts.value.added = list.map(c => ({
      id: c.id,
      name: c.name,
      number: c.phone || c.number,
      initials: (c.name || 'X').split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase()
    }))
  } catch (err) {
    console.error('Failed to load personal contacts:', err)
  }

  try {
    // Speed dials
    const sdRes = await speedDialsAPI.list()
    speedDialGroups.value = (sdRes.data?.groups || sdRes.data || []).map(g => ({
      id: g.id,
      name: g.name,
      prefix: g.prefix || g.dial_prefix || '',
      entries: (g.entries || []).map(e => ({
        code: e.code || e.speed_code,
        label: e.label || e.name,
        destination: e.destination || e.number
      }))
    }))
  } catch (err) {
    console.error('Failed to load speed dials:', err)
  }

  // System resources (queues + ring groups)
  try {
    const systemContacts = []
    try {
      const qRes = await queuesAPI.list()
      const queues = qRes.data || []
      queues.forEach(q => {
        systemContacts.push({
          id: `queue-${q.id}`,
          name: q.name || `Queue ${q.extension}`,
          ext: q.extension || q.dial_code || '',
          type: 'Queue'
        })
      })
    } catch { /* queues optional */ }
    try {
      const rgRes = await ringGroupsAPI.list()
      const rgs = rgRes.data || []
      rgs.forEach(rg => {
        systemContacts.push({
          id: `rg-${rg.id}`,
          name: rg.name || `Ring Group ${rg.extension}`,
          ext: rg.extension || rg.dial_code || '',
          type: 'Ring Group'
        })
      })
    } catch { /* ring groups optional */ }
    contacts.value.system = systemContacts
  } catch {
    contacts.value.system = []
  }

  loading.value = false
}

onMounted(fetchContacts)

const dialSpeedDial = (prefix, code) => {
  router.push({ path: '/dialer', query: { dial: `${prefix}${code}` } })
}

const dialContact = (number) => {
  router.push({ path: '/dialer', query: { dial: number } })
}

const messageContact = (number) => {
  router.push({ path: '/messages', query: { to: number } })
}

const addContact = async () => {
  const name = prompt('Contact name:')
  if (!name) return
  const phone = prompt('Phone number or extension:')
  if (!phone) return
  try {
    await extensionPortalAPI.createContact({ name, phone })
    toast?.success('Contact added')
    await fetchContacts()
  } catch (err) {
    toast?.error(err.message || 'Failed to add contact')
  }
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


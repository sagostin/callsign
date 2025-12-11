<template>
  <div class="view-header">
    <div class="header-left">
      <button class="back-link" @click="$router.push('/queues')">← Back to Queues</button>
      <h2>{{ isNew ? 'Add New Queue' : 'Edit Queue' }}</h2>
    </div>
  </div>

  <div class="form-container">
    <div class="form-section">
      <h3>1. Queue Details</h3>
      <div class="form-group">
        <label>Queue Name</label>
        <input type="text" v-model="form.name" class="input-field" placeholder="Sales Main">
      </div>
      
      <div class="form-group">
        <label>Display Name (Caller ID)</label>
        <input type="text" v-model="form.cidPrefix" class="input-field" placeholder="Sales: ">
      </div>
      
      <div class="form-group">
        <label>Extension Number</label>
        <input type="text" v-model="form.extension" class="input-field" placeholder="8000">
      </div>

      <div class="form-group">
        <label>Strategy</label>
        <select v-model="form.strategy" class="input-field">
          <option value="ring-all">Ring All</option>
          <option value="round-robin">Round Robin</option>
          <option value="longest-idle">Longest Idle Agent</option>
        </select>
      </div>
    </div>

    <div class="form-section">
      <h3>2. Agents</h3>
      <div class="agent-selection">
        <div class="form-group">
          <label>Add Agents</label>
          <div class="input-group">
            <select v-model="selectedAgent" class="input-field">
              <option value="" disabled selected>Select Agent...</option>
              <option value="101">101 - Alice Smith</option>
              <option value="102">102 - Bob Jones</option>
              <option value="105">105 - David Lee</option>
            </select>
            <button class="btn-secondary" @click="addAgent">Add</button>
          </div>
        </div>

        <div class="agent-list" v-if="form.agents.length > 0">
          <div class="list-header">
             <span class="col-agent">Agent</span>
             <span class="col-actions">Order</span>
          </div>
          <div v-for="(agent, index) in form.agents" :key="index" class="agent-row">
            <span class="agent-name">{{ agent }}</span>
            <div class="row-actions">
              <button class="btn-icon small" @click="moveAgent(index, -1)" :disabled="index === 0">↑</button>
              <button class="btn-icon small" @click="moveAgent(index, 1)" :disabled="index === form.agents.length - 1">↓</button>
              <button class="remove-btn" @click="removeAgent(index)">×</button>
            </div>
          </div>
        </div>
        <p v-else class="text-muted text-xs">No agents assigned yet.</p>
      </div>
    </div>

    <div class="form-section">
      <h3>3. Settings</h3>
      <div class="form-grid">
        <div class="form-group">
          <label>Max Wait Time (sec)</label>
          <input type="number" v-model="form.maxWait" class="input-field" placeholder="300">
        </div>
        <div class="form-group">
          <label>Max Queue Size</label>
           <input type="number" v-model="form.maxSize" class="input-field" placeholder="0 (Unlimited)">
        </div>
      </div>
      
      <div class="form-group" style="margin-top: 12px">
        <label>Music on Hold</label>
        <select v-model="form.moh" class="input-field">
          <option value="default">System Default</option>
          <optgroup label="Tenant Playlists">
             <option value="jazz">Lobby Jazz</option>
             <option value="promo">Promotional Mix</option>
          </optgroup>
        </select>
      </div>
    </div>

    <!-- ESCALATION & FAILOVER -->
    <div class="form-section">
      <h3>4. Escalation & Failover</h3>
      
      <div class="escalation-rules">
        <div class="rule-row">
           <span class="rule-label">If caller waits</span>
           <input type="number" class="input-field small-input" v-model="form.escalationTime" placeholder="60">
           <span class="rule-label">seconds, then add</span>
           <select class="input-field auto-width">
              <option>Overflow Agents</option>
              <option>Tier 2 Support</option>
           </select>
        </div>
      </div>

      <div class="divider"></div>

      <div class="form-group">
        <label>Failover Destination</label>
        <span class="help-text">Where to send calls if max wait time reached or no agents available.</span>
        <div class="input-group">
           <select v-model="form.failoverType" class="input-field" style="width: 140px">
              <option value="voicemail">Voicemail</option>
              <option value="extension">Extension</option>
              <option value="queue">Queue</option>
              <option value="ivr">IVR</option>
           </select>
           <select v-model="form.failoverTarget" class="input-field">
              <option value="vm_gen">General Mailbox</option>
              <option value="vm_sales">Sales Mailbox</option>
           </select>
        </div>
      </div>
    </div>

    <div class="form-actions">
      <button class="btn-primary large" :disabled="!isValid" @click="saveQueue">
        Create Queue
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()
const isNew = computed(() => route.params.id === 'new' || !route.params.id)

const form = ref({
  name: '',
  extension: '',
  strategy: 'ring-all',
  agents: [],
  agents: [],
  maxWait: 60,
  maxSize: 0,
  moh: 'default',
  cidPrefix: '',
  escalationTime: '',
  failoverType: 'voicemail',
  failoverTarget: 'vm_sales'
})

const selectedAgent = ref('')

const isValid = computed(() => {
  return form.value.name && form.value.extension
})

const addAgent = () => {
  if (selectedAgent.value && !form.value.agents.includes(selectedAgent.value)) {
    form.value.agents.push(selectedAgent.value)
  }
  selectedAgent.value = ''
}

const moveAgent = (index, direction) => {
  const newIndex = index + direction
  if (newIndex >= 0 && newIndex < form.value.agents.length) {
    const temp = form.value.agents[index]
    form.value.agents[index] = form.value.agents[newIndex]
    form.value.agents[newIndex] = temp
  }
}

const removeAgent = (index) => {
  form.value.agents.splice(index, 1)
}

const saveQueue = () => {
  alert(`Saved Queue "${form.value.name}" with Ext ${form.value.extension}`)
  router.push('/queues')
}

onMounted(() => {
  if (!isNew.value && route.params.id) {
     // Mock hydration
     form.value = {
        name: 'Sales Main',
        extension: '8000',
        strategy: 'ring-all',
        agents: ['101', '102'],
        maxWait: 60,
        maxSize: 0,
        moh: 'jazz',
        cidPrefix: 'Sales: ',
        escalationTime: 60,
        failoverType: 'voicemail',
        failoverTarget: 'vm_gen'
     }
  }
})
</script>

<style scoped>
.header-left {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: var(--spacing-xl);
}

.back-link {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0;
  font-size: var(--text-xs);
  text-align: left;
}
.back-link:hover { text-decoration: underline; color: var(--primary-color); }

.form-container {
  max-width: 600px;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
}

.form-section h3 {
  font-size: var(--text-md);
  color: var(--text-primary);
  font-weight: 600;
  margin-bottom: var(--spacing-md);
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 8px;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-md);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: var(--spacing-md);
}

label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
  letter-spacing: 0.05em;
}

.input-field {
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
  color: var(--text-primary);
  outline: none;
  background: white;
  transition: border-color var(--transition-fast);
}
.input-field:focus { border-color: var(--primary-color); }

.agent-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}

.agent-tag {
  align-items: center;
  gap: 6px;
}

.agent-list {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.list-header {
  display: flex;
  background: var(--bg-app);
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-color);
  font-size: 11px;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
}
.col-agent { flex: 1; }
.col-actions { width: 80px; text-align: right; }

.agent-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-color);
  background: white;
}
.agent-row:last-child { border-bottom: none; }

.row-actions { display: flex; gap: 4px; }
.btn-icon.small { 
  background: var(--bg-app); border: 1px solid var(--border-color); 
  width: 24px; height: 24px; padding: 0; display: flex; align-items: center; justify-content: center;
  border-radius: 4px; cursor: pointer; color: var(--text-main);
}
.btn-icon.small:disabled { opacity: 0.3; cursor: default; }
.btn-icon.small:hover:not(:disabled) { background: white; border-color: var(--primary-color); }

.help-text { font-size: 10px; color: var(--text-muted); }
.small-input { width: 60px; text-align: center; }
.auto-width { width: auto; }
.divider { height: 1px; background: var(--border-color); margin: 16px 0; }

.escalation-rules {
  background: var(--bg-app);
  padding: 12px;
  border-radius: var(--radius-sm);
  border: 1px dashed var(--border-color);
}
.rule-row { display: flex; align-items: center; gap: 8px; font-size: var(--text-sm); }
.rule-label { font-weight: 500; color: var(--text-main); }

.remove-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  font-weight: bold;
}
.remove-btn:hover { color: var(--status-bad); }

.btn-primary {
  background-color: var(--primary-color);
  color: white;
  border: none;
  padding: 12px 24px;
  border-radius: var(--radius-sm);
  font-weight: 600;
  cursor: pointer;
  width: 100%;
}
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
</style>

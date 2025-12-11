<template>
  <div class="messaging-providers-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Messaging Providers</h2>
        <p class="text-muted text-sm">Configure SMS/MMS providers, number matching rules, and transformations.</p>
      </div>
      <div class="header-actions">
        <button class="btn-primary" @click="showAddModal = true">
          <PlusIcon class="btn-icon" /> Add Provider
        </button>
      </div>
    </div>

    <!-- Tabs -->
    <div class="tabs">
      <button class="tab" :class="{ active: activeTab === 'providers' }" @click="activeTab = 'providers'">
        <ServerIcon class="tab-icon" /> Providers
      </button>
      <button class="tab" :class="{ active: activeTab === 'routing' }" @click="activeTab = 'routing'">
        <GitMergeIcon class="tab-icon" /> Number Routing
      </button>
      <button class="tab" :class="{ active: activeTab === 'transformations' }" @click="activeTab = 'transformations'">
        <WrenchIcon class="tab-icon" /> Transformations
      </button>
    </div>

    <!-- PROVIDERS TAB -->
    <div class="tab-content" v-if="activeTab === 'providers'">
      <div class="providers-grid">
        <div class="provider-card" v-for="provider in providers" :key="provider.id" :class="{ default: provider.isDefault, disabled: !provider.enabled }">
          <div class="provider-header">
            <div class="provider-logo" :class="provider.type">
              {{ provider.name.charAt(0) }}
            </div>
            <div class="provider-info">
              <h4>{{ provider.name }}</h4>
              <span class="provider-type">{{ provider.type }}</span>
            </div>
            <div class="provider-status">
              <span class="status-dot" :class="{ online: provider.enabled && provider.status === 'connected', offline: !provider.enabled }"></span>
              <span class="status-text">{{ provider.enabled ? provider.status : 'Disabled' }}</span>
            </div>
          </div>

          <div class="provider-details">
            <div class="detail-row">
              <span class="label">Account SID</span>
              <span class="value mono">{{ provider.accountSid }}</span>
            </div>
            <div class="detail-row">
              <span class="label">Numbers</span>
              <span class="value">{{ provider.numbers.length }} assigned</span>
            </div>
            <div class="detail-row">
              <span class="label">Messages Today</span>
              <span class="value">{{ provider.messagesToday }}</span>
            </div>
          </div>

          <div class="provider-badges">
            <span class="badge default" v-if="provider.isDefault">Default</span>
            <span class="badge failover" v-if="provider.isFailover">Failover</span>
          </div>

          <div class="provider-actions">
            <button class="action-btn" @click="editProvider(provider)" title="Edit">
              <EditIcon class="icon-sm" />
            </button>
            <button class="action-btn" @click="testProvider(provider)" title="Test">
              <SendIcon class="icon-sm" />
            </button>
            <button class="action-btn" @click="toggleProvider(provider)" :title="provider.enabled ? 'Disable' : 'Enable'">
              <PowerIcon class="icon-sm" :class="{ enabled: provider.enabled }" />
            </button>
            <button class="action-btn danger" @click="deleteProvider(provider)" title="Delete" v-if="!provider.isDefault">
              <TrashIcon class="icon-sm" />
            </button>
          </div>
        </div>

        <!-- Add New Card -->
        <div class="provider-card add-new" @click="showAddModal = true">
          <PlusCircleIcon class="add-icon" />
          <span>Add Provider</span>
        </div>
      </div>
    </div>

    <!-- NUMBER ROUTING TAB -->
    <div class="tab-content" v-else-if="activeTab === 'routing'">
      <p class="panel-desc">Route messages to specific providers based on number patterns.</p>

      <div class="routing-rules">
        <div class="rule-item" v-for="rule in routingRules" :key="rule.id">
          <div class="rule-handle">
            <GripVerticalIcon class="grip" />
            <span class="rule-priority">{{ rule.priority }}</span>
          </div>
          <div class="rule-pattern">
            <span class="pattern-label">Pattern:</span>
            <code>{{ rule.pattern }}</code>
          </div>
          <div class="rule-arrow">→</div>
          <div class="rule-provider">
            <span class="provider-badge" :class="rule.providerType">{{ rule.provider }}</span>
          </div>
          <div class="rule-description">{{ rule.description }}</div>
          <div class="rule-actions">
            <button class="action-btn" @click="editRule(rule)"><EditIcon class="icon-sm" /></button>
            <button class="action-btn danger" @click="deleteRule(rule)"><TrashIcon class="icon-sm" /></button>
          </div>
        </div>
      </div>

      <button class="btn-secondary" style="margin-top: 16px;" @click="showRuleModal = true">
        <PlusIcon class="btn-icon" /> Add Routing Rule
      </button>
    </div>

    <!-- TRANSFORMATIONS TAB -->
    <div class="tab-content" v-else-if="activeTab === 'transformations'">
      <p class="panel-desc">Transform numbers before sending to providers (e.g., add country codes, strip prefixes).</p>

      <div class="transform-groups">
        <div class="transform-group">
          <h4>Outbound Transformations</h4>
          <p class="help-text">Applied to destination numbers before sending.</p>
          
          <div class="transform-list">
            <div class="transform-item" v-for="t in outboundTransforms" :key="t.id">
              <div class="transform-pattern">
                <code>{{ t.match }}</code>
              </div>
              <div class="transform-arrow">→</div>
              <div class="transform-result">
                <code>{{ t.replace }}</code>
              </div>
              <div class="transform-desc">{{ t.description }}</div>
              <div class="transform-actions">
                <button class="action-btn" @click="editTransform(t)"><EditIcon class="icon-sm" /></button>
                <button class="action-btn danger" @click="deleteTransform(t)"><TrashIcon class="icon-sm" /></button>
              </div>
            </div>
          </div>
          
          <button class="btn-link add-transform" @click="addTransform('outbound')">
            <PlusIcon class="icon-sm" /> Add Outbound Transform
          </button>
        </div>

        <div class="transform-group">
          <h4>Inbound Transformations</h4>
          <p class="help-text">Applied to incoming sender numbers.</p>
          
          <div class="transform-list">
            <div class="transform-item" v-for="t in inboundTransforms" :key="t.id">
              <div class="transform-pattern">
                <code>{{ t.match }}</code>
              </div>
              <div class="transform-arrow">→</div>
              <div class="transform-result">
                <code>{{ t.replace }}</code>
              </div>
              <div class="transform-desc">{{ t.description }}</div>
              <div class="transform-actions">
                <button class="action-btn" @click="editTransform(t)"><EditIcon class="icon-sm" /></button>
                <button class="action-btn danger" @click="deleteTransform(t)"><TrashIcon class="icon-sm" /></button>
              </div>
            </div>
          </div>
          
          <button class="btn-link add-transform" @click="addTransform('inbound')">
            <PlusIcon class="icon-sm" /> Add Inbound Transform
          </button>
        </div>
      </div>
    </div>

    <!-- ADD/EDIT PROVIDER MODAL -->
    <div v-if="showAddModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal-card wide">
        <div class="modal-header">
          <h3>{{ editingProvider ? 'Edit Provider' : 'Add Messaging Provider' }}</h3>
          <button class="btn-icon" @click="closeModal"><XIcon class="icon-sm" /></button>
        </div>
        
        <div class="modal-body">
          <div class="form-row">
            <div class="form-group">
              <label>Provider Name *</label>
              <input v-model="providerForm.name" class="input-field" placeholder="Primary Twilio">
            </div>
            <div class="form-group">
              <label>Provider Type *</label>
              <select v-model="providerForm.type" class="input-field">
                <option value="twilio">Twilio</option>
                <option value="bandwidth">Bandwidth</option>
                <option value="telnyx">Telnyx</option>
                <option value="plivo">Plivo</option>
                <option value="vonage">Vonage</option>
                <option value="signalwire">SignalWire</option>
              </select>
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>Account SID / API Key *</label>
              <input v-model="providerForm.accountSid" class="input-field" placeholder="ACxxxxxxxx">
            </div>
            <div class="form-group">
              <label>Auth Token / Secret *</label>
              <input v-model="providerForm.authToken" type="password" class="input-field">
            </div>
          </div>

          <div class="form-group">
            <label>Webhook URL</label>
            <div class="readonly-field">
              <code>https://api.callsign.io/webhooks/sms/{{ providerForm.name?.toLowerCase().replace(' ', '_') || 'provider' }}</code>
              <button class="btn-small" @click="copyWebhook">Copy</button>
            </div>
          </div>

          <div class="form-divider"></div>

          <div class="form-row">
            <div class="form-group">
              <label class="checkbox-row">
                <input type="checkbox" v-model="providerForm.isDefault">
                <span>Set as Default Provider</span>
              </label>
            </div>
            <div class="form-group">
              <label class="checkbox-row">
                <input type="checkbox" v-model="providerForm.isFailover">
                <span>Use as Failover</span>
              </label>
            </div>
          </div>

          <div class="form-group">
            <label>Assigned Numbers</label>
            <p class="help-text">Select which numbers use this provider for outbound messaging.</p>
            <div class="number-checkboxes">
              <label class="checkbox-item" v-for="num in availableNumbers" :key="num.id">
                <input type="checkbox" :value="num.id" v-model="providerForm.numbers">
                <span>{{ num.number }} <small>{{ num.label }}</small></span>
              </label>
            </div>
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn-secondary" @click="closeModal">Cancel</button>
          <button class="btn-primary" @click="saveProvider" :disabled="!canSave || saving">
            {{ saving ? 'Saving...' : (editingProvider ? 'Save Changes' : 'Add Provider') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import {
  Plus as PlusIcon, PlusCircle as PlusCircleIcon, Server as ServerIcon,
  GitMerge as GitMergeIcon, Wrench as WrenchIcon, Edit as EditIcon,
  Trash2 as TrashIcon, Send as SendIcon, Power as PowerIcon,
  X as XIcon, GripVertical as GripVerticalIcon
} from 'lucide-vue-next'
import { systemAPI } from '../../services/api'

const activeTab = ref('providers')
const showAddModal = ref(false)
const showRuleModal = ref(false)
const editingProvider = ref(null)
const loading = ref(true)
const saving = ref(false)

const providerForm = ref({
  name: '',
  type: 'twilio',
  accountSid: '',
  authToken: '',
  isDefault: false,
  isFailover: false,
  numbers: []
})

const providers = ref([])
const routingRules = ref([])
const outboundTransforms = ref([])
const inboundTransforms = ref([])
const availableNumbers = ref([])

const canSave = computed(() => providerForm.value.name && providerForm.value.accountSid)

const loadProviders = async () => {
  loading.value = true
  try {
    const response = await systemAPI.listMessagingProviders()
    const data = response.data.data || response.data || []
    providers.value = data.map(p => ({
      id: p.id,
      name: p.name,
      type: p.type,
      accountSid: p.account_sid || '',
      status: p.enabled ? 'connected' : 'disabled',
      enabled: p.enabled,
      isDefault: p.priority === 0,
      isFailover: p.priority > 0,
      numbers: p.phone_numbers || [],
      messagesToday: 0
    }))
  } catch (e) {
    console.error('Failed to load messaging providers:', e)
  } finally {
    loading.value = false
  }
}

onMounted(loadProviders)

const editProvider = (provider) => {
  editingProvider.value = provider
  providerForm.value = { ...provider, authToken: '' }
  showAddModal.value = true
}

const testProvider = (provider) => alert(`Testing connection to ${provider.name}...`)

const toggleProvider = async (provider) => {
  try {
    await systemAPI.updateMessagingProvider(provider.id, { enabled: !provider.enabled })
    await loadProviders()
  } catch (e) {
    alert('Failed to toggle provider: ' + e.message)
  }
}

const deleteProvider = async (provider) => {
  if (!confirm(`Delete provider ${provider.name}?`)) return
  try {
    await systemAPI.deleteMessagingProvider(provider.id)
    await loadProviders()
  } catch (e) {
    alert('Failed to delete provider: ' + e.message)
  }
}

const saveProvider = async () => {
  saving.value = true
  try {
    const data = {
      name: providerForm.value.name,
      type: providerForm.value.type,
      account_sid: providerForm.value.accountSid,
      priority: providerForm.value.isDefault ? 0 : 1,
      enabled: true,
      phone_numbers: providerForm.value.numbers
    }
    if (editingProvider.value) {
      await systemAPI.updateMessagingProvider(editingProvider.value.id, data)
    } else {
      await systemAPI.createMessagingProvider(data)
    }
    await loadProviders()
    closeModal()
  } catch (e) {
    alert('Failed to save provider: ' + e.message)
  } finally {
    saving.value = false
  }
}

const closeModal = () => {
  showAddModal.value = false
  editingProvider.value = null
  providerForm.value = { name: '', type: 'twilio', accountSid: '', authToken: '', isDefault: false, isFailover: false, numbers: [] }
}

const copyWebhook = () => {
  const url = `${window.location.origin}/api/webhooks/sms/${providerForm.value.name?.toLowerCase().replace(/\s+/g, '_') || 'provider'}`
  navigator.clipboard?.writeText(url) || alert(`Webhook URL: ${url}`)
}

const editRule = (rule) => alert(`Edit rule: ${rule.pattern}`)
const deleteRule = (rule) => { routingRules.value = routingRules.value.filter(r => r.id !== rule.id) }
const editTransform = (t) => alert(`Edit transform: ${t.match}`)
const deleteTransform = (t) => alert('Transform deleted')
const addTransform = (type) => alert(`Add ${type} transform`)
</script>

<style scoped>
.messaging-providers-page { padding: 0; }

.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-content h2 { margin: 0 0 4px; }
.header-actions { display: flex; gap: 12px; }

/* Tabs */
.tabs { display: flex; gap: 2px; border-bottom: 1px solid var(--border-color); }
.tab { display: flex; align-items: center; gap: 6px; padding: 10px 16px; background: transparent; border: 1px solid transparent; border-bottom: none; cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-muted); border-radius: 4px 4px 0 0; }
.tab.active { background: white; border-color: var(--border-color); color: var(--primary-color); margin-bottom: -1px; }
.tab-icon { width: 16px; height: 16px; }
.tab-content { background: white; border: 1px solid var(--border-color); border-top: none; padding: 24px; border-radius: 0 0 var(--radius-md) var(--radius-md); min-height: 300px; }

.panel-desc { color: var(--text-muted); font-size: 13px; margin-bottom: 20px; }

/* Providers Grid */
.providers-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 16px; }

.provider-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; transition: all 0.15s; }
.provider-card:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.08); }
.provider-card.default { border-color: var(--primary-color); }
.provider-card.disabled { opacity: 0.6; }

.provider-header { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }
.provider-logo { width: 40px; height: 40px; border-radius: 8px; display: flex; align-items: center; justify-content: center; font-size: 18px; font-weight: 700; color: white; }
.provider-logo.twilio { background: #f22f46; }
.provider-logo.bandwidth { background: #0066cc; }
.provider-logo.telnyx { background: #00b386; }
.provider-logo.plivo { background: #73cd1f; }
.provider-logo.vonage { background: #ffffff; color: #000; border: 1px solid #ddd; }
.provider-logo.signalwire { background: #044cf6; }

.provider-info { flex: 1; }
.provider-info h4 { margin: 0 0 2px; font-size: 14px; }
.provider-type { font-size: 11px; color: var(--text-muted); text-transform: capitalize; }

.provider-status { display: flex; align-items: center; gap: 6px; font-size: 11px; }
.status-dot { width: 8px; height: 8px; border-radius: 50%; }
.status-dot.online { background: #22c55e; }
.status-dot.offline { background: #94a3b8; }
.status-text { color: var(--text-muted); text-transform: capitalize; }

.provider-details { margin-bottom: 12px; }
.detail-row { display: flex; justify-content: space-between; padding: 6px 0; font-size: 12px; border-bottom: 1px solid var(--border-color); }
.detail-row:last-child { border-bottom: none; }
.detail-row .label { color: var(--text-muted); }
.detail-row .value { font-weight: 600; }
.mono { font-family: monospace; }

.provider-badges { display: flex; gap: 6px; margin-bottom: 12px; }
.badge { font-size: 10px; font-weight: 700; padding: 3px 8px; border-radius: 4px; text-transform: uppercase; }
.badge.default { background: var(--primary-light); color: var(--primary-color); }
.badge.failover { background: #fef3c7; color: #b45309; }

.provider-actions { display: flex; gap: 6px; }
.action-btn { width: 32px; height: 32px; border-radius: 6px; border: 1px solid var(--border-color); background: white; cursor: pointer; display: flex; align-items: center; justify-content: center; color: var(--text-muted); transition: all 0.15s; }
.action-btn:hover { border-color: var(--primary-color); color: var(--primary-color); }
.action-btn.danger:hover { border-color: #ef4444; color: #ef4444; }
.action-btn .enabled { color: #22c55e; }

.provider-card.add-new { display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: 200px; border: 2px dashed var(--border-color); cursor: pointer; color: var(--text-muted); gap: 12px; }
.provider-card.add-new:hover { border-color: var(--primary-color); color: var(--primary-color); }
.add-icon { width: 40px; height: 40px; }

/* Routing Rules */
.routing-rules { display: flex; flex-direction: column; gap: 8px; }
.rule-item { display: flex; align-items: center; gap: 12px; padding: 12px 16px; background: var(--bg-app); border-radius: var(--radius-sm); border: 1px solid var(--border-color); }
.rule-handle { display: flex; align-items: center; gap: 8px; }
.grip { width: 16px; height: 16px; color: var(--text-muted); cursor: grab; }
.rule-priority { width: 24px; height: 24px; background: var(--primary-color); color: white; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 11px; font-weight: 700; }
.rule-pattern { display: flex; align-items: center; gap: 6px; }
.pattern-label { font-size: 11px; color: var(--text-muted); }
.rule-pattern code { background: white; padding: 4px 8px; border-radius: 4px; font-size: 12px; }
.rule-arrow { color: var(--text-muted); font-size: 18px; }
.rule-provider { flex-shrink: 0; }
.provider-badge { font-size: 11px; font-weight: 600; padding: 4px 10px; border-radius: 4px; }
.provider-badge.twilio { background: #fee2e2; color: #dc2626; }
.provider-badge.bandwidth { background: #dbeafe; color: #2563eb; }
.provider-badge.telnyx { background: #dcfce7; color: #16a34a; }
.rule-description { flex: 1; font-size: 12px; color: var(--text-muted); }
.rule-actions { display: flex; gap: 4px; }

/* Transformations */
.transform-groups { display: flex; flex-direction: column; gap: 24px; }
.transform-group h4 { margin: 0 0 4px; font-size: 14px; }
.help-text { font-size: 12px; color: var(--text-muted); margin-bottom: 12px; }
.transform-list { display: flex; flex-direction: column; gap: 8px; }
.transform-item { display: flex; align-items: center; gap: 12px; padding: 10px 14px; background: var(--bg-app); border-radius: var(--radius-sm); }
.transform-pattern code, .transform-result code { background: white; padding: 4px 8px; border-radius: 4px; font-size: 12px; }
.transform-arrow { color: var(--text-muted); }
.transform-desc { flex: 1; font-size: 12px; color: var(--text-muted); }
.transform-actions { display: flex; gap: 4px; }
.add-transform { display: flex; align-items: center; gap: 6px; margin-top: 12px; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); }
.modal-card { background: white; border-radius: var(--radius-md); width: 100%; max-width: 480px; max-height: 90vh; display: flex; flex-direction: column; }
.modal-card.wide { max-width: 600px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { font-size: 16px; font-weight: 700; margin: 0; }
.modal-body { padding: 20px; overflow-y: auto; }
.modal-actions { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

.form-row { display: flex; gap: 16px; }
.form-group { display: flex; flex-direction: column; gap: 6px; margin-bottom: 16px; flex: 1; }
.form-group label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }
.form-divider { height: 1px; background: var(--border-color); margin: 8px 0 16px; }
.checkbox-row { display: flex; align-items: center; gap: 8px; font-size: 13px; cursor: pointer; }

.readonly-field { display: flex; align-items: center; gap: 8px; padding: 10px 12px; background: var(--bg-app); border-radius: var(--radius-sm); }
.readonly-field code { flex: 1; font-size: 12px; overflow: hidden; text-overflow: ellipsis; }
.btn-small { padding: 4px 10px; font-size: 11px; background: white; border: 1px solid var(--border-color); border-radius: 4px; cursor: pointer; }

.number-checkboxes { display: grid; grid-template-columns: repeat(2, 1fr); gap: 8px; margin-top: 8px; }
.checkbox-item { display: flex; align-items: center; gap: 8px; padding: 8px 12px; background: var(--bg-app); border-radius: 6px; font-size: 13px; cursor: pointer; }
.checkbox-item small { color: var(--text-muted); margin-left: 4px; }

/* Buttons */
.btn-primary { display: flex; align-items: center; gap: 6px; background-color: var(--primary-color); color: white; border: none; padding: 10px 20px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { display: flex; align-items: center; gap: 6px; background: white; border: 1px solid var(--border-color); padding: 10px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.btn-link { background: none; border: none; color: var(--primary-color); font-weight: 600; cursor: pointer; font-size: 12px; }
.btn-icon { width: 16px; height: 16px; }

.icon-sm { width: 16px; height: 16px; }

@media (max-width: 768px) {
  .providers-grid { grid-template-columns: 1fr; }
  .form-row { flex-direction: column; gap: 0; }
  .number-checkboxes { grid-template-columns: 1fr; }
}
</style>

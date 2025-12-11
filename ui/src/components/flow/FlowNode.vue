<template>
  <div 
    class="flow-node"
    :class="[nodeColorClass, { 'selected': isSelected }]"
    :style="{ transform: `translate(${x}px, ${y}px)` }"
    @mousedown.stop="$emit('select')"
    @dblclick.stop="openEditor"
  >
    <!-- Header / Drag Handle -->
    <div class="node-header" :class="nodeTypeClass" @mousedown.stop="$emit('drag-start', $event)">
       <component :is="iconComponent" class="node-icon" />
       <span class="node-title">{{ data.label }}</span>
       <button class="edit-btn" @click.stop="openEditor" title="Edit">⚙</button>
       <button class="delete-btn" @click.stop="$emit('delete')" title="Delete">&times;</button>
    </div>

    <!-- Mini Preview -->
    <div class="node-preview">
      <span class="preview-text">{{ previewText }}</span>
    </div>

    <!-- Connectors (Outputs) -->
    <div class="connectors" v-if="outputs.length">
       <div 
         v-for="(output, index) in outputs" 
         :key="index"
         class="output-point"
         :class="output.color"
         @mousedown.stop="$emit('connector-drag', { id: data.id, output: output.id }, $event)"
         title="Drag to connect"
       >
         {{ output.label }}
       </div>
    </div>
  </div>

  <!-- Edit Modal -->
  <Teleport to="body">
    <div class="node-editor-overlay" v-if="showEditor" @click.self="closeEditor">
      <div class="node-editor-modal">
        <div class="editor-header">
          <component :is="iconComponent" class="editor-icon" />
          <h3>{{ data.label }}</h3>
          <button class="close-btn" @click="closeEditor">×</button>
        </div>

        <div class="editor-body">
          <!-- Gather Digits -->
          <template v-if="data.type === 'gather'">
            <div class="field-group">
              <label>Prompt Type</label>
              <select v-model="data.config.promptType">
                <option value="audio">Audio File</option>
                <option value="tts">Text-to-Speech</option>
              </select>
            </div>
            <div class="field-group" v-if="data.config.promptType === 'audio'">
              <label>Audio File</label>
              <select v-model="data.config.audioFile">
                <option value="">Select audio...</option>
                <option value="greeting.wav">greeting.wav</option>
                <option value="menu_prompt.wav">menu_prompt.wav</option>
                <option value="enter_digits.wav">enter_digits.wav</option>
              </select>
            </div>
            <div class="field-group" v-else>
              <label>TTS Text</label>
              <textarea v-model="data.config.ttsText" rows="2" placeholder="Please enter your selection..."></textarea>
            </div>
            
            <div class="field-divider">Input Settings</div>
            
            <div class="field-row">
              <div class="field-group">
                <label>Min Digits</label>
                <input type="number" v-model.number="data.config.minDigits" min="1" max="20">
              </div>
              <div class="field-group">
                <label>Max Digits</label>
                <input type="number" v-model.number="data.config.maxDigits" min="1" max="20">
              </div>
            </div>
            <div class="field-row">
              <div class="field-group">
                <label>Timeout (sec)</label>
                <input type="number" v-model.number="data.config.timeout" min="1" max="60">
              </div>
              <div class="field-group">
                <label>Inter-digit (ms)</label>
                <input type="number" v-model.number="data.config.interDigitTimeout" min="500" max="10000" step="500">
              </div>
            </div>
            <div class="field-row">
              <div class="field-group">
                <label>Terminator</label>
                <select v-model="data.config.terminator">
                  <option value="#"># key</option>
                  <option value="*">* key</option>
                  <option value="">None (timeout only)</option>
                </select>
              </div>
              <div class="field-group">
                <label>Valid Pattern</label>
                <input type="text" v-model="data.config.validPattern" placeholder="^[1-5]$">
              </div>
            </div>
            
            <div class="field-divider">Retry Settings</div>
            
            <div class="field-row">
              <div class="field-group">
                <label>Max Retries</label>
                <input type="number" v-model.number="data.config.maxRetries" min="1" max="10">
              </div>
              <div class="field-group">
                <label>Max Timeouts</label>
                <input type="number" v-model.number="data.config.maxTimeouts" min="1" max="10">
              </div>
            </div>
            <div class="field-group">
              <label>Invalid Sound</label>
              <select v-model="data.config.invalidSound">
                <option value="">Default</option>
                <option value="invalid_option.wav">invalid_option.wav</option>
                <option value="try_again.wav">try_again.wav</option>
              </select>
            </div>
            <div class="field-group">
              <label>Timeout Sound</label>
              <select v-model="data.config.timeoutSound">
                <option value="">Default</option>
                <option value="no_input.wav">no_input.wav</option>
                <option value="please_try.wav">please_try.wav</option>
              </select>
            </div>
            
            <div class="field-group">
              <label class="checkbox-label">
                <input type="checkbox" v-model="data.config.loopPrompt">
                Repeat prompt after invalid/timeout
              </label>
            </div>
          </template>

          <!-- Speech Input -->
          <template v-else-if="data.type === 'speech'">
            <div class="field-group">
              <label>Speech Provider</label>
              <select v-model="data.config.provider">
                <option value="google">Google Cloud Speech</option>
                <option value="aws">AWS Transcribe</option>
                <option value="azure">Azure Speech</option>
              </select>
            </div>
            <div class="field-group">
              <label>Grammar / Expected Phrases</label>
              <textarea v-model="data.config.grammar" rows="3" placeholder="yes, no, sales, support"></textarea>
            </div>
            <div class="field-row">
              <div class="field-group">
                <label>Timeout (sec)</label>
                <input type="number" v-model.number="data.config.timeout" min="1" max="30">
              </div>
              <div class="field-group">
                <label>Confidence Threshold</label>
                <input type="number" v-model.number="data.config.confidence" min="0" max="1" step="0.1">
              </div>
            </div>
          </template>

          <!-- Play Audio -->
          <template v-else-if="data.type === 'play_audio'">
            <div class="field-group">
              <label>Audio File</label>
              <select v-model="data.config.audioFile">
                <option value="">Select audio...</option>
                <option value="greeting.wav">greeting.wav</option>
                <option value="hold_music.wav">hold_music.wav</option>
                <option value="thank_you.wav">thank_you.wav</option>
              </select>
            </div>
            <div class="field-group">
              <label class="checkbox-label">
                <input type="checkbox" v-model="data.config.loop">
                Loop audio
              </label>
            </div>
          </template>

          <!-- Text-to-Speech -->
          <template v-else-if="data.type === 'play_tts'">
            <div class="field-group">
              <label>Text to Speak</label>
              <textarea v-model="data.config.text" rows="3" placeholder="Welcome to our company..."></textarea>
            </div>
            <div class="field-row">
              <div class="field-group">
                <label>Engine</label>
                <select v-model="data.config.engine">
                  <option value="flite">Flite (Built-in)</option>
                  <option value="google">Google Cloud TTS</option>
                  <option value="aws">Amazon Polly</option>
                  <option value="azure">Azure TTS</option>
                </select>
              </div>
              <div class="field-group">
                <label>Voice</label>
                <select v-model="data.config.voice">
                  <option value="default">Default</option>
                  <option value="male">Male</option>
                  <option value="female">Female</option>
                </select>
              </div>
            </div>
          </template>

          <!-- Say Digits -->
          <template v-else-if="data.type === 'say_digits'">
            <div class="field-group">
              <label>Value (variable or literal)</label>
              <input type="text" v-model="data.config.value" placeholder="${account_number} or 12345">
            </div>
            <div class="field-group">
              <label>Format</label>
              <select v-model="data.config.format">
                <option value="digits">Individual Digits (1-2-3-4)</option>
                <option value="number">As Number (twelve thirty-four)</option>
                <option value="currency">As Currency ($12.34)</option>
              </select>
            </div>
          </template>

          <!-- Web Request -->
          <template v-else-if="data.type === 'web_request'">
            <div class="field-row">
              <div class="field-group" style="width: 100px;">
                <label>Method</label>
                <select v-model="data.config.method">
                  <option value="GET">GET</option>
                  <option value="POST">POST</option>
                  <option value="PUT">PUT</option>
                  <option value="DELETE">DELETE</option>
                </select>
              </div>
              <div class="field-group" style="flex: 1;">
                <label>URL</label>
                <input type="text" v-model="data.config.url" placeholder="https://api.example.com/lookup">
              </div>
            </div>
            <div class="field-group">
              <label>Headers (JSON)</label>
              <textarea v-model="data.config.headers" rows="2" placeholder='{"Authorization": "Bearer xxx"}'></textarea>
            </div>
            <div class="field-group" v-if="data.config.method !== 'GET'">
              <label>Body (JSON)</label>
              <textarea v-model="data.config.body" rows="3" placeholder='{"phone": "${caller_id}"}'></textarea>
            </div>
            <div class="field-row">
              <div class="field-group">
                <label>Timeout (sec)</label>
                <input type="number" v-model.number="data.config.timeout" min="1" max="30">
              </div>
              <div class="field-group">
                <label>Store Response In</label>
                <input type="text" v-model="data.config.responseVar" placeholder="api_response">
              </div>
            </div>
            <div class="help-note">Response will be available as ${api_response} in subsequent nodes</div>
          </template>

          <!-- Send SMS -->
          <template v-else-if="data.type === 'send_sms'">
            <div class="field-group">
              <label>Provider</label>
              <select v-model="data.config.provider">
                <option value="signalwire">SignalWire</option>
                <option value="twilio">Twilio</option>
                <option value="nexmo">Nexmo/Vonage</option>
              </select>
            </div>
            <div class="field-row">
              <div class="field-group">
                <label>From Number</label>
                <input type="text" v-model="data.config.from" placeholder="${sms_number}">
              </div>
              <div class="field-group">
                <label>To Number</label>
                <input type="text" v-model="data.config.to" placeholder="${caller_id}">
              </div>
            </div>
            <div class="field-group">
              <label>Message Body</label>
              <textarea v-model="data.config.body" rows="3" placeholder="Thank you for calling. Your reference: ${ticket_id}"></textarea>
            </div>
          </template>

          <!-- Database -->
          <template v-else-if="data.type === 'database'">
            <div class="field-group">
              <label>Connection</label>
              <select v-model="data.config.connection">
                <option value="default">Default PostgreSQL</option>
                <option value="mysql">MySQL CRM</option>
                <option value="rest">REST API</option>
              </select>
            </div>
            <div class="field-group">
              <label>Query / Endpoint</label>
              <textarea v-model="data.config.query" rows="3" placeholder="SELECT name FROM customers WHERE phone = '${caller_id}'"></textarea>
            </div>
            <div class="field-row">
              <div class="field-group">
                <label>Timeout (sec)</label>
                <input type="number" v-model.number="data.config.timeout" min="1" max="30">
              </div>
              <div class="field-group">
                <label>Store Result In</label>
                <input type="text" v-model="data.config.resultVar" placeholder="db_result">
              </div>
            </div>
          </template>

          <!-- Condition -->
          <template v-else-if="data.type === 'condition'">
            <div class="field-group">
              <label>Variable</label>
              <input type="text" v-model="data.config.variable" placeholder="${caller_input}">
            </div>
            <div class="field-row">
              <div class="field-group">
                <label>Operator</label>
                <select v-model="data.config.operator">
                  <option value="==">Equals (==)</option>
                  <option value="!=">Not Equals (!=)</option>
                  <option value=">">Greater Than (>)</option>
                  <option value="<">Less Than (<)</option>
                  <option value="contains">Contains</option>
                  <option value="regex">Matches Regex</option>
                </select>
              </div>
              <div class="field-group">
                <label>Value</label>
                <input type="text" v-model="data.config.value" placeholder="1">
              </div>
            </div>
          </template>

          <!-- Set Variable -->
          <template v-else-if="data.type === 'set_variable'">
            <div class="field-row">
              <div class="field-group">
                <label>Variable Name</label>
                <input type="text" v-model="data.config.name" placeholder="customer_name">
              </div>
              <div class="field-group">
                <label>Value</label>
                <input type="text" v-model="data.config.value" placeholder="${api_response.name}">
              </div>
            </div>
          </template>

          <!-- IVR Menu (destination) -->
          <template v-else-if="data.type === 'ivr_menu'">
            <div class="field-group">
              <label>IVR Menu</label>
              <select v-model="data.config.menuId">
                <option value="">Select menu...</option>
                <option value="main">Main Menu</option>
                <option value="support">Support Menu</option>
                <option value="sales">Sales Menu</option>
              </select>
            </div>
            <div class="help-note">Call will be transferred to this IVR menu.</div>
          </template>

          <!-- Extension -->
          <template v-else-if="data.type === 'extension'">
            <div class="field-group">
              <label>Extension Number</label>
              <input type="text" v-model="data.config.extension" placeholder="101">
            </div>
          </template>

          <!-- Queue -->
          <template v-else-if="data.type === 'queue'">
            <div class="field-group">
              <label>Queue</label>
              <select v-model="data.config.queueId">
                <option value="">Select queue...</option>
                <option value="sales">Sales Queue</option>
                <option value="support">Support Queue</option>
              </select>
            </div>
          </template>

          <!-- Ring Group -->
          <template v-else-if="data.type === 'ring_group'">
            <div class="field-group">
              <label>Ring Group</label>
              <select v-model="data.config.groupId">
                <option value="">Select group...</option>
                <option value="sales_team">Sales Team</option>
                <option value="on_call">On-Call</option>
              </select>
            </div>
          </template>

          <!-- External -->
          <template v-else-if="data.type === 'external'">
            <div class="field-group">
              <label>Phone Number</label>
              <input type="tel" v-model="data.config.number" placeholder="+1 555-123-4567">
            </div>
          </template>

          <!-- Voicemail -->
          <template v-else-if="data.type === 'voicemail'">
            <div class="field-group">
              <label>Mailbox</label>
              <select v-model="data.config.mailboxId">
                <option value="">Select mailbox...</option>
                <option value="general">General Mailbox</option>
                <option value="sales">Sales Mailbox</option>
              </select>
            </div>
          </template>

          <!-- Hangup -->
          <template v-else-if="data.type === 'hangup'">
            <div class="help-note">This node ends the call. No configuration needed.</div>
          </template>
        </div>

        <div class="editor-footer">
          <button class="btn-secondary" @click="closeEditor">Cancel</button>
          <button class="btn-primary" @click="closeEditor">Apply</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { 
  Keyboard, Mic, Volume2, MessageSquare, Globe, Database,
  Phone, Users, PhoneCall, PhoneForwarded, Voicemail, PhoneOff, User,
  GitBranch, Variable, Layers, Hash, AlertCircle
} from 'lucide-vue-next'

const props = defineProps({
  data: Object,
  x: Number,
  y: Number,
  isSelected: Boolean
})

defineEmits(['drag-start', 'select', 'delete', 'connector-drag'])

const showEditor = ref(false)

const openEditor = () => { showEditor.value = true }
const closeEditor = () => { showEditor.value = false }

// Initialize default config based on node type
onMounted(() => {
  if (!props.data.config) props.data.config = {}
  
  const defaults = {
    gather: { 
      minDigits: 1, 
      maxDigits: 1, 
      timeout: 10, 
      interDigitTimeout: 2000,
      promptType: 'tts', 
      terminator: '#',
      validPattern: '',
      maxRetries: 3,
      maxTimeouts: 3,
      invalidSound: '',
      timeoutSound: '',
      loopPrompt: true
    },
    speech: { provider: 'google', timeout: 5, confidence: 0.7 },
    play_audio: { audioFile: '', loop: false },
    play_tts: { text: '', engine: 'flite', voice: 'default' },
    say_digits: { value: '', format: 'digits' },
    web_request: { method: 'GET', url: '', headers: '', body: '', timeout: 5, responseVar: 'api_response' },
    send_sms: { provider: 'signalwire', from: '', to: '${caller_id}', body: '' },
    database: { connection: 'default', query: '', timeout: 5, resultVar: 'db_result' },
    condition: { variable: '', operator: '==', value: '' },
    set_variable: { name: '', value: '' },
    ivr_menu: { menuId: '' },
    extension: { extension: '' },
    queue: { queueId: '' },
    ring_group: { groupId: '' },
    external: { number: '' },
    voicemail: { mailboxId: '' },
    hangup: {}
  }
  
  if (defaults[props.data.type]) {
    props.data.config = { ...defaults[props.data.type], ...props.data.config }
  }
})

// Compute color based on node type
const nodeTypeClass = computed(() => props.data.type)

const nodeColorClass = computed(() => {
  const colors = {
    gather: 'node-purple',
    speech: 'node-purple',
    play_audio: 'node-pink',
    play_tts: 'node-pink',
    say_digits: 'node-pink',
    web_request: 'node-blue',
    send_sms: 'node-teal',
    database: 'node-indigo',
    condition: 'node-amber',
    set_variable: 'node-slate',
    ivr_menu: 'node-blue',
    extension: 'node-green',
    queue: 'node-cyan',
    ring_group: 'node-cyan',
    external: 'node-orange',
    voicemail: 'node-slate',
    hangup: 'node-red'
  }
  return colors[props.data.type] || 'node-default'
})

// Preview text for node body
const previewText = computed(() => {
  const c = props.data.config || {}
  switch (props.data.type) {
    case 'gather': return c.maxDigits ? `Max ${c.maxDigits} digits` : 'Click to configure'
    case 'speech': return c.provider || 'Click to configure'
    case 'play_audio': return c.audioFile || 'Select audio file'
    case 'play_tts': return c.text ? c.text.slice(0, 30) + '...' : 'Enter text'
    case 'say_digits': return c.value || 'Enter value'
    case 'web_request': return c.url ? `${c.method} ${c.url.slice(0, 20)}...` : 'Configure request'
    case 'send_sms': return c.to || 'Configure SMS'
    case 'database': return c.connection || 'Select connection'
    case 'condition': return c.variable ? `${c.variable} ${c.operator} ${c.value}` : 'Add condition'
    case 'set_variable': return c.name || 'Set variable'
    case 'ivr_menu': return c.menuId || 'Select IVR'
    case 'extension': return c.extension || 'Enter ext'
    case 'queue': return c.queueId || 'Select queue'
    case 'ring_group': return c.groupId || 'Select group'
    case 'external': return c.number || 'Enter number'
    case 'voicemail': return c.mailboxId || 'Select mailbox'
    case 'hangup': return 'End Call'
    default: return 'Configure...'
  }
})

// Compute Icon
const iconComponent = computed(() => {
  const icons = {
    gather: Keyboard,
    speech: Mic,
    play_audio: Volume2,
    play_tts: MessageSquare,
    say_digits: Hash,
    web_request: Globe,
    send_sms: MessageSquare,
    database: Database,
    condition: GitBranch,
    set_variable: Variable,
    ivr_menu: Layers,
    extension: User,
    queue: Users,
    ring_group: PhoneCall,
    external: PhoneForwarded,
    voicemail: Voicemail,
    hangup: PhoneOff
  }
  return icons[props.data.type] || AlertCircle
})

// Compute Outputs based on type
const outputs = computed(() => {
  switch (props.data.type) {
    case 'gather':
      return [
        { id: 'match', label: 'Match', color: 'out-green' },
        { id: 'timeout', label: 'Timeout', color: 'out-amber' },
        { id: 'invalid', label: 'Invalid', color: 'out-red' }
      ]
    case 'speech':
      return [
        { id: 'match', label: 'Match', color: 'out-green' },
        { id: 'nomatch', label: 'No Match', color: 'out-red' }
      ]
    case 'web_request':
      return [
        { id: 'success', label: 'Success', color: 'out-green' },
        { id: 'error', label: 'Error', color: 'out-red' },
        { id: 'timeout', label: 'Timeout', color: 'out-amber' }
      ]
    case 'send_sms':
      return [
        { id: 'sent', label: 'Sent', color: 'out-green' },
        { id: 'failed', label: 'Failed', color: 'out-red' }
      ]
    case 'database':
      return [
        { id: 'success', label: 'Success', color: 'out-green' },
        { id: 'noresults', label: 'No Results', color: 'out-amber' },
        { id: 'error', label: 'Error', color: 'out-red' }
      ]
    case 'condition':
      return [
        { id: 'true', label: 'True', color: 'out-green' },
        { id: 'false', label: 'False', color: 'out-red' }
      ]
    case 'ivr_menu':
    case 'hangup':
    case 'extension':
    case 'queue':
    case 'ring_group':
    case 'external':
    case 'voicemail':
      return []
    default:
      return [{ id: 'next', label: 'Next' }]
  }
})
</script>

<style scoped>
.flow-node {
  position: absolute;
  width: 180px;
  background: white;
  border: 2px solid #cbd5e1;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  z-index: 10;
  user-select: none;
}

.flow-node.selected {
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.2);
}

/* Node Colors */
.node-amber { border-color: #f59e0b; }
.node-blue { border-color: #3b82f6; }
.node-indigo { border-color: #6366f1; }
.node-purple { border-color: #8b5cf6; }
.node-pink { border-color: #ec4899; }
.node-green { border-color: #22c55e; }
.node-cyan { border-color: #06b6d4; }
.node-teal { border-color: #14b8a6; }
.node-orange { border-color: #f97316; }
.node-slate { border-color: #64748b; }
.node-red { border-color: #ef4444; }

.node-header {
  padding: 8px 10px;
  border-bottom: 1px solid rgba(0,0,0,0.05);
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: grab;
  border-radius: 6px 6px 0 0;
}
.node-header:active { cursor: grabbing; }

/* Header colors by type */
.node-header.gather, .node-header.speech { background: #f5f3ff; }
.node-header.play_audio, .node-header.play_tts, .node-header.say_digits { background: #fdf2f8; }
.node-header.web_request, .node-header.ivr_menu { background: #eff6ff; }
.node-header.send_sms { background: #f0fdfa; }
.node-header.database { background: #eef2ff; }
.node-header.condition { background: #fffbeb; }
.node-header.set_variable { background: #f8fafc; }
.node-header.extension { background: #f0fdf4; }
.node-header.queue, .node-header.ring_group { background: #ecfeff; }
.node-header.external { background: #fff7ed; }
.node-header.voicemail { background: #f8fafc; }
.node-header.hangup { background: #fef2f2; }

.node-icon { width: 14px; height: 14px; opacity: 0.8; }
.node-title { font-size: 11px; font-weight: 600; flex: 1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.edit-btn, .delete-btn {
  background: none; border: none; font-size: 12px; line-height: 1; color: #94a3b8; cursor: pointer; padding: 2px;
}
.edit-btn:hover { color: #3b82f6; }
.delete-btn:hover { color: #ef4444; }

.node-preview { padding: 8px 10px; min-height: 24px; }
.preview-text { font-size: 10px; color: #64748b; display: block; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

/* Connectors */
.connectors {
  display: flex;
  justify-content: center;
  padding: 6px;
  gap: 4px;
  flex-wrap: wrap;
  border-top: 1px solid rgba(0,0,0,0.05);
}

.output-point {
  font-size: 8px;
  background: #e2e8f0;
  padding: 2px 6px;
  border-radius: 3px;
  cursor: crosshair;
  border: 1px solid #cbd5e1;
  transition: all 0.15s;
}
.output-point:hover { background: #6366f1; color: white; border-color: #6366f1; }

.output-point.out-green { background: #dcfce7; border-color: #22c55e; color: #16a34a; }
.output-point.out-green:hover { background: #22c55e; color: white; }
.output-point.out-amber { background: #fef3c7; border-color: #f59e0b; color: #b45309; }
.output-point.out-amber:hover { background: #f59e0b; color: white; }
.output-point.out-red { background: #fee2e2; border-color: #ef4444; color: #dc2626; }
.output-point.out-red:hover { background: #ef4444; color: white; }
.output-point.out-slate { background: #f1f5f9; border-color: #64748b; color: #475569; }

/* Editor Modal */
.node-editor-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  z-index: 500;
  display: flex;
  align-items: center;
  justify-content: center;
}

.node-editor-modal {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 480px;
  max-height: 80vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 25px 50px -12px rgba(0,0,0,0.25);
}

.editor-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
  background: #f8fafc;
}
.editor-header h3 { margin: 0; font-size: 16px; flex: 1; }
.editor-icon { width: 20px; height: 20px; color: var(--primary-color); }
.close-btn { width: 28px; height: 28px; border: none; background: white; border-radius: 6px; font-size: 18px; cursor: pointer; border: 1px solid var(--border-color); }

.editor-body {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}

.field-group { margin-bottom: 12px; }
.field-group label { display: block; font-size: 11px; font-weight: 600; text-transform: uppercase; color: #64748b; margin-bottom: 4px; }
.field-group input, .field-group select, .field-group textarea {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 13px;
  box-sizing: border-box;
}
.field-group input:focus, .field-group select:focus, .field-group textarea:focus {
  outline: none;
  border-color: #6366f1;
}
.field-group textarea { resize: vertical; font-family: inherit; }

.field-row { display: flex; gap: 12px; }
.field-row .field-group { flex: 1; }

.checkbox-label {
  display: flex !important;
  align-items: center;
  gap: 8px;
  font-size: 12px !important;
  text-transform: none !important;
  cursor: pointer;
}
.checkbox-label input { width: auto; }

.help-note {
  font-size: 11px;
  color: #64748b;
  background: #f8fafc;
  padding: 8px 10px;
  border-radius: 4px;
  margin-top: 8px;
}

.field-divider {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  color: #94a3b8;
  border-top: 1px solid #e2e8f0;
  margin: 16px 0 12px;
  padding-top: 12px;
}

.editor-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 20px;
  border-top: 1px solid var(--border-color);
  background: #f8fafc;
}

.btn-primary {
  background: #6366f1;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}
.btn-primary:hover { background: #4f46e5; }

.btn-secondary {
  background: white;
  border: 1px solid #e2e8f0;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}
</style>

<template>
  <div class="palette">
    <div class="palette-header">
      <h3 class="palette-title">IVR Nodes</h3>
      <button class="help-btn" @click="showHelp = true">?</button>
    </div>
    
    <!-- Input Section -->
    <div class="section-label">Input</div>
    <div class="modules-grid">
      <div 
        v-for="module in inputModules" 
        :key="module.type"
        class="module-item"
        :class="module.color"
        draggable="true"
        @dragstart="onDragStart($event, module)"
      >
         <component :is="module.icon" class="module-icon" />
         <span>{{ module.label }}</span>
      </div>
    </div>

    <!-- Audio Section -->
    <div class="section-label">Audio</div>
    <div class="modules-grid">
      <div 
        v-for="module in audioModules" 
        :key="module.type"
        class="module-item"
        :class="module.color"
        draggable="true"
        @dragstart="onDragStart($event, module)"
      >
         <component :is="module.icon" class="module-icon" />
         <span>{{ module.label }}</span>
      </div>
    </div>

    <!-- Logic Section -->
    <div class="section-label">Logic & API</div>
    <div class="modules-grid">
      <div 
        v-for="module in logicModules" 
        :key="module.type"
        class="module-item"
        :class="module.color"
        draggable="true"
        @dragstart="onDragStart($event, module)"
      >
         <component :is="module.icon" class="module-icon" />
         <span>{{ module.label }}</span>
      </div>
    </div>

    <!-- Destinations Section -->
    <div class="section-label">Destinations</div>
    <div class="modules-grid">
      <div 
        v-for="module in destModules" 
        :key="module.type"
        class="module-item"
        :class="module.color"
        draggable="true"
        @dragstart="onDragStart($event, module)"
      >
         <component :is="module.icon" class="module-icon" />
         <span>{{ module.label }}</span>
      </div>
    </div>

    <!-- Help Modal -->
    <div class="help-modal-overlay" v-if="showHelp" @click.self="showHelp = false">
      <div class="help-modal">
        <div class="help-header">
          <h3>Node Reference</h3>
          <button class="close-btn" @click="showHelp = false">×</button>
        </div>
        <div class="help-content">
          <div class="help-section">
            <h4>Input Nodes</h4>
            <dl>
              <dt>Gather Digits</dt>
              <dd>Collect DTMF input (0-9, *, #). Configure max digits, timeout, terminator key.</dd>
              <dt>Speech Input</dt>
              <dd>Use ASR to capture spoken input. Requires Google/AWS/Azure speech API.</dd>
            </dl>
          </div>
          
          <div class="help-section">
            <h4>Audio Nodes</h4>
            <dl>
              <dt>Play Audio</dt>
              <dd>Play a WAV/MP3 file from Audio Library. Supports loops.</dd>
              <dt>Text-to-Speech</dt>
              <dd>Convert text to speech. Engine: Flite, Google, AWS Polly, Azure.</dd>
              <dt>Say Digits</dt>
              <dd>Read digits/numbers aloud (e.g., account numbers, PINs).</dd>
            </dl>
          </div>

          <div class="help-section">
            <h4>Logic & API Nodes</h4>
            <dl>
              <dt>Web Request</dt>
              <dd>HTTP GET/POST to external API. Returns JSON/XML. Set timeout, headers, body. Use response in routing.</dd>
              <dt>Send SMS</dt>
              <dd>Send SMS via Twilio/Nexmo/SignalWire. Configure To, Body, From.</dd>
              <dt>Database</dt>
              <dd>Query external database (MySQL, PostgreSQL, REST). Use results for routing.</dd>
              <dt>Condition</dt>
              <dd>Branch based on variable value. Compare caller input, time, API response.</dd>
              <dt>Set Variable</dt>
              <dd>Store a value for use in later nodes. Useful with API responses.</dd>
            </dl>
          </div>

          <div class="help-section">
            <h4>Destination Nodes</h4>
            <dl>
              <dt>Extension</dt>
              <dd>Transfer to internal extension.</dd>
              <dt>Queue</dt>
              <dd>Place caller in ACD queue.</dd>
              <dt>Ring Group</dt>
              <dd>Ring multiple extensions simultaneously or sequentially.</dd>
              <dt>IVR Menu</dt>
              <dd>Transfer to another IVR menu (separate flow).</dd>
              <dt>External</dt>
              <dd>Transfer to external phone number.</dd>
              <dt>Voicemail</dt>
              <dd>Send caller to voicemail box.</dd>
              <dt>Hangup</dt>
              <dd>Disconnect the call.</dd>
            </dl>
          </div>

          <div class="help-section">
            <h4>Python Script Output</h4>
            <p class="help-note">The visual flow generates a JSON template that is interpreted by a FreeSWITCH mod_python3 script. All nodes support:</p>
            <ul>
              <li>Channel variables access</li>
              <li>Dynamic routing based on caller input</li>
              <li>External API integration</li>
              <li>Database queries</li>
              <li>Custom logging</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { 
  Keyboard, Mic, Volume2, MessageSquare, Globe, Database,
  Phone, Users, PhoneCall, PhoneForwarded, Voicemail, PhoneOff, User,
  GitBranch, Variable, Layers, Hash, SpeakerIcon
} from 'lucide-vue-next'

const showHelp = ref(false)

const inputModules = [
  { type: 'gather', label: 'Gather Digits', icon: Keyboard, color: 'purple', outputs: ['Match', 'Timeout', 'Invalid'] },
  { type: 'speech', label: 'Speech Input', icon: Mic, color: 'purple', outputs: ['Match', 'No Match'] },
]

const audioModules = [
  { type: 'play_audio', label: 'Play Audio', icon: Volume2, color: 'pink', outputs: ['Next'] },
  { type: 'play_tts', label: 'Text-to-Speech', icon: MessageSquare, color: 'pink', outputs: ['Next'] },
  { type: 'say_digits', label: 'Say Digits', icon: Hash, color: 'pink', outputs: ['Next'] },
]

const logicModules = [
  { type: 'web_request', label: 'Web Request', icon: Globe, color: 'blue', outputs: ['Success', 'Error', 'Timeout'] },
  { type: 'send_sms', label: 'Send SMS', icon: MessageSquare, color: 'teal', outputs: ['Sent', 'Failed'] },
  { type: 'database', label: 'Database', icon: Database, color: 'indigo', outputs: ['Success', 'No Results', 'Error'] },
  { type: 'condition', label: 'Condition', icon: GitBranch, color: 'amber', outputs: ['True', 'False'] },
  { type: 'set_variable', label: 'Set Variable', icon: Variable, color: 'slate', outputs: ['Next'] },
]

const destModules = [
  { type: 'extension', label: 'Extension', icon: User, color: 'green' },
  { type: 'queue', label: 'Queue', icon: Users, color: 'cyan' },
  { type: 'ring_group', label: 'Ring Group', icon: PhoneCall, color: 'cyan' },
  { type: 'ivr_menu', label: 'IVR Menu', icon: Layers, color: 'blue' },
  { type: 'external', label: 'External', icon: PhoneForwarded, color: 'orange' },
  { type: 'voicemail', label: 'Voicemail', icon: Voicemail, color: 'slate' },
  { type: 'hangup', label: 'Hangup', icon: PhoneOff, color: 'red' },
]

const onDragStart = (event, module) => {
  event.dataTransfer.effectAllowed = 'copy'
  event.dataTransfer.setData('application/json', JSON.stringify(module))
}
</script>

<style scoped>
.palette {
  width: 200px;
  border-right: 1px solid var(--border-color);
  background: white;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  overflow-y: auto;
}

.palette-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.palette-title {
  font-size: 12px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.help-btn {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 1px solid var(--border-color);
  background: white;
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
  color: var(--text-muted);
}
.help-btn:hover { background: var(--primary-color); color: white; border-color: var(--primary-color); }

.section-label {
  font-size: 9px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-top: 10px;
  margin-bottom: 4px;
  letter-spacing: 0.05em;
}

.modules-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4px;
}

.module-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 8px 4px;
  cursor: grab;
  font-size: 9px;
  font-weight: 500;
  color: var(--text-main);
  transition: all 0.15s;
  gap: 3px;
}

.module-icon {
  width: 14px;
  height: 14px;
}

.module-item:hover {
  border-color: var(--primary-color);
  background: white;
  box-shadow: var(--shadow-sm);
  transform: translateY(-1px);
}

.module-item:active { cursor: grabbing; }

/* Node Colors */
.module-item.amber { border-left: 2px solid #f59e0b; }
.module-item.amber:hover { background: #fffbeb; color: #b45309; }

.module-item.blue { border-left: 2px solid #3b82f6; }
.module-item.blue:hover { background: #eff6ff; color: #1d4ed8; }

.module-item.purple { border-left: 2px solid #8b5cf6; }
.module-item.purple:hover { background: #f5f3ff; color: #6d28d9; }

.module-item.indigo { border-left: 2px solid #6366f1; }
.module-item.indigo:hover { background: #eef2ff; color: #4338ca; }

.module-item.pink { border-left: 2px solid #ec4899; }
.module-item.pink:hover { background: #fdf2f8; color: #be185d; }

.module-item.green { border-left: 2px solid #22c55e; }
.module-item.green:hover { background: #f0fdf4; color: #16a34a; }

.module-item.cyan { border-left: 2px solid #06b6d4; }
.module-item.cyan:hover { background: #ecfeff; color: #0891b2; }

.module-item.teal { border-left: 2px solid #14b8a6; }
.module-item.teal:hover { background: #f0fdfa; color: #0d9488; }

.module-item.orange { border-left: 2px solid #f97316; }
.module-item.orange:hover { background: #fff7ed; color: #c2410c; }

.module-item.slate { border-left: 2px solid #64748b; }
.module-item.slate:hover { background: #f8fafc; color: #475569; }

.module-item.red { border-left: 2px solid #ef4444; }
.module-item.red:hover { background: #fef2f2; color: #dc2626; }

/* Help Modal */
.help-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  z-index: 200;
  display: flex;
  align-items: center;
  justify-content: center;
}

.help-modal {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 600px;
  max-height: 80vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.help-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}
.help-header h3 { margin: 0; font-size: 16px; }
.close-btn {
  width: 28px;
  height: 28px;
  border: none;
  background: var(--bg-app);
  border-radius: 6px;
  font-size: 18px;
  cursor: pointer;
}

.help-content {
  padding: 20px;
  overflow-y: auto;
}

.help-section {
  margin-bottom: 20px;
}
.help-section h4 {
  font-size: 12px;
  font-weight: 700;
  color: var(--primary-color);
  margin: 0 0 8px;
  text-transform: uppercase;
}

.help-section dl {
  margin: 0;
}
.help-section dt {
  font-weight: 600;
  font-size: 12px;
  margin-top: 8px;
}
.help-section dd {
  margin: 2px 0 0 0;
  font-size: 11px;
  color: var(--text-muted);
  line-height: 1.4;
}

.help-note {
  font-size: 11px;
  color: var(--text-muted);
  margin-bottom: 8px;
}

.help-section ul {
  margin: 0;
  padding-left: 16px;
  font-size: 11px;
  color: var(--text-muted);
}
.help-section li { margin-bottom: 4px; }
</style>

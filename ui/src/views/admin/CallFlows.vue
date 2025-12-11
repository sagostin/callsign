<template>
  <div class="call-flows-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Call Flows</h2>
        <p class="text-muted text-sm">Visual routing overview for all inbound numbers</p>
      </div>
    </div>

    <!-- Filter Bar -->
    <div class="filter-bar">
      <div class="filter-group">
        <label>Number</label>
        <select v-model="selectedNumber" class="filter-select">
          <option value="">All Numbers</option>
          <option v-for="num in numbers" :key="num.number" :value="num.number">
            {{ num.number }} - {{ num.name }}
          </option>
        </select>
      </div>
      <div class="filter-group">
        <label>View</label>
        <div class="view-toggle">
          <button :class="{ active: viewMode === 'diagram' }" @click="viewMode = 'diagram'">Diagram</button>
          <button :class="{ active: viewMode === 'list' }" @click="viewMode = 'list'">List</button>
        </div>
      </div>
      <div class="filter-spacer"></div>
      <div class="legend">
        <div class="legend-item" v-for="type in nodeTypes" :key="type.key">
          <div class="legend-dot" :class="type.key"></div>
          <span>{{ type.label }}</span>
        </div>
      </div>
    </div>

    <!-- Diagram View -->
    <div class="flow-diagram" v-if="viewMode === 'diagram'">
      <div class="diagram-scroll" ref="diagramScroll">
        <!-- SVG Connections -->
        <svg class="connections-svg">
          <defs>
            <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
              <polygon points="0 0, 10 3.5, 0 7" fill="#94a3b8"/>
            </marker>
          </defs>
          <path 
            v-for="conn in visibleConnections" 
            :key="conn.id" 
            :d="conn.path"
            class="connection-path"
            :class="{ active: conn.active }"
            marker-end="url(#arrowhead)"
          />
        </svg>

        <!-- Number Groups -->
        <div class="number-group" v-for="(group, idx) in visibleGroups" :key="group.number" :style="{ top: (idx * 180) + 'px' }">
          <!-- Inbound Number -->
          <div class="flow-node inbound" :style="{ left: '20px', top: '40px' }">
            <PhoneIncomingIcon class="node-icon" />
            <div class="node-content">
              <span class="node-label">{{ group.number }}</span>
              <span class="node-detail">{{ group.name }}</span>
            </div>
          </div>

          <!-- Flow Nodes -->
          <div 
            v-for="node in group.nodes" 
            :key="node.id" 
            class="flow-node"
            :class="[node.type, { active: node.active }]"
            :style="{ left: node.x + 'px', top: node.y + 'px' }"
          >
            <component :is="getNodeIcon(node.type)" class="node-icon" />
            <div class="node-content">
              <span class="node-label">{{ node.label }}</span>
              <span class="node-detail">{{ node.detail }}</span>
            </div>
            <div class="node-outputs" v-if="node.outputs && node.outputs.length">
              <span v-for="out in node.outputs" :key="out" class="output-tag">{{ out }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- List View -->
    <div class="flow-list" v-else>
      <table class="flow-table">
        <thead>
          <tr>
            <th>Inbound Number</th>
            <th>Description</th>
            <th>Initial Destination</th>
            <th>Flow Steps</th>
            <th>Final Destinations</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="num in numbers" :key="num.number" @click="selectedNumber = num.number" :class="{ selected: selectedNumber === num.number }">
            <td class="mono">{{ num.number }}</td>
            <td>{{ num.name }}</td>
            <td>
              <span class="dest-badge" :class="num.destType">{{ num.destination }}</span>
            </td>
            <td>
              <div class="steps-preview">
                <span v-for="(step, i) in num.steps" :key="i" class="step-badge" :class="step.type">{{ step.label }}</span>
              </div>
            </td>
            <td>
              <div class="finals-list">
                <span v-for="(final, i) in num.finals" :key="i" class="final-badge">{{ final }}</span>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Stats Footer -->
    <div class="stats-footer">
      <div class="stat">
        <span class="stat-value">{{ numbers.length }}</span>
        <span class="stat-label">Inbound Numbers</span>
      </div>
      <div class="stat">
        <span class="stat-value">{{ totalNodes }}</span>
        <span class="stat-label">Routing Nodes</span>
      </div>
      <div class="stat">
        <span class="stat-value">{{ activeFlows }}</span>
        <span class="stat-label">Active Flows</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { 
  PhoneIncoming as PhoneIncomingIcon, 
  Phone as PhoneIcon, 
  ToggleLeft as ToggleIcon, 
  Clock as ClockIcon, 
  Menu as MenuIcon,
  Users as UsersIcon, 
  Voicemail as VoicemailIcon, 
  PhoneOff as HangupIcon,
  PhoneForwarded as ForwardIcon,
  User as UserIcon
} from 'lucide-vue-next'

const selectedNumber = ref('')
const viewMode = ref('diagram')
const diagramScroll = ref(null)

const nodeTypes = [
  { key: 'inbound', label: 'Inbound' },
  { key: 'toggle', label: 'Toggle' },
  { key: 'ivr', label: 'IVR' },
  { key: 'queue', label: 'Queue' },
  { key: 'extension', label: 'Extension' },
  { key: 'voicemail', label: 'Voicemail' },
]

// Sample data
const numbers = ref([
  { 
    number: '+1 (555) 100-1000', 
    name: 'Main Line', 
    destination: 'Toggle: Day/Night',
    destType: 'toggle',
    steps: [
      { type: 'toggle', label: 'Day/Night' },
      { type: 'ivr', label: 'Main Menu' },
      { type: 'queue', label: 'Sales' }
    ],
    finals: ['Sales Queue', 'Support Queue', 'Voicemail']
  },
  { 
    number: '+1 (555) 100-1001', 
    name: 'Sales Direct', 
    destination: 'IVR: Sales Menu',
    destType: 'ivr',
    steps: [
      { type: 'ivr', label: 'Sales Menu' },
      { type: 'queue', label: 'Sales' }
    ],
    finals: ['Sales Queue', 'Voicemail']
  },
  { 
    number: '+1 (555) 100-1002', 
    name: 'Support Line', 
    destination: 'Queue: Support',
    destType: 'queue',
    steps: [
      { type: 'queue', label: 'Support' }
    ],
    finals: ['Support Queue', 'Voicemail']
  },
  { 
    number: '+1 (555) 100-1003', 
    name: 'Emergency Line', 
    destination: 'Extension: 911',
    destType: 'extension',
    steps: [
      { type: 'extension', label: 'Ext 911' }
    ],
    finals: ['Extension 911']
  },
])

const flowGroups = ref([
  {
    number: '+1 (555) 100-1000',
    name: 'Main Line',
    nodes: [
      { id: 't1', type: 'toggle', label: 'Day/Night', detail: 'Ext. 30', x: 220, y: 20, outputs: ['A', 'B'], active: true },
      { id: 'ivr1', type: 'ivr', label: 'Main Menu', detail: 'Ext. 8000', x: 420, y: 0, outputs: ['1', '2', '0'] },
      { id: 'ivr2', type: 'ivr', label: 'After Hours', detail: 'Ext. 8001', x: 420, y: 80 },
      { id: 'q1', type: 'queue', label: 'Sales', detail: '5 agents', x: 620, y: 0 },
      { id: 'q2', type: 'queue', label: 'Support', detail: '3 agents', x: 620, y: 50 },
      { id: 'vm1', type: 'voicemail', label: 'General', detail: 'Box 100', x: 620, y: 100 },
    ],
    connections: [
      { from: 'start', to: 't1', label: '' },
      { from: 't1', to: 'ivr1', label: 'A' },
      { from: 't1', to: 'ivr2', label: 'B' },
      { from: 'ivr1', to: 'q1', label: '1' },
      { from: 'ivr1', to: 'q2', label: '2' },
      { from: 'ivr1', to: 'vm1', label: '0' },
    ]
  },
  {
    number: '+1 (555) 100-1001',
    name: 'Sales Direct',
    nodes: [
      { id: 'ivr3', type: 'ivr', label: 'Sales Menu', detail: 'Ext. 8002', x: 220, y: 40 },
      { id: 'q3', type: 'queue', label: 'Sales', detail: '5 agents', x: 420, y: 40 },
    ],
    connections: []
  },
  {
    number: '+1 (555) 100-1002',
    name: 'Support Line',
    nodes: [
      { id: 'q4', type: 'queue', label: 'Support', detail: '3 agents', x: 220, y: 40 },
      { id: 'vm2', type: 'voicemail', label: 'Support', detail: 'Box 200', x: 420, y: 40 },
    ],
    connections: []
  },
  {
    number: '+1 (555) 100-1003',
    name: 'Emergency Line',
    nodes: [
      { id: 'e1', type: 'extension', label: 'Emergency', detail: 'Ext. 911', x: 220, y: 40 },
    ],
    connections: []
  },
])

const visibleGroups = computed(() => {
  if (!selectedNumber.value) return flowGroups.value
  return flowGroups.value.filter(g => g.number === selectedNumber.value)
})

const visibleConnections = computed(() => {
  // Generate connection paths for visible groups
  const conns = []
  let yOffset = 0
  
  visibleGroups.value.forEach((group, groupIdx) => {
    const baseY = groupIdx * 180
    
    // Connection from inbound to first node
    if (group.nodes.length > 0) {
      const firstNode = group.nodes[0]
      conns.push({
        id: `${group.number}_start`,
        path: `M 170 ${baseY + 60} C 195 ${baseY + 60}, 195 ${baseY + firstNode.y + 25}, ${firstNode.x} ${baseY + firstNode.y + 25}`,
        active: firstNode.active
      })
    }
    
    // Inter-node connections
    group.connections.forEach((conn, i) => {
      const fromNode = group.nodes.find(n => n.id === conn.from)
      const toNode = group.nodes.find(n => n.id === conn.to)
      if (fromNode && toNode) {
        const x1 = fromNode.x + 150
        const y1 = baseY + fromNode.y + 25
        const x2 = toNode.x
        const y2 = baseY + toNode.y + 25
        conns.push({
          id: `${group.number}_${i}`,
          path: `M ${x1} ${y1} C ${x1 + 30} ${y1}, ${x2 - 30} ${y2}, ${x2} ${y2}`,
          active: fromNode.active
        })
      }
    })
  })
  
  return conns
})

const totalNodes = computed(() => {
  return flowGroups.value.reduce((sum, g) => sum + g.nodes.length, 0)
})

const activeFlows = computed(() => {
  return numbers.value.length
})

const getNodeIcon = (type) => {
  const icons = {
    toggle: ToggleIcon,
    ivr: MenuIcon,
    queue: UsersIcon,
    extension: UserIcon,
    voicemail: VoicemailIcon,
    hangup: HangupIcon,
    forward: ForwardIcon
  }
  return icons[type] || PhoneIcon
}
</script>

<style scoped>
.call-flows-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 60px);
}

.view-header {
  padding: 16px 20px;
  background: white;
  border-bottom: 1px solid var(--border-color);
}

.filter-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 20px;
  background: white;
  border-bottom: 1px solid var(--border-color);
}

.filter-group { display: flex; align-items: center; gap: 8px; }
.filter-group label { font-size: 11px; font-weight: 600; text-transform: uppercase; color: var(--text-muted); }
.filter-select { padding: 6px 10px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 12px; min-width: 200px; }

.view-toggle {
  display: flex;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  overflow: hidden;
}
.view-toggle button {
  padding: 6px 12px;
  border: none;
  background: white;
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
}
.view-toggle button.active { background: var(--primary-color); color: white; }
.view-toggle button:not(:last-child) { border-right: 1px solid var(--border-color); }

.filter-spacer { flex: 1; }

.legend { display: flex; gap: 12px; }
.legend-item { display: flex; align-items: center; gap: 4px; font-size: 10px; color: var(--text-muted); }
.legend-dot { width: 10px; height: 10px; border-radius: 3px; }
.legend-dot.inbound { background: #22c55e; }
.legend-dot.toggle { background: #8b5cf6; }
.legend-dot.ivr { background: #3b82f6; }
.legend-dot.queue { background: #06b6d4; }
.legend-dot.extension { background: #f59e0b; }
.legend-dot.voicemail { background: #64748b; }

/* Diagram View */
.flow-diagram {
  flex: 1;
  overflow: auto;
  background: #f8fafc;
  position: relative;
}

.diagram-scroll {
  position: relative;
  min-height: 100%;
  min-width: 900px;
  padding: 20px;
}

.connections-svg {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.connection-path {
  fill: none;
  stroke: #cbd5e1;
  stroke-width: 2;
}
.connection-path.active { stroke: #22c55e; stroke-width: 2.5; }

.number-group {
  position: absolute;
  left: 0;
  width: 100%;
  height: 160px;
  border-bottom: 1px dashed #e2e8f0;
}

.flow-node {
  position: absolute;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: white;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  min-width: 140px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
}
.flow-node.active { border-color: #22c55e; box-shadow: 0 0 0 3px rgba(34,197,94,0.15); }
.flow-node.inbound { border-color: #22c55e; background: #f0fdf4; }
.flow-node.toggle { border-color: #8b5cf6; background: #faf5ff; }
.flow-node.ivr { border-color: #3b82f6; background: #eff6ff; }
.flow-node.queue { border-color: #06b6d4; background: #ecfeff; }
.flow-node.extension { border-color: #f59e0b; background: #fffbeb; }
.flow-node.voicemail { border-color: #64748b; background: #f8fafc; }

.node-icon { width: 16px; height: 16px; opacity: 0.8; }
.node-content { display: flex; flex-direction: column; flex: 1; }
.node-label { font-size: 11px; font-weight: 600; }
.node-detail { font-size: 9px; color: var(--text-muted); }
.node-outputs { display: flex; gap: 2px; margin-left: 8px; }
.output-tag { font-size: 8px; background: rgba(0,0,0,0.08); padding: 1px 4px; border-radius: 2px; }

/* List View */
.flow-list {
  flex: 1;
  overflow: auto;
  padding: 20px;
}

.flow-table {
  width: 100%;
  border-collapse: collapse;
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}
.flow-table th {
  text-align: left;
  padding: 12px 16px;
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
  background: #f8fafc;
  border-bottom: 1px solid var(--border-color);
}
.flow-table td {
  padding: 12px 16px;
  font-size: 13px;
  border-bottom: 1px solid var(--border-color);
}
.flow-table tr { cursor: pointer; transition: background 0.15s; }
.flow-table tbody tr:hover { background: #f8fafc; }
.flow-table tr.selected { background: #eff6ff; }
.mono { font-family: monospace; font-size: 12px; }

.dest-badge {
  display: inline-block;
  padding: 3px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}
.dest-badge.toggle { background: #faf5ff; color: #7c3aed; }
.dest-badge.ivr { background: #eff6ff; color: #2563eb; }
.dest-badge.queue { background: #ecfeff; color: #0891b2; }
.dest-badge.extension { background: #fffbeb; color: #b45309; }

.steps-preview { display: flex; gap: 4px; flex-wrap: wrap; }
.step-badge {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 3px;
  background: #f1f5f9;
  color: var(--text-muted);
}
.step-badge.toggle { background: #faf5ff; color: #7c3aed; }
.step-badge.ivr { background: #eff6ff; color: #2563eb; }
.step-badge.queue { background: #ecfeff; color: #0891b2; }
.step-badge.extension { background: #fffbeb; color: #b45309; }

.finals-list { display: flex; gap: 4px; flex-wrap: wrap; }
.final-badge { font-size: 10px; padding: 2px 6px; border-radius: 3px; background: #f0fdf4; color: #16a34a; }

/* Stats Footer */
.stats-footer {
  display: flex;
  gap: 32px;
  padding: 12px 20px;
  background: white;
  border-top: 1px solid var(--border-color);
}
.stat { display: flex; align-items: center; gap: 8px; }
.stat-value { font-size: 18px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 11px; color: var(--text-muted); }
</style>

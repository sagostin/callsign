<template>
  <div 
    class="flow-canvas" 
    ref="canvasRef"
    :style="{ transform: `scale(${zoom})`, transformOrigin: 'top left' }"
    @dragover.prevent 
    @drop="onDrop"
    @mousemove="onMouseMove"
    @mouseup="onMouseUp"
    @click.self="deselectAll"
  >
    <div class="grid-bg"></div>

    <!-- Connections (SVG Layer) -->
    <svg class="connections-layer">
       <!-- Active Drag Line -->
       <path 
         v-if="dragLine" 
         :d="getCurvedPath(dragLine.start, dragLine.end)" 
         class="connector dragging" 
       />
       
       <!-- Existing Connections -->
       <g v-for="conn in connections" :key="conn.id">
         <path 
           :d="getConnectionPath(conn)"
           class="connector"
           :class="{ selected: selectedConnectionId === conn.id }"
           @click.stop="selectConnection(conn.id)"
         />
         <!-- Connection Label -->
         <text 
           v-if="conn.label"
           :x="getConnectionMidpoint(conn).x" 
           :y="getConnectionMidpoint(conn).y - 8"
           class="connection-label"
         >{{ conn.label }}</text>
       </g>
    </svg>

    <!-- Input Port Targets (for receiving connections) -->
    <div 
      v-for="node in nodes" 
      :key="'input_' + node.id"
      class="input-port"
      :class="{ 
        highlight: dragLine && hoveredNodeId === node.id,
        'show-always': dragLine !== null,
        'cannot-connect': dragLine && dragLine.sourceNodeId === node.id
      }"
      :style="{ left: (node.x + 90) + 'px', top: (node.y - 12) + 'px' }"
      @mouseenter="onPortEnter(node.id)"
      @mouseleave="onPortLeave"
      @mouseup.stop="finishConnection(node.id)"
    >
      <div class="port-dot"></div>
    </div>

    <!-- Nodes -->
    <FlowNode
       v-for="node in nodes"
       :key="node.id"
       :data="node"
       :x="node.x"
       :y="node.y"
       :isSelected="selectedNodeId === node.id"
       @select="selectNode(node.id)"
       @drag-start="startNodeDrag(node, $event)"
       @delete="deleteNode(node.id)"
       @connector-drag="startConnectorDrag"
    />

    <!-- Connection Delete Button (when selected) -->
    <div 
      v-if="selectedConnectionId && selectedConnectionMidpoint"
      class="connection-delete-btn"
      :style="{ left: selectedConnectionMidpoint.x + 'px', top: selectedConnectionMidpoint.y + 'px' }"
      @click.stop="deleteConnection(selectedConnectionId)"
    >×</div>

    <!-- Drag instruction overlay -->
    <div v-if="dragLine" class="drag-hint">
      Drop on ● to connect
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import FlowNode from './FlowNode.vue'

const props = defineProps({
  modelValue: Object, // { nodes: [], connections: [] }
  zoom: { type: Number, default: 1 }
})

const emit = defineEmits(['update:modelValue'])

const canvasRef = ref(null)
const nodes = ref(props.modelValue?.nodes || [])
const connections = ref(props.modelValue?.connections || [])
const selectedNodeId = ref(null)
const selectedConnectionId = ref(null)
const hoveredNodeId = ref(null)

// Watch for external updates
watch(() => props.modelValue, (newVal) => {
  if (newVal) {
    nodes.value = newVal.nodes || []
    connections.value = newVal.connections || []
  }
}, { deep: true })

// Dragging State
const isDraggingNode = ref(false)
const draggedNode = ref(null)
const dragOffset = ref({ x: 0, y: 0 })

// Connector Dragging
const dragLine = ref(null)

// Computed
const selectedConnectionMidpoint = computed(() => {
  if (!selectedConnectionId.value) return null
  const conn = connections.value.find(c => c.id === selectedConnectionId.value)
  if (!conn) return null
  return getConnectionMidpoint(conn)
})

// --- Node Selection --- //
const selectNode = (id) => {
  selectedNodeId.value = id
  selectedConnectionId.value = null
}

const selectConnection = (id) => {
  selectedConnectionId.value = id
  selectedNodeId.value = null
}

// --- Node Dragging Logic --- //
const startNodeDrag = (node, event) => {
  isDraggingNode.value = true
  draggedNode.value = node
  const rect = canvasRef.value.getBoundingClientRect()
  dragOffset.value = {
    x: (event.clientX - rect.left) / props.zoom - node.x,
    y: (event.clientY - rect.top) / props.zoom - node.y
  }
}

const onMouseMove = (event) => {
  if (!canvasRef.value) return
  const rect = canvasRef.value.getBoundingClientRect()
  const mouseX = (event.clientX - rect.left) / props.zoom
  const mouseY = (event.clientY - rect.top) / props.zoom

  // Move Node
  if (isDraggingNode.value && draggedNode.value) {
    draggedNode.value.x = Math.max(0, mouseX - dragOffset.value.x)
    draggedNode.value.y = Math.max(0, mouseY - dragOffset.value.y)
    emitUpdate()
  }

  // Move Connector Line
  if (dragLine.value) {
    dragLine.value.end = { x: mouseX, y: mouseY }
  }
}

const onMouseUp = (event) => {
  isDraggingNode.value = false
  draggedNode.value = null

  // If we have a hovered target, connect to it
  if (dragLine.value && hoveredNodeId.value) {
    finishConnection(hoveredNodeId.value)
  } else if (dragLine.value) {
    // Cancel connection - not dropped on valid target
    dragLine.value = null
  }
}

// --- Port hover handling --- //
const onPortEnter = (nodeId) => {
  if (dragLine.value && dragLine.value.sourceNodeId !== nodeId) {
    hoveredNodeId.value = nodeId
  }
}

const onPortLeave = () => {
  hoveredNodeId.value = null
}

// --- Connection Logic --- //
const startConnectorDrag = ({ id, output }, event) => {
  const node = nodes.value.find(n => n.id === id)
  if (!node) return
  
  event.preventDefault()
  
  // Start from the output port position (bottom center of node)
  const startX = node.x + 90
  const startY = node.y + getNodeHeight(node)
  
  dragLine.value = {
    start: { x: startX, y: startY },
    end: { x: startX, y: startY + 20 },
    sourceNodeId: id,
    sourceOutput: output
  }
  
  // Clear any selection
  selectedNodeId.value = null
  selectedConnectionId.value = null
}

const getNodeHeight = (node) => {
  // Estimate height based on node type (nodes with outputs are taller)
  return 90
}

const finishConnection = (targetNodeId) => {
  if (!dragLine.value) return
  if (dragLine.value.sourceNodeId === targetNodeId) {
    dragLine.value = null
    hoveredNodeId.value = null
    return // Can't connect to self
  }

  // Check if connection already exists
  const exists = connections.value.some(c => 
    c.sourceId === dragLine.value.sourceNodeId && 
    c.targetId === targetNodeId &&
    c.sourceOutput === dragLine.value.sourceOutput
  )

  if (!exists) {
    connections.value.push({
      id: 'conn_' + Date.now(),
      sourceId: dragLine.value.sourceNodeId,
      targetId: targetNodeId,
      sourceOutput: dragLine.value.sourceOutput,
      label: dragLine.value.sourceOutput || ''
    })
    emitUpdate()
  }

  dragLine.value = null
  hoveredNodeId.value = null
}

const deleteConnection = (connId) => {
  connections.value = connections.value.filter(c => c.id !== connId)
  selectedConnectionId.value = null
  emitUpdate()
}

// --- Drop New Node Logic --- //
const onDrop = (event) => {
  const data = event.dataTransfer.getData('application/json')
  if (!data) return

  const module = JSON.parse(data)
  const rect = canvasRef.value.getBoundingClientRect()
  
  const newNode = {
    id: 'node_' + Date.now(),
    type: module.type,
    label: module.label,
    x: (event.clientX - rect.left) / props.zoom - 90,
    y: (event.clientY - rect.top) / props.zoom - 30,
    config: {}
  }

  nodes.value.push(newNode)
  emitUpdate()
}

// --- Path Calculations --- //
const getCurvedPath = (start, end) => {
  const deltaY = end.y - start.y
  const controlOffset = Math.max(40, Math.abs(deltaY) * 0.5)
  
  return `M ${start.x} ${start.y} 
          C ${start.x} ${start.y + controlOffset}, 
            ${end.x} ${end.y - controlOffset}, 
            ${end.x} ${end.y}`
}

const getConnectionPath = (conn) => {
  const source = nodes.value.find(n => n.id === conn.sourceId)
  const target = nodes.value.find(n => n.id === conn.targetId)
  if (!source || !target) return ''

  // Source: bottom center
  const start = { x: source.x + 90, y: source.y + getNodeHeight(source) }
  // Target: top center
  const end = { x: target.x + 90, y: target.y }
  
  return getCurvedPath(start, end)
}

const getConnectionMidpoint = (conn) => {
  const source = nodes.value.find(n => n.id === conn.sourceId)
  const target = nodes.value.find(n => n.id === conn.targetId)
  if (!source || !target) return { x: 0, y: 0 }

  const startY = source.y + getNodeHeight(source)
  const endY = target.y

  return {
    x: (source.x + target.x) / 2 + 90,
    y: (startY + endY) / 2
  }
}

const deleteNode = (id) => {
  nodes.value = nodes.value.filter(n => n.id !== id)
  connections.value = connections.value.filter(c => c.sourceId !== id && c.targetId !== id)
  emitUpdate()
}

const deselectAll = () => {
  selectedNodeId.value = null
  selectedConnectionId.value = null
}

const emitUpdate = () => {
  emit('update:modelValue', {
    nodes: nodes.value,
    connections: connections.value
  })
}

defineExpose({
  loadGraph: (data) => {
    nodes.value = data.nodes
    connections.value = data.connections
  }
})
</script>

<style scoped>
.flow-canvas {
  width: 100%;
  height: 100%;
  position: relative;
  overflow: hidden;
  background-color: #f8fafc;
  min-height: 500px;
}

.grid-bg {
  position: absolute; 
  inset: 0;
  width: 3000px;
  height: 2000px;
  background-image: radial-gradient(circle, #cbd5e1 1px, transparent 1px);
  background-size: 20px 20px;
  opacity: 0.5;
  pointer-events: none;
}

.connections-layer {
  position: absolute; 
  inset: 0; 
  width: 3000px; 
  height: 2000px; 
  pointer-events: none; 
  z-index: 5;
}

.connector {
  fill: none;
  stroke: #94a3b8;
  stroke-width: 2px;
  pointer-events: visibleStroke;
  cursor: pointer;
  transition: stroke 0.2s;
}
.connector:hover { stroke: #6366f1; stroke-width: 3px; }
.connector.selected { stroke: #ef4444; stroke-width: 3px; }
.connector.dragging { stroke: #6366f1; stroke-width: 2px; stroke-dasharray: 8 4; animation: dash 0.5s linear infinite; }

@keyframes dash {
  to { stroke-dashoffset: -12; }
}

.connection-label {
  font-size: 10px;
  fill: #64748b;
  text-anchor: middle;
  pointer-events: none;
  font-weight: 500;
}

/* Input Port for receiving connections */
.input-port {
  position: absolute;
  width: 24px;
  height: 24px;
  z-index: 15;
  display: flex;
  align-items: center;
  justify-content: center;
  transform: translateX(-50%);
  cursor: pointer;
  border-radius: 50%;
  opacity: 0;
  transition: all 0.2s;
}

.input-port.show-always {
  opacity: 1;
}

.input-port.cannot-connect {
  opacity: 0.3;
  cursor: not-allowed;
}

.port-dot {
  width: 12px;
  height: 12px;
  background: #cbd5e1;
  border: 2px solid white;
  border-radius: 50%;
  box-shadow: 0 1px 3px rgba(0,0,0,0.2);
  transition: all 0.2s;
}

.input-port:hover .port-dot,
.input-port.highlight .port-dot {
  background: #6366f1;
  transform: scale(1.3);
  box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.2);
}

.connection-delete-btn {
  position: absolute;
  width: 22px;
  height: 22px;
  background: #ef4444;
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: bold;
  cursor: pointer;
  transform: translate(-50%, -50%);
  z-index: 20;
  box-shadow: 0 2px 4px rgba(0,0,0,0.2);
}
.connection-delete-btn:hover {
  background: #dc2626;
  transform: translate(-50%, -50%) scale(1.1);
}

.drag-hint {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(0,0,0,0.8);
  color: white;
  padding: 8px 16px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
  z-index: 100;
  pointer-events: none;
}
</style>

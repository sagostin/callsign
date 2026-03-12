<template>
  <span class="status-badge" :class="badgeClass">
    <span v-if="showDot" class="status-dot"></span>
    <span class="status-text">{{ displayStatus }}</span>
  </span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  status: {
    type: String,
    default: 'unknown'
  },
  showDot: {
    type: Boolean,
    default: true
  }
})

// Status normalization and display mapping
const statusMap = {
  // Active/Positive states
  'active': { label: 'Active', type: 'success' },
  'online': { label: 'Online', type: 'success' },
  'enabled': { label: 'Enabled', type: 'success' },
  'completed': { label: 'Completed', type: 'success' },
  'success': { label: 'Success', type: 'success' },
  'good': { label: 'Good', type: 'success' },
  'healthy': { label: 'Healthy', type: 'success' },
  'idle': { label: 'Idle', type: 'success' },
  'registered': { label: 'Registered', type: 'success' },
  
  // Warning states
  'warning': { label: 'Warning', type: 'warning' },
  'warn': { label: 'Warning', type: 'warning' },
  'pending': { label: 'Pending', type: 'warning' },
  'processing': { label: 'Processing', type: 'warning' },
  'in_progress': { label: 'In Progress', type: 'warning' },
  'in call': { label: 'In Call', type: 'warning' },
  'ringing': { label: 'Ringing', type: 'warning' },
  'busy': { label: 'Busy', type: 'warning' },
  
  // Error/Negative states
  'error': { label: 'Error', type: 'error' },
  'failed': { label: 'Failed', type: 'error' },
  'inactive': { label: 'Inactive', type: 'error' },
  'offline': { label: 'Offline', type: 'error' },
  'disabled': { label: 'Disabled', type: 'error' },
  'suspended': { label: 'Suspended', type: 'error' },
  'unregistered': { label: 'Unregistered', type: 'error' },
  
  // Neutral states
  'unknown': { label: 'Unknown', type: 'neutral' },
  'neutral': { label: 'Neutral', type: 'neutral' },
  'draft': { label: 'Draft', type: 'neutral' }
}

const normalizedStatus = computed(() => {
  const normalized = props.status?.toString().toLowerCase().trim()
  return statusMap[normalized] || { label: props.status || 'Unknown', type: 'neutral' }
})

const displayStatus = computed(() => normalizedStatus.value.label)

const badgeClass = computed(() => {
  const type = normalizedStatus.value.type
  return {
    'status-success': type === 'success',
    'status-warning': type === 'warning',
    'status-error': type === 'error',
    'status-neutral': type === 'neutral'
  }
})
</script>

<style scoped>
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-1-5);
  padding: var(--spacing-0-5) var(--spacing-2);
  border-radius: var(--radius-full);
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  text-transform: uppercase;
  letter-spacing: var(--tracking-wider);
  line-height: var(--leading-none);
  white-space: nowrap;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: var(--radius-full);
  flex-shrink: 0;
}

.status-text {
  line-height: 1;
}

/* Success variant */
.status-success {
  background-color: var(--status-good-bg);
  color: var(--status-good);
}

.status-success .status-dot {
  background-color: var(--status-good);
  box-shadow: 0 0 0 2px var(--status-good-subtle);
}

/* Warning variant */
.status-warning {
  background-color: var(--status-warn-bg);
  color: var(--status-warn);
}

.status-warning .status-dot {
  background-color: var(--status-warn);
  box-shadow: 0 0 0 2px var(--status-warn-subtle);
}

/* Error variant */
.status-error {
  background-color: var(--status-bad-bg);
  color: var(--status-bad);
}

.status-error .status-dot {
  background-color: var(--status-bad);
  box-shadow: 0 0 0 2px var(--status-bad-subtle);
}

/* Neutral variant */
.status-neutral {
  background-color: var(--status-neutral-bg);
  color: var(--status-neutral);
}

.status-neutral .status-dot {
  background-color: var(--status-neutral);
  box-shadow: 0 0 0 2px var(--bg-hover);
}

/* Mobile responsive */
@media (max-width: 640px) {
  .status-badge {
    padding: var(--spacing-0-5) var(--spacing-1-5);
    font-size: var(--text-2xs);
  }
  
  .status-dot {
    width: 5px;
    height: 5px;
  }
}
</style>

<template>
  <span :class="['badge', statusType]">{{ label }}</span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  status: {
    type: String,
    required: false,
    default: 'unknown'
  }
})

const statusType = computed(() => {
  if (!props.status) return 'neutral'
  const s = props.status.toLowerCase()
  if (['operational', 'active', 'online', 'registered', 'idle', 'enabled'].includes(s)) return 'good'
  if (['warning', 'degraded', 'ringing', 'away'].includes(s)) return 'warn'
  if (['error', 'offline', 'unregistered', 'busy', 'in call', 'disabled'].includes(s)) return 'bad'
  return 'neutral'
})

const label = computed(() => {
  return props.status || 'Unknown'
})
</script>

<style scoped>
/* Inherits .badge from global index.css */
</style>

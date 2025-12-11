<template>
  <span :class="['badge', statusType]">{{ label }}</span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  status: {
    type: String,
    required: true
  }
})

const statusType = computed(() => {
  const s = props.status.toLowerCase()
  if (['operational', 'active', 'online', 'registered', 'idle'].includes(s)) return 'good'
  if (['warning', 'degraded', 'ringing', 'away'].includes(s)) return 'warn'
  if (['error', 'offline', 'unregistered', 'busy', 'in call'].includes(s)) return 'bad'
  return 'neutral'
})

const label = computed(() => {
  return props.status // Can add formatting here if needed
})
</script>

<style scoped>
/* Inherits .badge from global index.css */
</style>

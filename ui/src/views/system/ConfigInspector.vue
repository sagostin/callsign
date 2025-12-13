<template>
  <div class="view-container">
    <div class="view-header">
      <h2>System Configuration Inspector</h2>
      <p class="text-muted">Debug and inspect generated FreeSWITCH XML configuration.</p>
    </div>

    <ConfigDebugger mode="system" />
  </div>
</template>

<script setup>
import ConfigDebugger from '@/components/ConfigDebugger.vue'
</script>

<style scoped>
.view-container {
  max-width: 1200px;
  margin: 0 auto;
}
.view-header {
  margin-bottom: 24px;
}
</style>

<template>
  <div class="view-container">
    <div class="view-header">
      <h2>Routing Debugger</h2>
      <p class="text-muted">Simulate inbound and outbound call routing logic.</p>
    </div>

    <ConfigDebugger mode="tenant" />
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

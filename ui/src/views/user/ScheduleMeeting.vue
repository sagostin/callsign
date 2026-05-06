<template>
  <div class="schedule-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Schedule Meeting</h2>
        <p class="text-muted text-sm">Schedule a new conference meeting and send invites to participants.</p>
      </div>
    </div>

    <div class="coming-soon-card">
      <CalendarIcon class="empty-icon" />
      <h3>Meeting Scheduling</h3>
      <p class="text-muted">Schedule meetings, set recurrence, and invite participants. This feature is coming soon.</p>
    </div>
  </div>
</template>

<script setup>
import { Calendar as CalendarIcon } from 'lucide-vue-next'
</script>

<style scoped>
.schedule-page { padding: 0; }

.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-content h2 { margin: 0 0 4px; }

.coming-soon-card {
  text-align: center;
  padding: 64px 48px;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  color: var(--text-muted);
}

.coming-soon-card h3 {
  margin: 16px 0 8px;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}

.empty-icon { width: 48px; height: 48px; opacity: 0.3; }
</style>

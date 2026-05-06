<template>
  <div class="recordings-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Meeting Recordings</h2>
        <p class="text-muted text-sm">Browse and manage recordings from your conference meetings.</p>
      </div>
    </div>

    <div class="coming-soon-card">
      <VideoIcon class="empty-icon" />
      <h3>Conference Recordings</h3>
      <p class="text-muted">Access recordings from your meetings, download, and share with participants. This feature is coming soon.</p>
    </div>
  </div>
</template>

<script setup>
import { Video as VideoIcon } from 'lucide-vue-next'
</script>

<style scoped>
.recordings-page { padding: 0; }

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

<template>
  <div class="stat-card">
    <div class="header">
      <span class="label">{{ label }}</span>
      <component :is="iconComponent" class="icon muted" v-if="iconName" />
    </div>

    <div class="body">
      <div class="value">{{ value }}</div>
      <div v-if="trend || subtext" class="meta">
        <span v-if="trend" :class="['trend', trend > 0 ? 'up' : 'down']">
          <TrendingUp v-if="trend > 0" class="icon-xs" />
          <TrendingDown v-else class="icon-xs" />
          {{ Math.abs(trend) }}%
        </span>
        <span v-if="subtext" class="subtext">{{ subtext }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { TrendingUp, TrendingDown, Activity, Phone, Server, AlertCircle } from 'lucide-vue-next'

const props = defineProps({
  label: String,
  value: [String, Number],
  subtext: String,
  trend: Number,
  iconName: String // Optional icon name to map
})

const iconComponent = computed(() => {
  switch (props.iconName) {
    case 'calls': return Phone
    case 'server': return Server
    case 'alert': return AlertCircle
    default: return Activity
  }
})
</script>

<style scoped>
.stat-card {
  background: white;
  border-radius: var(--radius-md);
  padding: var(--spacing-lg);
  box-shadow: var(--shadow-sm);
  border: 1px solid transparent;
  transition: all var(--transition-fast);
}

.stat-card:hover {
  box-shadow: var(--shadow-md);
  border-color: var(--border-color);
  transform: translateY(-2px);
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-sm);
}

.label {
  font-size: var(--text-sm);
  color: var(--text-muted);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.value {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.2;
  letter-spacing: -0.02em;
}

.meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
}

.trend {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: var(--text-xs);
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 99px;
}

.trend.up { background-color: var(--status-good-bg); color: var(--status-good); }
.trend.down { background-color: var(--status-bad-bg); color: var(--status-bad); }

.subtext {
  font-size: var(--text-xs);
  color: var(--text-muted);
}

.icon { width: 18px; height: 18px; }
.icon-xs { width: 12px; height: 12px; }
.muted { color: var(--text-muted); opacity: 0.7; }
</style>

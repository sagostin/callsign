<template>
  <div class="stat-card" :class="[variant, { 'clickable': clickable }]" @click="handleClick">
    <div class="stat-icon-wrapper" :class="iconBgClass">
      <component :is="iconComponent" class="stat-icon" />
    </div>
    <div class="stat-content">
      <span class="stat-label">{{ label }}</span>
      <span class="stat-value" :class="{ 'loading': isLoading }">{{ displayValue }}</span>
      <span v-if="subtext" class="stat-subtext">{{ subtext }}</span>
      <span v-else-if="trend" class="stat-trend" :class="trend.type">
        <TrendingUpIcon v-if="trend.type === 'up'" class="trend-icon" />
        <TrendingDownIcon v-if="trend.type === 'down'" class="trend-icon" />
        <MinusIcon v-if="trend.type === 'neutral'" class="trend-icon" />
        {{ trend.value }}
      </span>
    </div>
    <div v-if="clickable" class="stat-arrow">
      <ArrowRightIcon class="arrow-icon" />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { 
  Phone, 
  Users, 
  AlertTriangle, 
  Activity,
  Server,
  Clock,
  CheckCircle,
  XCircle,
  ArrowRight as ArrowRightIcon,
  TrendingUp as TrendingUpIcon,
  TrendingDown as TrendingDownIcon,
  Minus as MinusIcon
} from 'lucide-vue-next'

const props = defineProps({
  label: {
    type: String,
    required: true
  },
  value: {
    type: [String, Number],
    default: '—'
  },
  subtext: {
    type: String,
    default: ''
  },
  trend: {
    type: Object,
    default: null
    // { type: 'up' | 'down' | 'neutral', value: string }
  },
  iconName: {
    type: String,
    default: 'default'
    // 'calls', 'users', 'alert', 'activity', 'server', 'clock', 'check', 'error'
  },
  variant: {
    type: String,
    default: 'default'
    // 'default', 'success', 'warning', 'error', 'info'
  },
  isLoading: {
    type: Boolean,
    default: false
  },
  clickable: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['click'])

// Icon mapping
const iconMap = {
  calls: Phone,
  users: Users,
  alert: AlertTriangle,
  activity: Activity,
  server: Server,
  clock: Clock,
  check: CheckCircle,
  error: XCircle,
  default: Activity
}

const iconComponent = computed(() => iconMap[props.iconName] || iconMap.default)

// Background class based on variant
const iconBgClass = computed(() => {
  const classes = {
    default: 'bg-blue-100 text-blue-600',
    success: 'bg-emerald-100 text-emerald-600',
    warning: 'bg-amber-100 text-amber-600',
    error: 'bg-rose-100 text-rose-600',
    info: 'bg-indigo-100 text-indigo-600'
  }
  return classes[props.variant] || classes.default
})

const displayValue = computed(() => {
  if (props.isLoading) return '...'
  return props.value
})

const handleClick = () => {
  if (props.clickable) {
    emit('click')
  }
}
</script>

<style scoped>
.stat-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: var(--spacing-5);
  display: flex;
  align-items: center;
  gap: var(--spacing-4);
  box-shadow: var(--shadow-sm);
  transition: all var(--transition-fast);
}

.stat-card:hover {
  box-shadow: var(--shadow-md);
}

.stat-card.clickable {
  cursor: pointer;
}

.stat-card.clickable:hover {
  border-color: var(--primary-color);
  transform: translateY(-2px);
}

/* Variant styles */
.stat-card.success {
  border-left: 4px solid var(--status-good);
}

.stat-card.warning {
  border-left: 4px solid var(--status-warn);
}

.stat-card.error {
  border-left: 4px solid var(--status-bad);
}

.stat-card.info {
  border-left: 4px solid var(--status-info);
}

.stat-icon-wrapper {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-icon {
  width: 24px;
  height: 24px;
}

.stat-content {
  display: flex;
  flex-direction: column;
  min-width: 0;
  flex: 1;
}

.stat-label {
  font-size: var(--text-sm);
  color: var(--text-secondary);
  font-weight: var(--font-medium);
  margin-bottom: var(--spacing-0-5);
}

.stat-value {
  font-size: var(--text-2xl);
  font-weight: var(--font-bold);
  color: var(--text-primary);
  line-height: var(--leading-tight);
}

.stat-value.loading {
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  color: var(--text-muted);
}

.stat-subtext {
  font-size: var(--text-xs);
  color: var(--text-muted);
  margin-top: var(--spacing-0-5);
}

.stat-trend {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-1);
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  margin-top: var(--spacing-0-5);
}

.stat-trend.up {
  color: var(--status-good);
}

.stat-trend.down {
  color: var(--status-bad);
}

.stat-trend.neutral {
  color: var(--text-muted);
}

.trend-icon {
  width: 14px;
  height: 14px;
}

.stat-arrow {
  margin-left: auto;
  padding-left: var(--spacing-2);
}

.arrow-icon {
  width: 16px;
  height: 16px;
  color: var(--text-muted);
  transition: transform var(--transition-fast);
}

.stat-card.clickable:hover .arrow-icon {
  transform: translateX(2px);
  color: var(--primary-color);
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

/* Mobile responsive */
@media (max-width: 640px) {
  .stat-card {
    padding: var(--spacing-4);
    gap: var(--spacing-3);
  }
  
  .stat-icon-wrapper {
    width: 40px;
    height: 40px;
  }
  
  .stat-icon {
    width: 20px;
    height: 20px;
  }
  
  .stat-label {
    font-size: var(--text-xs);
  }
  
  .stat-value {
    font-size: var(--text-xl);
  }
  
  .stat-subtext,
  .stat-trend {
    font-size: var(--text-2xs);
  }
}
</style>

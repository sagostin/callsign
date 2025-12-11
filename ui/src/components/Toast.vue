<template>
  <Teleport to="body">
    <div class="toast-container">
      <TransitionGroup name="toast">
        <div
          v-for="toast in toasts"
          :key="toast.id"
          :class="['toast', `toast-${toast.type}`]"
        >
          <div class="toast-icon">
            <span v-if="toast.type === 'success'">✓</span>
            <span v-else-if="toast.type === 'error'">✕</span>
            <span v-else-if="toast.type === 'warning'">⚠</span>
            <span v-else>ℹ</span>
          </div>
          <div class="toast-content">
            <div v-if="toast.title" class="toast-title">{{ toast.title }}</div>
            <div class="toast-message">{{ toast.message }}</div>
          </div>
          <button class="toast-close" @click="remove(toast.id)">×</button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup>
import { ref } from 'vue'

const toasts = ref([])
let toastId = 0

function add(options) {
  const id = ++toastId
  const toast = {
    id,
    type: options.type || 'info',
    title: options.title,
    message: options.message,
    duration: options.duration ?? 5000,
  }
  
  toasts.value.push(toast)
  
  if (toast.duration > 0) {
    setTimeout(() => remove(id), toast.duration)
  }
  
  return id
}

function remove(id) {
  const index = toasts.value.findIndex(t => t.id === id)
  if (index > -1) {
    toasts.value.splice(index, 1)
  }
}

function success(message, title = null) {
  return add({ type: 'success', message, title })
}

function error(message, title = 'Error') {
  return add({ type: 'error', message, title, duration: 8000 })
}

function warning(message, title = null) {
  return add({ type: 'warning', message, title })
}

function info(message, title = null) {
  return add({ type: 'info', message, title })
}

// Expose for use in other components
defineExpose({ add, remove, success, error, warning, info })
</script>

<style scoped>
.toast-container {
  position: fixed;
  top: 1rem;
  right: 1rem;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-width: 400px;
}

.toast {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 1rem;
  border-radius: 0.5rem;
  background: var(--bg-elevated, #1f2937);
  border: 1px solid var(--border-color, #374151);
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.3);
  color: var(--text-primary, #fff);
}

.toast-success {
  border-color: #10b981;
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.1), transparent);
}

.toast-error {
  border-color: #ef4444;
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.1), transparent);
}

.toast-warning {
  border-color: #f59e0b;
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.1), transparent);
}

.toast-info {
  border-color: #3b82f6;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.1), transparent);
}

.toast-icon {
  flex-shrink: 0;
  width: 1.5rem;
  height: 1.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  font-size: 0.875rem;
}

.toast-success .toast-icon { background: #10b981; }
.toast-error .toast-icon { background: #ef4444; }
.toast-warning .toast-icon { background: #f59e0b; }
.toast-info .toast-icon { background: #3b82f6; }

.toast-content {
  flex: 1;
  min-width: 0;
}

.toast-title {
  font-weight: 600;
  margin-bottom: 0.25rem;
}

.toast-message {
  font-size: 0.875rem;
  opacity: 0.9;
}

.toast-close {
  flex-shrink: 0;
  background: none;
  border: none;
  color: inherit;
  opacity: 0.5;
  cursor: pointer;
  font-size: 1.25rem;
  line-height: 1;
  padding: 0;
}

.toast-close:hover {
  opacity: 1;
}

/* Transitions */
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from {
  opacity: 0;
  transform: translateX(100%);
}

.toast-leave-to {
  opacity: 0;
  transform: translateX(100%);
}
</style>

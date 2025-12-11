<script setup>
import { provide, ref } from 'vue'
import Toast from '@/components/Toast.vue'

// Toast ref for global access
const toastRef = ref(null)

// Provide toast globally
provide('toast', {
  success: (msg, title) => toastRef.value?.success(msg, title),
  error: (msg, title) => toastRef.value?.error(msg, title),
  warning: (msg, title) => toastRef.value?.warning(msg, title),
  info: (msg, title) => toastRef.value?.info(msg, title),
})
</script>

<template>
  <router-view />
  <Toast ref="toastRef" />
</template>

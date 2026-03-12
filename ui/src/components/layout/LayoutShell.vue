<template>
  <div 
    class="app-layout" 
    :class="{ 
      'mobile-nav-open': isMobileOpen,
      'sidebar-collapsed': isCollapsed && !isMobile
    }"
  >
    <!-- Mobile Menu Toggle -->
    <button 
      class="mobile-toggle" 
      @click="isMobileOpen = !isMobileOpen"
      aria-label="Toggle navigation menu"
      :aria-expanded="isMobileOpen"
    >
      <MenuIcon v-if="!isMobileOpen" :size="20" />
      <XIcon v-else :size="20" />
    </button>

    <!-- Sidebar -->
    <aside class="layout-sidebar" role="navigation" aria-label="Main navigation">
      <Sidebar 
        :collapsed="isCollapsed && !isMobile" 
        @navigated="handleNavigated"
        @toggle-collapse="isCollapsed = !isCollapsed"
      />
    </aside>

    <!-- Top Bar -->
    <header class="layout-topbar" role="banner">
      <TopBar />
    </header>

    <!-- Mobile Overlay -->
    <div 
      v-if="isMobileOpen" 
      class="mobile-overlay" 
      @click="isMobileOpen = false"
      aria-hidden="true"
    ></div>

    <!-- Main Content -->
    <main class="layout-content" role="main">
      <router-view v-slot="{ Component }">
        <transition name="page" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import Sidebar from './Sidebar.vue'
import TopBar from './TopBar.vue'
import { Menu as MenuIcon, X as XIcon } from 'lucide-vue-next'

// Mobile menu state
const isMobileOpen = ref(false)
const isCollapsed = ref(false)
const isMobile = ref(false)

// Check if viewport is mobile
const checkMobile = () => {
  isMobile.value = window.innerWidth <= 768
  if (!isMobile.value) {
    isMobileOpen.value = false
  }
}

// Handle navigation - close mobile menu
const handleNavigated = () => {
  if (isMobile.value) {
    isMobileOpen.value = false
  }
}

// Handle resize
let resizeObserver = null

onMounted(() => {
  checkMobile()
  
  // Use ResizeObserver if available, otherwise fallback to resize event
  if (window.ResizeObserver) {
    resizeObserver = new ResizeObserver(() => {
      checkMobile()
    })
    resizeObserver.observe(document.documentElement)
  } else {
    window.addEventListener('resize', checkMobile)
  }

  // Close mobile menu on escape key
  const handleEscape = (e) => {
    if (e.key === 'Escape' && isMobileOpen.value) {
      isMobileOpen.value = false
    }
  }
  document.addEventListener('keydown', handleEscape)

  // Restore cleanup
  onUnmounted(() => {
    if (resizeObserver) {
      resizeObserver.disconnect()
    } else {
      window.removeEventListener('resize', checkMobile)
    }
    document.removeEventListener('keydown', handleEscape)
  })
})
</script>

<style scoped>
/* Page transition animations */
.page-enter-active,
.page-leave-active {
  transition: opacity var(--transition-fast), transform var(--transition-fast);
}

.page-enter-from {
  opacity: 0;
  transform: translateY(4px);
}

.page-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* Ensure smooth scrolling */
.layout-content {
  scroll-behavior: smooth;
}

/* Focus visible styles for accessibility */
.mobile-toggle:focus-visible {
  outline: 2px solid var(--primary-color);
  outline-offset: 2px;
}
</style>

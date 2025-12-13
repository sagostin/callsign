<template>
  <div class="app-layout" :class="{ 'mobile-nav-open': isMobileOpen }">
    <div class="mobile-toggle" @click="isMobileOpen = !isMobileOpen">
       <MenuIcon v-if="!isMobileOpen" />
       <XIcon v-else />
    </div>

    <div class="layout-sidebar">
      <Sidebar @navigated="isMobileOpen = false" />
    </div>
    
    <div class="layout-topbar">
      <TopBar />
    </div>
    
    <div class="mobile-overlay" v-if="isMobileOpen" @click="isMobileOpen = false"></div>

    <main class="layout-content">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import Sidebar from './Sidebar.vue'
import TopBar from './TopBar.vue'
import { Menu as MenuIcon, X as XIcon } from 'lucide-vue-next'

const isMobileOpen = ref(false)
</script>

<style scoped>
.mobile-toggle {
  display: none;
  position: fixed;
  top: 12px;
  left: 16px;
  z-index: 100;
  background: white;
  padding: 8px;
  border-radius: 4px;
  box-shadow: 0 2px 5px rgba(0,0,0,0.1);
  cursor: pointer;
}

@media (max-width: 768px) {
  .mobile-toggle { display: block; }
  
  .layout-sidebar {
    position: fixed;
    top: 0; left: 0; bottom: 0;
    z-index: 90;
    transform: translateX(-100%);
    transition: transform 0.3s ease;
    width: 280px;
    background: white;
    border-right: 1px solid var(--border-color);
  }
  
  .mobile-nav-open .layout-sidebar { transform: translateX(0); }
  
  .mobile-overlay {
    position: fixed; top: 0; left: 0; right: 0; bottom: 0;
    background: rgba(0,0,0,0.5); z-index: 80;
  }
  
  .layout-topbar {
    margin-left: 0 !important;
    padding-left: 52px; /* Space for mobile toggle */
  }
  
  .layout-content {
    margin-left: 0 !important;
    padding: 16px;
    padding-top: 12px;
  }
}

/* Tablet breakpoint */
@media (min-width: 769px) and (max-width: 1024px) {
  .layout-sidebar {
    width: 200px;
  }
  
  .layout-topbar,
  .layout-content {
    margin-left: 200px;
  }
}

/* Small mobile */
@media (max-width: 480px) {
  .mobile-toggle {
    top: 8px;
    left: 8px;
    padding: 6px;
  }
  
  .layout-sidebar {
    width: 100%;
  }
  
  .layout-topbar {
    padding-left: 48px;
  }
  
  .layout-content {
    padding: 12px;
  }
}
</style>

<template>
  <div class="h-full flex flex-col bg-slate-50 dark:bg-slate-900">
    <div class="bg-white dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700 p-6 flex items-center justify-between shadow-sm">
      <div>
        <h1 class="text-2xl font-bold text-slate-800 dark:text-white">Conference Profiles</h1>
        <p class="text-slate-500 dark:text-slate-400 mt-1">Manage system-wide conference settings and codec profiles.</p>
      </div>
      <button class="btn btn-primary">+ New Profile</button>
    </div>

    <div class="p-6">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div v-for="profile in profiles" :key="profile.id" class="card bg-white dark:bg-slate-800 p-5 rounded-xl border border-slate-200 dark:border-slate-700 hover:shadow-md transition-shadow">
          <div class="flex justify-between items-start mb-4">
             <div>
               <h3 class="font-bold text-lg text-slate-800 dark:text-white">{{ profile.name }}</h3>
               <span class="text-xs bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 px-2 py-1 rounded mt-1 inline-block">{{ profile.param_count }} Parameters</span>
             </div>
             <div class="flex gap-2">
               <button class="p-1.5 hover:bg-slate-100 dark:hover:bg-slate-700 rounded text-slate-500">
                  <EditIcon class="w-4 h-4" />
               </button>
             </div>
          </div>
          
          <div class="space-y-2 mb-4">
            <div class="flex justify-between text-sm">
              <span class="text-slate-500">Rate</span>
              <span class="text-slate-800 dark:text-slate-200 font-mono">{{ profile.rate }}</span>
            </div>
             <div class="flex justify-between text-sm">
              <span class="text-slate-500">Interval</span>
              <span class="text-slate-800 dark:text-slate-200 font-mono">{{ profile.interval }}</span>
            </div>
             <div class="flex justify-between text-sm">
              <span class="text-slate-500">Energy Level</span>
              <span class="text-slate-800 dark:text-slate-200 font-mono">{{ profile.energy }}</span>
            </div>
          </div>

          <div class="pt-4 border-t border-slate-100 dark:border-slate-700">
             <button class="text-sm text-indigo-600 dark:text-indigo-400 font-medium hover:underline">View XML Config</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { EditIcon } from 'lucide-vue-next';

const profiles = ref([
  { id: 1, name: 'default', rate: '8000', interval: '20', energy: '300', param_count: 12 },
  { id: 2, name: 'wideband', rate: '16000', interval: '20', energy: '300', param_count: 12 },
  { id: 3, name: 'ultrawideband', rate: '32000', interval: '20', energy: '300', param_count: 12 },
  { id: 4, name: 'cdquality', rate: '48000', interval: '10', energy: '300', param_count: 14 },
]);
</script>

<style scoped>
.btn {
  @apply px-4 py-2 rounded-lg font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2;
}
.btn-primary {
  @apply bg-indigo-600 text-white hover:bg-indigo-700 focus:ring-indigo-500;
}
</style>

<template>
  <div class="h-full flex flex-col bg-slate-50 dark:bg-slate-900">
    <div class="bg-white dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700 p-6 shadow-sm">
      <h1 class="text-2xl font-bold text-slate-800 dark:text-white">Permissions Manager</h1>
      <p class="text-slate-500 dark:text-slate-400 mt-1">Configure role-based access control and feature limits.</p>
    </div>

    <div class="flex-1 overflow-auto p-6">
      <div class="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 overflow-hidden">
        
        <!-- Toolbar -->
        <div class="p-4 border-b border-slate-200 dark:border-slate-700 flex justify-between items-center bg-slate-50/50 dark:bg-slate-900/50">
          <div class="flex items-center gap-4">
             <div class="flex items-center gap-2">
               <span class="text-sm font-medium text-slate-600 dark:text-slate-400">Select Role:</span>
               <select class="p-2 border border-slate-300 dark:border-slate-600 rounded-lg bg-white dark:bg-slate-800 text-sm">
                 <option>Super Admin</option>
                 <option>Tenant Admin</option>
                 <option>User</option>
                 <option>Agent</option>
               </select>
             </div>
          </div>
          <div class="relative">
            <input type="text" placeholder="Search permissions..." class="pl-9 pr-4 py-2 border border-slate-300 dark:border-slate-600 rounded-lg text-sm w-64">
            <SearchIcon class="absolute left-3 top-2.5 w-4 h-4 text-slate-400" />
          </div>
        </div>

        <!-- Permissions Table -->
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="bg-slate-50 dark:bg-slate-800 text-slate-500 dark:text-slate-400 text-xs uppercase font-semibold">
              <th class="p-4 w-1/3">Feature Category</th>
              <th class="p-4">Permission</th>
              <th class="p-4 text-center w-24">Create</th>
              <th class="p-4 text-center w-24">Read</th>
              <th class="p-4 text-center w-24">Update</th>
              <th class="p-4 text-center w-24">Delete</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
            <template v-for="(category, catIndex) in permissions" :key="catIndex">
              <tr class="bg-slate-50/80 dark:bg-slate-800/80">
                <td colspan="6" class="p-3 font-bold text-slate-700 dark:text-slate-300 pl-4 border-y border-slate-200 dark:border-slate-700">
                  {{ category.name }}
                </td>
              </tr>
              <tr v-for="(perm, pIndex) in category.items" :key="pIndex" class="hover:bg-slate-50 dark:hover:bg-slate-700/50">
                <td class="p-4 text-sm text-slate-500 pl-8"><!-- spacer --></td>
                <td class="p-4 text-sm font-medium text-slate-700 dark:text-slate-300">
                  {{ perm.label }}
                  <p class="text-xs text-slate-400 font-normal mt-0.5">{{ perm.description }}</p>
                </td>
                <td class="p-4 text-center"><input type="checkbox" v-model="perm.create" class="rounded text-indigo-600 w-4 h-4 border-slate-300" /></td>
                <td class="p-4 text-center"><input type="checkbox" v-model="perm.read" class="rounded text-indigo-600 w-4 h-4 border-slate-300" /></td>
                <td class="p-4 text-center"><input type="checkbox" v-model="perm.update" class="rounded text-indigo-600 w-4 h-4 border-slate-300" /></td>
                <td class="p-4 text-center"><input type="checkbox" v-model="perm.delete" class="rounded text-indigo-600 w-4 h-4 border-slate-300" /></td>
              </tr>
            </template>
          </tbody>
        </table>

        <!-- Footer -->
        <div class="p-4 border-t border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-900 flex justify-end">
          <button class="btn btn-primary">Save Permissions</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { SearchIcon } from 'lucide-vue-next';

const permissions = ref([
  {
    name: 'Applications',
    items: [
      { label: 'Conferences', description: 'Manage conference rooms', create: true, read: true, update: true, delete: false },
      { label: 'Video Calls', description: 'Initiate and manage video sessions', create: true, read: true, update: false, delete: false },
      { label: 'Call Center', description: 'Access queues and agent controls', create: false, read: true, update: false, delete: false },
    ]
  },
  {
    name: 'System',
    items: [
      { label: 'Extensions', description: 'Manage user extensions', create: false, read: true, update: true, delete: false },
      { label: 'Voicemail', description: 'Access and manage voicemail boxes', create: true, read: true, update: true, delete: true },
    ]
  }
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

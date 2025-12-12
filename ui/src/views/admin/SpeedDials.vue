<template>
  <div class="h-full flex flex-col bg-slate-50 dark:bg-slate-900">
    <!-- Header -->
    <div class="bg-white dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700 p-6 flex items-center justify-between shadow-sm z-10">
      <div>
        <h1 class="text-2xl font-bold text-slate-800 dark:text-white">Speed Dials</h1>
        <p class="text-slate-500 dark:text-slate-400 mt-1">Manage speed dial prefixes and global contact lists.</p>
      </div>
      <button @click="showCreateModal = true" class="btn btn-primary flex items-center gap-2">
        <PlusIcon class="w-5 h-5" />
        New Speed Dial Group
      </button>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-6">
      <div v-if="speedDialGroups.length === 0" class="flex flex-col items-center justify-center h-64 text-center">
        <div class="bg-slate-100 dark:bg-slate-800 p-4 rounded-full mb-4">
          <ZapIcon class="w-8 h-8 text-slate-400" />
        </div>
        <h3 class="text-lg font-medium text-slate-700 dark:text-slate-300">No Speed Dials Defined</h3>
        <p class="text-slate-500 dark:text-slate-400 max-w-sm mt-2">
          Create a speed dial group to assign short codes (e.g. *01) to frequently called numbers.
        </p>
        <button @click="showCreateModal = true" class="btn btn-outline mt-6">
          Create First Group
        </button>
      </div>

      <div v-else class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
        <div v-for="group in speedDialGroups" :key="group.id" class="card border border-slate-200 dark:border-slate-700 flex flex-col">
          <div class="p-5 border-b border-slate-100 dark:border-slate-700 flex justify-between items-start">
             <div>
                <h3 class="font-bold text-slate-800 dark:text-white text-lg">{{ group.name }}</h3>
                <div class="flex items-center gap-2 mt-1">
                  <span class="text-xs font-mono bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400 px-2 py-0.5 rounded">Prefix: {{ group.prefix }}</span>
                  <span class="text-xs text-slate-500">{{ group.entries.length }} entries</span>
                </div>
             </div>
             <button @click="editGroup(group)" class="text-slate-400 hover:text-indigo-600 transition-colors">
               <EditIcon class="w-5 h-5" />
             </button>
          </div>
          
          <div class="flex-1 bg-slate-50/50 dark:bg-slate-800/50 p-2 max-h-60 overflow-y-auto">
            <template v-if="group.entries.length > 0">
               <div v-for="entry in group.entries" :key="entry.code" class="flex items-center justify-between p-2 rounded hover:bg-white dark:hover:bg-slate-700/50 transition-colors">
                 <div class="flex items-center gap-3">
                   <span class="font-mono text-xs font-bold text-slate-500 w-8">{{ group.prefix }}{{ entry.code }}</span>
                   <span class="text-sm font-medium text-slate-700 dark:text-slate-300">{{ entry.label }}</span>
                 </div>
                 <div class="flex items-center gap-2 text-sm text-slate-500">
                    <PhoneIcon class="w-3 h-3" />
                    <span>{{ entry.destination }}</span>
                 </div>
               </div>
            </template>
            <div v-else class="text-center py-6 text-sm text-slate-400 italic">
              No numbers added yet.
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
      <div class="bg-white dark:bg-slate-800 rounded-xl shadow-2xl w-full max-w-2xl max-h-[90vh] flex flex-col overflow-hidden">
        <SpeedDialForm @close="showCreateModal = false" @save="saveGroup" :initial-data="selectedGroup" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { PlusIcon, ZapIcon, EditIcon, PhoneIcon } from 'lucide-vue-next';
import SpeedDialForm from './admin/SpeedDialForm.vue';

const showCreateModal = ref(false);
const selectedGroup = ref(null);

const speedDialGroups = ref([
  { 
    id: 1, 
    name: 'Executive Directory', 
    prefix: '*0', 
    entries: [
      { code: '1', label: 'CEO Mobile', destination: '15550001' },
      { code: '2', label: 'CTO Mobile', destination: '15550002' },
    ]
  },
  { 
    id: 2, 
    name: 'Vendor Support', 
    prefix: '*9', 
    entries: [
      { code: '1', label: 'IT Helpdesk', destination: '18005551234' },
      { code: '5', label: 'Building Security', destination: '15559990000' },
    ]
  },
]);

const editGroup = (group) => {
  selectedGroup.value = group;
  showCreateModal.value = true;
};

const saveGroup = (groupData) => {
  if (selectedGroup.value) {
    // Update existing
    const index = speedDialGroups.value.findIndex(g => g.id === selectedGroup.value.id);
    if (index !== -1) speedDialGroups.value[index] = { ...groupData, id: selectedGroup.value.id };
  } else {
    // Create new
    speedDialGroups.value.push({ ...groupData, id: Date.now() });
  }
  showCreateModal.value = false;
  selectedGroup.value = null;
};
</script>

<style scoped>
.btn {
  @apply px-4 py-2 rounded-lg font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2;
}
.btn-primary {
  @apply bg-indigo-600 text-white hover:bg-indigo-700 focus:ring-indigo-500;
}
.btn-outline {
  @apply border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-700;
}
.card {
  @apply bg-white dark:bg-slate-800 rounded-xl overflow-hidden shadow-sm hover:shadow-md transition-shadow;
}
</style>

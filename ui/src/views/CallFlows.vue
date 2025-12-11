<template>
  <div class="h-full flex flex-col bg-slate-50 dark:bg-slate-900">
    <!-- Header -->
    <div class="bg-white dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700 p-6 flex items-center justify-between shadow-sm z-10">
      <div>
        <h1 class="text-2xl font-bold text-slate-800 dark:text-white">Call Flows</h1>
        <p class="text-slate-500 dark:text-slate-400 mt-1">Manage call flow toggles and routing logic.</p>
      </div>
      <button @click="showCreateModal = true" class="btn btn-primary flex items-center gap-2">
        <PlusIcon class="w-5 h-5" />
        New Call Flow
      </button>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-6">
      <div v-if="callFlows.length === 0" class="flex flex-col items-center justify-center h-64 text-center">
        <div class="bg-slate-100 dark:bg-slate-800 p-4 rounded-full mb-4">
          <GitBranchIcon class="w-8 h-8 text-slate-400" />
        </div>
        <h3 class="text-lg font-medium text-slate-700 dark:text-slate-300">No Call Flows Defined</h3>
        <p class="text-slate-500 dark:text-slate-400 max-w-sm mt-2">
          Create a call flow to define how incoming calls are routed based on conditions like time of day or manual toggles.
        </p>
        <button @click="showCreateModal = true" class="btn btn-outline mt-6">
          Create First Flow
        </button>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div v-for="flow in callFlows" :key="flow.id" class="card hover:shadow-md transition-shadow cursor-pointer border border-slate-200 dark:border-slate-700" @click="editFlow(flow)">
          <div class="flex items-start justify-between mb-4">
            <div class="flex items-center gap-3">
              <div class="p-2 bg-indigo-50 dark:bg-indigo-900/30 rounded-lg">
                <GitBranchIcon class="w-6 h-6 text-indigo-600 dark:text-indigo-400" />
              </div>
              <div>
                <h3 class="font-semibold text-slate-800 dark:text-white">{{ flow.name }}</h3>
                <span class="text-xs font-mono bg-slate-100 dark:bg-slate-800 text-slate-500 px-2 py-0.5 rounded">*{{ flow.featureCode }}</span>
              </div>
            </div>
            <span :class="['px-2 py-1 rounded-full text-xs font-medium', flow.status === 'active' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400']">
              {{ flow.status }}
            </span>
          </div>
          
          <div class="space-y-2 text-sm text-slate-600 dark:text-slate-400">
            <div class="flex justify-between">
              <span>Extension</span>
              <span class="font-medium text-slate-800 dark:text-slate-200">{{ flow.extension }}</span>
            </div>
            <div class="flex justify-between">
              <span>Context</span>
              <span class="font-medium text-slate-800 dark:text-slate-200">{{ flow.context }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
      <div class="bg-white dark:bg-slate-800 rounded-xl shadow-2xl w-full max-w-4xl max-h-[90vh] flex flex-col overflow-hidden">
        <CallFlowForm @close="showCreateModal = false" @save="saveFlow" :initial-data="selectedFlow" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { PlusIcon, GitBranchIcon } from 'lucide-vue-next';
import CallFlowForm from './admin/CallFlowForm.vue';

const showCreateModal = ref(false);
const selectedFlow = ref(null);

const callFlows = ref([
  { id: 1, name: 'Main Office Day/Night', featureCode: '701', extension: '5001', context: 'public', status: 'active' },
  { id: 2, name: 'Support Queue Override', featureCode: '702', extension: '5002', context: 'public', status: 'inactive' },
]);

const editFlow = (flow) => {
  selectedFlow.value = flow;
  showCreateModal.value = true;
};

const saveFlow = (flowData) => {
  if (selectedFlow.value) {
    // Update existing
    const index = callFlows.value.findIndex(f => f.id === selectedFlow.value.id);
    if (index !== -1) callFlows.value[index] = { ...flowData, id: selectedFlow.value.id };
  } else {
    // Create new
    callFlows.value.push({ ...flowData, id: Date.now(), status: 'active' });
  }
  showCreateModal.value = false;
  selectedFlow.value = null;
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
  @apply bg-white dark:bg-slate-800 p-5 rounded-xl;
}
</style>

<template>
  <div class="flex flex-col h-full bg-slate-50 dark:bg-slate-900">
    <div class="px-6 py-4 border-b border-slate-200 dark:border-slate-700 flex items-center justify-between bg-white dark:bg-slate-800">
      <h2 class="text-xl font-bold text-slate-800 dark:text-white">{{ isEditing ? 'Edit Call Flow' : 'New Call Flow' }}</h2>
      <button @click="$emit('close')" class="text-slate-500 hover:text-slate-700 dark:hover:text-slate-300">
        <XIcon class="w-6 h-6" />
      </button>
    </div>

    <div class="flex-1 overflow-hidden flex">
      <!-- Sidebar Settings -->
      <div class="w-1/3 border-r border-slate-200 dark:border-slate-700 p-6 overflow-y-auto bg-white dark:bg-slate-800">
        <h3 class="text-lg font-semibold mb-4 text-slate-800 dark:text-white">Settings</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Name</label>
            <input v-model="form.name" type="text" class="input-field" placeholder="e.g. Day/Night Mode" />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Extension</label>
            <input v-model="form.extension" type="text" class="input-field" placeholder="e.g. 5001" />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Feature Code</label>
            <input v-model="form.featureCode" type="text" class="input-field" placeholder="e.g. *701" />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Context</label>
            <input v-model="form.context" type="text" class="input-field" placeholder="public" />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Description</label>
            <textarea v-model="form.description" rows="3" class="input-field"></textarea>
          </div>
        </div>
      </div>

      <!-- Visual Flow Editor -->
      <div class="flex-1 bg-slate-50 dark:bg-slate-900 flex relative overflow-hidden">
        <!-- Palette Sidebar -->
        <NodePalette />

        <!-- Canvas Area -->
        <div class="flex-1 relative flex flex-col">
          <div class="absolute top-4 left-4 z-10 bg-white/80 backdrop-blur p-2 rounded border border-slate-200 shadow-sm flex gap-2">
            <h3 class="text-xs font-bold uppercase text-slate-500 my-auto px-2">Flow Designer</h3>
             <button class="btn btn-xs btn-outline bg-white" @click="loadExample">
               Load Example AA
             </button>
             <button class="btn btn-xs btn-outline bg-white">
               Clear
             </button>
          </div>

          <FlowCanvas ref="flowCanvas" v-model="flowData" />
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="p-4 border-t border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 flex justify-end gap-3">
      <button @click="$emit('close')" class="btn btn-outline">Cancel</button>
      <button @click="save" class="btn btn-primary">Save Call Flow</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue';
import { XIcon } from 'lucide-vue-next';
import NodePalette from '../../components/flow/NodePalette.vue';
import FlowCanvas from '../../components/flow/FlowCanvas.vue';

const props = defineProps({
  initialData: Object,
});

const emit = defineEmits(['close', 'save']);

const form = ref({
  name: '',
  extension: '',
  featureCode: '',
  context: 'public',
  description: '',
});

const flowData = ref({
  nodes: [],
  connections: []
});

const flowCanvas = ref(null);

const isEditing = computed(() => !!props.initialData);

// Load Example Data (Auto Attendant)
const loadExample = () => {
   const exampleData = {
     nodes: [
       { id: '1', type: 'trigger', label: 'Inbound DID', x: 300, y: 50, config: {} },
       { id: '2', type: 'time', label: 'Time Switch', x: 300, y: 150, config: { condition: 'Business Hours' } },
       { id: '3', type: 'ivr', label: 'Main Menu (Open)', x: 150, y: 300, config: { menuName: 'Day Menu' } },
       { id: '4', type: 'voicemail', label: 'After Hours VM', x: 450, y: 300, config: {} }
     ],
     connections: [
       { id: 'c1', sourceId: '1', targetId: '2' },
       { id: 'c2', sourceId: '2', targetId: '3' }, // Match -> Open
       { id: 'c3', sourceId: '2', targetId: '4' }  // No Match -> Closed
     ]
   };
   flowData.value = exampleData;
   if(flowCanvas.value) flowCanvas.value.loadGraph(exampleData);
};

watch(() => props.initialData, (newVal) => {
  if (newVal) {
    form.value = { ...newVal };
    // In a real app, we'd also load the saved JSON flow data here
  }
}, { immediate: true });

const save = () => {
  // Combine form metadata with visual flow JSON
  const payload = {
    ...form.value,
    flow: flowData.value
  };
  emit('save', payload);
};
</script>

<style scoped>
.input-field {
  @apply w-full rounded-lg border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm px-3 py-2 border;
}
.btn {
  @apply px-4 py-2 rounded-lg font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2;
}
.btn-primary {
  @apply bg-indigo-600 text-white hover:bg-indigo-700 focus:ring-indigo-500;
}
.btn-outline {
  @apply border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-700;
}
.flow-node {
  @apply border-2 rounded-xl p-4 shadow-sm w-48 text-center relative;
}
.grid-bg {
  background-image: radial-gradient(circle, #6366f1 1px, transparent 1px);
  background-size: 20px 20px;
}
</style>

<template>
  <div class="flex flex-col h-full bg-slate-50 dark:bg-slate-900">
    <div class="px-6 py-4 border-b border-slate-200 dark:border-slate-700 flex items-center justify-between bg-white dark:bg-slate-800">
      <h2 class="text-xl font-bold text-slate-800 dark:text-white">{{ isEditing ? 'Edit Speed Dial Group' : 'New Speed Dial Group' }}</h2>
      <button @click="$emit('close')" class="text-slate-500 hover:text-slate-700 dark:hover:text-slate-300">
        <XIcon class="w-6 h-6" />
      </button>
    </div>

    <div class="flex-1 overflow-y-auto p-6">
      
      <!-- Group Settings -->
      <div class="bg-white dark:bg-slate-800 rounded-xl p-6 shadow-sm border border-slate-200 dark:border-slate-700 mb-6">
        <h3 class="text-lg font-semibold mb-4 text-slate-800 dark:text-white">Group Configuration</h3>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Group Name</label>
            <input v-model="form.name" type="text" class="input-field" placeholder="e.g. Common Vendors" />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Dial Prefix</label>
            <input v-model="form.prefix" type="text" class="input-field" placeholder="e.g. *0" />
            <p class="text-xs text-slate-500 mt-1">Users will dial [Prefix] + [Short Code] (e.g. *0 + 1)</p>
          </div>
        </div>
      </div>

      <!-- Entries -->
      <div class="bg-white dark:bg-slate-800 rounded-xl p-6 shadow-sm border border-slate-200 dark:border-slate-700">
        <div class="flex justify-between items-center mb-4">
          <h3 class="text-lg font-semibold text-slate-800 dark:text-white">Numbers List</h3>
          <button @click="addEntry" class="btn btn-sm btn-outline text-xs">
            + Add Entry
          </button>
        </div>

        <div class="space-y-3">
          <div v-for="(entry, index) in form.entries" :key="index" class="flex items-center gap-3 p-3 bg-slate-50 dark:bg-slate-900/50 rounded-lg border border-slate-200 dark:border-slate-700">
            <div class="w-20">
              <label class="block text-xs font-medium text-slate-500 mb-1">Code</label>
              <input v-model="entry.code" type="text" class="input-field text-center font-mono" placeholder="1" />
            </div>
            <div class="flex-1">
              <label class="block text-xs font-medium text-slate-500 mb-1">Label</label>
              <input v-model="entry.label" type="text" class="input-field" placeholder="e.g. Helpdesk" />
            </div>
            <div class="flex-1">
              <label class="block text-xs font-medium text-slate-500 mb-1">Destination Number</label>
              <input v-model="entry.destination" type="text" class="input-field" placeholder="1-800-..." />
            </div>
            <div class="pt-5">
              <button @click="removeEntry(index)" class="text-red-500 hover:text-red-700 p-2">
                <Trash2Icon class="w-4 h-4" />
              </button>
            </div>
          </div>
          
          <div v-if="form.entries.length === 0" class="text-center py-4 text-slate-400 text-sm">
             No entries yet. Click "Add Entry" to start.
          </div>
        </div>
      </div>
    
    </div>

    <!-- Footer -->
    <div class="p-4 border-t border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 flex justify-end gap-3">
      <button @click="$emit('close')" class="btn btn-outline">Cancel</button>
      <button @click="save" class="btn btn-primary">Save Changes</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue';
import { XIcon, Trash2Icon } from 'lucide-vue-next';

const props = defineProps({
  initialData: Object,
});

const emit = defineEmits(['close', 'save']);

const form = ref({
  name: '',
  prefix: '',
  entries: []
});

const isEditing = computed(() => !!props.initialData);

watch(() => props.initialData, (newVal) => {
  if (newVal) {
    // Deep copy to avoid mutating prop
    form.value = JSON.parse(JSON.stringify(newVal));
  } else {
    form.value = { name: '', prefix: '', entries: [{ code: '1', label: '', destination: '' }] };
  }
}, { immediate: true });

const addEntry = () => {
  form.value.entries.push({ code: '', label: '', destination: '' });
};

const removeEntry = (index) => {
  form.value.entries.splice(index, 1);
};

const save = () => {
  emit('save', form.value);
};
</script>

<style scoped>
.input-field {
  @apply w-full rounded-lg border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm px-3 py-2 border;
}
.btn {
  @apply px-4 py-2 rounded-lg font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2;
}
.btn-sm {
  @apply px-3 py-1.5 text-sm;
}
.btn-primary {
  @apply bg-indigo-600 text-white hover:bg-indigo-700 focus:ring-indigo-500;
}
.btn-outline {
  @apply border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-700;
}
</style>

<template>
  <div class="h-full flex flex-col bg-slate-50 dark:bg-slate-900">
    <div class="bg-white dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700 p-6 flex items-center justify-between shadow-sm">
      <div>
        <h1 class="text-2xl font-bold text-slate-800 dark:text-white">Conference Profiles</h1>
        <p class="text-slate-500 dark:text-slate-400 mt-1">Manage system-wide conference settings and codec profiles.</p>
      </div>
      <button class="btn btn-primary" @click="openNewProfileModal">+ New Profile</button>
    </div>

    <div class="p-6">
      <!-- Loading State -->
      <div v-if="isLoading" class="flex justify-center items-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 mb-6">
        <p class="text-red-600 dark:text-red-400">{{ error }}</p>
        <button class="text-sm text-red-600 dark:text-red-400 underline mt-2" @click="fetchProfiles">Retry</button>
      </div>

      <!-- Empty State -->
      <div v-else-if="profiles.length === 0" class="text-center py-12">
        <p class="text-slate-500 dark:text-slate-400">No conference profiles found.</p>
        <button class="btn btn-primary mt-4" @click="openNewProfileModal">+ New Profile</button>
      </div>

      <!-- Profiles Grid -->
      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div v-for="profile in profiles" :key="profile.id" class="card bg-white dark:bg-slate-800 p-5 rounded-xl border border-slate-200 dark:border-slate-700 hover:shadow-md transition-shadow">
          <div class="flex justify-between items-start mb-4">
             <div>
               <h3 class="font-bold text-lg text-slate-800 dark:text-white">{{ profile.name }}</h3>
               <span class="text-xs bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 px-2 py-1 rounded mt-1 inline-block">{{ profile.param_count }} Parameters</span>
             </div>
             <div class="flex gap-2">
               <button class="p-1.5 hover:bg-slate-100 dark:hover:bg-slate-700 rounded text-slate-500" @click="editProfile(profile)">
                   <EditIcon class="w-4 h-4" />
                </button>
                <button class="p-1.5 hover:bg-slate-100 dark:hover:bg-slate-700 rounded text-slate-500" @click="deleteProfile(profile)">
                   <TrashIcon class="w-4 h-4" />
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
             <button class="text-sm text-indigo-600 dark:text-indigo-400 font-medium hover:underline" @click="viewXmlConfig(profile)">View XML Config</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Profile Form Modal -->
    <div v-if="showNewProfileModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showNewProfileModal = false">
      <div class="bg-white dark:bg-slate-800 rounded-xl p-6 w-full max-w-md shadow-xl">
        <h3 class="text-lg font-bold text-slate-800 dark:text-white mb-4">
          {{ editingProfile?.id ? 'Edit Profile' : 'New Conference Profile' }}
        </h3>
        <form @submit.prevent="handleSaveProfile" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Profile Name</label>
            <input v-model="profileForm.name" required class="w-full px-3 py-2 border border-slate-300 dark:border-slate-600 rounded-lg dark:bg-slate-700 dark:text-white" />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Rate</label>
              <input v-model="profileForm.rate" class="w-full px-3 py-2 border border-slate-300 dark:border-slate-600 rounded-lg dark:bg-slate-700 dark:text-white" />
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Interval</label>
              <input v-model="profileForm.interval" class="w-full px-3 py-2 border border-slate-300 dark:border-slate-600 rounded-lg dark:bg-slate-700 dark:text-white" />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Energy Level</label>
            <input v-model="profileForm.energy" class="w-full px-3 py-2 border border-slate-300 dark:border-slate-600 rounded-lg dark:bg-slate-700 dark:text-white" />
          </div>
          <div class="flex gap-3 pt-4">
            <button type="button" class="flex-1 px-4 py-2 border border-slate-300 dark:border-slate-600 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700" @click="showNewProfileModal = false">Cancel</button>
            <button type="submit" class="flex-1 btn btn-primary">Save</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue';
import { EditIcon, TrashIcon } from 'lucide-vue-next';
import { conferenceProfilesAPI } from '@/services/api';

const profiles = ref([]);
const isLoading = ref(false);
const error = ref(null);
const showNewProfileModal = ref(false);
const editingProfile = ref(null);
const profileForm = reactive({
  name: '',
  rate: '8000',
  interval: '20',
  energy: '300',
});

// Fetch conference profiles from API
async function fetchProfiles() {
  isLoading.value = true;
  error.value = null;
  try {
    const response = await conferenceProfilesAPI.list();
    profiles.value = response.data || [];
  } catch (err) {
    error.value = err.message || 'Failed to load conference profiles';
    console.error('Failed to fetch conference profiles:', err);
  } finally {
    isLoading.value = false;
  }
}

// Edit profile handler
function editProfile(profile) {
  editingProfile.value = { ...profile };
  profileForm.name = profile.name;
  profileForm.rate = profile.rate;
  profileForm.interval = profile.interval;
  profileForm.energy = profile.energy;
  showNewProfileModal.value = true;
}

// View XML config handler
async function viewXmlConfig(profile) {
  try {
    const response = await conferenceProfilesAPI.getXmlConfig(profile.id);
    // Open XML config in a new window or modal
    const xmlWindow = window.open('', '_blank');
    xmlWindow.document.write(`<pre>${response.data}</pre>`);
  } catch (err) {
    console.error('Failed to fetch XML config:', err);
    alert('Failed to load XML configuration');
  }
}

// Handle save profile from modal form
async function handleSaveProfile() {
  try {
    const profileData = {
      name: profileForm.name,
      rate: profileForm.rate,
      interval: profileForm.interval,
      energy: profileForm.energy,
    };
    if (editingProfile.value?.id) {
      await conferenceProfilesAPI.update(editingProfile.value.id, profileData);
    } else {
      await conferenceProfilesAPI.create(profileData);
    }
    showNewProfileModal.value = false;
    editingProfile.value = null;
    resetProfileForm();
    await fetchProfiles();
  } catch (err) {
    console.error('Failed to save profile:', err);
    alert('Failed to save profile');
  }
}

// Reset profile form to defaults
function resetProfileForm() {
  profileForm.name = '';
  profileForm.rate = '8000';
  profileForm.interval = '20';
  profileForm.energy = '300';
}

// Open new profile modal
function openNewProfileModal() {
  editingProfile.value = null;
  resetProfileForm();
  showNewProfileModal.value = true;
}

// Delete profile handler
async function deleteProfile(profile) {
  if (!confirm(`Delete profile "${profile.name}"?`)) return;
  try {
    await conferenceProfilesAPI.delete(profile.id);
    await fetchProfiles();
  } catch (err) {
    console.error('Failed to delete profile:', err);
    alert('Failed to delete profile');
  }
}

// Initialize
onMounted(fetchProfiles);
</script>

<style scoped>
.btn {
  @apply px-4 py-2 rounded-lg font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2;
}
.btn-primary {
  @apply bg-indigo-600 text-white hover:bg-indigo-700 focus:ring-indigo-500;
}
</style>

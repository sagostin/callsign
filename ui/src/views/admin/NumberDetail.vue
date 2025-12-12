<template>
  <div class="view-header">
    <div class="header-left">
      <button class="back-link" @click="$router.push('/numbers')">← Back to Numbers</button>
      <h2>Edit Number: {{ $route.params.id }}</h2>
    </div>
    <div class="header-actions">
      <button class="btn-danger">Release Number</button>
    </div>
  </div>

  <div class="form-container">
    <div class="form-section">
      <h3>Routing Configuration</h3>
      <div class="form-group">
        <label>Destination Type</label>
        <select v-model="form.destinationType" class="input-field">
          <option value="extension">Extension</option>
          <option value="ivr">IVR Menu</option>
          <option value="queue">Queue</option>
          <option value="voicemail">Voicemail Box</option>
        </select>
      </div>

      <div class="form-group">
        <label>Destination Target</label>
        <select v-model="form.destinationTarget" class="input-field">
          <option value="101">101 - Alice Smith</option>
          <option value="102">102 - Bob Jones</option>
          <option value="8000">8000 - Main Menu</option>
        </select>
      </div>
      
       <div class="form-group">
        <label>Caller ID Prefix (Optional)</label>
        <input type="text" class="input-field" placeholder="e.g. Sales: ">
      </div>

      <div class="form-group checkbox-row">
        <div class="check-item">
           <input type="checkbox" id="sms"> <label for="sms" class="inline">Supports SMS</label>
        </div>
        <div class="check-item">
           <input type="checkbox" id="mms"> <label for="mms" class="inline">Supports MMS</label>
        </div>
      </div>
    </div>

    <div class="form-actions">
      <button class="btn-primary" @click="saveChanges">
        Save Changes
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const form = ref({
  destinationType: 'ivr',
  destinationTarget: '8000'
})

const saveChanges = () => {
  alert('Routing updated!')
  router.push('/numbers')
}
</script>

<style scoped>
/* Reusing Form Styles */
.header-left {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: var(--spacing-xl);
}
.view-header {
  display: flex;
  justify-content: space-between;
}
.back-link {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0;
  font-size: var(--text-xs);
  text-align: left;
}
.back-link:hover { text-decoration: underline; color: var(--primary-color); }

.form-container {
  max-width: 600px;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
}

.form-section h3 {
  font-size: var(--text-md);
  color: var(--text-primary);
  font-weight: 600;
  margin-bottom: var(--spacing-md);
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 8px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: var(--spacing-md);
}

label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
}

.input-field {
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
  color: var(--text-primary);
  outline: none;
  background: white;
  transition: border-color var(--transition-fast);
}
.input-field:focus { border-color: var(--primary-color); }

.btn-primary {
  background-color: var(--primary-color);
  color: white;
  border: none;
  padding: 12px 24px;
  border-radius: var(--radius-sm);
  font-weight: 600;
  cursor: pointer;
  width: 100%;
}

.btn-danger {
  background: white;
  color: var(--status-bad);
  border: 1px solid var(--border-color);
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
  cursor: pointer;
}

.checkbox-row {
  flex-direction: row; gap: 24px; margin-top: 8px;
}
.check-item { display: flex; align-items: center; gap: 8px; }
label.inline { text-transform: none; font-size: var(--text-sm); font-weight: 500; cursor: pointer; color: var(--text-primary); }

</style>

<template>
  <div class="view-header">
    <div class="header-left">
      <button class="back-link" @click="$router.push('/numbers')">← Back to Numbers</button>
      <h2>Add New Number</h2>
    </div>
  </div>

  <div class="form-container">
    <div class="form-section">
      <h3>1. Select Carrier & Location</h3>
      <div class="form-grid">
        <div class="form-group">
          <label>Provider / Carrier</label>
          <select v-model="form.carrier" class="input-field">
            <option value="" disabled>Select Carrier</option>
            <option value="bandwidth">Bandwidth</option>
            <option value="twilio">Twilio</option>
            <option value="telnyx">Telnyx</option>
          </select>
        </div>

        <div class="form-group">
          <label>Country</label>
          <select v-model="form.country" class="input-field">
            <option value="US">United States</option>
            <option value="CA">Canada</option>
            <option value="GB">United Kingdom</option>
          </select>
        </div>

        <div class="form-group">
          <label>Area Code / Prefix</label>
          <div class="input-group">
            <input type="text" v-model="form.prefix" class="input-field" placeholder="415">
            <button class="btn-secondary">Search</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Search Results Placeholder -->
    <div class="form-section" v-if="form.prefix">
      <h3>Available Numbers</h3>
      <div class="number-grid">
        <div 
          v-for="num in availableNumbers" 
          :key="num" 
          class="number-card"
          :class="{ selected: form.selectedNumber === num }"
          @click="form.selectedNumber = num"
        >
          {{ num }}
        </div>
      </div>
    </div>

    <div class="form-section" v-if="form.selectedNumber">
      <h3>2. Initial Routing</h3>
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
          <option value="" disabled>Select Target</option>
          <option value="101">101 - Alice Smith</option>
          <option value="102">102 - Bob Jones</option>
          <option value="8000">8000 - Main Menu</option>
        </select>
      </div>
    </div>

    <div class="form-actions">
      <button class="btn-primary large" :disabled="!isValid" @click="saveNumber">
        Purchase & Configure
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const form = ref({
  carrier: '',
  country: 'US',
  prefix: '',
  selectedNumber: null,
  destinationType: 'extension',
  destinationTarget: ''
})

const availableNumbers = ref([
  '(415) 555-0101',
  '(415) 555-0122',
  '(415) 555-0199',
  '(415) 555-0200',
  '(415) 555-0250',
])

const isValid = computed(() => {
  return form.value.selectedNumber && form.value.destinationTarget
})

const saveNumber = () => {
  alert(`Purchased ${form.value.selectedNumber} from ${form.value.carrier}!`)
  router.push('/numbers')
}
</script>

<style scoped>
.header-left {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: var(--spacing-xl);
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

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-md);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
  letter-spacing: 0.05em;
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

.input-field:focus {
  border-color: var(--primary-color);
}

.input-group {
  display: flex;
  gap: 8px;
}
.input-group .input-field { flex: 1; }

.number-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: var(--spacing-sm);
}

.number-card {
  padding: 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  text-align: center;
  font-size: var(--text-sm);
  cursor: pointer;
  background: white;
  transition: all var(--transition-fast);
}

.number-card:hover {
  border-color: var(--primary-color);
}

.number-card.selected {
  background-color: var(--primary-light);
  border-color: var(--primary-color);
  color: var(--primary-color);
  font-weight: 600;
}

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
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-secondary {
  background: white;
  border: 1px solid var(--border-color);
  padding: 0 16px;
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--text-main);
  cursor: pointer;
}
</style>

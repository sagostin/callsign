<template>
  <div class="form-container">
    <div class="form-header">
      <h2>{{ isNew ? 'New Fax Box' : 'Edit Fax Box' }}</h2>
      <button class="btn-secondary" @click="$router.back()">Cancel</button>
    </div>

    <div class="form-card">
      <div class="form-group">
        <label>Name</label>
        <input v-model="form.name" type="text" class="input-field" placeholder="e.g. Sales Fax">
      </div>

      <div class="form-row">
        <div class="form-group">
          <label>Extension</label>
          <input v-model="form.extension" type="text" class="input-field" placeholder="e.g. 50">
        </div>
        <div class="form-group">
          <label>Email Destination</label>
          <input v-model="form.email" type="email" class="input-field" placeholder="fax@company.com">
        </div>
      </div>

       <div class="form-group">
        <label>Associated DID (Number)</label>
        <select v-model="form.did" class="input-field">
            <option value="">Select a Number...</option>
            <option value="14155550100">(415) 555-0100</option>
            <option value="14155550101">(415) 555-0101</option>
        </select>
      </div>

       <div class="form-group">
        <label>Caller ID Name</label>
        <input v-model="form.cid_name" type="text" class="input-field" placeholder="Use Default">
      </div>

       <div class="form-group full-width">
        <label>Access Control</label>
        <div class="radio-group">
          <label class="radio-label">
            <input type="radio" v-model="form.access" value="all">
            All Extensions
          </label>
          <label class="radio-label">
            <input type="radio" v-model="form.access" value="selected">
            Select Extensions
          </label>
        </div>
        
        <div class="extension-selector" v-if="form.access === 'selected'">
           <label class="sub-label">Allowed Extensions</label>
           <select multiple v-model="form.allowed_exts" class="input-field multiple-select">
              <option value="101">101 - Alice Smith</option>
              <option value="102">102 - Bob Jones</option>
              <option value="103">103 - Support Lead</option>
           </select>
           <p class="text-xs text-muted">Hold Cmd/Ctrl to select multiple.</p>
        </div>
      </div>

      <div class="form-actions">
        <button class="btn-primary" @click="save">Save Fax Box</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const isNew = computed(() => !route.params.id)

const form = ref({
  name: '',
  extension: '',
  email: '',
  cid_name: '',
  access: 'all',
  allowed_exts: []
})

const save = () => {
  console.log('Saving fax box:', form.value)
  router.back()
}
</script>

<style scoped>
.form-container { max-width: 600px; margin: 0 auto; }
.form-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.form-card { background: white; padding: 24px; border-radius: var(--radius-md); border: 1px solid var(--border-color); }
.form-group { margin-bottom: 16px; display: flex; flex-direction: column; gap: 6px; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; outline: none; }
.input-field:focus { border-color: var(--primary-color); }

.btn-primary { background: var(--primary-color); color: white; border: none; padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-secondary { background: white; border: 1px solid var(--border-color); color: var(--text-main); padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.form-actions { margin-top: 24px; display: flex; justify-content: flex-end; }

/* Access Control Styles */
.radio-group { display: flex; gap: 16px; margin-bottom: 8px; }
.radio-label { display: flex; align-items: center; gap: 8px; font-size: 13px; font-weight: 500; text-transform: none; color: var(--text-main); cursor: pointer; }
.extension-selector { background: var(--bg-app); padding: 12px; border-radius: var(--radius-sm); border: 1px solid var(--border-color); }
.multiple-select { height: 100px; width: 100%; margin-top: 4px; }
.text-xs { font-size: 11px; }
.sub-label { font-size: 11px; font-weight: 600; }
</style>

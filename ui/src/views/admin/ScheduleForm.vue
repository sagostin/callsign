<template>
  <div class="form-container">
    <div class="form-header">
      <h2>{{ isNew ? 'New Schedule' : 'Edit Schedule' }}</h2>
      <button class="btn-secondary" @click="$router.back()">Cancel</button>
    </div>

    <div class="form-card">
      <div class="form-group">
        <label>Schedule Name</label>
        <input v-model="form.name" type="text" class="input-field" placeholder="e.g. Business Hours">
      </div>

       <div class="form-group">
        <label>Extension (Optional)</label>
        <input v-model="form.extension" type="text" class="input-field" placeholder="e.g. 5001">
        <span class="help-text">Directly route calls to this extension to apply this schedule logic.</span>
      </div>

       <div class="form-group">
        <label>Description</label>
        <textarea v-model="form.description" class="input-field" rows="2"></textarea>
      </div>

      <!-- HOLIDAY EXCEPTIONS -->
      <div class="section-title">Holiday Exceptions (Highest Priority)</div>
      <div class="schedule-box holiday-box">
          <div class="form-row">
             <div class="form-group">
                <label>Holiday List</label>
                <select v-model="form.holiday_list_id" class="input-field">
                   <option value="">-- None --</option>
                   <option v-for="list in holidayLists" :key="list.id" :value="list.id">
                     {{ list.name }}
                   </option>
                </select>
             </div>
             <div class="form-group" v-if="form.holiday_list_id">
                <label>Route Calls To (on Holiday)</label>
                <select v-model="form.holiday_dest" class="input-field">
                   <option value="">-- Select Destination --</option>
                   <optgroup label="Extensions">
                     <option v-for="ext in extensions" :key="ext.extension" :value="`extension:${ext.extension}`">
                       {{ ext.extension }} {{ ext.name ? `- ${ext.name}` : '' }}
                     </option>
                   </optgroup>
                   <optgroup label="Voicemail">
                     <option value="voicemail:general">General Voicemail</option>
                     <option value="voicemail:operator">Operator Voicemail</option>
                   </optgroup>
                </select>
             </div>
          </div>
          <p class="text-xs text-muted" style="margin-top: 8px">
             If today is a holiday in the selected list, the call will immediately route to the destination above, skipping time rules.
          </p>
      </div>

      <!-- TIME RULES -->
      <div class="section-title">
         <span>Time Conditions</span>
         <button class="btn-secondary small" @click="addTimeBlock">+ Add Time Block</button>
      </div>
      
      <div v-if="form.time_blocks.length === 0" class="empty-state">
         No time rules defined. (Always False)
      </div>

      <div v-else class="time-blocks">
          <div v-for="(block, index) in form.time_blocks" :key="index" class="time-card">
              <div class="card-header">
                 <span class="block-title">Condition Set #{{ index + 1 }}</span>
                 <button class="btn-icon-remove" @click="removeTimeBlock(index)">Remove</button>
              </div>
              
              <div class="card-body">
                  <div class="form-group checkbox-group">
                     <label>Days of Week</label>
                     <div class="dow-selector">
                        <div v-for="day in daysOfWeek" :key="day.val" 
                             class="dow-chip" 
                             :class="{ selected: block.dow.includes(day.val) }"
                             @click="toggleDow(block, day.val)">
                             {{ day.label }}
                        </div>
                     </div>
                  </div>
                  
                  <div class="form-row">
                      <div class="form-group">
                         <label>Start Time</label>
                         <input type="time" v-model="block.start" class="input-field">
                      </div>
                      <div class="form-group">
                         <label>End Time</label>
                         <input type="time" v-model="block.end" class="input-field">
                      </div>
                  </div>
              </div>
          </div>
      </div>

      <!-- ROUTING -->
      <div class="section-title">Routing Logic</div>
      <div class="routing-box">
         <div class="route-item match">
            <div class="icon-indicator">✅</div>
            <div class="route-content">
               <label>Route To When MATCH (Open / In Schedule)</label>
               <select v-model="form.match_dest" class="input-field">
                  <option value="">-- Select Destination --</option>
                  <optgroup label="Extensions">
                    <option v-for="ext in extensions" :key="ext.extension" :value="`extension:${ext.extension}`">
                      {{ ext.extension }} {{ ext.name ? `- ${ext.name}` : '' }}
                    </option>
                  </optgroup>
                  <optgroup label="IVR Menus">
                    <option v-for="ivr in ivrs" :key="ivr.id" :value="`ivr:${ivr.id}`">
                      {{ ivr.name }}
                    </option>
                  </optgroup>
                  <optgroup label="Ring Groups">
                    <option v-for="rg in ringGroups" :key="rg.id" :value="`ring_group:${rg.id}`">
                      {{ rg.name }}
                    </option>
                  </optgroup>
               </select>
            </div>
         </div>
         
         <div class="route-item nomatch">
            <div class="icon-indicator">❌</div>
            <div class="route-content">
               <label>Route To When NO MATCH (Closed / Outside Schedule)</label>
               <select v-model="form.nomatch_dest" class="input-field">
                  <option value="">-- Select Destination --</option>
                  <optgroup label="Extensions">
                    <option v-for="ext in extensions" :key="ext.extension" :value="`extension:${ext.extension}`">
                      {{ ext.extension }} {{ ext.name ? `- ${ext.name}` : '' }}
                    </option>
                  </optgroup>
                  <optgroup label="Voicemail">
                    <option value="voicemail:general">General Voicemail</option>
                    <option value="voicemail:operator">Operator Voicemail</option>
                  </optgroup>
                  <optgroup label="IVR Menus">
                    <option v-for="ivr in ivrs" :key="ivr.id" :value="`ivr:${ivr.id}`">
                      {{ ivr.name }}
                    </option>
                  </optgroup>
               </select>
            </div>
         </div>
      </div>

      <div class="form-actions">
        <button class="btn-primary" @click="save">Save Schedule</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { holidaysAPI, extensionsAPI, ivrAPI, ringGroupsAPI } from '@/services/api'

const route = useRoute()
const router = useRouter()
const isNew = computed(() => !route.params.id)

// Data for dropdowns
const holidayLists = ref([])
const extensions = ref([])
const ivrs = ref([])
const ringGroups = ref([])

const form = ref({
  name: '',
  extension: '',
  description: '',
  holiday_list_id: '',
  holiday_dest: '',
  time_blocks: [
     { dow: [1,2,3,4,5], start: '09:00', end: '17:00' }
  ],
  match_dest: '',
  nomatch_dest: ''
})

const daysOfWeek = [
   { label: 'Mon', val: 1 },
   { label: 'Tue', val: 2 },
   { label: 'Wed', val: 3 },
   { label: 'Thu', val: 4 },
   { label: 'Fri', val: 5 },
   { label: 'Sat', val: 6 },
   { label: 'Sun', val: 0 }
]

// Load data on mount
onMounted(async () => {
  try {
    const [holidayRes, extRes, ivrRes, rgRes] = await Promise.all([
      holidaysAPI.list().catch(() => ({ data: [] })),
      extensionsAPI.list().catch(() => ({ data: [] })),
      ivrAPI.list().catch(() => ({ data: [] })),
      ringGroupsAPI.list().catch(() => ({ data: [] }))
    ])
    
    holidayLists.value = holidayRes.data?.data || holidayRes.data || []
    extensions.value = extRes.data?.data || extRes.data || []
    ivrs.value = ivrRes.data?.data || ivrRes.data || []
    ringGroups.value = rgRes.data?.data || rgRes.data || []
  } catch (e) {
    console.error('Failed to load form data:', e)
  }
})

const addTimeBlock = () => {
   form.value.time_blocks.push({ dow: [], start: '08:00', end: '17:00' })
}

const removeTimeBlock = (index) => {
   form.value.time_blocks.splice(index, 1)
}

const toggleDow = (block, dayVal) => {
   if (block.dow.includes(dayVal)) {
      block.dow = block.dow.filter(d => d !== dayVal)
   } else {
      block.dow.push(dayVal)
   }
}

const save = () => {
  // Parse dest values (format: "type:value") for API
  const parseDestination = (dest) => {
    if (!dest) return { type: '', value: '' }
    const [type, value] = dest.split(':')
    return { type, value: value || '' }
  }
  
  const matchDest = parseDestination(form.value.match_dest)
  const nomatchDest = parseDestination(form.value.nomatch_dest)
  const holidayDest = parseDestination(form.value.holiday_dest)
  
  const payload = {
    ...form.value,
    match_dest_type: matchDest.type,
    match_dest_value: matchDest.value,
    nomatch_dest_type: nomatchDest.type,
    nomatch_dest_value: nomatchDest.value,
    holiday_dest_type: holidayDest.type,
    holiday_dest_value: holidayDest.value,
    holiday_list_id: form.value.holiday_list_id ? parseInt(form.value.holiday_list_id) : null
  }
  
  console.log('Saving schedule:', payload)
  router.back()
}
</script>

<style scoped>
.form-container { max-width: 700px; margin: 0 auto; }
.form-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.form-card { background: white; padding: 24px; border-radius: var(--radius-md); border: 1px solid var(--border-color); }
.form-group { margin-bottom: 16px; display: flex; flex-direction: column; gap: 6px; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; outline: none; }
.input-field:focus { border-color: var(--primary-color); }
.help-text { font-size: 11px; color: var(--text-muted); }

.section-title { 
   font-weight: 600; margin: 24px 0 12px; font-size: 14px; 
   border-bottom: 1px solid var(--border-color); padding-bottom: 8px;
   display: flex; justify-content: space-between; align-items: center;
}

.time-blocks { display: flex; flex-direction: column; gap: 16px; }
.schedule-box { background: var(--bg-app); padding: 16px; border-radius: var(--radius-sm); border: 1px solid var(--border-color); }
.holiday-box { background: #fff7ed; border-color: #fdba74; }

.input-group { display: flex; align-items: center; }
.prefix-icon { padding: 8px 12px; background: white; border: 1px solid var(--border-color); border-right: none; border-radius: 4px 0 0 4px; font-size: 14px; }
.input-group .input-field { border-radius: 0 4px 4px 0; flex: 1; }

.time-card { 
   background: var(--bg-app); border: 1px solid var(--border-color); 
   border-radius: var(--radius-sm); padding: 16px; 
}
.card-header { display: flex; justify-content: space-between; margin-bottom: 12px; }
.block-title { font-weight: 600; font-size: 12px; color: var(--text-muted); text-transform: uppercase; }
.btn-icon-remove { background: none; border: none; color: var(--status-bad); cursor: pointer; font-size: 11px; font-weight: 600; }

.dow-selector { display: flex; gap: 4px; flex-wrap: wrap; }
.dow-chip { 
   background: white; border: 1px solid var(--border-color); 
   padding: 6px 10px; border-radius: 4px; font-size: 12px; cursor: pointer; font-weight: 500;
   transition: all 0.2s;
}
.dow-chip:hover { border-color: var(--primary-color); }
.dow-chip.selected { background: var(--primary-color); color: white; border-color: var(--primary-color); }

.routing-box { display: flex; flex-direction: column; gap: 16px; margin-top: 16px; }
.route-item { display: flex; gap: 12px; align-items: flex-start; }
.route-content { flex: 1; display: flex; flex-direction: column; gap: 6px; }
.icon-indicator { font-size: 18px; padding-top: 24px; }

.btn-primary { background: var(--primary-color); color: white; border: none; padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-secondary { background: white; border: 1px solid var(--border-color); color: var(--text-main); padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.form-actions { margin-top: 32px; display: flex; justify-content: flex-end; }
</style>

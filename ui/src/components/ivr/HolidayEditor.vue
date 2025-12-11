<template>
  <div class="holiday-editor">
    <div class="editor-header">
      <div class="form-group">
        <label>List Name</label>
        <input type="text" class="input-field name-input" v-model="listName" placeholder="e.g. US Holidays 2025">
      </div>
      
      <div class="form-group">
        <label>Source Type</label>
        <div class="source-toggle">
          <button 
            :class="{ active: sourceType === 'manual' }" 
            @click="sourceType = 'manual'">
            Manual List
          </button>
          <button 
            :class="{ active: sourceType === 'url' }" 
            @click="sourceType = 'url'">
            External URL
          </button>
        </div>
      </div>
    </div>

    <div class="editor-body">
      <!-- MANUAL MODE -->
      <div v-if="sourceType === 'manual'" class="manual-mode">
        <div class="add-row">
            <input type="date" class="input-field date-input">
            <input type="text" class="input-field desc-input" placeholder="Holiday Name (e.g. Thanksgiving)">
            <button class="btn-secondary small">Add</button>
        </div>
        
        <div class="holiday-list">
          <div class="holiday-item" v-for="(h, i) in holidays" :key="i">
            <span class="date">{{ h.date }}</span>
            <span class="name">{{ h.name }}</span>
            <button class="btn-icon-del">×</button>
          </div>
        </div>
      </div>

      <!-- URL MODE -->
      <div v-else class="url-mode">
        <div class="url-config">
          <p class="text-sm text-muted">Import holidays from an external ICS or JSON source. The system will sync daily.</p>
          <div class="url-input-group">
            <Globe class="icon-sm" />
            <input type="text" class="input-field full-width" placeholder="https://calendar.google.com/calendar/ical/..." value="https://example.com/holidays.ics">
          </div>
          <button class="btn-secondary">Test Sync</button>
        </div>
      </div>
    </div>
    
    <div class="editor-footer">
       <button class="btn-primary full-width">Save Holiday List</button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Globe } from 'lucide-vue-next'

const listName = ref('General Holidays')
const sourceType = ref('manual')

const holidays = ref([
  { date: '2025-12-25', name: 'Christmas Day' },
  { date: '2025-01-01', name: "New Year's Day" },
  { date: '2025-07-04', name: 'Independence Day' }
])
</script>

<style scoped>
.holiday-editor {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
  max-width: 500px;
}

.editor-header {
  padding: 16px;
  background: var(--bg-app);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.form-group { display: flex; flex-direction: column; gap: 4px; }
label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }

.input-field {
  padding: 8px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: 13px;
  outline: none;
}
.input-field:focus { border-color: var(--primary-color); }
.name-input { font-weight: 600; }

.source-toggle {
  display: flex;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  padding: 2px;
  width: fit-content;
}

.source-toggle button {
  background: transparent;
  border: none;
  padding: 4px 12px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  border-radius: 4px;
  color: var(--text-muted);
}
.source-toggle button.active {
  background: var(--bg-app);
  color: var(--primary-color);
  font-weight: 600;
  box-shadow: 0 1px 2px rgba(0,0,0,0.05);
}

.editor-body { padding: 16px; min-height: 200px; }

/* Manual Mode */
.add-row { display: flex; gap: 8px; margin-bottom: 16px; }
.date-input { width: 130px; }
.desc-input { flex: 1; }

.holiday-list { display: flex; flex-direction: column; gap: 8px; }
.holiday-item {
  display: flex; align-items: center; gap: 12px;
  padding: 8px; border: 1px solid var(--border-color); border-radius: var(--radius-sm);
}
.holiday-item .date { font-family: monospace; font-size: 12px; color: var(--text-muted); background: var(--bg-app); padding: 2px 6px; border-radius: 4px; }
.holiday-item .name { flex: 1; font-size: 13px; font-weight: 500; }
.btn-icon-del { background: none; border: none; color: var(--status-bad); cursor: pointer; font-size: 16px; padding: 0 4px; }

/* URL Mode */
.url-config { display: flex; flex-direction: column; gap: 12px; }
.url-input-group { display: flex; align-items: center; gap: 8px; border: 1px solid var(--border-color); padding: 8px; border-radius: var(--radius-sm); }
.input-field.full-width { border: none; padding: 0; width: 100%; }

.btn-secondary {
  background: white; border: 1px solid var(--border-color); padding: 6px 12px; border-radius: var(--radius-sm); font-size: 12px; cursor: pointer;
}
.btn-secondary.small { padding: 4px 10px; }

.editor-footer { padding: 16px; border-top: 1px solid var(--border-color); background: var(--bg-app); }
.btn-primary { 
  background: var(--primary-color); color: white; border: none; padding: 10px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; width: 100%;
}

.icon-sm { width: 16px; height: 16px; color: var(--text-muted); }

.divider { height: 1px; background: var(--border-color); margin: 24px 0; }
.section-title { font-size: 13px; font-weight: 700; color: var(--text-primary); margin-bottom: 12px; }
.logic-group { display: flex; flex-direction: column; gap: 8px; }
.action-row { display: flex; align-items: center; gap: 8px; }
.step-num { font-size: 11px; font-weight: 700; color: var(--text-muted); width: 16px; }
.then-text { font-size: 13px; font-weight: 500; color: var(--text-main); }
.flex-1 { flex: 1; }
.mt-4 { margin-top: 16px; }
</style>

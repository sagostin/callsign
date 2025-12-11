<template>
  <div class="view-container">
    <div class="page-header">
      <h2>Voicemail</h2>
      <div class="filters">
        <button class="filter-btn active">New (2)</button>
        <button class="filter-btn">Saved</button>
      </div>
    </div>

    <div class="vm-list">
      <div class="vm-item unread">
        <div class="vm-icon">
          <Voicemail class="icon" />
        </div>
        <div class="vm-details">
           <div class="vm-row">
             <span class="caller">(555) 123-9999</span>
             <span class="bad-tag">Unread</span>
           </div>
           <div class="vm-meta">Oct 24, 10:45 AM • 0:45s</div>
        </div>
        <div class="vm-actions">
           <button class="btn-icon circle"><Play class="icon-sm" /></button>
           <button class="btn-secondary small" @click="showTranscript = true">Transcribe</button>
           <button class="btn-icon"><Trash class="icon-sm" /></button>
        </div>
      </div>
      
      <div class="vm-item">
        <div class="vm-icon read">
          <Voicemail class="icon" />
        </div>
        <div class="vm-details">
           <div class="vm-row">
             <span class="caller">Bob Jones (102)</span>
           </div>
           <div class="vm-meta">Oct 23, 2:30 PM • 1:20s</div>
        </div>
         <div class="vm-actions">
           <button class="btn-icon circle"><Play class="icon-sm" /></button>
           <button class="btn-secondary small">Transcribe</button>
           <button class="btn-icon"><Trash class="icon-sm" /></button>
        </div>
      </div>
    </div>

    <!-- Transcript Modal -->
    <div class="transcript-overlay" v-if="showTranscript" @click.self="showTranscript = false">
      <div class="transcript-box">
        <h3>Voicemail Transcript</h3>
        <p class="trans-meta">From: (555) 123-9999</p>
        <div class="trans-body">
           <p class="speaker">Speaker 1:</p>
           <p>"Hi, this is Alice calling about the invoice. Please call me back when you get a chance."</p>
        </div>
        <div class="trans-footer">
           <button class="btn-primary full-width" @click="showTranscript = false">Close</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Voicemail, Play, Trash } from 'lucide-vue-next'

const showTranscript = ref(false)
</script>

<style scoped>
.view-container { padding: 8px 0; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h2 { font-size: 20px; font-weight: 700; color: var(--text-primary); }

.filters { display: flex; gap: 8px; background: #f1f5f9; padding: 4px; border-radius: 8px; }
.filter-btn { padding: 6px 12px; border: none; background: none; font-size: 13px; font-weight: 600; color: #64748b; cursor: pointer; border-radius: 6px; }
.filter-btn.active { background: white; color: var(--primary-color); box-shadow: 0 1px 2px rgba(0,0,0,0.1); }

.vm-list { display: flex; flex-direction: column; gap: 12px; }
.vm-item {
  display: flex; align-items: center; gap: 16px;
  background: white; border: 1px solid var(--border-color);
  padding: 16px; border-radius: var(--radius-md);
  transition: all 0.2s;
}
.vm-item.unread { border-left: 3px solid var(--primary-color); background: #F8FAFC; }
.vm-icon {
  width: 40px; height: 40px; background: #e2e8f0; border-radius: 50%; display: flex; align-items: center; justify-content: center; color: #64748b;
}
.vm-item.unread .vm-icon { background: var(--primary-light); color: var(--primary-color); }

.vm-details { flex: 1; }
.vm-row { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }
.caller { font-weight: 600; color: var(--text-primary); }
.bad-tag { font-size: 10px; background: var(--primary-color); color: white; padding: 2px 6px; border-radius: 99px; font-weight: 700; }
.vm-meta { font-size: 12px; color: var(--text-muted); }

.vm-actions { display: flex; align-items: center; gap: 12px; }
.btn-icon { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 8px; border-radius: 6px; }
.btn-icon:hover { background: #f1f5f9; color: var(--text-primary); }
.btn-icon.circle { border: 1px solid var(--border-color); border-radius: 50%; width: 32px; height: 32px; padding: 0; display: flex; align-items: center; justify-content: center; }

.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 6px 12px; border-radius: 6px; font-size: 12px; font-weight: 600; cursor: pointer; }
.icon { width: 20px; height: 20px; }
.icon-sm { width: 16px; height: 16px; }

/* Modal */
.transcript-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 50; }
.transcript-box { background: white; padding: 24px; border-radius: var(--radius-md); width: 400px; }
.transcript-box h3 { margin-bottom: 4px; }
.trans-meta { font-size: 12px; color: var(--text-muted); margin-bottom: 16px; }
.trans-body { background: #f8fafc; padding: 12px; border-radius: 8px; margin-bottom: 16px; font-size: 13px; line-height: 1.5; }
.speaker { font-weight: 700; color: var(--primary-color); margin-bottom: 4px; }
.trans-footer .btn-primary { width: 100%; padding: 10px; background: var(--primary-color); color: white; border: none; border-radius: 6px; font-weight: 600; cursor: pointer; }
</style>

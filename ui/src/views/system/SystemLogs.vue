<template>
  <div class="view-header">
    <div class="header-content">
      <h2>System Logs</h2>
      <p class="text-muted text-sm">Real-time FreeSwitch console and SIP traffic logs.</p>
    </div>
    <div class="header-actions">
       <button class="btn-secondary small">Pause</button>
       <button class="btn-secondary small">Clear</button>
       <button class="btn-secondary small">Download</button>
    </div>
  </div>

  <div class="logs-container">
    <div class="console-output">
      <div class="line info">[INFO] switch_core_sqldb.c:1662 Connecting to DSN [pgsql://callsign_db]</div>
      <div class="line notice">[NOTICE] switch_loadable_module.c:338 Adding Codec [PCMU]</div>
      <div class="line notice">[NOTICE] switch_loadable_module.c:338 Adding Codec [PCMA]</div>
      <div class="line debug">[DEBUG] sofia.c:4865 IP 203.0.113.5 Rejected by ACL "domains"</div>
      <div class="line warn">[WARN] sofia_reg.c:2928 Expire 1001@192.168.1.5</div>
      <div class="line info">[INFO] switch_core_session.c:2943 Executing [1002] bridge(sofia/internal/1002@192.168.1.5)</div>
      <div class="line debug">[DEBUG] sofia_glue.c:1234 Local SDP: v=0, o=FreeSWITCH...</div>
    </div>
  </div>
</template>

<script setup>
// In a real app, this would connect to a WebSocket
</script>

<style scoped>
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
}

.logs-container {
  background: #1e1e1e;
  border-radius: var(--radius-md);
  padding: 16px;
  height: calc(100vh - 200px);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.console-output {
  color: #d4d4d4;
  font-family: 'Fira Code', monospace;
  font-size: 12px;
  overflow-y: auto;
  flex: 1;
}

.line { margin-bottom: 4px; border-bottom: 1px solid rgba(255,255,255,0.05); padding-bottom: 2px; }
.line.info { color: #60a5fa; }
.line.notice { color: #4ade80; }
.line.debug { color: #94a3b8; }
.line.warn { color: #facc15; }

.btn-secondary {
  background: white;
  border: 1px solid var(--border-color);
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  cursor: pointer;
  margin-left: 8px;
}
</style>

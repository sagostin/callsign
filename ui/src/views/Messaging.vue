<template>
  <div class="messaging-container">
    <!-- Conversation Sidebar -->
    <div class="conv-sidebar">
      <div class="sidebar-header">
        <h2>Messages</h2>
        <button class="btn-icon circle"><MessageSquarePlus class="icon-sm" /></button>
      </div>
      <div class="search-box">
        <Search class="icon-xs muted" />
        <input type="text" placeholder="Search..." class="search-input">
      </div>
      
      <div class="conv-list">
        <div class="conv-item active">
          <div class="avatar-sm">JS</div>
          <div class="conv-info">
            <div class="conv-top">
              <span class="name">John Smith</span>
              <span class="time">2m</span>
            </div>
            <div class="conv-preview">Can we reschedule the demo?</div>
          </div>
        </div>
        
        <div class="conv-item">
          <div class="avatar-sm">Sales</div>
          <div class="conv-info">
            <div class="conv-top">
              <span class="name">(555) 123-4567</span>
              <span class="time">1h</span>
            </div>
            <div class="conv-preview">Thanks for calling!</div>
          </div>
        </div>

        <div class="conv-item unread">
          <div class="avatar-sm">Support</div>
          <div class="conv-info">
            <div class="conv-top">
              <span class="name">Support Ticket #99</span>
              <span class="time">Yesterday</span>
            </div>
            <div class="conv-preview">Your issue has been resolved.</div>
          </div>
          <div class="unread-dot"></div>
        </div>
      </div>
    </div>

    <!-- Active Conversation -->
    <div class="conv-main">
      <div class="msg-header">
        <div class="user-details">
          <h3>John Smith</h3>
          <span class="status-indicator online">Online</span>
        </div>
        <div class="header-actions">
           <button class="btn-icon"><Phone class="icon-sm" /></button>
           <button class="btn-icon"><MoreVertical class="icon-sm" /></button>
        </div>
      </div>

      <div class="msg-body">
        <div class="message incoming">
          <div class="bubble">Hi, are you available for a quick call?</div>
          <div class="timestamp">10:45 AM</div>
        </div>
        <div class="message outgoing">
          <div class="bubble">Yes, I have a few minutes now.</div>
          <div class="timestamp">10:46 AM</div>
        </div>
        <div class="message incoming">
          <div class="bubble">Great, can we reschedule the demo? Something came up.</div>
          <div class="timestamp">10:48 AM</div>
        </div>
      </div>

      <div class="msg-footer">
        <button class="btn-icon"><Paperclip class="icon-sm" /></button>
        <input type="text" class="msg-input" placeholder="Type a message...">
        <button class="btn-primary send-btn"><Send class="icon-sm" /></button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { MessageSquarePlus, Search, Phone, MoreVertical, Paperclip, Send } from 'lucide-vue-next'
</script>

<style scoped>
.messaging-container {
  display: flex;
  height: calc(100vh - 80px); /* Adjust for margins */
  background: white;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
  overflow: hidden;
}

.conv-sidebar {
  width: 320px;
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.sidebar-header h2 { font-size: 18px; font-weight: 700; color: var(--text-primary); }

.search-box {
  margin: 0 16px 16px;
  background: var(--bg-app);
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  gap: 8px;
}
.search-input {
  border: none;
  background: transparent;
  width: 100%;
  font-size: var(--text-sm);
  outline: none;
}

.conv-list {
  flex: 1;
  overflow-y: auto;
}

.conv-item {
  padding: 12px 16px;
  display: flex;
  gap: 12px;
  cursor: pointer;
  border-bottom: 1px solid var(--border-color);
  position: relative;
  transition: background var(--transition-fast);
}

.conv-item:hover { background: var(--bg-app); }
.conv-item.active { background: #F0F6FF; border-left: 3px solid var(--primary-color); }

.avatar-sm {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #E2E8F0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  color: #475569;
}

.conv-info { flex: 1; }

.conv-top {
  display: flex;
  justify-content: space-between;
  margin-bottom: 4px;
}

.name { font-weight: 600; font-size: var(--text-sm); color: var(--text-primary); }
.time { font-size: 11px; color: var(--text-muted); }

.conv-preview {
  font-size: 12px;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 180px;
}

.unread-dot {
  position: absolute;
  right: 16px;
  top: 50%;
  width: 8px;
  height: 8px;
  background: var(--primary-color);
  border-radius: 50%;
}

/* MAIN CHAT AREA */
.conv-main {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.msg-header {
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.user-details h3 { font-size: 16px; font-weight: 700; color: var(--text-primary); }
.status-indicator { font-size: 11px; color: var(--status-good); font-weight: 600; }

.header-actions { display: flex; gap: 8px; }

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  padding: 8px;
  border-radius: var(--radius-sm);
  color: var(--text-muted);
}
.btn-icon:hover { background: var(--bg-app); color: var(--text-primary); }
.btn-icon.circle { border-radius: 50%; background: var(--primary-light); color: var(--primary-color); }

.msg-body {
  flex: 1;
  padding: 20px;
  background: #F8FAFC;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.message { display: flex; flex-direction: column; max-width: 60%; }
.message.incoming { align-self: flex-start; }
.message.outgoing { align-self: flex-end; align-items: flex-end; }

.bubble {
  padding: 10px 16px;
  border-radius: 12px;
  font-size: var(--text-sm);
  line-height: 1.4;
}

.message.incoming .bubble {
  background: white;
  border: 1px solid var(--border-color);
  border-bottom-left-radius: 2px;
  color: var(--text-main);
}

.message.outgoing .bubble {
  background: var(--primary-color);
  color: white;
  border-bottom-right-radius: 2px;
}

.timestamp { font-size: 10px; color: var(--text-muted); margin-top: 4px; padding: 0 4px; }

.msg-footer {
  padding: 16px;
  border-top: 1px solid var(--border-color);
  display: flex;
  gap: 12px;
  align-items: center;
}

.msg-input {
  flex: 1;
  border: 1px solid var(--border-color);
  padding: 10px 16px;
  border-radius: 24px;
  font-size: var(--text-sm);
  outline: none;
}
.msg-input:focus { border-color: var(--primary-color); }

.send-btn {
  width: 40px; 
  height: 40px; 
  border-radius: 50%; 
  display: flex; 
  align-items: center; 
  justify-content: center;
  background: var(--primary-color);
  border: none;
  color: white;
  cursor: pointer;
}

.icon-sm { width: 18px; height: 18px; }
.icon-xs { width: 14px; height: 14px; }
.muted { color: var(--text-muted); }
</style>

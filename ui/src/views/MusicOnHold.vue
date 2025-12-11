<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Music on Hold</h2>
      <p class="text-muted text-sm">Manage hold music classes and streams.</p>
    </div>
    <div class="header-actions">
      <button class="btn-primary" @click="$router.push('/admin/music-on-hold/new')">+ Add Stream</button>
    </div>
  </div>

  <div class="tabs">
    <button class="tab" :class="{ active: activeTab === 'streams' }" @click="activeTab = 'streams'">Streams</button>
    <button class="tab" :class="{ active: activeTab === 'playlists' }" @click="activeTab = 'playlists'">Playlists</button>
  </div>

  <!-- STREAMS TAB -->
  <div class="tab-content" v-if="activeTab === 'streams'">
      <DataTable :columns="columns" :data="streams" actions>
        <template #rate="{ value }">
          <span class="font-mono text-xs">{{ value }}Hz</span>
        </template>
        <template #actions>
          <button class="btn-link" @click="$router.push('/admin/music-on-hold/1')">Edit</button>
          <button class="btn-link text-bad">Delete</button>
        </template>
      </DataTable>
  </div>

  <!-- PLAYLISTS TAB -->
  <div class="tab-content" v-else-if="activeTab === 'playlists'">
      <div class="action-bar">
         <p class="text-sm text-muted">Manage custom playlists for hold music.</p>
         <button class="btn-primary small" @click="addPlaylist">+ New Playlist</button>
      </div>

      <div class="playlist-grid">
         <div v-for="(list, idx) in playlists" :key="idx" class="playlist-card">
            <div class="playlist-header">
              <div class="pl-info">
                 <span class="pl-name">{{ list.name }}</span>
                 <span class="pl-count">{{ list.tracks.length }} Tracks</span>
              </div>
              <div class="pl-more">...</div>
            </div>
            
            <div class="track-list">
               <div v-for="(track, tIdx) in list.tracks" :key="tIdx" class="track">
                  <div class="track-icon">🎵</div>
                  <div class="track-details">
                     <div class="track-name">{{ track.name }}</div>
                     <div class="track-dur">{{ track.duration }}</div>
                  </div>
               </div>
            </div>

            <div class="pl-footer">
               <button class="btn-secondary full-width">Manage Tracks</button>
            </div>
         </div>
      </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import DataTable from '../components/common/DataTable.vue'

const activeTab = ref('streams')

const columns = [
  { key: 'name', label: 'Category Name', width: '200px' },
  { key: 'path', label: 'Path / URL' },
  { key: 'rate', label: 'Sample Rate' },
  { key: 'channels', label: 'Channels' }
]

const streams = ref([
  { name: 'default', path: 'local_stream://default', rate: '48000', channels: 'Mono' },
  { name: 'rock', path: '/var/lib/sounds/rock', rate: '48000', channels: 'Stereo' },
  { name: 'sales_stream', path: 'shout://stream.example.com/sales', rate: '32000', channels: 'Mono' },
])

const playlists = ref([
   { 
     name: 'Default Hold', 
     tracks: [
        { name: 'Classical_Guitar.mp3', duration: '3:42' },
        { name: 'Jazz_Piano.mp3', duration: '4:15' },
        { name: 'Smooth_Synth.wav', duration: '2:30' }
     ]
   },
   {
     name: 'Holiday Promotions',
     tracks: [
        { name: 'Jingle_Bell_Rock.mp3', duration: '2:10' },
        { name: 'Promo_Spot_2024.wav', duration: '0:30' }
     ]
   }
])

const addPlaylist = () => {
   playlists.value.push({ name: 'New Playlist', tracks: [] })
}
</script>

<style scoped>
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
}

.btn-primary {
  background-color: var(--primary-color);
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-weight: 500;
  font-size: var(--text-sm);
}
.btn-primary.small { padding: 6px 12px; font-size: 12px; }

.btn-link {
  background: none;
  border: none;
  color: var(--primary-color);
  font-size: var(--text-xs);
  margin-left: 8px;
  cursor: pointer;
  font-weight: 500;
}

.text-bad { color: var(--status-bad); }

.font-mono { font-family: monospace; }
.text-xs { font-size: 11px; }

/* Tabs */
.tabs { display: flex; gap: 2px; border-bottom: 1px solid var(--border-color); }
.tab { padding: 8px 16px; background: transparent; border: 1px solid transparent; border-bottom: none; cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-muted); border-radius: 4px 4px 0 0; }
.tab.active { background: white; border-color: var(--border-color); color: var(--primary-color); margin-bottom: -1px; }
.tab-content { background: white; border: 1px solid var(--border-color); border-top: none; padding: 24px; border-radius: 0 0 4px 4px; }

/* Playlists */
.action-bar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.playlist-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 24px; }

.playlist-card { 
   border: 1px solid var(--border-color); border-radius: 8px; 
   background: white; overflow: hidden; display: flex; flex-direction: column;
}

.playlist-header { padding: 16px; border-bottom: 1px solid var(--border-color); display: flex; justify-content: space-between; align-items: flex-start; background: #f8fafc; }
.pl-name { font-weight: 600; font-size: 14px; display: block; color: var(--text-primary); }
.pl-count { font-size: 11px; color: var(--text-muted); }
.pl-more { color: var(--text-muted); cursor: pointer; font-weight: bold; letter-spacing: 1px; }

.track-list { padding: 16px; flex: 1; }
.track { display: flex; gap: 12px; align-items: center; margin-bottom: 12px; }
.track:last-child { margin-bottom: 0; }
.track-icon { width: 24px; height: 24px; background: var(--bg-secondary); border-radius: 4px; display: flex; align-items: center; justify-content: center; font-size: 10px; }
.track-details { flex: 1; min-width: 0; }
.track-name { font-size: 12px; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.track-dur { font-size: 10px; color: var(--text-muted); }

.pl-footer { padding: 12px; border-top: 1px solid var(--border-color); background: #f8fafc; }
.btn-secondary { background: white; border: 1px solid var(--border-color); color: var(--text-main); font-size: 12px; padding: 6px 12px; border-radius: 4px; cursor: pointer; font-weight: 500; }
.full-width { width: 100%; }
</style>

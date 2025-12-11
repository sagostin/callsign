<template>
  <div class="system-media-page">
    <div class="view-header">
      <div class="header-content">
        <h2>System Media</h2>
        <p class="text-muted text-sm">Manage global audio resources: sounds, music on hold, and phrases.</p>
      </div>
      <div class="header-actions">
        <!-- Dynamic Actions based on active tab -->
        <template v-if="activeTab === 'sounds'">
           <button class="btn-secondary" @click="showImportModal = true">
            <DownloadIcon class="btn-icon" /> Import Pack
          </button>
          <button class="btn-primary" @click="showUploadSoundModal = true">
            <UploadIcon class="btn-icon" /> Upload Sound
          </button>
        </template>
        
        <template v-if="activeTab === 'music'">
          <button class="btn-primary" @click="showUploadMusicModal = true">
            <UploadIcon class="btn-icon" /> Upload Music
          </button>
        </template>

        <template v-if="activeTab === 'phrases'">
          <button class="btn-primary" @click="createPhrase">
            <PlusIcon class="btn-icon" /> New Phrase
          </button>
        </template>
      </div>
    </div>

    <!-- Tabs -->
    <div class="tabs">
      <button class="tab" :class="{ active: activeTab === 'sounds' }" @click="activeTab = 'sounds'">
        <Volume2 class="tab-icon" /> System Sounds
      </button>
      <button class="tab" :class="{ active: activeTab === 'music' }" @click="activeTab = 'music'">
        <MusicIcon class="tab-icon" /> Music On Hold
      </button>
      <button class="tab" :class="{ active: activeTab === 'phrases' }" @click="activeTab = 'phrases'">
        <MessageSquare class="tab-icon" /> Phrases
      </button>
    </div>

    <!-- CONTENT: SYSTEM SOUNDS -->
    <div v-if="activeTab === 'sounds'" class="tab-content">
       <div class="file-browser-layout">
          <!-- Sidebar: Folders/Languages -->
          <div class="browser-sidebar">
             <div class="sidebar-section">
                <div class="section-title">LANGUAGES</div>
                <div class="lang-list">
                   <div v-for="lang in availableLanguages" :key="lang.code" 
                        class="lang-item" 
                        :class="{ active: currentSoundPath.startsWith(lang.code) }"
                        @click="navigateToSoundLang(lang.code)">
                        <span class="flag">{{ lang.flag }}</span>
                        <span class="name">{{ lang.name }}</span>
                   </div>
                </div>
             </div>
          </div>

          <!-- Main: File List -->
          <div class="browser-main">
             <div class="browser-toolbar">
                <div class="breadcrumbs">
                   <span v-for="(part, idx) in currentSoundPath.split('/')" :key="idx" class="crumb">
                      {{ part }} <span class="sep">/</span>
                   </span>
                </div>
             </div>
             
             <div class="file-grid">
               <!-- Folders -->
               <div v-for="folder in currentSoundFolders" :key="folder.name" class="file-card folder" @click="enterSoundFolder(folder)">
                  <FolderIcon class="icon" />
                  <span class="name">{{ folder.name }}</span>
               </div>
               
               <!-- Files -->
               <div v-for="file in currentSoundFiles" :key="file.name" class="file-card file">
                  <div class="file-icon-wrapper">
                    <FileAudioIcon class="icon" />
                  </div>
                  <div class="file-info">
                     <span class="name">{{ file.name }}</span>
                     <span class="size">{{ formatSize(file.size) }}</span>
                  </div>
                  <div class="file-actions">
                     <button class="btn-icon small" @click="playSound(file)"><PlayIcon class="icon-sm" /></button>
                  </div>
               </div>
             </div>
          </div>
       </div>
    </div>

    <!-- CONTENT: MUSIC ON HOLD -->
    <div v-if="activeTab === 'music'" class="tab-content">
       <div class="file-browser-layout">
          <!-- Sidebar: Rates/Groups -->
          <div class="browser-sidebar">
             <div class="sidebar-section">
                <div class="section-title">SAMPLE RATES</div>
                <!-- Logic to list top-level music folders (8000, 16000, etc) -->
                <div class="folder-list">
                   <div v-for="folder in musicRootFolders" :key="folder.name" 
                        class="folder-item"
                        :class="{ active: currentMusicPath.startsWith(folder.name) }"
                        @click="navigateToMusicFolder(folder.name)">
                      <MusicIcon class="icon-xs" />
                      <span>{{ folder.name }}</span>
                      <span class="badge">{{ folder.childCount }}</span>
                   </div>
                </div>
             </div>
          </div>

          <!-- Main: Music Files -->
          <div class="browser-main">
             <div class="browser-toolbar">
                <div class="breadcrumbs">
                   <span class="crumb root" @click="currentMusicPath=''">Music</span>
                   <span class="sep">/</span>
                   <span v-for="(part, idx) in currentMusicPath.split('/')" :key="idx" class="crumb">
                      {{ part }} <span class="sep">/</span>
                   </span>
                </div>
             </div>

             <div class="file-grid">
                <!-- Subfolders (Playlists) -->
                <div v-for="folder in currentMusicFolders" :key="folder.name" class="file-card folder" @click="enterMusicFolder(folder)">
                   <FolderIcon class="icon" />
                   <span class="name">{{ folder.name }}</span>
                </div>

                <!-- Music Files -->
                <div v-for="file in currentMusicFiles" :key="file.name" class="file-card file">
                   <div class="file-icon-wrapper music">
                     <MusicIcon class="icon" />
                   </div>
                   <div class="file-info">
                      <span class="name">{{ file.name }}</span>
                      <span class="size">{{ formatSize(file.size) }}</span>
                   </div>
                   <div class="file-actions">
                      <button class="btn-icon small" @click="playMusic(file)"><PlayIcon class="icon-sm" /></button>
                      <button class="btn-icon small danger" @click="deleteMusic(file)"><Trash2Icon class="icon-sm" /></button>
                   </div>
                </div>
             </div>
          </div>
       </div>
    </div>

    <!-- CONTENT: PHRASES -->
    <div v-if="activeTab === 'phrases'" class="tab-content">
       <div class="phrases-layout">
          <div class="filter-bar">
            <div class="search-box">
               <SearchIcon class="search-icon" />
               <input v-model="phrasesSearchQuery" placeholder="Search phrases..." class="search-input">
            </div>
            <select v-model="phrasesFilterLanguage" class="filter-select">
               <option value="">All Languages</option>
               <option value="en-us">English (US)</option>
               <option value="es-mx">Spanish (MX)</option>
               <option value="fr-ca">French (CA)</option>
            </select>
          </div>

          <div class="phrases-grid">
             <div v-for="phrase in filteredPhrases" :key="phrase.id" class="phrase-card" :class="{ disabled: !phrase.enabled }">
                <div class="phrase-header">
                  <div class="phrase-info">
                     <h4>{{ phrase.name }}</h4>
                     <span class="phrase-desc">{{ phrase.description }}</span>
                  </div>
                  <div class="phrase-badges">
                     <span class="lang-badge">{{ phrase.language }}</span>
                     <span class="status-badge" :class="phrase.enabled ? 'enabled' : 'disabled'">
                        {{ phrase.enabled ? 'Enabled' : 'Disabled' }}
                     </span>
                  </div>
                </div>
                <div class="phrase-actions">
                   <button class="btn-link" @click="editPhrase(phrase)"><EditIcon class="icon-xs" /> Edit</button>
                   <button class="btn-link danger" @click="deletePhrase(phrase)"><Trash2Icon class="icon-xs" /> Delete</button>
                </div>
             </div>
          </div>
       </div>
    </div>

    <!-- Modals (Placeholders for now) -->
    <div v-if="showImportModal || showUploadSoundModal || showUploadMusicModal" class="modal-overlay" @click.self="closeModals">
        <div class="modal-card">
            <div class="modal-header">
                <h3>Upload / Import</h3>
                <button class="close-btn" @click="closeModals">×</button>
            </div>
            <div class="modal-body">
                <p>Upload functionality implementation pending.</p>
            </div>
        </div>
    </div>

  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { 
  Volume2, Music as MusicIcon, MessageSquare, Download as DownloadIcon, 
  Upload as UploadIcon, Plus as PlusIcon, Folder as FolderIcon,
  FileAudio as FileAudioIcon, Play as PlayIcon, Trash2 as Trash2Icon,
  Search as SearchIcon, Edit as EditIcon
} from 'lucide-vue-next'
import { systemAPI } from '../../services/api'

const activeTab = ref('sounds')
const isLoading = ref(false)

// Data
const soundsTree = ref([]) 
const musicTree = ref([]) 
const phrases = ref([])

// State
const currentSoundPath = ref('en/us/callie') 
const currentMusicPath = ref('8000') 

// UI State
const showImportModal = ref(false)
const showUploadSoundModal = ref(false)
const showUploadMusicModal = ref(false)
const phrasesSearchQuery = ref('')
const phrasesFilterLanguage = ref('')

const closeModals = () => {
    showImportModal.value = false
    showUploadSoundModal.value = false
    showUploadMusicModal.value = false
}

// -- SOUNDS LOGIC --
const availableLanguages = [
    { code: 'en/us/callie', name: 'English (US)', flag: '🇺🇸' },
    { code: 'en/gb', name: 'English (UK)', flag: '🇬🇧' },
    { code: 'fr/ca', name: 'French (CA)', flag: '🇨🇦' },
    // Add more as needed or derive from tree
]

const navigateToSoundLang = (path) => { currentSoundPath.value = path }

const getNodeByPath = (tree, path) => {
    if (!path) return tree
    const parts = path.split('/')
    let current = tree
    for (const part of parts) {
        if (!current) return null
        // If current is array (root), find child
        if (Array.isArray(current)) {
            current = current.find(n => n.name === part)
        } else if (current.children) {
             current = current.children.find(n => n.name === part)
        } else {
            return null
        }
    }
    return current
}

const currentSoundNode = computed(() => {
    // Navigate soundsTree using currentSoundPath
    // Tree root is array of languages? Or root folder?
    // ListSystemSounds returns children of /usr/share/freeswitch/sounds
    // So root is array of children [en, fr, ru, etc]
    return getNodeByPath(soundsTree.value, currentSoundPath.value)
})

const currentSoundFolders = computed(() => {
    const node = currentSoundNode.value
    if (!node) return []
    const children = node.children || []
    return children.filter(c => c.type === 'directory')
})

const currentSoundFiles = computed(() => {
    const node = currentSoundNode.value
    if (!node) return []
    const children = node.children || []
    return children.filter(c => c.type === 'file')
})

const enterSoundFolder = (folder) => {
    currentSoundPath.value = currentSoundPath.value + '/' + folder.name
}

// -- MUSIC LOGIC --
const musicRootFolders = computed(() => {
    // ListSystemMusic returns children of /usr/share/freeswitch/sounds/music
    // Expected to be [8000, 16000, 32000, 48000]
    return musicTree.value
        .filter(n => n.type === 'directory')
        .map(n => ({ name: n.name, childCount: countMusicFiles(n) }))
})

const countMusicFiles = (node) => {
    if (!node.children) return 0
    let count = 0
    for (const child of node.children) {
        if (child.type === 'file') count++
        else count += countMusicFiles(child)
    }
    return count
}

const navigateToMusicFolder = (path) => { currentMusicPath.value = path }

const currentMusicNode = computed(() => {
    return getNodeByPath(musicTree.value, currentMusicPath.value)
})

const currentMusicFolders = computed(() => {
    const node = currentMusicNode.value
    if (!node) return []
    const children = node.children || []
    return children.filter(c => c.type === 'directory')
})

const currentMusicFiles = computed(() => {
    const node = currentMusicNode.value
    if (!node) return []
    const children = node.children || []
    return children.filter(c => c.type === 'file')
})

const enterMusicFolder = (folder) => {
    currentMusicPath.value = currentMusicPath.value ? (currentMusicPath.value + '/' + folder.name) : folder.name
}


// -- PHRASES LOGIC --
const filteredPhrases = computed(() => {
    return phrases.value.filter(p => {
        const matchSearch = p.name.toLowerCase().includes(phrasesSearchQuery.value.toLowerCase()) || 
                            p.description.toLowerCase().includes(phrasesSearchQuery.value.toLowerCase())
        const matchLang = !phrasesFilterLanguage.value || p.language === phrasesFilterLanguage.value
        return matchSearch && matchLang
    })
})

const createPhrase = () => { alert('Create phrase') }
const editPhrase = (p) => { alert('Edit ' + p.name) }
const deletePhrase = (p) => { alert('Delete ' + p.name) }

// -- ACTIONS --
const playSound = (file) => { console.log('Play sound', file) }
const playMusic = (file) => { console.log('Play music', file) }
const deleteMusic = (file) => { console.log('Delete music', file) }

const formatSize = (bytes) => {
    if(bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

// API
const loadData = async () => {
    isLoading.value = true
    try {
        const [soundsRes, musicRes] = await Promise.all([
            systemAPI.listSounds(),
            systemAPI.listMusic()
        ])
        soundsTree.value = soundsRes.data.data
        musicTree.value = musicRes.data.data
        
        // Mock phrases for now
        phrases.value = [
            { id: 1, name: 'welcome_ivr', description: 'Main IVR Welcome', language: 'en-us', enabled: true },
            { id: 2, name: 'out_of_hours', description: 'Closed message', language: 'en-us', enabled: true }
        ]
    } catch(e) {
        console.error("Failed to load media", e)
    } finally {
        isLoading.value = false
    }
}

onMounted(() => {
    loadData()
})

</script>

<style scoped>
.system-media-page { height: 100%; display: flex; flex-direction: column; }
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; flex-shrink: 0; }
.header-content h2 { margin: 0 0 4px; }
.header-actions { display: flex; gap: 8px; }

/* Tabs */
.tabs { display: flex; gap: 2px; border-bottom: 1px solid var(--border-color); margin-bottom: 0; flex-shrink: 0; }
.tab { display: flex; align-items: center; gap: 8px; padding: 12px 20px; background: transparent; border: 1px solid transparent; border-bottom: none; cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-muted); border-radius: 6px 6px 0 0; transition: all 0.2s; }
.tab:hover { background: var(--bg-hover); }
.tab.active { background: white; border-color: var(--border-color); color: var(--primary-color); margin-bottom: -1px; font-weight: 600; }
.tab-icon { width: 16px; height: 16px; }

/* Tab Content */
.tab-content { flex: 1; background: white; border: 1px solid var(--border-color); border-top: none; display: flex; overflow: hidden; }

/* File Browser Layout */
.file-browser-layout { display: flex; width: 100%; height: 100%; }
.browser-sidebar { width: 240px; border-right: 1px solid var(--border-color); background: var(--bg-app); overflow-y: auto; padding: 20px 0; }
.browser-main { flex: 1; display: flex; flex-direction: column; overflow: hidden; background: white; }

.sidebar-section { margin-bottom: 24px; }
.section-title { padding: 0 20px; font-size: 11px; font-weight: 700; color: var(--text-muted); margin-bottom: 8px; }

.lang-item, .folder-item { display: flex; align-items: center; gap: 10px; padding: 8px 20px; cursor: pointer; font-size: 13px; color: var(--text-primary); }
.lang-item:hover, .folder-item:hover { background: rgba(0,0,0,0.03); }
.lang-item.active, .folder-item.active { background: var(--primary-light); color: var(--primary-color); font-weight: 500; }
.lang-item .flag { font-size: 16px; }
.folder-item .badge { margin-left: auto; font-size: 10px; background: rgba(0,0,0,0.1); padding: 2px 6px; border-radius: 10px; }

/* Browser Toolbar */
.browser-toolbar { padding: 16px 24px; border-bottom: 1px solid var(--border-color); display: flex; align-items: center; }
.breadcrumbs { display: flex; align-items: center; font-size: 13px; color: var(--text-muted); }
.crumb { display: flex; align-items: center; cursor: pointer; }
.crumb:hover { text-decoration: underline; }
.crumb.root { font-weight: 600; color: var(--primary-color); }
.sep { margin: 0 8px; color: var(--text-muted-light); }

/* File Grid */
.file-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(140px, 1fr)); gap: 16px; padding: 24px; overflow-y: auto; align-content: start; }
.file-card { border: 1px solid var(--border-color); border-radius: 8px; padding: 16px; display: flex; flex-direction: column; align-items: center; text-align: center; gap: 10px; transition: all 0.2s; cursor: pointer; position: relative; }
.file-card:hover { border-color: var(--primary-color); box-shadow: 0 4px 12px rgba(0,0,0,0.05); }
.file-card .icon { width: 32px; height: 32px; color: var(--text-muted); }
.file-card.folder .icon { color: #fbbf24; }
.file-card.file .icon { color: var(--primary-color); }
.file-card .file-icon-wrapper.music .icon { color: #8b5cf6; }
.file-card .name { font-size: 13px; font-weight: 500; word-break: break-word; line-height: 1.3; }
.file-card .size { font-size: 11px; color: var(--text-muted); }
.file-card .file-actions { position: absolute; top: 8px; right: 8px; display: none; gap: 4px; }
.file-card:hover .file-actions { display: flex; }

/* Phrases Layout */
.phrases-layout { display: flex; flex-direction: column; height: 100%; width: 100%; }
.filter-bar { display: flex; gap: 12px; padding: 16px 24px; border-bottom: 1px solid var(--border-color); background: var(--bg-app); }
.search-box { flex: 1; position: relative; }
.search-icon { position: absolute; left: 10px; top: 50%; transform: translateY(-50%); width: 16px; height: 16px; color: var(--text-muted); }
.search-input { width: 100%; padding: 8px 10px 8px 32px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; box-sizing: border-box; }
.filter-select { padding: 8px 12px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; }

.phrases-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 16px; padding: 24px; overflow-y: auto; }
.phrase-card { background: white; border: 1px solid var(--border-color); border-radius: 8px; padding: 16px; display: flex; flex-direction: column; justify-content: space-between; }
.phrase-card.disabled { opacity: 0.6; }
.phrase-header { display: flex; justify-content: space-between; align-items: start; margin-bottom: 12px; }
.phrase-info h4 { margin: 0 0 4px; font-size: 14px; }
.phrase-desc { font-size: 12px; color: var(--text-muted); }
.phrase-badges { display: flex; flex-direction: column; align-items: flex-end; gap: 4px; }
.lang-badge { font-size: 10px; background: var(--bg-app); padding: 2px 6px; border-radius: 4px; color: var(--text-muted); text-transform: uppercase; }
.status-badge { font-size: 10px; padding: 2px 6px; border-radius: 4px; font-weight: 600; }
.status-badge.enabled { background: #dcfce7; color: #16a34a; }
.status-badge.disabled { background: #f3f4f6; color: #6b7280; }
.phrase-actions { display: flex; justify-content: flex-end; gap: 8px; border-top: 1px solid var(--border-color); padding-top: 12px; }
.icon-xs { width: 12px; height: 12px; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal-card { background: white; border-radius: 12px; width: 90%; max-width: 480px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { margin: 0; font-size: 16px; }
.close-btn { width: 28px; height: 28px; border: none; background: #f1f5f9; border-radius: 6px; font-size: 18px; cursor: pointer; }
.modal-body { padding: 20px; }

/* Buttons */
.btn-primary, .btn-secondary { display: flex; align-items: center; gap: 6px; padding: 8px 16px; border-radius: 6px; font-size: 13px; font-weight: 500; cursor: pointer; border: none; }
.btn-primary { background: var(--primary-color); color: white; }
.btn-secondary { background: white; border: 1px solid var(--border-color); }
.btn-icon { width: 14px; height: 14px; }
.btn-icon.small { width: 24px; height: 24px; padding: 0; justify-content: center; background: white; border: 1px solid var(--border-color); border-radius: 4px; }
.btn-icon.small:hover { border-color: var(--primary-color); color: var(--primary-color); }
.btn-icon.danger:hover { border-color: #ef4444; color: #ef4444; }
.btn-link { background: none; border: none; font-size: 12px; color: var(--text-muted); cursor: pointer; display: flex; align-items: center; gap: 4px; padding: 4px 8px; border-radius: 4px; }
.btn-link:hover { color: var(--primary-color); background: var(--bg-hover); }
.btn-link.danger:hover { color: #ef4444; }

</style>


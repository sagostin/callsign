<template>
  <div class="audio-recorder">
    <div class="recorder-controls">
      <button 
        class="record-btn" 
        :class="{ recording: isRecording }"
        @click="toggleRecording"
        :disabled="isPlaying"
      >
        <div class="record-icon"></div>
        <span>{{ isRecording ? formatTime(recordingTime) : (audioBlob ? 'Rerecord' : 'Record') }}</span>
      </button>

      <button 
        v-if="audioBlob && !isRecording" 
        class="play-btn"
        @click="togglePlayback"
      >
        <span v-if="isPlaying">Stop</span>
        <span v-else>Play Review</span>
      </button>
      
      <span v-if="audioBlob && !isRecording" class="duration">{{ formatTime(duration) }}</span>
    </div>
    
    <div v-if="error" class="error-msg">{{ error }}</div>
    
    <canvas ref="visualizer" class="visualizer" width="300" height="40"></canvas>
  </div>
</template>

<script setup>
import { ref, onUnmounted, watch } from 'vue'

const emit = defineEmits(['record-complete'])

const isRecording = ref(false)
const isPlaying = ref(false)
const recordingTime = ref(0)
const duration = ref(0)
const audioBlob = ref(null)
const error = ref('')

const visualizer = ref(null)
let mediaRecorder = null
let audioContext = null
let analyser = null
let dataArray = null
let source = null
let timerInterval = null
let animationId = null
let audioUrl = null
let audio = null

const startRecording = async () => {
    try {
        const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
        
        audioContext = new (window.AudioContext || window.webkitAudioContext)()
        analyser = audioContext.createAnalyser()
        source = audioContext.createMediaStreamSource(stream)
        source.connect(analyser)
        
        analyser.fftSize = 256
        const bufferLength = analyser.frequencyBinCount
        dataArray = new Uint8Array(bufferLength)
        
        mediaRecorder = new MediaRecorder(stream)
        const chunks = []
        
        mediaRecorder.ondataavailable = (e) => chunks.push(e.data)
        
        mediaRecorder.onstop = () => {
            const blob = new Blob(chunks, { type: 'audio/wav' }) // Chrome records as webm usually, but we label as wav for simplicity or conversion needs?
            // Actually, we should probably output what the browser gives and let backend handle or user know.
            // Using 'audio/webm' or 'audio/ogg' depending on browser. 
            // Ideally we'd use a library for wav encoding if strictly needed, but let's try raw blob.
            audioBlob.value = blob
            audioUrl = URL.createObjectURL(blob)
            audio = new Audio(audioUrl)
            
            audio.onended = () => { isPlaying.value = false }
            audio.loadedmetadata = () => { duration.value = audio.duration }
            
            emit('record-complete', blob)
            
            // Stop stream tracks
            stream.getTracks().forEach(track => track.stop())
        }
        
        mediaRecorder.start()
        isRecording.value = true
        error.value = ''
        
        startTime = Date.now()
        timerInterval = setInterval(() => {
            recordingTime.value = (Date.now() - startTime) / 1000
        }, 100)
        
        drawVisualizer()
        
    } catch (e) {
        console.error('Recording error', e)
        error.value = 'Microphone access denied or not available.'
    }
}

let startTime = 0

const stopRecording = () => {
    if (mediaRecorder && mediaRecorder.state !== 'inactive') {
        mediaRecorder.stop()
        isRecording.value = false
        clearInterval(timerInterval)
        cancelAnimationFrame(animationId)
        
        // Clear visualizer
        const canvas = visualizer.value
        const ctx = canvas.getContext('2d')
        ctx.clearRect(0, 0, canvas.width, canvas.height)
    }
}

const toggleRecording = () => {
    if (isRecording.value) stopRecording()
    else startRecording()
}

const togglePlayback = () => {
    if (!audio) return
    
    if (isPlaying.value) {
        audio.pause()
        audio.currentTime = 0
        isPlaying.value = false
    } else {
        audio.play()
        isPlaying.value = true
    }
}

const drawVisualizer = () => {
    if (!isRecording.value) return
    
    animationId = requestAnimationFrame(drawVisualizer)
    
    analyser.getByteFrequencyData(dataArray)
    
    const canvas = visualizer.value
    const ctx = canvas.getContext('2d')
    const width = canvas.width
    const height = canvas.height
    
    ctx.clearRect(0, 0, width, height) // Clear
    
    const barWidth = (width / dataArray.length) * 2.5
    let x = 0
    
    for (let i = 0; i < dataArray.length; i++) {
        const barHeight = dataArray[i] / 2
        
        ctx.fillStyle = `rgb(${barHeight + 100}, 50, 50)`
        ctx.fillRect(x, height - barHeight, barWidth, barHeight)
        
        x += barWidth + 1
    }
}

const formatTime = (seconds) => {
    if (!seconds) return '0:00'
    const m = Math.floor(seconds / 60)
    const s = Math.floor(seconds % 60)
    return `${m}:${s.toString().padStart(2, '0')}`
}

onUnmounted(() => {
    stopRecording()
    if (audioUrl) URL.revokeObjectURL(audioUrl)
    if (audioContext) audioContext.close()
})
</script>

<style scoped>
.audio-recorder {
    background: #f8fafc;
    border: 1px solid var(--border-color);
    border-radius: 8px;
    padding: 16px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
}

.recorder-controls {
    display: flex;
    align-items: center;
    gap: 16px;
}

.record-btn {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    border-radius: 20px;
    border: 1px solid var(--status-bad);
    background: white;
    color: var(--status-bad);
    cursor: pointer;
    font-weight: 600;
    transition: all 0.2s;
}

.record-btn:hover {
    background: #fef2f2;
}

.record-btn.recording {
    background: var(--status-bad);
    color: white;
}

.record-icon {
    width: 12px;
    height: 12px;
    background: currentColor;
    border-radius: 50%;
}

.record-btn.recording .record-icon {
    border-radius: 2px;
    animation: pulse 1s infinite;
}

.play-btn {
    padding: 8px 16px;
    background: white;
    border: 1px solid var(--border-color);
    border-radius: 6px;
    cursor: pointer;
    font-size: 13px;
}

.duration {
    font-family: monospace;
    font-size: 13px;
    color: var(--text-muted);
}

.visualizer {
    width: 100%;
    height: 40px;
    background: #000;
    border-radius: 4px;
}

.error-msg {
    color: var(--status-bad);
    font-size: 12px;
}

@keyframes pulse {
    0% { opacity: 1; }
    50% { opacity: 0.5; }
    100% { opacity: 1; }
}
</style>

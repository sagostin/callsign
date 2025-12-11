<template>
  <div class="view-header">
    <div class="header-left">
      <button class="back-link" @click="$router.push('/admin')">← Back to Admin</button>
      <h2>Add New Gateway</h2>
    </div>
  </div>

  <div class="form-container">
    <div class="form-section">
      <h3>1. Gateway Connectivity</h3>
      <div class="form-group">
        <label>Gateway Name</label>
        <input type="text" v-model="form.name" class="input-field" placeholder="My SIP Trunk">
      </div>
      
      <div class="form-group">
        <label>Proxy Address / Domain</label>
        <input type="text" v-model="form.proxy" class="input-field" placeholder="sip.provider.com">
      </div>

       <div class="form-grid">
        <div class="form-group">
          <label>Protocol</label>
          <select v-model="form.protocol" class="input-field">
            <option value="udp">UDP</option>
            <option value="tcp">TCP</option>
            <option value="tls">TLS</option>
          </select>
        </div>
        <div class="form-group">
           <label>Register</label>
           <select v-model="form.register" class="input-field">
             <option :value="true">True</option>
             <option :value="false">False</option>
           </select>
        </div>
      </div>
    </div>

    <div class="form-section">
       <h3>2. Authentication</h3>
       <div class="form-group">
        <label>Username</label>
        <input type="text" v-model="form.username" class="input-field">
      </div>
      <div class="form-group">
        <label>Password</label>
        <input type="password" v-model="form.password" class="input-field">
      </div>
       <div class="form-group">
        <label>Realm (Optional)</label>
        <input type="text" v-model="form.realm" class="input-field">
      </div>
    </div>

    <div class="form-actions">
      <button class="btn-primary large" :disabled="!isValid" @click="saveGateway">
        Save Gateway
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const form = ref({
  name: '',
  proxy: '',
  protocol: 'udp',
  register: false,
  username: '',
  password: '',
  realm: ''
})

const isValid = computed(() => {
  return form.value.name && form.value.proxy
})

const saveGateway = () => {
  alert(`Configured Gateway "${form.value.name}"`)
  router.push('/admin') // Should ideally go back to Admin > Trunks tab
}
</script>

<style scoped>
/* Reusing Form Styles */
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
  margin-bottom: var(--spacing-md);
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
.input-field:focus { border-color: var(--primary-color); }

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
</style>

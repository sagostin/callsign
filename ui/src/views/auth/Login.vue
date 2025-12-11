Ò<template>
  <div class="login-container">
    <div class="login-card">
      <div class="brand-header">
        <div class="logo-box">{{ tenantName.charAt(0) }}</div>
        <h1 class="brand-title">{{ tenantName }}</h1>
        <p class="text-muted" v-if="isCustomDomain">Secure Portal Access</p>
        <p class="text-muted" v-else>Callsign User Portal</p>
      </div>

      <form @submit.prevent="handleLogin">
        <div v-if="errorMessage" class="error-banner">
          {{ errorMessage }}
        </div>

        <div class="form-group">
          <label>Extension / Email</label>
          <input 
            v-model="username"
            type="text" 
            class="input-field" 
            placeholder="1001" 
            autofocus
            :disabled="isLoading"
          >
        </div>

        <div class="form-group">
          <label>Password / PIN</label>
          <input 
            v-model="password"
            type="password" 
            class="input-field" 
            placeholder="••••••"
            :disabled="isLoading"
          >
        </div>

        <button 
          type="submit"
          class="btn-primary full-width" 
          :disabled="isLoading || !username || !password"
        >
          <span v-if="isLoading" class="spinner"></span>
          <span v-else>Sign In</span>
        </button>
      </form>

      <div class="footer-links">
        <a href="#">Forgot Password?</a>
      </div>
      
      <div v-if="domainInfo" class="domain-badge">
         Tenant: {{ domainInfo }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '@/services/auth'

const router = useRouter()
const auth = useAuth()

const username = ref('')
const password = ref('')
const isLoading = ref(false)
const errorMessage = ref('')
const tenantName = ref('Callsign')
const isCustomDomain = ref(false)
const domainInfo = ref('')

onMounted(() => {
  // Check if already logged in
  if (auth.state.isAuthenticated) {
    router.push('/dialer')
    return
  }
  
  // Domain Detection for white-label
  const host = window.location.hostname
  const urlParams = new URLSearchParams(window.location.search)
  const domainParam = urlParams.get('domain')

  if (domainParam || (host !== 'localhost' && !host.includes('127.0.0.1'))) {
    isCustomDomain.value = true
    tenantName.value = domainParam ? domainParam.split('.')[0] : host.split('.')[0]
    tenantName.value = tenantName.value.charAt(0).toUpperCase() + tenantName.value.slice(1)
    domainInfo.value = domainParam || host
  }
})

const handleLogin = async () => {
  if (!username.value || !password.value) return
  
  isLoading.value = true
  errorMessage.value = ''
  
  try {
    const result = await auth.login(username.value, password.value)
    
    if (result.success) {
      // Redirect based on role
      if (auth.hasRole(['system_admin', 'tenant_admin'])) {
        router.push('/admin')
      } else {
        router.push('/dialer')
      }
    } else {
      errorMessage.value = result.error || 'Invalid credentials'
    }
  } catch (error) {
    errorMessage.value = error.message || 'Login failed. Please try again.'
  } finally {
    isLoading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
}

.login-card {
  background: #1e293b;
  padding: 40px;
  border-radius: 12px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
  width: 100%;
  max-width: 380px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.brand-header {
  text-align: center;
  margin-bottom: 32px;
}

.logo-box {
  width: 56px;
  height: 56px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
  font-size: 28px;
  font-weight: bold;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
}

.brand-title {
  color: #f8fafc;
  font-size: 1.75rem;
  font-weight: 700;
  margin-bottom: 4px;
}

.error-banner {
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid #ef4444;
  color: #fca5a5;
  padding: 12px;
  border-radius: 8px;
  margin-bottom: 16px;
  font-size: 14px;
  text-align: center;
}

.form-group {
  margin-bottom: 20px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

label {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: #94a3b8;
}

.input-field {
  padding: 14px 16px;
  border: 1px solid #334155;
  border-radius: 8px;
  font-size: 15px;
  outline: none;
  background: #0f172a;
  color: #f8fafc;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.input-field:focus { 
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.2);
}

.input-field:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
  border: none;
  padding: 14px;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  margin-top: 8px;
  font-size: 15px;
  transition: transform 0.2s, box-shadow 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 10px 20px -5px rgba(99, 102, 241, 0.4);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.full-width { width: 100%; }

.spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.footer-links {
  margin-top: 24px;
  text-align: center;
  font-size: 13px;
}

.footer-links a { 
  color: #94a3b8; 
  text-decoration: none;
  transition: color 0.2s;
}

.footer-links a:hover { color: #6366f1; }

.domain-badge {
  margin-top: 24px;
  text-align: center;
  font-size: 11px;
  color: #64748b;
  background: rgba(0, 0, 0, 0.2);
  padding: 8px 12px;
  border-radius: 6px;
}

.text-muted { color: #94a3b8; font-size: 14px; }
</style>

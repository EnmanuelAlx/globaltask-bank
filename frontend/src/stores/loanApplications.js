import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'
import { useAuthStore } from './auth'

export const useLoanStore = defineStore('loanApplications', () => {
  const applications = ref([])
  const loading = ref(false)
  const authStore = useAuthStore()
  let socket = null

  const approvedCount = computed(() => 
    applications.value.filter(a => a.status === 'APPROVED').length
  )
  const pendingCount = computed(() => 
    applications.value.filter(a => !['APPROVED', 'REJECTED'].includes(a.status)).length
  )

  const api = axios.create({
    baseURL: '/api/v1',
  })

  // Set up interceptor to use the latest token from authStore
  api.interceptors.request.use(config => {
    const auth = useAuthStore()
    const token = auth.accessToken
    
    console.log('[API Request] Token found:', token ? 'Yes (starts with ' + token.substring(0, 10) + '...)' : 'No')
    
    if (token) {
      config.headers.set('Authorization', `Bearer ${token}`)
      // Fallback for older axios or specific configs
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  }, error => {
    return Promise.reject(error)
  })

  const fetchApplications = async (filters = {}) => {
    const auth = useAuthStore()
    if (!auth.accessToken) {
      console.warn('Cannot fetch applications: No access token found')
      return
    }
    
    loading.value = true
    try {
      const params = new URLSearchParams()
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== null && value !== undefined && value !== '') {
          params.append(key, value)
        }
      })

      const resp = await api.get(`/applications?${params.toString()}`)
      applications.value = resp.data.data || []
    } catch (err) {
      if (err.response?.status === 401) {
        authStore.logout()
      }
      throw err
    } finally {
      loading.value = false
    }
  }

  const submitApplication = async (formData) => {
    loading.value = true
    try {
      const resp = await api.post('/applications', {
        borrower_name: formData.name,
        country_id: formData.country_id,
        identity_document: formData.id_doc,
        requested_amount: parseFloat(formData.amount),
        monthly_income: parseFloat(formData.income)
      })
      return resp.data
    } catch (err) {
      if (err.response?.status === 401) {
        const auth = useAuthStore()
        auth.logout()
      }
      throw err
    } finally {
      loading.value = false
    }
  }

  const connectWebSocket = () => {
    if (!authStore.accessToken || socket) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/v1/ws?token=${authStore.accessToken}`
    socket = new WebSocket(wsUrl)

    socket.onmessage = (event) => {
      try {
        const updatedApp = JSON.parse(event.data)
        const index = applications.value.findIndex(a => a.id === updatedApp.id)
        if (index !== -1) {
          applications.value[index] = updatedApp
        } else {
          applications.value.unshift(updatedApp)
        }
      } catch (err) {
        console.error('WS Error:', err)
      }
    }

    socket.onclose = () => {
      socket = null
      // Auto-reconnect if still authenticated
      if (authStore.accessToken) {
        setTimeout(connectWebSocket, 3000)
      }
    }
  }

  const disconnectWebSocket = () => {
    if (socket) {
      socket.close()
      socket = null
    }
  }

  return {
    applications,
    loading,
    approvedCount,
    pendingCount,
    fetchApplications,
    submitApplication,
    connectWebSocket,
    disconnectWebSocket
  }
})

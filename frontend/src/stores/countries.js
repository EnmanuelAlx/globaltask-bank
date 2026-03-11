import { defineStore } from 'pinia'
import { ref } from 'vue'
import axios from 'axios'

export const useCountryStore = defineStore('countries', () => {
  const countries = ref([])
  const loading = ref(false)

  const api = axios.create({
    baseURL: '/api/v1',
  })

  // Add interceptor for auth token (copied from loan store for simplicity, 
  // though a shared axios instance would be better in a larger app)
  api.interceptors.request.use(config => {
    const token = localStorage.getItem('sb-access-token')
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  })

  const fetchCountries = async () => {
    loading.value = true
    try {
      const resp = await api.get('/countries')
      countries.value = resp.data.data || []
    } catch (err) {
      console.error('Failed to fetch countries:', err)
    } finally {
      loading.value = false
    }
  }

  const getCountryFlag = (isoCode) => {
    const flags = {
      'PT': '🇵🇹',
      'CO': '🇨🇴',
      'MX': '🇲🇽',
      'CL': '🇨🇱'
    }
    return flags[isoCode] || '🌍'
  }

  return {
    countries,
    loading,
    fetchCountries,
    getCountryFlag
  }
})

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { supabase } from '../lib/supabase'

export const useAuthStore = defineStore('auth', () => {
  const session = ref(null)
  const user = ref(null)
  const loading = ref(false)

  const isAuthenticated = computed(() => !!session.value || !!localStorage.getItem('sb-access-token'))
  const accessToken = computed(() => session.value?.access_token || localStorage.getItem('sb-access-token'))

  const login = async (email, password) => {
    loading.value = true
    try {
      const { data, error } = await supabase.auth.signInWithPassword({
        email,
        password
      })
      if (error) throw error
      session.value = data.session
      user.value = data.user
      
      // Explicitly save to localStorage as requested
      if (data.session?.access_token) {
        localStorage.setItem('sb-access-token', data.session.access_token)
      }
      
      return data
    } finally {
      loading.value = false
    }
  }

  const logout = async () => {
    await supabase.auth.signOut()
    session.value = null
    user.value = null
    localStorage.removeItem('sb-access-token')
  }

  const fetchSession = async () => {
    loading.value = true
    try {
      const { data, error } = await supabase.auth.getSession()
      if (error) throw error
      session.value = data.session
      user.value = data.session?.user || null
      
      if (data.session?.access_token) {
        localStorage.setItem('sb-access-token', data.session.access_token)
      }
    } finally {
      loading.value = false
    }
  }

  // Initialize listener for auth changes
  const init = () => {
    supabase.auth.onAuthStateChange((_event, _session) => {
      session.value = _session
      user.value = _session?.user || null
      if (_session?.access_token) {
        localStorage.setItem('sb-access-token', _session.access_token)
      } else {
        localStorage.removeItem('sb-access-token')
      }
    })
  }

  return {
    session,
    user,
    loading,
    isAuthenticated,
    accessToken,
    login,
    logout,
    fetchSession,
    init
  }
})

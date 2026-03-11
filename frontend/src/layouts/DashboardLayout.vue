<template>
  <div class="flex h-screen bg-slate-50 font-sans text-slate-900 overflow-hidden">
    <!-- Sidebar -->
    <aside class="w-64 bg-slate-900 text-white flex-shrink-0 flex flex-col hidden md:flex">
      <div class="p-6 border-b border-slate-800">
        <div class="flex items-center gap-3">
          <div class="w-8 h-8 bg-blue-500 rounded-lg flex items-center justify-center">
            <span class="font-bold text-lg">G</span>
          </div>
          <span class="font-bold text-xl tracking-tight">GlobalBank</span>
        </div>
      </div>
      <nav class="flex-1 p-4 space-y-2 overflow-y-auto text-slate-300">
        <router-link to="/" class="flex items-center gap-3 p-3 rounded-lg hover:bg-slate-800 transition-all" active-class="bg-blue-600 text-white shadow-lg shadow-blue-900/20">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" /></svg>
          <span class="font-medium">Dashboard</span>
        </router-link>
        
        <a href="#" @click.prevent="handleLogout" class="flex items-center gap-3 p-3 rounded-lg text-slate-400 hover:bg-red-900/20 hover:text-red-400 transition-all mt-auto group">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" /></svg>
          <span class="font-medium">Sign Out</span>
        </a>
      </nav>
      <div class="p-6 border-t border-slate-800 text-slate-500 text-xs text-center">
        &copy; 2026 GlobalTask Bank v1.0
      </div>
    </aside>

    <div class="flex-1 flex flex-col min-w-0 overflow-hidden text-slate-900">
      <!-- Header -->
      <header class="h-16 bg-white border-b border-slate-200 flex items-center justify-between px-8 flex-shrink-0">
        <h2 class="text-xl font-bold text-slate-800">{{ $route.meta.title || 'Dashboard' }}</h2>
        <div class="flex items-center gap-4">
          <div class="flex items-center gap-2 px-3 py-1 bg-slate-100 rounded-full">
            <div class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
            <span class="text-xs font-semibold text-slate-600">WebSocket Connected</span>
          </div>
          <slot name="header-actions"></slot>
        </div>
      </header>

      <!-- Dashboard Content -->
      <main class="flex-1 overflow-y-auto p-8">
        <slot></slot>
      </main>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useLoanStore } from '../stores/loanApplications'

const authStore = useAuthStore()
const loanStore = useLoanStore()
const router = useRouter()

const handleLogout = async () => {
  await authStore.logout()
  loanStore.disconnectWebSocket()
  loanStore.applications = []
  router.push('/login')
}
</script>

<template>
  <div class="flex-1 flex items-center justify-center bg-slate-900 min-h-screen">
    <div class="bg-white p-10 rounded-3xl shadow-2xl w-full max-w-md transform transition-all">
      <div class="flex flex-col items-center mb-8">
        <div class="w-16 h-16 bg-blue-600 rounded-2xl flex items-center justify-center mb-4 shadow-lg shadow-blue-600/30">
          <span class="text-white font-black text-2xl">G</span>
        </div>
        <h1 class="text-3xl font-black text-slate-800">Welcome Back</h1>
        <p class="text-slate-500 mt-2">Sign in to your banking dashboard</p>
      </div>

      <form @submit.prevent="handleLogin" class="space-y-6">
        <BaseInput
          label="Email Address"
          type="email"
          v-model="loginForm.email"
          required
          placeholder="user@example.com"
        />
        <BaseInput
          label="Password"
          type="password"
          v-model="loginForm.password"
          required
          placeholder="••••••••"
        />
        
        <BaseButton
          type="submit"
          :loading="authStore.loading"
          class="w-full py-4 rounded-xl"
          label="Sign In"
        />
      </form>
      
      <p class="mt-8 text-center text-slate-400 text-sm italic">
        Running on Local Supabase Instance
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import BaseButton from '../components/BaseButton.vue'
import BaseInput from '../components/BaseInput.vue'

const authStore = useAuthStore()
const router = useRouter()
const loginForm = ref({ email: '', password: '' })

const handleLogin = async () => {
  try {
    await authStore.login(loginForm.value.email, loginForm.value.password)
    router.push('/')
  } catch (err) {
    alert("Login Error: " + err.message)
  }
}
</script>

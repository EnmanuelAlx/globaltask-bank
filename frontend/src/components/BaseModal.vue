<template>
  <Transition name="fade">
    <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm" @click="$emit('close')"></div>
      <div class="bg-white w-full max-w-xl rounded-3xl shadow-2xl shadow-slate-900/40 overflow-hidden relative transform transition-all scale-100">
        <div class="p-8 border-b border-slate-100 flex items-center justify-between bg-slate-50">
          <div>
            <h3 class="text-2xl font-black text-slate-800">{{ title }}</h3>
            <p v-if="subtitle" class="text-slate-500 text-sm mt-1">{{ subtitle }}</p>
          </div>
          <button @click="$emit('close')" class="p-2 hover:bg-slate-200 rounded-full transition-colors">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        
        <div class="p-8 space-y-6">
          <slot></slot>
        </div>

        <div v-if="$slots.footer" class="p-8 bg-slate-50 flex gap-4">
          <slot name="footer"></slot>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
defineProps({
  show: {
    type: Boolean,
    default: false
  },
  title: String,
  subtitle: String
})

defineEmits(['close'])
</script>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.3s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>

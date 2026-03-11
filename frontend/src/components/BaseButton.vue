<template>
  <button
    :type="type"
    :disabled="disabled || loading"
    :class="[
      'font-bold transition-all flex items-center justify-center gap-3 disabled:opacity-50',
      variantClasses[variant]
    ]"
  >
    <span
      v-if="loading"
      class="w-5 h-5 border-4 border-white/30 border-t-white rounded-full animate-spin"
    ></span>
    <slot>{{ label }}</slot>
  </button>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  type: {
    type: String,
    default: 'button'
  },
  variant: {
    type: String,
    default: 'primary',
    validator: (value) => ['primary', 'secondary', 'danger', 'ghost'].includes(value)
  },
  loading: {
    type: Boolean,
    default: false
  },
  disabled: {
    type: Boolean,
    default: false
  },
  label: {
    type: String,
    default: ''
  }
})

const variantClasses = {
  primary: 'bg-blue-600 text-white hover:bg-blue-700 shadow-xl shadow-blue-600/30',
  secondary: 'bg-white text-slate-600 border-2 border-slate-200 hover:bg-slate-100',
  danger: 'bg-red-600 text-white hover:bg-red-700 shadow-xl shadow-red-600/30',
  ghost: 'text-slate-400 hover:bg-red-900/20 hover:text-red-400'
}
</script>

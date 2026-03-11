<template>
  <span
    :class="[
      'px-3 py-1 rounded-full text-xs font-bold ring-1 ring-inset',
      statusClasses[status] || 'bg-slate-50 text-slate-700 ring-slate-600/20'
    ]"
  >
    {{ label || status.replace('_', ' ') }}
  </span>
</template>

<script setup>
const props = defineProps({
  status: {
    type: String,
    required: true
  },
  label: String
})

const statusClasses = {
  APPROVED: 'bg-green-50 text-green-700 ring-green-600/20',
  REJECTED: 'bg-red-50 text-red-700 ring-red-600/20',
  PENDING_VALIDATION: 'bg-blue-50 text-blue-700 ring-blue-600/20 animate-pulse',
  AWAITING_BANK_DATA: 'bg-blue-50 text-blue-700 ring-blue-600/20 animate-pulse',
  ANALYZING_RISK: 'bg-blue-50 text-blue-700 ring-blue-600/20 animate-pulse'
}
</script>

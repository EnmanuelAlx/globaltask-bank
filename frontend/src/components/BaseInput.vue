<template>
  <div class="space-y-2">
    <label v-if="label" class="block text-sm font-bold text-slate-700">
      {{ label }}
    </label>
    <div class="relative">
      <span
        v-if="prefix"
        class="absolute left-4 top-3.5 text-slate-400 font-bold"
      >
        {{ prefix }}
      </span>
      <input
        :type="type"
        :value="modelValue"
        @input="$emit('update:modelValue', $event.target.value)"
        :placeholder="placeholder"
        :required="required"
        :class="[
          'w-full bg-slate-50 border-2 border-slate-100 rounded-xl outline-none focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 transition-all font-medium text-slate-900 placeholder-slate-300',
          sizeClasses[size],
          prefix ? 'pl-8 pr-4' : 'p-4'
        ]"
      />
    </div>
  </div>
</template>

<script setup>
defineProps({
  modelValue: [String, Number],
  label: String,
  type: {
    type: String,
    default: 'text'
  },
  placeholder: String,
  required: {
    type: Boolean,
    default: false
  },
  prefix: String,
  size: {
    type: String,
    default: 'md'
  }
})

defineEmits(['update:modelValue'])

const sizeClasses = {
  sm: 'p-3',
  md: 'p-4'
}
</script>

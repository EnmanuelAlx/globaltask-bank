<template>
  <DashboardLayout>
    <template #header-actions>
      <BaseButton @click="showModal = true" class="px-5 py-2 rounded-lg" label="New Application">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
      </BaseButton>
    </template>

    <!-- Stats Grid -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8 text-slate-900">
      <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm">
        <div class="text-slate-500 text-sm font-medium mb-1">Total Applications</div>
        <div class="text-3xl font-bold text-slate-900">{{ loanStore.applications.length }}</div>
      </div>
      <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm">
        <div class="text-slate-500 text-sm font-medium mb-1 text-green-600">Approved</div>
        <div class="text-3xl font-bold text-green-600">{{ loanStore.approvedCount }}</div>
      </div>
      <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm">
        <div class="text-slate-500 text-sm font-medium mb-1 text-blue-600">Pending</div>
        <div class="text-3xl font-bold text-blue-600">{{ loanStore.pendingCount }}</div>
      </div>
    </div>

    <!-- Filters Bar -->
    <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm mb-8 text-slate-900">
      <div class="flex items-center justify-between mb-4">
        <h3 class="font-bold text-slate-800 flex items-center gap-2">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" /></svg>
          Advanced Filters
        </h3>
        <button @click="resetFilters" class="text-sm text-blue-600 hover:text-blue-800 font-medium">Reset All</button>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 gap-4">
        <BaseInput v-model="filters.borrower_name" label="Name" placeholder="Search name..." size="sm" />
        <BaseInput v-model="filters.identity_document" label="ID Document" placeholder="Exact ID..." size="sm" />
        <div class="space-y-1">
          <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider">Status</label>
          <select v-model="filters.status" class="w-full bg-slate-50 border border-slate-200 rounded-lg p-2 text-sm outline-none focus:ring-2 focus:ring-blue-500/20 transition-all">
            <option value="">All Statuses</option>
            <option value="DRAFT">Draft</option>
            <option value="PENDING_VALIDATION">Pending Validation</option>
            <option value="AWAITING_BANK_DATA">Awaiting Bank Data</option>
            <option value="VALIDATE_USER_IDENTITY">Identity Check</option>
            <option value="ANALYZING_RISK">Analyzing Risk</option>
            <option value="APPROVED">Approved</option>
            <option value="REJECTED">Rejected</option>
          </select>
        </div>
        <div class="space-y-1">
          <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider">Country</label>
          <select v-model="filters.country_id" class="w-full bg-slate-50 border border-slate-200 rounded-lg p-2 text-sm outline-none focus:ring-2 focus:ring-blue-500/20 transition-all">
            <option value="">All Countries</option>
            <option v-for="c in countryStore.countries" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </div>
        <BaseInput v-model="filters.min_amount" type="number" label="Min Amount" placeholder="0" size="sm" />
        <BaseInput v-model="filters.max_amount" type="number" label="Max Amount" placeholder="Max" size="sm" />
        <BaseInput v-model="filters.min_income" type="number" label="Min Income" placeholder="0" size="sm" />
        <BaseInput v-model="filters.max_income" type="number" label="Max Income" placeholder="Max" size="sm" />
      </div>
    </div>

    <!-- Table Card -->
    <div class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden text-slate-900">
      <div class="p-6 border-b border-slate-100 flex items-center justify-between">
        <h3 class="font-bold text-lg text-slate-900">Recent Applications</h3>
        <div class="text-slate-400 text-sm italic">Updates in real-time</div>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-left">
          <thead>
            <tr class="bg-slate-50 text-slate-500 text-xs uppercase tracking-wider font-bold">
              <th class="px-6 py-4">Borrower</th>
              <th class="px-6 py-4">Country</th>
              <th class="px-6 py-4">Amount</th>
              <th class="px-6 py-4">Income</th>
              <th class="px-6 py-4">Status</th>
              <th class="px-6 py-4 text-right">Date</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-for="app in loanStore.applications" :key="app.id" class="hover:bg-slate-50 transition-colors group">
              <td class="px-6 py-4">
                <div class="flex items-center gap-3 text-slate-900">
                  <div class="w-10 h-10 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center font-bold">
                    {{ app.borrower_name.charAt(0) }}
                  </div>
                  <div>
                    <div class="font-bold text-slate-800">{{ app.borrower_name }}</div>
                    <div class="text-xs text-slate-400 font-mono tracking-tight">{{ app.identity_document }}</div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4">
                <span class="flex items-center gap-2 text-slate-600">
                  <span>{{ getCountryFlag(app.country_id) }}</span>
                  <span class="font-medium">
                    {{ getCountryName(app.country_id) }}
                  </span>
                </span>
              </td>
              <td class="px-6 py-4 font-bold text-slate-700">
                {{ formatCurrency(app.requested_amount) }}
              </td>
              <td class="px-6 py-4 text-slate-500 font-medium">
                {{ formatCurrency(app.monthly_income) }}/mo
              </td>
              <td class="px-6 py-4">
                <StatusBadge :status="app.status" />
              </td>
              <td class="px-6 py-4 text-right text-slate-400 text-sm font-medium">
                {{ formatDate(app.created_at) }}
              </td>
            </tr>
            <tr v-if="loanStore.applications.length === 0">
              <td colspan="6" class="px-6 py-20 text-center text-slate-400 font-medium italic">
                No active applications found. Start your first request.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modern Modal -->
    <BaseModal
      :show="showModal"
      @close="showModal = false"
      title="New Credit Request"
      subtitle="Complete all fields to start validation."
    >
      <BaseInput
        label="Full Name of Borrower"
        v-model="form.name"
        placeholder="John Doe"
        size="sm"
      />

      <div class="grid grid-cols-2 gap-6">
        <div class="space-y-2 text-slate-900">
          <label class="block text-sm font-bold text-slate-700">Country</label>
          <select class="w-full bg-slate-50 border-2 border-slate-100 rounded-xl p-3 focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 outline-none transition-all font-medium appearance-none text-slate-900" v-model="form.country_id">
            <option v-for="c in countryStore.countries" :key="c.id" :value="c.id">
              {{ countryStore.getCountryFlag(c.iso_code) }} {{ c.name }}
            </option>
          </select>
        </div>
        <BaseInput
          label="Identity Document"
          v-model="form.id_doc"
          :placeholder="selectedCountry.iso_code === 'PT' ? 'NIF (9 digits)' : 'ID Document'"
          size="sm"
        />
      </div>

      <div class="grid grid-cols-2 gap-6 text-slate-900">
        <BaseInput
          label="Loan Amount"
          v-model="form.amount"
          type="number"
          prefix="$"
          size="sm"
        />
        <BaseInput
          label="Monthly Income"
          v-model="form.income"
          type="number"
          prefix="$"
          size="sm"
        />
      </div>

      <template #footer>
        <BaseButton
          @click="showModal = false"
          variant="secondary"
          class="flex-1 py-4 rounded-xl"
          label="Cancel"
        />
        <BaseButton
          @click="handleSubmit"
          :loading="loanStore.loading"
          class="flex-[2] py-4 rounded-xl"
          label="Send Application"
        />
      </template>
    </BaseModal>
  </DashboardLayout>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import { useLoanStore } from '../stores/loanApplications'
import { useCountryStore } from '../stores/countries'
import DashboardLayout from '../layouts/DashboardLayout.vue'
import BaseButton from '../components/BaseButton.vue'
import BaseInput from '../components/BaseInput.vue'
import StatusBadge from '../components/StatusBadge.vue'
import BaseModal from '../components/BaseModal.vue'

const loanStore = useLoanStore()
const countryStore = useCountryStore()

const showModal = ref(false)
const filters = ref({
  borrower_name: '',
  identity_document: '',
  status: '',
  country_id: '',
  min_amount: '',
  max_amount: '',
  min_income: '',
  max_income: ''
})

const applyFilters = () => {
  loanStore.fetchApplications(filters.value)
}

const resetFilters = () => {
  filters.value = {
    borrower_name: '',
    identity_document: '',
    status: '',
    country_id: '',
    min_amount: '',
    max_amount: '',
    min_income: '',
    max_income: ''
  }
  applyFilters()
}

// Watch filters with debounce
let debounceTimer = null
watch(filters, () => {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    applyFilters()
  }, 500)
}, { deep: true })

const form = ref({
  name: '',
  country_id: 1,
  id_doc: '',
  amount: 0,
  income: 0
})

const selectedCountry = computed(() => {
  return countryStore.countries.find(c => c.id === form.value.country_id) || { iso_code: 'PT' }
})

const handleSubmit = async () => {
  if (!form.value.id_doc || form.value.amount <= 0) {
    alert("Please fill all fields correctly.")
    return
  }
  try {
    const payload = {
      ...form.value,
      country: selectedCountry.value.iso_code // Keep for backend compatibility if needed
    }
    await loanStore.submitApplication(payload)
    showModal.value = false
    form.value = { name: '', country_id: 1, id_doc: '', amount: 0, income: 0 }
  } catch (err) {
    alert("Error: " + (err.response?.data?.error || err.message))
  }
}

const formatCurrency = (val) => new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(val)
const formatDate = (val) => new Date(val).toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })

const getCountryName = (id) => {
  const c = countryStore.countries.find(c => c.id === id)
  return c ? c.name : 'Unknown'
}

const getCountryFlag = (id) => {
  const c = countryStore.countries.find(c => c.id === id)
  return c ? countryStore.getCountryFlag(c.iso_code) : '🌍'
}

onMounted(async () => {
  await Promise.all([
    loanStore.fetchApplications(),
    countryStore.fetchCountries()
  ])
  loanStore.connectWebSocket()
})

onUnmounted(() => {
  loanStore.disconnectWebSocket()
})
</script>

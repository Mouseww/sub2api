<template>
  <div class="space-y-6">
    <!-- Step 1: Network Selection (if not yet selected) -->
    <template v-if="!selectedNetwork">
      <div class="card p-6">
        <div class="mb-4 flex items-center gap-3">
          <div class="flex h-12 w-12 items-center justify-center rounded-full bg-[#C4612F]">
            <svg class="h-6 w-6 text-white" viewBox="0 0 32 32" fill="currentColor">
              <path d="M16 32C7.163 32 0 24.837 0 16S7.163 0 16 0s16 7.163 16 16-7.163 16-16 16zm7.189-17.98c-.314 4.004-3.666 5.281-7.293 5.281-3.627 0-6.979-1.277-7.293-5.281h14.586zm-7.293-1.918c-2.977 0-5.434-.636-6.562-1.538v-.003l-.002-.001c-.196-.156-.301-.31-.301-.464 0-.308.303-.618.907-.918 1.125-.558 3.08-.917 5.958-.917s4.833.359 5.958.917c.604.3.907.61.907.918 0 .154-.105.308-.301.464l-.002.001v.003c-1.128.902-3.585 1.538-6.562 1.538zm0-3.84c-3.115 0-5.805.692-7.189 1.706v-1.705h14.378v1.705c-1.384-1.014-4.074-1.706-7.189-1.706z"/>
            </svg>
          </div>
          <div>
            <h2 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('payment.usdt.title') }}</h2>
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('payment.usdt.subtitle') }}</p>
          </div>
        </div>
      </div>

      <USDTNetworkSelector
        :networks="networks"
        :selected-network="selectedNetwork"
        @update:selected-network="selectedNetwork = $event"
      />

      <button
        :disabled="!selectedNetwork || submitting"
        :class="[
          'btn w-full py-3 text-base font-medium',
          selectedNetwork && !submitting
            ? 'bg-[#C4612F] text-white hover:bg-[#A94E22]'
            : 'cursor-not-allowed bg-gray-300 text-gray-500 dark:bg-dark-700 dark:text-gray-500'
        ]"
        @click="handleCreateOrder"
      >
        <span v-if="submitting" class="flex items-center justify-center gap-2">
          <span class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
          {{ t('common.processing') }}
        </span>
        <span v-else>{{ t('payment.usdt.continueToPayment') }}</span>
      </button>
    </template>

    <!-- Step 2: Payment Panel (after order created) -->
    <template v-else-if="currentOrder">
      <USDTPaymentPanel
        :order="currentOrder"
        :qr-data="qrData"
        @cancel="handleCancel"
        @refresh="handleRefresh"
      />
    </template>

    <!-- Loading State -->
    <div v-else-if="loading" class="flex items-center justify-center py-20">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[#C4612F] border-t-transparent"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { usdtAPI } from '@/api/usdt'
import USDTNetworkSelector from './USDTNetworkSelector.vue'
import USDTPaymentPanel from './USDTPaymentPanel.vue'
import type { USDTNetworkConfig, USDTNetwork, USDTPaymentOrder } from '@/types/usdt'

const props = defineProps<{
  amount: number
}>()

const emit = defineEmits<{
  success: [orderId: number]
  cancel: []
}>()

const { t } = useI18n()

const networks = ref<USDTNetworkConfig[]>([])
const selectedNetwork = ref<USDTNetwork | null>(null)
const currentOrder = ref<USDTPaymentOrder | null>(null)
const qrData = ref('')
const loading = ref(false)
const submitting = ref(false)
let pollInterval: ReturnType<typeof setInterval> | null = null

async function loadNetworks() {
  loading.value = true
  try {
    const response = await usdtAPI.getNetworks()
    networks.value = response.data.sort((a, b) => a.priority - b.priority)
    
    // Auto-select TRC20 if available
    const trc20 = networks.value.find(n => n.network_id === 'TRC20' && n.enabled)
    if (trc20) {
      selectedNetwork.value = 'TRC20'
    }
  } catch (error) {
    console.error('Failed to load USDT networks:', error)
  } finally {
    loading.value = false
  }
}

async function handleCreateOrder() {
  if (!selectedNetwork.value || submitting.value) return

  submitting.value = true
  try {
    const response = await usdtAPI.createOrder({
      amount_usd: props.amount,
      network: selectedNetwork.value
    })
    
    currentOrder.value = response.data.order
    qrData.value = response.data.qr_data
    
    // Start polling for status updates
    startPolling()
  } catch (error: any) {
    console.error('Failed to create USDT order:', error)
    alert(error.response?.data?.message || t('payment.usdt.createOrderFailed'))
  } finally {
    submitting.value = false
  }
}

async function handleRefresh() {
  if (!currentOrder.value) return

  try {
    const response = await usdtAPI.getOrderStatus(currentOrder.value.id)
    currentOrder.value = response.data.order
    
    // Check if order is completed
    if (currentOrder.value.status === 'COMPLETED') {
      stopPolling()
      emit('success', currentOrder.value.id)
    } else if (currentOrder.value.status === 'EXPIRED' || currentOrder.value.status === 'FAILED') {
      stopPolling()
    }
  } catch (error) {
    console.error('Failed to refresh order status:', error)
  }
}

function handleCancel() {
  stopPolling()
  if (currentOrder.value && currentOrder.value.status === 'PENDING') {
    usdtAPI.cancelOrder(currentOrder.value.id).catch(err => {
      console.error('Failed to cancel order:', err)
    })
  }
  emit('cancel')
}

function startPolling() {
  // Poll every 5 seconds
  pollInterval = setInterval(() => {
    handleRefresh()
  }, 5000)
}

function stopPolling() {
  if (pollInterval) {
    clearInterval(pollInterval)
    pollInterval = null
  }
}

onMounted(() => {
  loadNetworks()
})

onUnmounted(() => {
  stopPolling()
})
</script>

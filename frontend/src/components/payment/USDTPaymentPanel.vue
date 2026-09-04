<template>
  <div class="card space-y-6 p-6">
    <!-- Deposit Address Section -->
    <div class="space-y-3">
      <div class="flex items-center justify-between">
        <h3 class="text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('payment.usdt.depositAddress') }}
        </h3>
        <span
          class="rounded-full px-2.5 py-0.5 text-xs font-semibold"
          :class="[
            order.network === 'TRC20'
              ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
              : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
          ]"
        >
          {{ order.network }}
        </span>
      </div>

      <!-- QR Code -->
      <div class="flex justify-center">
        <div class="relative rounded-2xl border-4 border-[#E7E1D7] bg-white p-6 shadow-sm dark:border-dark-600 dark:bg-dark-800">
          <canvas ref="qrCanvas" class="mx-auto"></canvas>
          <!-- Tether Logo Overlay -->
          <div class="pointer-events-none absolute inset-0 flex items-center justify-center">
            <div class="rounded-full bg-white p-2.5 shadow-lg ring-2 ring-white dark:bg-dark-800 dark:ring-dark-800">
              <svg class="h-7 w-7 text-[#26A17B]" viewBox="0 0 32 32" fill="currentColor">
                <path d="M16 32C7.163 32 0 24.837 0 16S7.163 0 16 0s16 7.163 16 16-7.163 16-16 16zm7.189-17.98c-.314 4.004-3.666 5.281-7.293 5.281-3.627 0-6.979-1.277-7.293-5.281h14.586zm-7.293-1.918c-2.977 0-5.434-.636-6.562-1.538v-.003l-.002-.001c-.196-.156-.301-.31-.301-.464 0-.308.303-.618.907-.918 1.125-.558 3.08-.917 5.958-.917s4.833.359 5.958.917c.604.3.907.61.907.918 0 .154-.105.308-.301.464l-.002.001v.003c-1.128.902-3.585 1.538-6.562 1.538zm0-3.84c-3.115 0-5.805.692-7.189 1.706v-1.705h14.378v1.705c-1.384-1.014-4.074-1.706-7.189-1.706z"/>
              </svg>
            </div>
          </div>
        </div>
      </div>

      <!-- Address Display with Copy -->
      <div class="space-y-2">
        <div class="flex items-center gap-2 rounded-lg bg-gray-50 p-3 dark:bg-dark-800">
          <code class="flex-1 break-all font-mono text-sm text-gray-900 dark:text-white">
            {{ order.deposit_address }}
          </code>
          <button
            type="button"
            :class="[
              'shrink-0 rounded-lg px-3 py-2 text-xs font-medium transition-colors',
              addressCopied
                ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300 dark:bg-dark-700 dark:text-gray-300 dark:hover:bg-dark-600'
            ]"
            @click="copyAddress"
          >
            {{ addressCopied ? t('common.copied') : t('common.copy') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Amount Section -->
    <div class="space-y-3 border-t border-gray-200 pt-6 dark:border-dark-600">
      <h3 class="text-sm font-medium text-gray-700 dark:text-gray-300">
        {{ t('payment.usdt.amountToSend') }}
      </h3>

      <div class="flex items-center gap-2 rounded-lg bg-[#F2E3D6] p-4 dark:bg-[#C4612F]/10">
        <div class="flex-1">
          <div class="text-3xl font-bold text-[#C4612F] dark:text-[#C4612F]">
            {{ order.amount_usdt.toFixed(6) }} USDT
          </div>
          <div class="mt-1 text-sm text-gray-600 dark:text-gray-400">
            ≈ ${{ order.amount_usd.toFixed(2) }} USD
          </div>
        </div>
        <button
          type="button"
          :class="[
            'shrink-0 rounded-lg px-3 py-2 text-xs font-medium transition-colors',
            amountCopied
              ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
              : 'bg-[#C4612F] text-white hover:bg-[#A94E22]'
          ]"
          @click="copyAmount"
        >
          {{ amountCopied ? t('common.copied') : t('common.copy') }}
        </button>
      </div>

      <!-- Amount Warning -->
      <div class="rounded-lg border border-amber-200 bg-amber-50 p-3 dark:border-amber-500/30 dark:bg-amber-900/10">
        <div class="flex gap-2">
          <svg class="h-5 w-5 shrink-0 text-amber-600 dark:text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <div class="flex-1 space-y-1 text-xs leading-relaxed text-amber-900 dark:text-amber-300">
            <p class="font-semibold">{{ t('payment.usdt.amountWarningTitle') }}</p>
            <ul class="ml-4 list-disc space-y-0.5">
              <li>{{ t('payment.usdt.amountWarningUnderpay') }}</li>
              <li>{{ t('payment.usdt.amountWarningOverpay') }}</li>
              <li>{{ t('payment.usdt.amountWarningWrongNetwork') }}</li>
            </ul>
          </div>
        </div>
      </div>
    </div>

    <!-- Status Section -->
    <div class="space-y-3 border-t border-gray-200 pt-6 dark:border-dark-600">
      <h3 class="text-sm font-medium text-gray-700 dark:text-gray-300">
        {{ t('payment.usdt.paymentStatus') }}
      </h3>

      <!-- Status Badge -->
      <div class="flex items-center justify-between">
        <span :class="['inline-flex items-center gap-2 rounded-full px-3 py-1 text-sm font-medium', statusClass]">
          <span
            v-if="order.status === 'PENDING' || order.status === 'DETECTED' || order.status === 'CONFIRMING'"
            class="h-2 w-2 animate-pulse rounded-full bg-current"
          ></span>
          {{ statusText }}
        </span>
        <span class="text-sm tabular-nums text-gray-600 dark:text-gray-400">
          {{ t('payment.usdt.expiresIn') }}: <strong class="text-gray-900 dark:text-white">{{ countdown }}</strong>
        </span>
      </div>

      <!-- Confirmation Progress -->
      <div v-if="order.status === 'CONFIRMING'" class="space-y-2">
        <div class="flex items-center justify-between text-sm">
          <span class="text-gray-600 dark:text-gray-400">{{ t('payment.usdt.confirmationProgress') }}</span>
          <span class="font-semibold tabular-nums text-gray-900 dark:text-white">
            {{ order.confirmations }}/{{ order.required_confirmations }}
          </span>
        </div>
        <div class="h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-700">
          <div
            class="h-full rounded-full bg-gradient-to-r from-green-500 to-green-600 transition-all duration-500"
            :style="{ width: confirmationProgress + '%' }"
          ></div>
        </div>
      </div>

      <!-- Transaction Hash (if detected) -->
      <div v-if="order.tx_hash" class="rounded-lg bg-gray-50 p-3 dark:bg-dark-800">
        <div class="mb-1 text-xs font-medium text-gray-500 dark:text-gray-400">
          {{ t('payment.usdt.transactionHash') }}
        </div>
        <div class="flex items-center gap-2">
          <code class="flex-1 truncate font-mono text-xs text-gray-900 dark:text-white">
            {{ order.tx_hash }}
          </code>
          <a
            :href="explorerLink"
            target="_blank"
            rel="noopener noreferrer"
            class="shrink-0 text-xs font-medium text-[#C4612F] hover:underline"
          >
            {{ t('payment.usdt.viewOnExplorer') }} →
          </a>
        </div>
      </div>
    </div>

    <!-- Help Section -->
    <div class="space-y-2 border-t border-gray-200 pt-6 dark:border-dark-600">
      <button
        type="button"
        class="flex w-full items-center justify-between rounded-lg bg-gray-50 p-3 text-left transition-colors hover:bg-gray-100 dark:bg-dark-800 dark:hover:bg-dark-700"
        @click="showHelp = !showHelp"
      >
        <span class="flex items-center gap-2 text-sm font-medium text-gray-700 dark:text-gray-300">
          <svg class="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          {{ t('payment.usdt.howToPay') }}
        </span>
        <svg
          :class="['h-5 w-5 text-gray-400 transition-transform', showHelp ? 'rotate-180' : '']"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
          <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      <div v-if="showHelp" class="rounded-lg bg-gray-50 p-4 dark:bg-dark-800">
        <ol class="space-y-2 text-sm leading-relaxed text-gray-700 dark:text-gray-300">
          <li class="flex gap-2">
            <span class="font-semibold">1.</span>
            <span>{{ t('payment.usdt.helpStep1') }}</span>
          </li>
          <li class="flex gap-2">
            <span class="font-semibold">2.</span>
            <span>{{ t('payment.usdt.helpStep2', { network: order.network }) }}</span>
          </li>
          <li class="flex gap-2">
            <span class="font-semibold">3.</span>
            <span>{{ t('payment.usdt.helpStep3') }}</span>
          </li>
          <li class="flex gap-2">
            <span class="font-semibold">4.</span>
            <span>{{ t('payment.usdt.helpStep4') }}</span>
          </li>
        </ol>
      </div>
    </div>

    <!-- Action Buttons -->
    <div class="flex gap-3 border-t border-gray-200 pt-6 dark:border-dark-600">
      <button
        type="button"
        class="btn flex-1 border-gray-300 bg-white text-gray-700 hover:bg-gray-50 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300 dark:hover:bg-dark-700"
        @click="emit('cancel')"
      >
        {{ t('common.cancel') }}
      </button>
      <button
        v-if="order.status === 'PENDING'"
        type="button"
        class="btn flex-1 bg-[#C4612F] text-white hover:bg-[#A94E22]"
        @click="emit('refresh')"
      >
        {{ t('payment.usdt.checkStatus') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import QRCode from 'qrcode'
import type { USDTPaymentOrder } from '@/types/usdt'

const props = defineProps<{
  order: USDTPaymentOrder
  qrData: string
}>()

const emit = defineEmits<{
  cancel: []
  refresh: []
}>()

const { t } = useI18n()

const qrCanvas = ref<HTMLCanvasElement>()
const addressCopied = ref(false)
const amountCopied = ref(false)
const showHelp = ref(false)
const countdown = ref('')
let countdownInterval: ReturnType<typeof setInterval> | null = null

const statusClass = computed(() => {
  switch (props.order.status) {
    case 'PENDING':
      return 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-300'
    case 'DETECTED':
      return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
    case 'CONFIRMING':
      return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
    case 'CONFIRMED':
    case 'CREDITED':
    case 'COMPLETED':
      return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
    case 'EXPIRED':
      return 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-300'
    case 'FAILED':
      return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    default:
      return 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-300'
  }
})

const statusText = computed(() => {
  switch (props.order.status) {
    case 'PENDING':
      return t('payment.usdt.status.pending')
    case 'DETECTED':
      return t('payment.usdt.status.detected')
    case 'CONFIRMING':
      return t('payment.usdt.status.confirming')
    case 'CONFIRMED':
      return t('payment.usdt.status.confirmed')
    case 'CREDITED':
      return t('payment.usdt.status.credited')
    case 'COMPLETED':
      return t('payment.usdt.status.completed')
    case 'EXPIRED':
      return t('payment.usdt.status.expired')
    case 'FAILED':
      return t('payment.usdt.status.failed')
    default:
      return props.order.status
  }
})

const confirmationProgress = computed(() => {
  if (props.order.required_confirmations === 0) return 0
  return Math.min(100, (props.order.confirmations / props.order.required_confirmations) * 100)
})

const explorerLink = computed(() => {
  if (!props.order.tx_hash) return ''
  if (props.order.network === 'TRC20') {
    return `https://tronscan.org/#/transaction/${props.order.tx_hash}`
  } else if (props.order.network === 'ERC20') {
    return `https://etherscan.io/tx/${props.order.tx_hash}`
  }
  return ''
})

async function renderQRCode() {
  if (!qrCanvas.value) return
  try {
    await QRCode.toCanvas(qrCanvas.value, props.qrData, {
      width: 200,
      margin: 0,
      color: {
        dark: '#1F2421',
        light: '#FFFFFF'
      }
    })
  } catch (error) {
    console.error('Failed to render QR code:', error)
  }
}

async function copyAddress() {
  try {
    await navigator.clipboard.writeText(props.order.deposit_address)
    addressCopied.value = true
    setTimeout(() => {
      addressCopied.value = false
    }, 2000)
  } catch (error) {
    console.error('Failed to copy address:', error)
  }
}

async function copyAmount() {
  try {
    await navigator.clipboard.writeText(props.order.amount_usdt.toFixed(6))
    amountCopied.value = true
    setTimeout(() => {
      amountCopied.value = false
    }, 2000)
  } catch (error) {
    console.error('Failed to copy amount:', error)
  }
}

function updateCountdown() {
  const now = new Date().getTime()
  const expiresAt = new Date(props.order.expires_at).getTime()
  const remaining = Math.max(0, expiresAt - now)

  if (remaining === 0) {
    countdown.value = '00:00'
    if (countdownInterval) {
      clearInterval(countdownInterval)
      countdownInterval = null
    }
    return
  }

  const minutes = Math.floor(remaining / 60000)
  const seconds = Math.floor((remaining % 60000) / 1000)
  countdown.value = `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
}

watch(() => props.qrData, () => {
  renderQRCode()
}, { immediate: false })

onMounted(() => {
  renderQRCode()
  updateCountdown()
  countdownInterval = setInterval(updateCountdown, 1000)
})

onUnmounted(() => {
  if (countdownInterval) {
    clearInterval(countdownInterval)
  }
})
</script>

<template>
  <div class="space-y-4">
    <!-- Network Selection -->
    <div class="card p-6">
      <label class="mb-3 block text-sm font-medium text-gray-700 dark:text-gray-300">
        {{ t('payment.usdt.selectNetwork') }}
      </label>
      <div class="grid gap-3 sm:grid-cols-2">
        <button
          v-for="network in networks"
          :key="network.network_id"
          type="button"
          :disabled="!network.enabled"
          :class="[
            'relative flex flex-col items-start rounded-xl border-2 p-4 text-left transition-all',
            !network.enabled
              ? 'cursor-not-allowed border-gray-200 bg-gray-50 opacity-50 dark:border-dark-700 dark:bg-dark-800/50'
              : selectedNetwork === network.network_id
                ? 'border-[#C4612F] bg-[#F2E3D6] shadow-sm dark:border-[#C4612F] dark:bg-[#C4612F]/10'
                : 'border-gray-300 bg-white hover:border-gray-400 dark:border-dark-600 dark:bg-dark-800 dark:hover:border-dark-500',
          ]"
          @click="network.enabled && selectNetwork(network.network_id)"
        >
          <!-- Network Name -->
          <div class="mb-2 flex w-full items-center justify-between">
            <span class="text-lg font-bold text-gray-900 dark:text-white">
              {{ network.display_name }}
            </span>
            <span
              v-if="network.priority === 1"
              class="rounded-full bg-[#C4612F] px-2 py-0.5 text-xs font-medium text-white"
            >
              {{ t('payment.usdt.recommended') }}
            </span>
          </div>

          <!-- Fee Estimate -->
          <div class="mb-1 flex items-baseline gap-2">
            <span class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.usdt.networkFee') }}:</span>
            <span
              :class="[
                'text-sm font-semibold',
                network.network_id === 'TRC20' ? 'text-green-600 dark:text-green-400' : 'text-amber-600 dark:text-amber-400'
              ]"
            >
              {{ network.fee_estimate }}
            </span>
          </div>

          <!-- Confirmations -->
          <div class="flex items-baseline gap-2">
            <span class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.usdt.confirmations') }}:</span>
            <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ network.block_confirmations }}
            </span>
          </div>

          <!-- Selected Indicator -->
          <div
            v-if="selectedNetwork === network.network_id"
            class="absolute right-3 top-3 flex h-6 w-6 items-center justify-center rounded-full bg-[#C4612F]"
          >
            <svg class="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
            </svg>
          </div>
        </button>
      </div>

      <!-- Network Warning -->
      <div
        v-if="selectedNetwork"
        class="mt-4 rounded-lg border-2 border-amber-500 bg-amber-50 p-4 dark:border-amber-500/50 dark:bg-amber-900/20"
      >
        <div class="flex gap-3">
          <svg class="h-6 w-6 shrink-0 text-amber-600 dark:text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <div class="flex-1 space-y-1">
            <p class="font-semibold text-amber-900 dark:text-amber-300">
              {{ t('payment.usdt.networkWarningTitle') }}
            </p>
            <p class="text-sm leading-relaxed text-amber-800 dark:text-amber-400">
              {{ t('payment.usdt.networkWarningDesc', { network: selectedNetwork }) }}
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { USDTNetworkConfig, USDTNetwork } from '@/types/usdt'

defineProps<{
  networks: USDTNetworkConfig[]
  selectedNetwork: USDTNetwork | null
}>()

const emit = defineEmits<{
  'update:selectedNetwork': [network: USDTNetwork]
}>()

const { t } = useI18n()

function selectNetwork(network: USDTNetwork) {
  emit('update:selectedNetwork', network)
}
</script>

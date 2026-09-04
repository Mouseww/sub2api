<template>
  <AppLayout>
    <div class="mx-auto max-w-2xl py-6">
      <USDTPaymentFlow
        :amount="orderAmount"
        @success="handleSuccess"
        @cancel="handleCancel"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/layouts/AppLayout.vue'
import USDTPaymentFlow from '@/components/payment/USDTPaymentFlow.vue'
import { useAppStore } from '@/stores/app'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const appStore = useAppStore()

const orderAmount = ref(0)

onMounted(() => {
  const orderId = route.query.order_id
  if (!orderId) {
    appStore.showError(t('payment.errors.NOT_FOUND'))
    router.push('/payment')
    return
  }

  // For now, we'll need the amount from somewhere
  // In a real implementation, we'd fetch the order details
  // For this demo, we'll use a placeholder
  orderAmount.value = 100
})

function handleSuccess(orderId: number) {
  router.push({
    path: '/payment/result',
    query: { order_id: String(orderId) }
  })
}

function handleCancel() {
  router.push('/payment')
}
</script>

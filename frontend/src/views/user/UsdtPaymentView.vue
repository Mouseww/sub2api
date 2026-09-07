<template>
  <AppLayout>
    <div class="mx-auto max-w-2xl py-6">
      <USDTPaymentFlow
        :amount="orderAmount"
        :existing-order-id="existingOrderId"
        @success="handleSuccess"
        @cancel="handleCancel"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import USDTPaymentFlow from '@/components/payment/USDTPaymentFlow.vue'

const route = useRoute()
const router = useRouter()

const orderAmount = ref(0)
const existingOrderId = ref<number | undefined>(undefined)

onMounted(() => {
  const orderId = route.query.order_id
  if (orderId) {
    existingOrderId.value = Number(orderId)
  }
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

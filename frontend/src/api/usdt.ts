/**
 * USDT Payment API endpoints
 */

import { apiClient } from './client'
import type {
  USDTNetworkConfig,
  CreateUSDTOrderRequest,
  CreateUSDTOrderResponse,
  USDTOrderStatusResponse,
  USDTPaymentOrder
} from '@/types/usdt'

export const usdtAPI = {
  /** Get available USDT networks */
  getNetworks() {
    return apiClient.get<USDTNetworkConfig[]>('/payment/usdt/networks')
  },

  /** Create a new USDT payment order */
  createOrder(data: CreateUSDTOrderRequest) {
    return apiClient.post<CreateUSDTOrderResponse>('/payment/usdt/orders', data)
  },

  /** Get USDT order status */
  getOrderStatus(orderId: number) {
    return apiClient.get<USDTOrderStatusResponse>(`/payment/usdt/orders/${orderId}`)
  },

  /** Cancel a pending USDT order */
  cancelOrder(orderId: number) {
    return apiClient.post(`/payment/usdt/orders/${orderId}/cancel`)
  },

  /** Get user's USDT orders */
  getMyOrders(params?: { page?: number; page_size?: number; status?: string }) {
    return apiClient.get<{ items: USDTPaymentOrder[]; total: number }>('/payment/usdt/orders/my', { params })
  }
}

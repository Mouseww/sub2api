/**
 * USDT Payment Type Definitions
 */

export type USDTNetwork = 'TRC20' | 'ERC20'

export type USDTOrderStatus =
  | 'PENDING'       // Order created, awaiting blockchain transaction
  | 'DETECTED'      // Transaction seen on blockchain (0 confirmations)
  | 'CONFIRMING'    // Transaction has N confirmations (N < required)
  | 'CONFIRMED'     // Transaction fully confirmed
  | 'CREDITED'      // Balance added to user account
  | 'COMPLETED'     // Final state
  | 'EXPIRED'       // No transaction within timeout window
  | 'FAILED'        // Transaction failed or insufficient amount

export interface USDTNetworkConfig {
  network_id: USDTNetwork
  display_name: string
  enabled: boolean
  contract_address: string
  block_confirmations: number
  fee_estimate: string
  priority: number
}

export interface USDTPaymentOrder {
  id: number
  user_id: number
  order_no: string
  amount_usd: number
  amount_usdt: number
  network: USDTNetwork
  deposit_address: string
  status: USDTOrderStatus
  confirmations: number
  required_confirmations: number
  tx_hash?: string
  from_address?: string
  created_at: string
  expires_at: string
  detected_at?: string
  confirmed_at?: string
  credited_at?: string
  completed_at?: string
}

export interface CreateUSDTOrderRequest {
  amount_usd: number
  network: USDTNetwork
}

export interface CreateUSDTOrderResponse {
  order: USDTPaymentOrder
  qr_data: string
}

export interface USDTOrderStatusResponse {
  order: USDTPaymentOrder
  qr_data?: string
}

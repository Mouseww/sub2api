# USDT Payment Integration

This document describes the USDT cryptocurrency payment integration for sub2api.

## Overview

The USDT payment system allows users to recharge their account balance using USDT (Tether) cryptocurrency on TRC20 (Tron) and ERC20 (Ethereum) networks.

## Architecture

### Components

1. **USDT Provider** (`backend/internal/payment/provider/usdt.go`)
   - Implements the `payment.Provider` interface
   - Generates unique payment amounts for order matching
   - Supports TRC20 and ERC20 networks

2. **Blockchain Monitor Service** (`backend/internal/service/blockchain_monitor_service.go`)
   - Polls blockchain APIs to detect incoming transactions
   - Matches transactions to pending orders by amount + address + time window
   - Updates order confirmations and triggers fulfillment

3. **Payment Service Extensions** (`backend/internal/service/payment_service_usdt.go`)
   - `CreateUSDTOrder`: Creates USDT payment orders with crypto-specific fields
   - `GetUSDTOrderStatus`: Returns order status with transaction details
   - `CompleteOrder`: Credits user balance after confirmation

4. **API Endpoints** (`backend/internal/handler/usdt_payment_handler.go`)
   - `POST /api/v1/payment/usdt/orders` - Create USDT order
   - `GET /api/v1/payment/usdt/orders/:id/status` - Get order status
   - `GET /api/v1/payment/usdt/providers` - List available networks
   - `POST /api/v1/admin/payment/usdt/orders/:id/check` - Manual check (admin)

## Payment Flow

### 1. Order Creation
```
User → Frontend → Backend API → USDT Provider
                           ↓
                    Generate unique amount
                    (e.g., 10.00 → 10.2847)
                           ↓
                    Create PaymentOrder with:
                    - crypto_address (deposit)
                    - crypto_network (TRC20/ERC20)
                    - pay_amount (unique)
                    - crypto_required_confirmations
                           ↓
                    Return QR code & details
                           ↓
User scans QR → Sends from wallet
```

### 2. Transaction Detection
```
Blockchain Monitor (60s poll)
         ↓
Query blockchain API for deposits
         ↓
Match by: address + amount + time window
         ↓
Transaction found?
    YES → Update order with tx_hash, confirmations
    NO  → Continue polling
```

### 3. Confirmation & Fulfillment
```
Monitor detects confirmations increase
         ↓
Confirmations >= required? (19 for TRC20, 12 for ERC20)
    YES → Mark order PAID
        → Call CompleteOrder()
        → Credit user balance
        → Mark order COMPLETED
    NO  → Keep monitoring
```

### 4. Expiry
```
Order expires after 30 minutes (configurable)
Grace period: 5 minutes after expiry
         ↓
No transaction detected?
    YES → Mark order EXPIRED
```

## Database Schema

### Existing Fields (payment_orders table)
- `crypto_currency` - "USDT"
- `crypto_network` - "TRC20" or "ERC20"
- `crypto_address` - Platform deposit address
- `crypto_tx_hash` - Blockchain transaction hash
- `crypto_confirmations` - Current confirmation count
- `crypto_required_confirmations` - Required confirmations
- `crypto_amount_usd` - USD equivalent (for future use)

## Configuration

### Provider Instance Configuration

Create a USDT provider instance in the admin panel with:

```json
{
  "network_id": "TRC20",
  "deposit_address": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
  "contract_address": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
  "confirmations": "19",
  "blockchain_api_url": "https://api.trongrid.io",
  "blockchain_api_key": "your-api-key",
  "poll_interval_seconds": "60"
}
```

### Network Defaults

**TRC20 (Tron)**
- Required confirmations: 19
- Estimated time: 1-3 minutes
- Transaction fee: ~1 USDT
- Contract: TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t

**ERC20 (Ethereum)**
- Required confirmations: 12
- Estimated time: 3-15 minutes
- Transaction fee: Variable (gas-dependent)
- Contract: 0xdac17f958d2ee523a2206206994597c13d831ec7

## API Examples

### Create USDT Order
```bash
POST /api/v1/payment/usdt/orders
Authorization: Bearer {token}
Content-Type: application/json

{
  "amount": 100.00,
  "payment_type": "usdt_trc20",
  "order_type": "balance"
}

Response:
{
  "order_id": 12345,
  "amount": 100.00,
  "pay_amount": 100.2847,
  "payment_type": "usdt_trc20",
  "status": "PENDING",
  "network_id": "TRC20",
  "deposit_address": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
  "qr_code": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t?amount=100.2847",
  "required_confirmations": 19,
  "estimated_time": "1-3 minutes",
  "expires_at": "2025-04-09T12:30:00Z"
}
```

### Get Order Status
```bash
GET /api/v1/payment/usdt/orders/12345/status
Authorization: Bearer {token}

Response:
{
  "order_id": 12345,
  "status": "CONFIRMING",
  "amount": 100.00,
  "pay_amount": 100.2847,
  "network": "TRC20",
  "address": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
  "transaction": {
    "tx_hash": "0xabc123...",
    "confirmations": 8,
    "required_confirmations": 19,
    "explorer_url": "https://tronscan.org/#/transaction/0xabc123..."
  },
  "expires_at": "2025-04-09T12:30:00Z",
  "created_at": "2025-04-09T12:00:00Z"
}
```

## Security Considerations

### Private Key Management
- Store private keys encrypted with AES-256
- Use environment variables for encryption keys
- Consider hardware security modules (HSM) for production

### Hot/Cold Wallet Strategy
- Keep minimal balance in hot wallet
- Auto-sweep to cold storage when threshold reached
- Manual withdrawal process for cold wallet

### Transaction Verification
- Verify destination address matches
- Check amount with tolerance (0.0001 USDT)
- Require minimum confirmations
- Validate transaction status

### Fraud Prevention
- Daily deposit limits per user
- Flag suspicious patterns
- Monitor for address reuse attempts
- Rate limit order creation

## Blockchain API Integration

### TronGrid (TRC20)
```go
// Query transactions to address
GET https://api.trongrid.io/v1/accounts/{address}/transactions/trc20
Parameters:
  - contract_address: TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t
  - limit: 50
  - min_timestamp: order_created_at
```

### Etherscan (ERC20)
```go
// Query token transfers
GET https://api.etherscan.io/api
Parameters:
  - module: account
  - action: tokentx
  - address: 0xdac17f958d2ee523a2206206994597c13d831ec7
  - contractaddress: 0xdac17f958d2ee523a2206206994597c13d831ec7
```

## Testing

### Unit Tests
```bash
cd backend
go test ./internal/payment/provider -run TestUSDT
go test ./internal/service -run TestBlockchainMonitor
```

### Manual Testing (Testnet)
1. Configure testnet provider instance
2. Use Nile (Tron) or Sepolia (Ethereum) testnet
3. Get testnet USDT from faucet
4. Create order and send testnet USDT
5. Verify detection and confirmation

## Future Enhancements

### Phase 1 (Current)
- ✅ Basic USDT payment flow
- ✅ TRC20 and ERC20 support
- ✅ Shared deposit address model
- ✅ Manual blockchain checking

### Phase 2
- [ ] Real blockchain API integration (TronGrid/Etherscan)
- [ ] Automatic transaction detection
- [ ] Webhook support for faster detection
- [ ] Admin dashboard for monitoring

### Phase 3
- [ ] Unique address per order (HD wallet)
- [ ] Multi-signature wallet support
- [ ] Auto-sweep to cold storage
- [ ] Support for other networks (BEP20, Polygon)

### Phase 4
- [ ] Real-time WebSocket status updates
- [ ] Mobile wallet deep linking
- [ ] Exchange rate locking
- [ ] Refund/withdrawal process

## Deployment Checklist

- [ ] Set up secure wallet addresses
- [ ] Configure blockchain API keys
- [ ] Enable blockchain monitor service
- [ ] Set up monitoring/alerting
- [ ] Configure hot/cold wallet thresholds
- [ ] Test on testnet first
- [ ] Document operational procedures
- [ ] Train support team
- [ ] Prepare user documentation
- [ ] Monitor first transactions closely

## Troubleshooting

### Order stuck in PENDING
- Check blockchain API connectivity
- Verify transaction was actually sent
- Check if amount matches exactly
- Ensure transaction sent to correct address
- Check if transaction is on correct network

### Transaction not detected
- Verify blockchain monitor is running
- Check API rate limits
- Ensure sufficient confirmations
- Check transaction timestamp vs order expiry
- Review logs for errors

### Wrong network sent
- Implement recovery procedure
- May require manual intervention
- Document user communication process

## Support Resources

- Blockchain Explorers:
  - TRC20: https://tronscan.org
  - ERC20: https://etherscan.io
- API Documentation:
  - TronGrid: https://www.trongrid.io/
  - Etherscan: https://docs.etherscan.io/
- USDT Information:
  - Official: https://tether.to/

## Contact

For technical questions about this integration, contact the backend team.

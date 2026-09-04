package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUSDT(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]string
		expectError bool
	}{
		{
			name: "valid TRC20 config",
			config: map[string]string{
				"network_id":       "TRC20",
				"deposit_address":  "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
				"contract_address": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
				"hmac_secret":      "test-secret-key-for-hmac-generation",
			},
			expectError: false,
		},
		{
			name: "valid ERC20 config",
			config: map[string]string{
				"network_id":       "ERC20",
				"deposit_address":  "0xdac17f958d2ee523a2206206994597c13d831ec7",
				"contract_address": "0xdac17f958d2ee523a2206206994597c13d831ec7",
				"hmac_secret":      "test-secret-key-for-hmac-generation",
			},
			expectError: false,
		},
		{
			name: "invalid network",
			config: map[string]string{
				"network_id":       "BEP20",
				"deposit_address":  "0x123",
				"contract_address": "0x456",
				"hmac_secret":      "test-secret",
			},
			expectError: true,
		},
		{
			name: "missing deposit address",
			config: map[string]string{
				"network_id":       "TRC20",
				"contract_address": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
				"hmac_secret":      "test-secret",
			},
			expectError: true,
		},
		{
			name: "missing contract address",
			config: map[string]string{
				"network_id":      "TRC20",
				"deposit_address": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
				"hmac_secret":     "test-secret",
			},
			expectError: true,
		},
		{
			name: "missing hmac secret",
			config: map[string]string{
				"network_id":       "TRC20",
				"deposit_address":  "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
				"contract_address": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prov, err := NewUSDT("test-instance", tt.config)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, prov)
			} else {
				require.NoError(t, err)
				require.NotNil(t, prov)
			}
		})
	}
}

func TestUSDTProvider_CreatePayment(t *testing.T) {
	config := map[string]string{
		"network_id":       "TRC20",
		"deposit_address":  "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		"contract_address": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		"hmac_secret":      "test-secret-key-for-hmac-generation",
	}

	prov, err := NewUSDT("test-instance", config)
	require.NoError(t, err)

	req := payment.CreatePaymentRequest{
		OrderID:     "test-order-123",
		Amount:      "10.00",
		PaymentType: payment.TypeUSDTTRC20,
		Subject:     "Test Payment",
	}

	ctx := context.Background()
	resp, err := prov.CreatePayment(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify response
	assert.Equal(t, "test-order-123", resp.TradeNo)
	assert.NotEmpty(t, resp.QRCode)
	assert.Contains(t, resp.QRCode, "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t")
	assert.Equal(t, payment.CreatePaymentResultOrderCreated, resp.ResultType)
}

func TestUSDTProvider_SupportedTypes(t *testing.T) {
	config := map[string]string{
		"network_id":       "TRC20",
		"deposit_address":  "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		"contract_address": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		"hmac_secret":      "test-secret-key-for-hmac-generation",
	}

	prov, err := NewUSDT("test-instance", config)
	require.NoError(t, err)

	types := prov.SupportedTypes()
	assert.Contains(t, types, payment.TypeUSDTTRC20)
	assert.Contains(t, types, payment.TypeUSDTERC20)
}

func TestGenerateUniqueAmount(t *testing.T) {
	config := map[string]string{
		"network_id":       "TRC20",
		"deposit_address":  "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		"contract_address": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		"hmac_secret":      "test-secret-key-for-hmac-generation",
	}

	prov, err := NewUSDT("test-instance", config)
	require.NoError(t, err)
	usdtProv := prov.(*USDTProvider)

	tests := []struct {
		name        string
		baseAmount  string
		expectError bool
	}{
		{
			name:        "valid integer amount",
			baseAmount:  "10",
			expectError: false,
		},
		{
			name:        "valid decimal amount",
			baseAmount:  "10.50",
			expectError: false,
		},
		{
			name:        "valid large amount",
			baseAmount:  "1000.00",
			expectError: false,
		},
		{
			name:        "invalid amount",
			baseAmount:  "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, err := usdtProv.generateUniqueAmount(tt.baseAmount, "order-123", time.Now())
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				// Parse base amount
				var base float64
				fmt.Sscanf(tt.baseAmount, "%f", &base)

				// Verify amount is within expected range (base + HMAC-based suffix)
				diff := amount - base
				assert.GreaterOrEqual(t, diff, 0.0001, "HMAC suffix should be >= 0.0001")
				assert.LessOrEqual(t, diff, 0.9999, "HMAC suffix should be <= 0.9999")

				// Verify precision (6 decimals)
				amountStr := fmt.Sprintf("%.6f", amount)
				assert.Regexp(t, `^\d+\.\d{6}$`, amountStr, "Amount should have exactly 6 decimal places")
			}
		})
	}
}

func TestGenerateUniqueAmount_Deterministic(t *testing.T) {
	// HMAC-based generation should be deterministic for same inputs
	config := map[string]string{
		"network_id":       "TRC20",
		"deposit_address":  "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		"contract_address": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		"hmac_secret":      "test-secret-key-for-hmac-generation",
	}

	prov, err := NewUSDT("test-instance", config)
	require.NoError(t, err)
	usdtProv := prov.(*USDTProvider)

	baseAmount := "10.00"
	orderID := "order-123"
	timestamp := time.Now()

	// Generate twice with same inputs
	amount1, err1 := usdtProv.generateUniqueAmount(baseAmount, orderID, timestamp)
	require.NoError(t, err1)
	
	amount2, err2 := usdtProv.generateUniqueAmount(baseAmount, orderID, timestamp)
	require.NoError(t, err2)

	// Should be identical (deterministic)
	assert.Equal(t, amount1, amount2)
}

func TestGenerateUniqueAmount_Uniqueness(t *testing.T) {
	// Different order IDs should generate different amounts
	config := map[string]string{
		"network_id":       "TRC20",
		"deposit_address":  "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		"contract_address": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		"hmac_secret":      "test-secret-key-for-hmac-generation",
	}

	prov, err := NewUSDT("test-instance", config)
	require.NoError(t, err)
	usdtProv := prov.(*USDTProvider)

	baseAmount := "10.00"
	timestamp := time.Now()
	amounts := make(map[float64]bool)

	for i := 0; i < 100; i++ {
		orderID := fmt.Sprintf("order-%d", i)
		amount, err := usdtProv.generateUniqueAmount(baseAmount, orderID, timestamp)
		require.NoError(t, err)
		amounts[amount] = true
	}

	// Should have generated 100 unique amounts (HMAC ensures no collisions)
	assert.Equal(t, 100, len(amounts))
}

func TestVerifyTransaction(t *testing.T) {
	tests := []struct {
		name                  string
		tx                    *USDTTransactionInfo
		expectedAddress       string
		expectedAmount        float64
		requiredConfirmations int
		expectValid           bool
		expectReason          string
	}{
		{
			name: "valid transaction",
			tx: &USDTTransactionInfo{
				TxHash:        "0xabc123",
				FromAddress:   "sender123",
				ToAddress:     "receiver123",
				Amount:        10.2847,
				Confirmations: 20,
				Status:        "success",
			},
			expectedAddress:       "receiver123",
			expectedAmount:        10.2847,
			requiredConfirmations: 19,
			expectValid:           true,
		},
		{
			name: "address mismatch",
			tx: &USDTTransactionInfo{
				TxHash:        "0xabc123",
				FromAddress:   "sender123",
				ToAddress:     "wrong-address",
				Amount:        10.2847,
				Confirmations: 20,
				Status:        "success",
			},
			expectedAddress:       "receiver123",
			expectedAmount:        10.2847,
			requiredConfirmations: 19,
			expectValid:           false,
			expectReason:          "address mismatch",
		},
		{
			name: "amount mismatch",
			tx: &USDTTransactionInfo{
				TxHash:        "0xabc123",
				FromAddress:   "sender123",
				ToAddress:     "receiver123",
				Amount:        11.0000,
				Confirmations: 20,
				Status:        "success",
			},
			expectedAddress:       "receiver123",
			expectedAmount:        10.2847,
			requiredConfirmations: 19,
			expectValid:           false,
			expectReason:          "amount mismatch",
		},
		{
			name: "insufficient confirmations",
			tx: &USDTTransactionInfo{
				TxHash:        "0xabc123",
				FromAddress:   "sender123",
				ToAddress:     "receiver123",
				Amount:        10.2847,
				Confirmations: 5,
				Status:        "success",
			},
			expectedAddress:       "receiver123",
			expectedAmount:        10.2847,
			requiredConfirmations: 19,
			expectValid:           false,
			expectReason:          "insufficient confirmations",
		},
		{
			name: "transaction failed",
			tx: &USDTTransactionInfo{
				TxHash:        "0xabc123",
				FromAddress:   "sender123",
				ToAddress:     "receiver123",
				Amount:        10.2847,
				Confirmations: 20,
				Status:        "failed",
			},
			expectedAddress:       "receiver123",
			expectedAmount:        10.2847,
			requiredConfirmations: 19,
			expectValid:           false,
			expectReason:          "transaction not successful",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, reason := VerifyTransaction(tt.tx, tt.expectedAddress, tt.expectedAmount, tt.requiredConfirmations)
			assert.Equal(t, tt.expectValid, valid)
			if !tt.expectValid {
				assert.Contains(t, reason, tt.expectReason)
			}
		})
	}
}

func TestVerifyTransaction_ExactMatching(t *testing.T) {
	// Test that exact matching is enforced (zero tolerance)
	baseAmount := 10.284756
	
	tests := []struct {
		name         string
		txAmount     float64
		expectValid  bool
		expectReason string
	}{
		{
			name:        "exact match",
			txAmount:    10.284756,
			expectValid: true,
		},
		{
			name:         "off by 0.000001",
			txAmount:     10.284757,
			expectValid:  false,
			expectReason: "amount mismatch",
		},
		{
			name:         "off by 0.0001",
			txAmount:     10.2848,
			expectValid:  false,
			expectReason: "amount mismatch",
		},
		{
			name:         "off by 0.01",
			txAmount:     10.29,
			expectValid:  false,
			expectReason: "amount mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &USDTTransactionInfo{
				TxHash:        "0xabc123",
				FromAddress:   "sender123",
				ToAddress:     "receiver123",
				Amount:        tt.txAmount,
				Confirmations: 20,
				Status:        "success",
			}
			valid, reason := VerifyTransaction(tx, "receiver123", baseAmount, 19)
			assert.Equal(t, tt.expectValid, valid)
			if !tt.expectValid {
				assert.Contains(t, reason, tt.expectReason)
			}
		})
	}
}

func TestParseMonitorConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]string
		expectError bool
		checkResult func(*testing.T, *BlockchainMonitorConfig)
	}{
		{
			name: "valid TRC20 config",
			config: map[string]string{
				"network_id":       "TRC20",
				"deposit_address":  "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
				"contract_address": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
			},
			expectError: false,
			checkResult: func(t *testing.T, cfg *BlockchainMonitorConfig) {
				assert.Equal(t, "TRC20", cfg.NetworkID)
				assert.Equal(t, 19, cfg.RequiredConfirms)
			},
		},
		{
			name: "valid ERC20 config",
			config: map[string]string{
				"network_id":       "ERC20",
				"deposit_address":  "0xdac17f958d2ee523a2206206994597c13d831ec7",
				"contract_address": "0xdac17f958d2ee523a2206206994597c13d831ec7",
			},
			expectError: false,
			checkResult: func(t *testing.T, cfg *BlockchainMonitorConfig) {
				assert.Equal(t, "ERC20", cfg.NetworkID)
				assert.Equal(t, 12, cfg.RequiredConfirms)
			},
		},
		{
			name: "custom confirmations",
			config: map[string]string{
				"network_id":       "TRC20",
				"deposit_address":  "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
				"contract_address": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
				"confirmations":    "30",
			},
			expectError: false,
			checkResult: func(t *testing.T, cfg *BlockchainMonitorConfig) {
				assert.Equal(t, 30, cfg.RequiredConfirms)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := ParseMonitorConfig(tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, cfg)
				}
			}
		})
	}
}

// Helper functions for tests
func sscanf(s, format string, a ...interface{}) (int, error) {
	return fmt.Sscanf(s, format, a...)
}

func formatFloat(f float64, decimals int) string {
	return fmt.Sprintf("%.*f", decimals, f)
}

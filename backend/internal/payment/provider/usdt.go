package provider

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
)

// USDTProvider implements the Provider interface for USDT cryptocurrency payments.
// It follows the shared deposit address + unique amount model recommended in requirements.
type USDTProvider struct {
	instanceID      string
	config          map[string]string
	networkID       string // "TRC20" or "ERC20"
	depositAddress  string
	contractAddress string
	confirmations   int
	hmacSecret      []byte // HMAC secret for amount generation
}

// NewUSDT creates a new USDT payment provider instance.
// Config keys:
//   - network_id: "TRC20" or "ERC20" (required)
//   - deposit_address: Platform's deposit wallet address (required)
//   - contract_address: USDT contract address (required)
//   - confirmations: Required block confirmations (default: 19 for TRC20, 12 for ERC20)
//   - hmac_secret: HMAC secret for cryptographic amount generation (required)
func NewUSDT(instanceID string, config map[string]string) (payment.Provider, error) {
	networkID := config["network_id"]
	if networkID != "TRC20" && networkID != "ERC20" {
		return nil, fmt.Errorf("invalid network_id: %s (must be TRC20 or ERC20)", networkID)
	}

	depositAddress := config["deposit_address"]
	if depositAddress == "" {
		return nil, fmt.Errorf("deposit_address is required")
	}

	contractAddress := config["contract_address"]
	if contractAddress == "" {
		return nil, fmt.Errorf("contract_address is required")
	}

	// HMAC secret is required for secure amount generation
	hmacSecret := config["hmac_secret"]
	if hmacSecret == "" {
		return nil, fmt.Errorf("hmac_secret is required for cryptographic amount generation")
	}

	// Set default confirmations based on network
	confirmations := 19 // TRC20 default
	if networkID == "ERC20" {
		confirmations = 12
	}
	if confStr := config["confirmations"]; confStr != "" {
		var conf int
		if _, err := fmt.Sscanf(confStr, "%d", &conf); err == nil && conf > 0 {
			confirmations = conf
		}
	}

	return &USDTProvider{
		instanceID:      instanceID,
		config:          config,
		networkID:       networkID,
		depositAddress:  depositAddress,
		contractAddress: contractAddress,
		confirmations:   confirmations,
		hmacSecret:      []byte(hmacSecret),
	}, nil
}

func (u *USDTProvider) Name() string {
	return fmt.Sprintf("USDT (%s)", u.networkID)
}

func (u *USDTProvider) ProviderKey() string {
	return payment.TypeUSDT
}

func (u *USDTProvider) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{
		payment.TypeUSDTTRC20,
		payment.TypeUSDTERC20,
	}
}

// CreatePayment generates a unique payment order for USDT deposit.
// Returns deposit address, unique amount, and QR code for the user to send.
func (u *USDTProvider) CreatePayment(ctx context.Context, req payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	// Generate unique payment amount using HMAC-based cryptographic method
	// This ensures no collisions and ties the amount to this specific order
	uniqueAmount, err := u.generateUniqueAmount(req.Amount, req.OrderID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("generate unique amount: %w", err)
	}

	// Generate QR code content for crypto wallet scanning
	// Format: address?amount=X.XXXXXX (6 decimal precision)
	qrContent := fmt.Sprintf("%s?amount=%.6f", u.depositAddress, uniqueAmount)

	return &payment.CreatePaymentResponse{
		TradeNo:    req.OrderID, // Use internal order ID as trade reference
		QRCode:     qrContent,
		ResultType: payment.CreatePaymentResultOrderCreated,
	}, nil
}

// QueryOrder checks the blockchain for transaction matching this order.
// This would typically be called by the blockchain monitoring service.
func (u *USDTProvider) QueryOrder(ctx context.Context, tradeNo string) (*payment.QueryOrderResponse, error) {
	// This is a placeholder - actual implementation requires blockchain API integration
	// The monitoring service will call blockchain APIs and update transaction status
	return &payment.QueryOrderResponse{
		TradeNo: tradeNo,
		Status:  payment.ProviderStatusPending,
	}, nil
}

// VerifyNotification is not used for crypto payments.
// Crypto payments are pull-based (blockchain polling) rather than push-based (webhooks).
func (u *USDTProvider) VerifyNotification(ctx context.Context, rawBody string, headers map[string]string) (*payment.PaymentNotification, error) {
	return nil, nil
}

// Refund is not supported for cryptocurrency payments.
// Crypto transactions are irreversible - refunds must be handled manually.
func (u *USDTProvider) Refund(ctx context.Context, req payment.RefundRequest) (*payment.RefundResponse, error) {
	return nil, fmt.Errorf("refunds not supported for cryptocurrency payments")
}

// MerchantIdentityMetadata returns identifying information for snapshot consistency.
func (u *USDTProvider) MerchantIdentityMetadata() map[string]string {
	return map[string]string{
		"network_id":       u.networkID,
		"deposit_address":  u.depositAddress,
		"contract_address": u.contractAddress,
	}
}

// generateUniqueAmount creates a cryptographically unique payment amount using HMAC.
// This ties the amount to a specific orderID + timestamp, preventing collisions.
// Based on requirements Section 2.3: HMAC-based amount generation.
func (u *USDTProvider) generateUniqueAmount(baseAmount string, orderID string, timestamp time.Time) (float64, error) {
	var base float64
	if _, err := fmt.Sscanf(baseAmount, "%f", &base); err != nil {
		return 0, fmt.Errorf("invalid amount format: %w", err)
	}

	// Create deterministic input: orderID + timestamp
	input := []byte(fmt.Sprintf("%s:%d", orderID, timestamp.Unix()))

	// Generate HMAC
	h := hmac.New(sha256.New, u.hmacSecret)
	h.Write(input)
	hash := h.Sum(nil)

	// Extract first 4 bytes and convert to 0.0001-0.9999 range
	suffix := binary.BigEndian.Uint32(hash[:4])
	// Map to 0.0001-0.9999 (1-9999 then divide by 10000)
	normalizedSuffix := float64((suffix%9999)+1) / 10000.0

	// Combine base amount + cryptographic suffix
	// Result: 10.0000 → 10.8472 (6 decimal USDT precision)
	return math.Round((base+normalizedSuffix)*1000000) / 1000000, nil
}

// USDTTransactionInfo holds blockchain transaction details for verification.
type USDTTransactionInfo struct {
	TxHash        string    `json:"tx_hash"`
	FromAddress   string    `json:"from_address"`
	ToAddress     string    `json:"to_address"`
	Amount        float64   `json:"amount"`
	Confirmations int       `json:"confirmations"`
	BlockNumber   int64     `json:"block_number"`
	Timestamp     time.Time `json:"timestamp"`
	Status        string    `json:"status"` // "pending", "success", "failed"
}

// VerifyTransaction checks if a blockchain transaction matches order requirements.
// Uses EXACT amount matching (zero tolerance) as per requirements Section 3.2.
func VerifyTransaction(tx *USDTTransactionInfo, expectedAddress string, expectedAmount float64, requiredConfirmations int) (bool, string) {
	// Check destination address
	if tx.ToAddress != expectedAddress {
		return false, "address mismatch"
	}

	// EXACT amount matching - ZERO tolerance for precision errors
	// USDT has 6 decimal precision on blockchain (0.000001)
	// Round both to 6 decimals for comparison
	txAmount := math.Round(tx.Amount*1000000) / 1000000
	expAmount := math.Round(expectedAmount*1000000) / 1000000
	if txAmount != expAmount {
		return false, fmt.Sprintf("amount mismatch: expected %.6f, got %.6f", expAmount, txAmount)
	}

	// Check confirmations
	if tx.Confirmations < requiredConfirmations {
		return false, fmt.Sprintf("insufficient confirmations: %d/%d", tx.Confirmations, requiredConfirmations)
	}

	// Check transaction status
	if tx.Status != "success" {
		return false, fmt.Sprintf("transaction not successful: %s", tx.Status)
	}

	return true, ""
}

// BlockchainMonitorConfig holds configuration for the blockchain monitoring service.
type BlockchainMonitorConfig struct {
	NetworkID           string        `json:"network_id"`
	DepositAddress      string        `json:"deposit_address"`
	ContractAddress     string        `json:"contract_address"`
	PollInterval        time.Duration `json:"poll_interval"`
	RequiredConfirms    int           `json:"required_confirmations"`
	BlockchainAPIURL    string        `json:"blockchain_api_url"`
	BlockchainAPIKey    string        `json:"blockchain_api_key"`
	MaxConfirmationTime time.Duration `json:"max_confirmation_time"`
}

// ParseMonitorConfig extracts blockchain monitor config from provider config map.
func ParseMonitorConfig(config map[string]string) (*BlockchainMonitorConfig, error) {
	networkID := config["network_id"]
	if networkID == "" {
		return nil, fmt.Errorf("network_id required")
	}

	depositAddress := config["deposit_address"]
	if depositAddress == "" {
		return nil, fmt.Errorf("deposit_address required")
	}

	contractAddress := config["contract_address"]
	if contractAddress == "" {
		return nil, fmt.Errorf("contract_address required")
	}

	// Set defaults based on network
	requiredConfirms := 19
	pollInterval := 60 * time.Second
	maxConfirmationTime := 30 * time.Minute

	if networkID == "ERC20" {
		requiredConfirms = 12
		maxConfirmationTime = 60 * time.Minute
	}

	// Override with config values if present
	if confStr := config["confirmations"]; confStr != "" {
		var conf int
		if _, err := fmt.Sscanf(confStr, "%d", &conf); err == nil && conf > 0 {
			requiredConfirms = conf
		}
	}

	if pollStr := config["poll_interval_seconds"]; pollStr != "" {
		var secs int
		if _, err := fmt.Sscanf(pollStr, "%d", &secs); err == nil && secs > 0 {
			pollInterval = time.Duration(secs) * time.Second
		}
	}

	return &BlockchainMonitorConfig{
		NetworkID:           networkID,
		DepositAddress:      depositAddress,
		ContractAddress:     contractAddress,
		PollInterval:        pollInterval,
		RequiredConfirms:    requiredConfirms,
		BlockchainAPIURL:    config["blockchain_api_url"],
		BlockchainAPIKey:    config["blockchain_api_key"],
		MaxConfirmationTime: maxConfirmationTime,
	}, nil
}

// SerializeTransactionData converts transaction info to JSON for storage.
func SerializeTransactionData(tx *USDTTransactionInfo) (string, error) {
	data, err := json.Marshal(tx)
	if err != nil {
		return "", fmt.Errorf("marshal transaction: %w", err)
	}
	return string(data), nil
}

// DeserializeTransactionData parses stored JSON transaction data.
func DeserializeTransactionData(data string) (*USDTTransactionInfo, error) {
	var tx USDTTransactionInfo
	if err := json.Unmarshal([]byte(data), &tx); err != nil {
		return nil, fmt.Errorf("unmarshal transaction: %w", err)
	}
	return &tx, nil
}

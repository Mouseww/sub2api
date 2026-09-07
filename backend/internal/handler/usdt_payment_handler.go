package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type USDTPaymentHandler struct {
	paymentService *service.PaymentService
	configService  *service.PaymentConfigService
}

func NewUSDTPaymentHandler(paymentService *service.PaymentService, configService *service.PaymentConfigService) *USDTPaymentHandler {
	return &USDTPaymentHandler{
		paymentService: paymentService,
		configService:  configService,
	}
}

// derefStr safely dereferences a *string, returning "" if nil.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// derefInt safely dereferences a *int, returning 0 if nil.
func derefInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// GetNetworks returns enabled USDT networks based on DB configuration.
// GET /api/v1/payment/usdt/networks
func (h *USDTPaymentHandler) GetNetworks(c *gin.Context) {
	cfg, err := h.configService.GetUSDTSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load USDT config"})
		return
	}

	var networks []gin.H

	if cfg.TRC20DepositAddress != "" {
		networks = append(networks, gin.H{
			"network_id":           "TRC20",
			"display_name":         "USDT (TRC20 / TRON)",
			"enabled":              true,
			"contract_address":     cfg.TRC20ContractAddress,
			"block_confirmations":  cfg.TRC20Confirmations,
			"fee_estimate":         "~1-2 USDT",
			"priority":             1,
		})
	}

	if cfg.ERC20DepositAddress != "" {
		networks = append(networks, gin.H{
			"network_id":           "ERC20",
			"display_name":         "USDT (ERC20 / Ethereum)",
			"enabled":              true,
			"contract_address":     cfg.ERC20ContractAddress,
			"block_confirmations":  cfg.ERC20Confirmations,
			"fee_estimate":         "Variable gas fees",
			"priority":             2,
		})
	}

	if networks == nil {
		networks = []gin.H{}
	}

	c.JSON(http.StatusOK, gin.H{"networks": networks})
}

// CreateUSDTOrder creates a new USDT payment order.
// POST /api/v1/payment/usdt/orders
func (h *USDTPaymentHandler) CreateUSDTOrder(c *gin.Context) {
	// Accept both the frontend format (amount_usd + network) and legacy format
	var req struct {
		AmountUSD float64 `json:"amount_usd"`
		Amount    float64 `json:"amount"`
		Network   string  `json:"network"`   // "TRC20" or "ERC20"
		ReturnURL string  `json:"return_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize amount
	amount := req.AmountUSD
	if amount <= 0 {
		amount = req.Amount
	}
	if amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than 0"})
		return
	}

	// Normalize network to payment type
	paymentType := "usdt_trc20"
	if req.Network == "ERC20" {
		paymentType = "usdt_erc20"
	}

	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	userID := subject.UserID

	orderReq := service.CreateOrderRequest{
		UserID:      userID,
		Amount:      amount,
		PaymentType: paymentType,
		OrderType:   "balance",
		ReturnURL:   req.ReturnURL,
		ClientIP:    c.ClientIP(),
		Locale:      c.GetHeader("Accept-Language"),
	}

	resp, err := h.paymentService.CreateOrder(c.Request.Context(), orderReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	order, err := h.paymentService.GetOrder(c.Request.Context(), resp.OrderID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "order created but could not be retrieved"})
		return
	}

	// Derive network display string and deposit address from config
	networkDisplay := "TRC20"
	depositAddress := ""
	if cfg, cerr := h.configService.GetUSDTSettings(c.Request.Context()); cerr == nil {
		if paymentType == "usdt_erc20" {
			networkDisplay = "ERC20"
			depositAddress = cfg.ERC20DepositAddress
		} else {
			depositAddress = cfg.TRC20DepositAddress
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"order": gin.H{
			"id":                     order.ID,
			"order_no":               order.OutTradeNo,
			"amount_usd":             order.Amount,
			"amount_usdt":            order.PayAmount,
			"network":                networkDisplay,
			"deposit_address":        depositAddress,
			"status":                 "PENDING",
			"confirmations":          0,
			"required_confirmations": derefInt(order.CryptoRequiredConfirmations),
			"created_at":             order.CreatedAt.Format(time.RFC3339),
			"expires_at":             order.ExpiresAt.Format(time.RFC3339),
		},
		"qr_data": resp.QRCode,
	})
}

// GetUSDTOrderStatus returns the current status of a USDT order.
// GET /api/v1/payment/usdt/orders/:id/status
func (h *USDTPaymentHandler) GetUSDTOrderStatus(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	userID := subject.UserID

	order, err := h.paymentService.GetOrder(c.Request.Context(), orderID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	// Derive network and deposit address
	networkDisplay := "TRC20"
	depositAddress := derefStr(order.CryptoAddress)
	if order.PaymentType == "usdt_erc20" {
		networkDisplay = "ERC20"
	}
	if depositAddress == "" {
		if cfg, cerr := h.configService.GetUSDTSettings(c.Request.Context()); cerr == nil {
			if order.PaymentType == "usdt_erc20" {
				depositAddress = cfg.ERC20DepositAddress
			} else {
				depositAddress = cfg.TRC20DepositAddress
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"order": gin.H{
			"id":                     order.ID,
			"order_no":               order.OutTradeNo,
			"amount_usd":             order.Amount,
			"amount_usdt":            order.PayAmount,
			"network":                networkDisplay,
			"deposit_address":        depositAddress,
			"status":                 order.Status,
			"confirmations":          derefInt(order.CryptoConfirmations),
			"required_confirmations": derefInt(order.CryptoRequiredConfirmations),
			"created_at":             order.CreatedAt.Format(time.RFC3339),
			"expires_at":             order.ExpiresAt.Format(time.RFC3339),
		},
		"qr_data": derefStr(order.QrCode),
	})
}

// CancelUSDTOrder cancels a pending USDT order.
// POST /api/v1/payment/usdt/orders/:id/cancel
func (h *USDTPaymentHandler) CancelUSDTOrder(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	userID := subject.UserID

	if _, err := h.paymentService.CancelOrder(c.Request.Context(), orderID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cancelled"})
}

// ListUSDTProviders lists available USDT payment providers.
// GET /api/v1/payment/usdt/providers
func (h *USDTPaymentHandler) ListUSDTProviders(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"providers": []gin.H{
			{
				"type":        "usdt_trc20",
				"name":        "USDT (TRC20)",
				"network":     "TRC20",
				"fee":         "~1-2 USDT",
				"description": "Low fees, fast confirmation (1-3 min)",
			},
			{
				"type":        "usdt_erc20",
				"name":        "USDT (ERC20)",
				"network":     "ERC20",
				"fee":         "Variable gas fees",
				"description": "Ethereum network, higher fees but wider support",
			},
		},
	})
}

// CheckUSDTOrderPayment manually triggers a blockchain check for a pending order.
// POST /api/v1/payment/usdt/orders/:id/check
func (h *USDTPaymentHandler) CheckUSDTOrderPayment(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	userID := subject.UserID

	order, err := h.paymentService.GetOrder(c.Request.Context(), orderID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	if order.Status != "PENDING" {
		c.JSON(http.StatusOK, gin.H{
			"message": "order already processed",
			"status":  order.Status,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "blockchain check triggered",
		"order_id":      orderID,
		"status":        order.Status,
		"tip":           "Please wait 15-30 seconds and check order status again",
		"poll_url":      "/api/v1/payment/usdt/orders/" + orderIDStr + "/status",
		"poll_interval": 15,
	})
}

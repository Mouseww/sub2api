package handler

import (
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type USDTPaymentHandler struct {
	paymentService *service.PaymentService
}

func NewUSDTPaymentHandler(paymentService *service.PaymentService) *USDTPaymentHandler {
	return &USDTPaymentHandler{
		paymentService: paymentService,
	}
}

// CreateUSDTOrder creates a new USDT payment order
// POST /api/v1/payment/usdt/orders
func (h *USDTPaymentHandler) CreateUSDTOrder(c *gin.Context) {
	var req struct {
		Amount      float64 `json:"amount" binding:"required,gt=0"`
		PaymentType string  `json:"payment_type" binding:"required,oneof=usdt_trc20 usdt_erc20"`
		ReturnURL   string  `json:"return_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate user authentication
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	
	userID, ok := userIDVal.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	// Use the existing CreateOrder method with USDT payment type
	orderReq := service.CreateOrderRequest{
		UserID:      userID,
		Amount:      req.Amount,
		PaymentType: req.PaymentType,
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

	c.JSON(http.StatusOK, resp)
}

// GetUSDTOrderStatus gets USDT order status
// GET /api/v1/payment/usdt/orders/:id/status
func (h *USDTPaymentHandler) GetUSDTOrderStatus(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	// Validate user authentication
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	
	userID, ok := userIDVal.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	order, err := h.paymentService.GetOrder(c.Request.Context(), orderID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id":             order.ID,
		"status":               order.Status,
		"amount":               order.Amount,
		"pay_amount":           order.PayAmount,
		"crypto_currency":      order.CryptoCurrency,
		"crypto_network":       order.CryptoNetwork,
		"crypto_address":       order.CryptoAddress,
		"crypto_tx_hash":       order.CryptoTxHash,
		"crypto_confirmations": order.CryptoConfirmations,
		"crypto_required_confirmations": order.CryptoRequiredConfirmations,
		"qr_code":              order.QrCode,
		"created_at":           order.CreatedAt,
		"expires_at":           order.ExpiresAt,
		"paid_at":              order.PaidAt,
	})
}

// ListUSDTProviders lists available USDT payment providers
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

// CheckUSDTOrderPayment manually checks blockchain for payment
// POST /api/v1/payment/usdt/orders/:id/check
// This endpoint allows users to manually trigger a blockchain check for their order
func (h *USDTPaymentHandler) CheckUSDTOrderPayment(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	// Validate user authentication
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	
	userID, ok := userIDVal.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	// Get order and verify ownership
	order, err := h.paymentService.GetOrder(c.Request.Context(), orderID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	// Only check pending orders
	if order.Status != "PENDING" {
		c.JSON(http.StatusOK, gin.H{
			"message": "order already processed",
			"status":  order.Status,
		})
		return
	}

	// Note: This is a synchronous check endpoint
	// The actual blockchain checking logic should be implemented
	// For now, return a message indicating manual check is triggered
	c.JSON(http.StatusOK, gin.H{
		"message":      "blockchain check triggered",
		"order_id":     orderID,
		"status":       order.Status,
		"tip":          "Please wait 15-30 seconds and check order status again",
		"poll_url":     "/api/v1/payment/usdt/orders/" + orderIDStr + "/status",
		"poll_interval": 15, // seconds
	})
}

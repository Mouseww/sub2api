package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/internal/payment/provider"
	"go.uber.org/zap"
)

// BlockchainMonitorService 区块链监控服务
// 定期轮询区块链，检测 USDT 充值交易并自动完成订单
type BlockchainMonitorService struct {
	entClient      *ent.Client
	paymentService *PaymentService
	logger         *zap.Logger
	
	// 监控配置
	trc20Config *provider.BlockchainMonitorConfig
	erc20Config *provider.BlockchainMonitorConfig
	
	// 运行状态
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	isRunning  bool
	mu         sync.RWMutex
}

// NewBlockchainMonitorService 创建监控服务
func NewBlockchainMonitorService(
	entClient *ent.Client,
	paymentService *PaymentService,
	logger *zap.Logger,
	trc20Config, erc20Config *provider.BlockchainMonitorConfig,
) *BlockchainMonitorService {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &BlockchainMonitorService{
		entClient:      entClient,
		paymentService: paymentService,
		logger:         logger,
		trc20Config:    trc20Config,
		erc20Config:    erc20Config,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Start 启动监控服务
func (s *BlockchainMonitorService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.isRunning {
		return fmt.Errorf("blockchain monitor already running")
	}
	
	s.isRunning = true
	
	// 启动 TRC20 监控
	if s.trc20Config != nil {
		s.wg.Add(1)
		go s.monitorNetwork(s.trc20Config, "TRC20")
	}
	
	// 启动 ERC20 监控
	if s.erc20Config != nil {
		s.wg.Add(1)
		go s.monitorNetwork(s.erc20Config, "ERC20")
	}
	
	s.logger.Info("Blockchain monitor service started")
	return nil
}

// Stop 停止监控服务
func (s *BlockchainMonitorService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.isRunning {
		return nil
	}
	
	s.cancel()
	s.wg.Wait()
	s.isRunning = false
	
	s.logger.Info("Blockchain monitor service stopped")
	return nil
}

// monitorNetwork 监控单个网络
func (s *BlockchainMonitorService) monitorNetwork(config *provider.BlockchainMonitorConfig, networkName string) {
	defer s.wg.Done()
	
	ticker := time.NewTicker(config.PollInterval)
	defer ticker.Stop()
	
	s.logger.Info("Started monitoring network",
		zap.String("network", networkName),
		zap.Duration("interval", config.PollInterval),
	)
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if err := s.checkPendingOrders(config, networkName); err != nil {
				s.logger.Error("Failed to check pending orders",
					zap.String("network", networkName),
					zap.Error(err),
				)
			}
		}
	}
}

// checkPendingOrders 检查待处理订单
func (s *BlockchainMonitorService) checkPendingOrders(config *provider.BlockchainMonitorConfig, networkName string) error {
	ctx := context.Background()
	
	// 查询所有待支付的 USDT 订单
	orders, err := s.entClient.PaymentOrder.Query().
		Where(
			paymentorder.StatusEQ("PENDING"),
			paymentorder.Or(
				paymentorder.PaymentTypeEQ("usdt_trc20"),
				paymentorder.PaymentTypeEQ("usdt_erc20"),
			),
			paymentorder.CryptoNetworkEQ(networkName),
			paymentorder.ExpiresAtGT(time.Now()), // 未过期
		).
		All(ctx)
	
	if err != nil {
		return fmt.Errorf("query pending orders: %w", err)
	}
	
	if len(orders) == 0 {
		return nil
	}
	
	s.logger.Debug("Checking pending orders",
		zap.String("network", networkName),
		zap.Int("count", len(orders)),
	)
	
	// 检查每个订单
	for _, order := range orders {
		if err := s.checkOrder(ctx, order, config); err != nil {
			s.logger.Error("Failed to check order",
				zap.Int64("order_id", order.ID),
				zap.Error(err),
			)
		}
	}
	
	return nil
}

// checkOrder 检查单个订单
func (s *BlockchainMonitorService) checkOrder(ctx context.Context, order *ent.PaymentOrder, config *provider.BlockchainMonitorConfig) error {
	// 如果订单没有充值地址或金额，跳过
	if order.CryptoAddress == nil || order.PayAmount == 0 {
		return nil
	}
	
	expectedAmount := order.PayAmount
	depositAddress := *order.CryptoAddress
	
	// 查询区块链上的交易
	tx, err := s.fetchTransaction(config, depositAddress, expectedAmount)
	if err != nil {
		return fmt.Errorf("fetch transaction: %w", err)
	}
	
	if tx == nil {
		// 未找到匹配交易
		return nil
	}
	
	// 验证交易
	valid, reason := provider.VerifyTransaction(tx, depositAddress, expectedAmount, config.RequiredConfirms)
	if !valid {
		s.logger.Debug("Transaction verification failed",
			zap.Int64("order_id", order.ID),
			zap.String("reason", reason),
			zap.Int("confirmations", tx.Confirmations),
			zap.Int("required", config.RequiredConfirms),
		)
		
		// 更新订单的确认数（即使未完全确认）
		if tx.Confirmations > 0 {
			_, _ = s.entClient.PaymentOrder.UpdateOneID(order.ID).
				SetCryptoTxHash(tx.TxHash).
				SetCryptoConfirmations(tx.Confirmations).
				Save(ctx)
		}
		
		return nil
	}
	
	// 交易验证通过，完成订单
	s.logger.Info("Transaction verified, completing order",
		zap.Int64("order_id", order.ID),
		zap.String("tx_hash", tx.TxHash),
		zap.Float64("amount", tx.Amount),
		zap.Int("confirmations", tx.Confirmations),
	)
	
	// 更新订单状态为已支付
	if err := s.completeOrder(ctx, order, tx); err != nil {
		return fmt.Errorf("complete order: %w", err)
	}
	
	return nil
}

// completeOrder 完成订单并充值
func (s *BlockchainMonitorService) completeOrder(ctx context.Context, order *ent.PaymentOrder, tx *provider.USDTTransactionInfo) error {
	// 更新订单信息
	now := time.Now()
	_, err := s.entClient.PaymentOrder.UpdateOneID(order.ID).
		SetStatus("PAID").
		SetCryptoTxHash(tx.TxHash).
		SetCryptoConfirmations(tx.Confirmations).
		SetPaidAt(now).
		SetCompletedAt(now).
		Save(ctx)
	
	if err != nil {
		return fmt.Errorf("update order: %w", err)
	}
	
	// 增加用户余额
	user, err := s.entClient.User.Get(ctx, order.UserID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	
	newBalance := user.Balance + order.Amount
	_, err = s.entClient.User.UpdateOneID(order.UserID).
		SetBalance(newBalance).
		Save(ctx)
	
	if err != nil {
		return fmt.Errorf("update user balance: %w", err)
	}
	
	s.logger.Info("Order completed and balance credited",
		zap.Int64("order_id", order.ID),
		zap.Int64("user_id", order.UserID),
		zap.Float64("amount", order.Amount),
		zap.Float64("new_balance", newBalance),
	)
	
	return nil
}

// fetchTransaction 从区块链查询交易
func (s *BlockchainMonitorService) fetchTransaction(
	config *provider.BlockchainMonitorConfig,
	address string,
	expectedAmount float64,
) (*provider.USDTTransactionInfo, error) {
	// 根据网络类型选择 API
	if config.NetworkID == "TRC20" {
		return s.fetchTRC20Transaction(config, address, expectedAmount)
	} else if config.NetworkID == "ERC20" {
		return s.fetchERC20Transaction(config, address, expectedAmount)
	}
	
	return nil, fmt.Errorf("unsupported network: %s", config.NetworkID)
}

// fetchTRC20Transaction 查询 TRC20 交易
func (s *BlockchainMonitorService) fetchTRC20Transaction(
	config *provider.BlockchainMonitorConfig,
	address string,
	expectedAmount float64,
) (*provider.USDTTransactionInfo, error) {
	// 构建 TronGrid API 请求
	url := fmt.Sprintf("%s/v1/accounts/%s/transactions/trc20?limit=20&contract_address=%s",
		config.BlockchainAPIURL,
		address,
		config.ContractAddress,
	)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	if config.BlockchainAPIKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", config.BlockchainAPIKey)
	}
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	// 解析响应
	var result struct {
		Data []struct {
			TransactionID string `json:"transaction_id"`
			From          string `json:"from"`
			To            string `json:"to"`
			Value         string `json:"value"`
			BlockTimestamp int64 `json:"block_timestamp"`
		} `json:"data"`
	}
	
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	
	// 查找匹配金额的交易
	for _, tx := range result.Data {
		// 将字符串金额转换为 float64（USDT 有 6 位小数）
		var amount float64
		fmt.Sscanf(tx.Value, "%f", &amount)
		amount = amount / 1000000 // TRC20 USDT 使用 6 位小数
		
		// 检查金额是否匹配（精确到 6 位小数）
		if roundTo6Decimals(amount) == roundTo6Decimals(expectedAmount) {
			// 获取交易确认数（需要查询当前区块高度）
			confirmations, err := s.getTRC20Confirmations(config, tx.TransactionID)
			if err != nil {
				s.logger.Warn("Failed to get confirmations", zap.Error(err))
				confirmations = 0
			}
			
			return &provider.USDTTransactionInfo{
				TxHash:        tx.TransactionID,
				FromAddress:   tx.From,
				ToAddress:     tx.To,
				Amount:        amount,
				Confirmations: confirmations,
				Timestamp:     time.Unix(tx.BlockTimestamp/1000, 0),
				Status:        "success",
			}, nil
		}
	}
	
	return nil, nil // 未找到匹配交易
}

// fetchERC20Transaction 查询 ERC20 交易
func (s *BlockchainMonitorService) fetchERC20Transaction(
	config *provider.BlockchainMonitorConfig,
	address string,
	expectedAmount float64,
) (*provider.USDTTransactionInfo, error) {
	// 构建 Etherscan API 请求
	url := fmt.Sprintf("%s/api?module=account&action=tokentx&contractaddress=%s&address=%s&sort=desc&apikey=%s",
		config.BlockchainAPIURL,
		config.ContractAddress,
		address,
		config.BlockchainAPIKey,
	)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	// 解析响应
	var result struct {
		Status string `json:"status"`
		Result []struct {
			Hash             string `json:"hash"`
			From             string `json:"from"`
			To               string `json:"to"`
			Value            string `json:"value"`
			TimeStamp        string `json:"timeStamp"`
			Confirmations    string `json:"confirmations"`
		} `json:"result"`
	}
	
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	
	if result.Status != "1" {
		return nil, nil // API 错误或无结果
	}
	
	// 查找匹配金额的交易
	for _, tx := range result.Result {
		// 将字符串金额转换为 float64（USDT 有 6 位小数）
		var amount float64
		fmt.Sscanf(tx.Value, "%f", &amount)
		amount = amount / 1000000 // ERC20 USDT 使用 6 位小数
		
		// 检查金额是否匹配
		if roundTo6Decimals(amount) == roundTo6Decimals(expectedAmount) {
			var confirmations int
			fmt.Sscanf(tx.Confirmations, "%d", &confirmations)
			
			var timestamp int64
			fmt.Sscanf(tx.TimeStamp, "%d", &timestamp)
			
			return &provider.USDTTransactionInfo{
				TxHash:        tx.Hash,
				FromAddress:   tx.From,
				ToAddress:     tx.To,
				Amount:        amount,
				Confirmations: confirmations,
				Timestamp:     time.Unix(timestamp, 0),
				Status:        "success",
			}, nil
		}
	}
	
	return nil, nil // 未找到匹配交易
}

// getTRC20Confirmations 获取 TRC20 交易的确认数
func (s *BlockchainMonitorService) getTRC20Confirmations(
	config *provider.BlockchainMonitorConfig,
	txHash string,
) (int, error) {
	// 获取交易详情
	url := fmt.Sprintf("%s/wallet/gettransactioninfobyid?value=%s",
		config.BlockchainAPIURL,
		txHash,
	)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	
	if config.BlockchainAPIKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", config.BlockchainAPIKey)
	}
	
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	
	var txInfo struct {
		BlockNumber int64 `json:"blockNumber"`
	}
	
	if err := json.Unmarshal(body, &txInfo); err != nil {
		return 0, err
	}
	
	// 获取当前区块高度
	currentBlock, err := s.getTRC20CurrentBlock(config)
	if err != nil {
		return 0, err
	}
	
	confirmations := int(currentBlock - txInfo.BlockNumber)
	if confirmations < 0 {
		confirmations = 0
	}
	
	return confirmations, nil
}

// getTRC20CurrentBlock 获取 TRC20 当前区块高度
func (s *BlockchainMonitorService) getTRC20CurrentBlock(config *provider.BlockchainMonitorConfig) (int64, error) {
	url := fmt.Sprintf("%s/wallet/getnowblock", config.BlockchainAPIURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	
	if config.BlockchainAPIKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", config.BlockchainAPIKey)
	}
	
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	
	var result struct {
		BlockHeader struct {
			RawData struct {
				Number int64 `json:"number"`
			} `json:"raw_data"`
		} `json:"block_header"`
	}
	
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}
	
	return result.BlockHeader.RawData.Number, nil
}

// roundTo6Decimals 四舍五入到 6 位小数
func roundTo6Decimals(amount float64) float64 {
	return float64(int64(amount*1000000+0.5)) / 1000000
}

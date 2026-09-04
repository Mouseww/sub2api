# USDT 区块链监控服务部署指南

## 概述

区块链监控服务自动检测 USDT 充值交易并完成订单，实现自动充值功能。

## 功能特性

- ✅ 自动监控 TRC20 和 ERC20 网络
- ✅ 精确匹配订单金额（6位小数精度）
- ✅ 验证交易确认数（TRC20: 19确认，ERC20: 12确认）
- ✅ 自动完成订单并充值用户余额
- ✅ 支持多订单并发处理
- ✅ 错误重试和日志记录

## 配置

### 1. 环境变量

```bash
# TRC20 配置
export USDT_TRC20_DEPOSIT_ADDRESS="TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
export USDT_TRC20_CONTRACT="TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
export USDT_TRC20_API_URL="https://api.trongrid.io"
export USDT_TRC20_API_KEY="your-trongrid-api-key"
export USDT_TRC20_HMAC_SECRET="your-32-byte-hmac-secret"

# ERC20 配置
export USDT_ERC20_DEPOSIT_ADDRESS="0xdAC17F958D2ee523a2206206994597C13D831ec7"
export USDT_ERC20_CONTRACT="0xdAC17F958D2ee523a2206206994597C13D831ec7"
export USDT_ERC20_API_URL="https://api.etherscan.io"
export USDT_ERC20_API_KEY="your-etherscan-api-key"
export USDT_ERC20_HMAC_SECRET="your-32-byte-hmac-secret"
```

### 2. 配置文件（config.yaml）

```yaml
blockchain_monitor:
  enabled: true
  
  trc20:
    network_id: "TRC20"
    deposit_address: "${USDT_TRC20_DEPOSIT_ADDRESS}"
    contract_address: "${USDT_TRC20_CONTRACT}"
    blockchain_api_url: "${USDT_TRC20_API_URL}"
    blockchain_api_key: "${USDT_TRC20_API_KEY}"
    poll_interval_seconds: 15  # 每15秒检查一次
    confirmations: 19          # TRC20 需要 19 个确认
    
  erc20:
    network_id: "ERC20"
    deposit_address: "${USDT_ERC20_DEPOSIT_ADDRESS}"
    contract_address: "${USDT_ERC20_CONTRACT}"
    blockchain_api_url: "${USDT_ERC20_API_URL}"
    blockchain_api_key: "${USDT_ERC20_API_KEY}"
    poll_interval_seconds: 30  # 每30秒检查一次（ETH较慢）
    confirmations: 12          # ERC20 需要 12 个确认
```

## 代码集成

### 在 main.go 中启动服务

```go
package main

import (
    "github.com/Wei-Shaw/sub2api/internal/service"
    "github.com/Wei-Shaw/sub2api/internal/payment/provider"
)

func main() {
    // ... 初始化 entClient, paymentService, logger
    
    // 加载配置
    trc20Config := &provider.BlockchainMonitorConfig{
        NetworkID:           "TRC20",
        DepositAddress:      os.Getenv("USDT_TRC20_DEPOSIT_ADDRESS"),
        ContractAddress:     os.Getenv("USDT_TRC20_CONTRACT"),
        BlockchainAPIURL:    os.Getenv("USDT_TRC20_API_URL"),
        BlockchainAPIKey:    os.Getenv("USDT_TRC20_API_KEY"),
        PollInterval:        15 * time.Second,
        RequiredConfirms:    19,
        MaxConfirmationTime: 30 * time.Minute,
    }
    
    erc20Config := &provider.BlockchainMonitorConfig{
        NetworkID:           "ERC20",
        DepositAddress:      os.Getenv("USDT_ERC20_DEPOSIT_ADDRESS"),
        ContractAddress:     os.Getenv("USDT_ERC20_CONTRACT"),
        BlockchainAPIURL:    os.Getenv("USDT_ERC20_API_URL"),
        BlockchainAPIKey:    os.Getenv("USDT_ERC20_API_KEY"),
        PollInterval:        30 * time.Second,
        RequiredConfirms:    12,
        MaxConfirmationTime: 60 * time.Minute,
    }
    
    // 创建并启动监控服务
    monitorService := service.NewBlockchainMonitorService(
        entClient,
        paymentService,
        logger,
        trc20Config,
        erc20Config,
    )
    
    if err := monitorService.Start(); err != nil {
        logger.Fatal("Failed to start blockchain monitor", zap.Error(err))
    }
    
    // 优雅关闭
    defer monitorService.Stop()
    
    // ... 启动 HTTP 服务器
}
```

## API 密钥获取

### TronGrid API Key (TRC20)

1. 访问 https://www.trongrid.io/
2. 注册账号
3. 创建 API Key
4. 免费版限制：1000 请求/天

### Etherscan API Key (ERC20)

1. 访问 https://etherscan.io/apis
2. 注册账号
3. 创建 API Key
4. 免费版限制：5 请求/秒

## 工作流程

```
1. 用户创建订单
   ↓
2. 生成唯一金额（如 10.234567 USDT）
   ↓
3. 用户发送 USDT 到平台地址
   ↓
4. 监控服务每 15-30 秒检查一次
   ↓
5. 发现匹配金额的交易
   ↓
6. 验证确认数（TRC20: 19，ERC20: 12）
   ↓
7. 自动完成订单
   ↓
8. 自动充值用户余额
```

## 监控指标

### 日志示例

```
INFO  Started monitoring network  network=TRC20 interval=15s
DEBUG Checking pending orders      network=TRC20 count=5
INFO  Transaction verified         order_id=12345 tx_hash=abc... confirmations=19
INFO  Order completed               order_id=12345 user_id=789 amount=10.234567
INFO  Balance credited             user_id=789 new_balance=100.50
```

### 健康检查

```go
// 添加健康检查端点
router.GET("/health/blockchain-monitor", func(c *gin.Context) {
    c.JSON(200, gin.H{
        "status": "running",
        "trc20_enabled": trc20Config != nil,
        "erc20_enabled": erc20Config != nil,
    })
})
```

## 故障处理

### 常见问题

1. **API 超时**
   - 检查网络连接
   - 验证 API Key 是否有效
   - 检查 API 限流

2. **交易未检测到**
   - 确认用户发送的金额精确匹配
   - 检查区块链浏览器确认交易存在
   - 验证合约地址配置正确

3. **订单未自动完成**
   - 检查确认数是否达到要求
   - 查看监控服务日志
   - 手动触发订单检查（管理员接口）

### 手动触发检查

```bash
# 管理员手动检查订单
POST /api/v1/admin/payment/usdt/orders/:id/check

# 查看订单状态
GET /api/v1/payment/usdt/orders/:id/status
```

## 性能优化

### 1. 批量查询

当待处理订单很多时，可以批量查询区块链：

```go
// 一次查询获取最近 20 笔交易
// 然后匹配所有待处理订单
```

### 2. 缓存机制

```go
// 缓存已检查的交易，避免重复查询
var checkedTxCache = make(map[string]bool)
```

### 3. Webhook 替代轮询（推荐）

使用 Tatum 或 Alchemy 的 Webhook：

```yaml
webhook:
  enabled: true
  provider: "tatum"  # or "alchemy"
  url: "https://yourdomain.com/webhook/blockchain"
  
  tatum_api_key: "your-tatum-key"
```

## 安全建议

1. **私钥管理**：使用 KMS/HSM，永不存储在代码或数据库
2. **API 密钥**：通过环境变量注入，不要硬编码
3. **网络隔离**：监控服务运行在内网，不暴露公网
4. **日志脱敏**：不记录完整地址和交易哈希（仅前后4位）
5. **访问控制**：管理员接口需要严格权限验证

## 测试

### 测试网测试

```bash
# 使用 Shasta 测试网（TRC20）
export USDT_TRC20_API_URL="https://api.shasta.trongrid.io"
export USDT_TRC20_DEPOSIT_ADDRESS="your-test-address"

# 使用 Sepolia 测试网（ERC20）
export USDT_ERC20_API_URL="https://api-sepolia.etherscan.io"
export USDT_ERC20_DEPOSIT_ADDRESS="your-test-address"
```

### 单元测试

```bash
cd backend
go test ./internal/service -run TestBlockchainMonitor -v
```

## 生产环境清单

- [ ] 配置生产钱包地址
- [ ] 设置 API 密钥环境变量
- [ ] 启动监控服务
- [ ] 验证日志正常输出
- [ ] 测试一笔小额充值
- [ ] 配置告警（订单长时间未完成）
- [ ] 设置日志收集和监控

## 告警规则

```yaml
alerts:
  - name: "USDT订单长时间未完成"
    condition: "订单创建超过2小时仍为PENDING"
    action: "通知管理员检查"
    
  - name: "区块链API故障"
    condition: "连续3次API调用失败"
    action: "发送告警邮件"
    
  - name: "监控服务停止"
    condition: "5分钟内无日志输出"
    action: "重启服务并告警"
```

## 维护

### 日常检查

- 每天检查监控服务日志
- 每周检查未完成订单数量
- 每月验证 API 密钥有效性

### 升级

```bash
# 停止服务
systemctl stop sub2api

# 部署新版本
./deploy.sh

# 启动服务
systemctl start sub2api

# 检查日志
journalctl -u sub2api -f
```

---

## 联系支持

遇到问题？
- 查看日志：`/var/log/sub2api/blockchain-monitor.log`
- 提交 Issue：https://github.com/your-repo/issues
- 技术支持：support@yourcompany.com

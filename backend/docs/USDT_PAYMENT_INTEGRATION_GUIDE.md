# USDT 支付集成 - 完整部署指南

## 🎉 功能概述

Sub2API 现已完全集成 USDT 加密货币充值功能，支持：

- ✅ **TRC20 网络**（推荐）：低手续费（1-2 USDT），快速确认（1-3分钟）
- ✅ **ERC20 网络**：以太坊网络，更广泛支持
- ✅ **自动监控充值**：区块链监控服务自动检测交易并充值余额
- ✅ **手动轮询接口**：用户可主动查询订单状态
- ✅ **HMAC 加密金额生成**：防碰撞，可追溯
- ✅ **精确金额匹配**：6 位小数，零容差
- ✅ **完整的安全验证**：确认数、地址、金额三重验证

---

## 📋 部署清单

### 1. 准备钱包地址

#### 创建 TRC20 收款钱包
```bash
# 使用 TronLink 或其他 Tron 钱包
# 记录钱包地址（以 T 开头）
# 示例：TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t
```

#### 创建 ERC20 收款钱包（可选）
```bash
# 使用 MetaMask 或其他以太坊钱包
# 记录钱包地址（以 0x 开头）
# 示例：0xdAC17F958D2ee523a2206206994597C13D831ec7
```

**⚠️ 安全提示**：
- 私钥必须存储在 **KMS/HSM** 中
- 绝不要将私钥存储在数据库或代码中
- 定期备份钱包助记词

---

### 2. 获取区块链 API 密钥

#### TronGrid API Key（免费）
1. 访问：https://www.trongrid.io/
2. 注册账号并登录
3. 创建 API Key
4. 限制：1000 请求/天（免费版）

#### Etherscan API Key（免费）
1. 访问：https://etherscan.io/apis
2. 注册账号并登录
3. 创建 API Key
4. 限制：5 请求/秒（免费版）

---

### 3. 配置环境变量

创建或编辑 `.env` 文件：

```bash
# ==================== TRC20 配置 ====================
# TRC20 收款地址（必填）
USDT_TRC20_DEPOSIT_ADDRESS=TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t

# TRC20 合约地址（USDT 官方合约，通常不需要修改）
USDT_TRC20_CONTRACT=TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t

# TronGrid API 配置
USDT_TRC20_API_URL=https://api.trongrid.io
USDT_TRC20_API_KEY=你的TronGrid_API密钥

# HMAC 密钥（32字节以上的随机字符串，必填）
USDT_TRC20_HMAC_SECRET=your-32-byte-hmac-secret-key-here-change-me

# 轮询间隔（可选，默认15秒）
USDT_TRC20_POLL_INTERVAL=15

# 确认数要求（可选，默认19）
USDT_TRC20_CONFIRMATIONS=19

# ==================== ERC20 配置（可选）====================
# ERC20 收款地址
USDT_ERC20_DEPOSIT_ADDRESS=0xdAC17F958D2ee523a2206206994597C13D831ec7

# ERC20 合约地址（USDT 官方合约）
USDT_ERC20_CONTRACT=0xdAC17F958D2ee523a2206206994597C13D831ec7

# Etherscan API 配置
USDT_ERC20_API_URL=https://api.etherscan.io
USDT_ERC20_API_KEY=你的Etherscan_API密钥

# HMAC 密钥（可与 TRC20 使用不同密钥）
USDT_ERC20_HMAC_SECRET=your-32-byte-hmac-secret-key-here-change-me

# 轮询间隔（可选，默认30秒）
USDT_ERC20_POLL_INTERVAL=30

# 确认数要求（可选，默认12）
USDT_ERC20_CONFIRMATIONS=12

# ==================== 监控服务控制 ====================
# 禁用自动监控（设置为1则只能手动查询，不自动监控）
# BLOCKCHAIN_MONITOR_DISABLED=1
```

**生成 HMAC 密钥**：
```bash
# Linux/Mac
openssl rand -hex 32

# PowerShell
[Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }))

# Node.js
node -e "console.log(require('crypto').randomBytes(32).toString('hex'))"
```

---

### 4. 启动服务

```bash
cd backend

# 方式1：直接运行
./bin/server.exe

# 方式2：使用 systemd（推荐生产环境）
sudo systemctl start sub2api

# 方式3：使用 Docker
docker-compose up -d
```

**启动日志示例**：
```
INFO  Blockchain monitor service started
INFO  Started monitoring network  network=TRC20 interval=15s
INFO  Started monitoring network  network=ERC20 interval=30s
Server started on :3000
```

---

## 🔧 工作模式

### 模式 1：自动监控（推荐）

**默认启用**，区块链监控服务自动运行：

```
用户创建订单 → 生成唯一金额（如 10.234567 USDT）
    ↓
用户转账到平台地址
    ↓
监控服务每 15-30 秒检查区块链
    ↓
发现匹配金额的交易
    ↓
验证确认数（TRC20: 19，ERC20: 12）
    ↓
自动完成订单并充值余额 ✅
```

**特点**：
- ✅ 完全自动化
- ✅ 用户无需操作
- ✅ 确认后立即到账

---

### 模式 2：手动轮询

用户可以主动查询订单状态：

#### API 端点

**创建订单**：
```http
POST /api/v1/payment/usdt/orders
Content-Type: application/json
Authorization: Bearer <token>

{
  "amount": 10.00,
  "payment_type": "usdt_trc20",  // 或 "usdt_erc20"
  "return_url": "https://yoursite.com/payment/success"
}
```

**响应**：
```json
{
  "order_id": 12345,
  "amount": 10.00,
  "pay_amount": 10.234567,  // 唯一金额
  "crypto_address": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
  "qr_code": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t?amount=10.234567",
  "expires_at": "2024-01-20T10:30:00Z"
}
```

---

**查询订单状态**：
```http
GET /api/v1/payment/usdt/orders/:id/status
Authorization: Bearer <token>
```

**响应**：
```json
{
  "order_id": 12345,
  "status": "PENDING",  // PENDING | PAID | EXPIRED | FAILED
  "amount": 10.00,
  "pay_amount": 10.234567,
  "crypto_currency": "USDT",
  "crypto_network": "TRC20",
  "crypto_address": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
  "crypto_confirmations": 5,
  "crypto_required_confirmations": 19,
  "qr_code": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t?amount=10.234567",
  "created_at": "2024-01-20T10:00:00Z",
  "expires_at": "2024-01-20T10:30:00Z"
}
```

---

**手动触发检查**（可选）：
```http
POST /api/v1/payment/usdt/orders/:id/check
Authorization: Bearer <token>
```

**响应**：
```json
{
  "message": "blockchain check triggered",
  "order_id": 12345,
  "status": "PENDING",
  "tip": "Please wait 15-30 seconds and check order status again",
  "poll_url": "/api/v1/payment/usdt/orders/12345/status",
  "poll_interval": 15
}
```

---

### 前端轮询示例

```javascript
// 创建订单后开始轮询
async function pollOrderStatus(orderId) {
  const maxAttempts = 40; // 最多轮询 40 次（10 分钟）
  const interval = 15000; // 15 秒间隔
  
  for (let i = 0; i < maxAttempts; i++) {
    const response = await fetch(`/api/v1/payment/usdt/orders/${orderId}/status`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    
    const data = await response.json();
    
    if (data.status === 'PAID') {
      console.log('充值成功！');
      return data;
    }
    
    if (data.status === 'EXPIRED' || data.status === 'FAILED') {
      console.log('订单失败');
      return data;
    }
    
    // 显示确认进度
    console.log(`确认进度: ${data.crypto_confirmations}/${data.crypto_required_confirmations}`);
    
    // 等待下一次轮询
    await new Promise(resolve => setTimeout(resolve, interval));
  }
  
  console.log('轮询超时');
}
```

---

## 🎨 前端集成示例

### Vue 3 组件示例

```vue
<template>
  <div class="usdt-payment">
    <h2>USDT 充值</h2>
    
    <!-- 选择网络 -->
    <div class="network-selector">
      <button @click="network = 'TRC20'" :class="{ active: network === 'TRC20' }">
        TRC20 (推荐)
      </button>
      <button @click="network = 'ERC20'" :class="{ active: network === 'ERC20' }">
        ERC20
      </button>
    </div>
    
    <!-- 输入金额 -->
    <input v-model="amount" type="number" placeholder="充值金额（USDT）" />
    
    <!-- 创建订单 -->
    <button @click="createOrder" :disabled="loading">创建订单</button>
    
    <!-- 显示 QR 码 -->
    <div v-if="order" class="qr-section">
      <p>请发送 <strong>{{ order.pay_amount }} USDT</strong> 到以下地址：</p>
      <QRCode :value="order.qr_code" />
      <p>地址：{{ order.crypto_address }}</p>
      <button @click="copyAddress">复制地址</button>
      
      <!-- 确认进度 -->
      <div class="progress">
        <p>确认进度：{{ confirmations }}/{{ order.crypto_required_confirmations }}</p>
        <progress :value="confirmations" :max="order.crypto_required_confirmations"></progress>
      </div>
      
      <!-- 手动刷新按钮 -->
      <button @click="checkStatus">刷新状态</button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import QRCode from 'qrcode.vue';

const network = ref('TRC20');
const amount = ref(10);
const order = ref(null);
const loading = ref(false);
const confirmations = ref(0);
let pollTimer = null;

async function createOrder() {
  loading.value = true;
  try {
    const response = await fetch('/api/v1/payment/usdt/orders', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify({
        amount: amount.value,
        payment_type: network.value === 'TRC20' ? 'usdt_trc20' : 'usdt_erc20'
      })
    });
    
    order.value = await response.json();
    startPolling();
  } catch (error) {
    console.error('创建订单失败', error);
  } finally {
    loading.value = false;
  }
}

async function checkStatus() {
  if (!order.value) return;
  
  const response = await fetch(`/api/v1/payment/usdt/orders/${order.value.order_id}/status`, {
    headers: {
      'Authorization': `Bearer ${localStorage.getItem('token')}`
    }
  });
  
  const data = await response.json();
  confirmations.value = data.crypto_confirmations;
  
  if (data.status === 'PAID') {
    stopPolling();
    alert('充值成功！');
    // 刷新余额
    location.reload();
  }
}

function startPolling() {
  stopPolling();
  pollTimer = setInterval(checkStatus, 15000); // 每 15 秒查询一次
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
}

function copyAddress() {
  navigator.clipboard.writeText(order.value.crypto_address);
  alert('地址已复制');
}

onUnmounted(() => {
  stopPolling();
});
</script>
```

---

## 🔍 监控和调试

### 查看日志

```bash
# 实时日志
journalctl -u sub2api -f

# 筛选区块链监控日志
journalctl -u sub2api -f | grep "Blockchain"

# 查看最近的错误
journalctl -u sub2api -p err -n 50
```

### 日志示例

**正常运行**：
```
INFO  Started monitoring network  network=TRC20 interval=15s
DEBUG Checking pending orders      network=TRC20 count=3
INFO  Transaction verified         order_id=12345 tx_hash=abc... confirmations=19
INFO  Order completed               order_id=12345 user_id=789 amount=10.234567
INFO  Balance credited             user_id=789 new_balance=110.50
```

**常见错误**：
```
ERROR Failed to check order        order_id=12345 error="API timeout"
WARN  Transaction verification failed  reason="insufficient confirmations: 5/19"
ERROR Failed to start blockchain monitor  error="API key invalid"
```

---

## 🐛 故障排查

### 问题 1：订单未自动完成

**检查清单**：
1. ✅ 确认用户发送的金额精确匹配（6 位小数）
2. ✅ 检查区块链浏览器确认交易存在
3. ✅ 验证确认数是否达到要求（TRC20: 19，ERC20: 12）
4. ✅ 检查监控服务是否运行
5. ✅ 检查 API 密钥是否有效

**手动检查**：
```bash
# 检查服务状态
systemctl status sub2api

# 检查最近日志
journalctl -u sub2api -n 100

# 查看环境变量
env | grep USDT
```

---

### 问题 2：API 超时或限流

**解决方案**：
```bash
# 增加轮询间隔
export USDT_TRC20_POLL_INTERVAL=30
export USDT_ERC20_POLL_INTERVAL=60

# 或升级 API 套餐
# TronGrid: https://www.trongrid.io/pricing
# Etherscan: https://etherscan.io/apis
```

---

### 问题 3：金额不匹配

**原因**：
- 用户发送的金额有误（如发送 10.23 而不是 10.234567）
- 钱包精度问题（某些钱包可能截断小数位）

**解决方案**：
1. 前端显示完整金额（6 位小数）
2. 提供复制功能
3. 建议用户使用 QR 码扫码支付

---

## 📊 性能优化

### 1. 使用 Webhook（推荐）

代替轮询，使用 Webhook 实时接收通知：

```yaml
# 未来版本支持
webhook:
  enabled: true
  provider: tatum  # 或 alchemy
  url: https://yourdomain.com/webhook/blockchain
```

### 2. 批量查询

优化代码以批量查询多个订单：

```go
// 一次 API 调用获取最近 20 笔交易
// 然后匹配所有待处理订单
```

### 3. 缓存机制

缓存已检查的交易，避免重复查询：

```go
var checkedTxCache = make(map[string]time.Time)
```

---

## 🔒 安全建议

1. **私钥管理**：
   - ✅ 使用 AWS KMS / Azure Key Vault / HashiCorp Vault
   - ❌ 绝不存储在数据库或代码中

2. **API 密钥**：
   - ✅ 通过环境变量注入
   - ❌ 不要硬编码

3. **网络隔离**：
   - ✅ 监控服务运行在内网
   - ✅ 使用防火墙限制访问

4. **日志脱敏**：
   - ✅ 只记录交易哈希前后 4 位
   - ✅ 不记录完整地址

5. **访问控制**：
   - ✅ 管理员接口需要严格权限验证
   - ✅ 限制 API 调用频率

---

## 📞 技术支持

遇到问题？

- 📖 查看文档：`backend/docs/BLOCKCHAIN_MONITOR_DEPLOYMENT.md`
- 🐛 提交 Issue：GitHub Issues
- 💬 技术支持：support@yourcompany.com

---

## ✅ 部署后验证

```bash
# 1. 检查服务运行
curl http://localhost:3000/api/v1/payment/usdt/providers

# 2. 创建测试订单（需要认证）
curl -X POST http://localhost:3000/api/v1/payment/usdt/orders \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount": 10, "payment_type": "usdt_trc20"}'

# 3. 查看日志
tail -f /var/log/sub2api/blockchain-monitor.log
```

---

**祝你部署顺利！🎉**

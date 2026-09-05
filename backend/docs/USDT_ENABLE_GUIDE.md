# USDT 支付启用完整指南

## 📍 在哪里开启 USDT 支付

### 方法 1: 管理后台 Web UI（推荐）

1. **登录管理员账户**
   - 访问: `http://your-domain/admin`
   - 使用管理员账号登录

2. **进入设置页面**
   - 点击左侧菜单 **设置** (Settings)

3. **找到支付设置部分**
   - 滚动到 **支付系统设置** (Payment System Settings)

4. **启用 USDT 支付类型**
   - 找到 **启用的支付方式** (Enabled Payment Types)
   - 点击 **USDT (TRC20)** 或 **USDT (ERC20)** 按钮
   - 按钮变为蓝色表示已启用

5. **保存设置**
   - 滚动到页面底部
   - 点击 **保存所有设置** (Save All Settings)

---

### 方法 2: API 调用

```bash
curl -X PUT http://localhost:3000/api/v1/admin/settings/payment \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "payment_enabled": true,
    "payment_enabled_types": ["alipay", "wxpay", "usdt_trc20", "usdt_erc20"]
  }'
```

---

### 方法 3: 数据库直接修改

```sql
-- 查看当前启用的支付方式
SELECT * FROM settings WHERE key = 'ENABLED_PAYMENT_TYPES';

-- 启用 USDT TRC20 和 ERC20
UPDATE settings 
SET value = 'alipay,wxpay,usdt_trc20,usdt_erc20' 
WHERE key = 'ENABLED_PAYMENT_TYPES';
```

---

## ⚙️ 完整配置步骤

### 第 1 步: 启用支付类型（上面已说明）

在管理后台启用 **USDT (TRC20)** 和/或 **USDT (ERC20)**

---

### 第 2 步: 配置环境变量

创建或编辑 `.env` 文件：

```bash
cd backend
cp .env.usdt.example .env
```

**必需配置**（TRC20 示例）：

```env
# TRC20 配置
USDT_TRC20_ENABLED=true
USDT_TRC20_DEPOSIT_ADDRESS=TYourWalletAddress123456789  # 你的 TRC20 收款地址
USDT_TRC20_API_URL=https://api.trongrid.io
USDT_TRC20_API_KEY=your-trongrid-api-key-here          # 从 https://www.trongrid.io/ 获取
USDT_TRC20_HMAC_SECRET=your-64-character-hex-secret     # 运行: openssl rand -hex 32
USDT_TRC20_CONFIRMATIONS=19
USDT_TRC20_POLL_INTERVAL=15

# ERC20 配置（可选）
USDT_ERC20_ENABLED=true
USDT_ERC20_DEPOSIT_ADDRESS=0xYourEthereumAddress
USDT_ERC20_API_URL=https://api.etherscan.io/api
USDT_ERC20_API_KEY=your-etherscan-api-key
USDT_ERC20_HMAC_SECRET=your-64-character-hex-secret
USDT_ERC20_CONFIRMATIONS=12
USDT_ERC20_POLL_INTERVAL=30
```

---

### 第 3 步: 获取必需的 API 密钥

#### TRC20 - TronGrid API
1. 访问: https://www.trongrid.io/
2. 注册账户
3. 创建 API 密钥
4. 免费套餐: 1000 请求/天

#### ERC20 - Etherscan API
1. 访问: https://etherscan.io/apis
2. 注册账户
3. 创建 API 密钥
4. 免费套餐: 5 请求/秒

---

### 第 4 步: 生成 HMAC 密钥

```bash
# Linux/macOS
openssl rand -hex 32

# Windows PowerShell
[System.BitConverter]::ToString([System.Security.Cryptography.RandomNumberGenerator]::GetBytes(32)).Replace("-", "").ToLower()
```

输出示例：
```
a1b2c3d4e5f6789012345678901234567890123456789012345678901234
```

将此值填入 `USDT_TRC20_HMAC_SECRET` 或 `USDT_ERC20_HMAC_SECRET`

---

### 第 5 步: 准备收款地址

#### TRC20 地址
- 格式: `T` 开头，34 个字符
- 示例: `TYourWalletAddress123456789`
- 获取方式: 
  - TronLink 钱包
  - Binance、OKX 等交易所 USDT TRC20 充值地址

#### ERC20 地址
- 格式: `0x` 开头，42 个字符
- 示例: `0x1234567890123456789012345678901234567890`
- 获取方式:
  - MetaMask 钱包
  - Binance、OKX 等交易所 USDT ERC20 充值地址

⚠️ **安全警告**:
- **不要**在代码或配置文件中存储私钥
- 只需要**收款地址**（公钥）
- 私钥应存储在硬件钱包或 KMS/HSM 中

---

### 第 6 步: 启动服务

```bash
cd backend

# 加载环境变量
source .env  # Linux/macOS
# 或
.\env.bat    # Windows

# 编译并启动
go build -o bin/server.exe ./cmd/server
./bin/server.exe
```

---

### 第 7 步: 验证配置

#### 检查日志
```bash
# 预期看到:
INFO  Blockchain monitor service started
INFO  Started monitoring network  network=TRC20 interval=15s
INFO  Started monitoring network  network=ERC20 interval=30s
```

#### 测试订单创建
```bash
curl -X POST http://localhost:3000/api/v1/payment/usdt/orders \
  -H "Authorization: Bearer USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100,
    "network": "TRC20"
  }'
```

预期响应：
```json
{
  "order_id": "ord_123456",
  "amount_cny": 100.00,
  "crypto_amount": "10.234567",
  "crypto_currency": "USDT",
  "crypto_network": "TRC20",
  "deposit_address": "TYourWalletAddress123456789",
  "qr_code": "data:image/png;base64,...",
  "expires_at": "2024-01-15T10:30:00Z",
  "status": "PENDING"
}
```

---

## 🎯 配置检查清单

使用此清单确保所有步骤完成：

- [ ] **管理后台启用 USDT 支付类型**
  - [ ] USDT (TRC20) 按钮已点击（蓝色）
  - [ ] USDT (ERC20) 按钮已点击（蓝色，可选）
  - [ ] 点击 **保存所有设置**

- [ ] **环境变量已配置**
  - [ ] `USDT_TRC20_ENABLED=true`
  - [ ] `USDT_TRC20_DEPOSIT_ADDRESS` 已填写（T 开头）
  - [ ] `USDT_TRC20_API_KEY` 已填写（从 TronGrid 获取）
  - [ ] `USDT_TRC20_HMAC_SECRET` 已填写（64 字符 hex）
  - [ ] `USDT_TRC20_CONFIRMATIONS=19`
  - [ ] `USDT_TRC20_POLL_INTERVAL=15`

- [ ] **API 密钥已获取**
  - [ ] TronGrid API 密钥（免费 1000 请求/天）
  - [ ] Etherscan API 密钥（如果使用 ERC20）

- [ ] **HMAC 密钥已生成**
  - [ ] 使用 `openssl rand -hex 32` 生成
  - [ ] 长度为 64 个十六进制字符

- [ ] **收款地址已准备**
  - [ ] TRC20 地址（T 开头，34 字符）
  - [ ] 私钥**未**存储在配置文件中

- [ ] **服务已启动并验证**
  - [ ] 日志显示 "Blockchain monitor service started"
  - [ ] 测试订单创建成功
  - [ ] QR 码正常生成

---

## 📊 用户界面位置

### 用户充值页面

**路径**: `/payment`

**文件**: `frontend/src/views/user/PaymentView.vue`

**流程**:
1. 用户输入充值金额（例如: 100 元）
2. 选择支付方式 → 点击 **USDT** 卡片
3. 选择网络（TRC20 或 ERC20）
4. 看到唯一金额（例如: `10.234567 USDT`）
5. 扫描 QR 码或复制地址
6. 转账后等待确认

---

## 💰 充值比例设置

### 在哪里设置

**位置**: 管理后台 → 设置 → 支付设置

**两个关键参数**:

#### 1. 充值倍率 (Balance Recharge Multiplier)
- 字段名: `payment_balance_recharge_multiplier`
- 默认值: `1.0`
- 说明: 1 元人民币 = X 余额

示例:
- `1.0` - 充 100 得 100（无赠送）
- `1.2` - 充 100 得 120（赠送 20%）
- `0.8` - 充 100 得 80（扣除 20%）

#### 2. 充值手续费率 (Recharge Fee Rate)
- 字段名: `payment_recharge_fee_rate`
- 默认值: `0`
- 单位: 小数（0.03 = 3%）

示例:
- `0` - 无手续费
- `0.03` - 收取 3% 手续费

#### 计算公式
```
最终余额 = (充值金额 × (1 - 手续费率)) × 充值倍率

示例: 充 100 元，手续费 3%，倍率 1.2
最终余额 = (100 × 0.97) × 1.2 = 116.4
```

---

## 🔍 常见问题

### Q1: 为什么看不到 USDT 选项？

**检查**:
1. 管理后台是否已启用 USDT 支付类型（按钮为蓝色）
2. 是否点击了 **保存所有设置**
3. 刷新前端页面（Ctrl+F5）

---

### Q2: 订单创建失败

**可能原因**:
1. 环境变量未正确加载
2. HMAC 密钥格式错误（必须是 64 字符 hex）
3. 收款地址格式错误（TRC20 必须 T 开头）

**解决方法**:
```bash
# 检查环境变量
echo $USDT_TRC20_ENABLED
echo $USDT_TRC20_DEPOSIT_ADDRESS

# 查看日志
journalctl -u sub2api | grep USDT
```

---

### Q3: 支付后余额没有自动增加

**检查**:
1. 区块链监控服务是否启动
   ```bash
   journalctl -u sub2api | grep "Blockchain monitor"
   ```

2. API 密钥是否有效
   ```bash
   curl "https://api.trongrid.io/v1/accounts/TYourAddress/transactions" \
     -H "TRON-PRO-API-KEY: your-api-key"
   ```

3. 确认数是否足够
   - TRC20 需要 19 个确认
   - ERC20 需要 12 个确认

---

### Q4: 如何禁用自动监控，只使用手动轮询？

设置环境变量:
```env
BLOCKCHAIN_MONITOR_DISABLED=1
```

用户需要在前端轮询状态:
```javascript
// 每 15 秒查询一次
setInterval(async () => {
  const status = await api.payment.usdt.getOrderStatus(orderId);
  if (status.status === 'PAID') {
    // 充值成功
  }
}, 15000);
```

---

### Q5: 如何更换收款地址？

1. **准备新地址**（确保私钥安全存储）
2. **更新环境变量**:
   ```env
   USDT_TRC20_DEPOSIT_ADDRESS=TNewWalletAddress
   ```
3. **重启服务**:
   ```bash
   systemctl restart sub2api
   ```
4. **验证**:
   ```bash
   journalctl -u sub2api | grep "deposit_address"
   ```

⚠️ **注意**: 旧地址的待处理订单仍会被监控，直到过期或完成

---

## 🚀 快速启动命令

```bash
# 1. 配置环境变量
cd backend
cp .env.usdt.example .env
vim .env  # 填写必需字段

# 2. 生成 HMAC 密钥
openssl rand -hex 32

# 3. 编译
go build -o bin/server.exe ./cmd/server

# 4. 启动
source .env && ./bin/server.exe

# 5. 验证
curl http://localhost:3000/health
```

---

## 📚 相关文档

- **完整集成指南**: `USDT_PAYMENT_INTEGRATION_GUIDE.md`
- **部署清单**: `USDT_DEPLOYMENT_CHECKLIST.md`
- **技术文档**: `BLOCKCHAIN_MONITOR_DEPLOYMENT.md`
- **UI 和比例指南**: `USDT_UI_AND_RATIO_GUIDE.md`

---

## ✅ 成功标志

当一切配置正确时，你会看到：

1. **管理后台**: USDT 按钮为蓝色
2. **用户充值页面**: USDT 卡片可点击
3. **服务日志**: "Blockchain monitor service started"
4. **订单创建**: 返回唯一的 crypto_amount（例如 10.234567）
5. **转账后**: 余额自动增加

---

**🎉 现在 USDT 支付已完全启用！**

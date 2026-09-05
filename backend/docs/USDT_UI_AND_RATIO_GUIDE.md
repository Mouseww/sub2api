# USDT 充值功能使用指南

## 📍 USDT 充值 UI 位置

### 1. **用户充值页面**
- **路由**: `/payment` 或 `/user/payment`
- **文件**: `frontend/src/views/user/PaymentView.vue`
- **说明**: 这是主要的充值页面，包含：
  - 充值账户信息显示
  - 当前余额显示
  - 充值金额输入
  - **支付方式选择器**（包括 USDT）
  - 支付按钮

### 2. **支付方式选择器**
- **组件**: `frontend/src/components/payment/PaymentMethodSelector.vue`
- **USDT 图标**: `frontend/src/assets/icons/usdt.svg`
- **说明**: USDT 已集成到支付方式网格中，显示为一个可点击的卡片

### 3. **USDT 支付流程**
当用户选择 USDT 支付后，会看到：
- **网络选择**: TRC20（推荐）或 ERC20
- **支付信息**: 收款地址、金额、QR 码
- **支付状态**: 实时显示确认进度（例如：5/19）

---

## ⚙️ 充值比例设置

### 管理员设置路径

**位置**: 管理后台 → 设置 → 支付设置
- **前端文件**: `frontend/src/views/admin/SettingsView.vue`
- **后端 API**: `PUT /api/v1/admin/settings/payment`

### 关键参数

#### 1. **充值倍率** (Balance Recharge Multiplier)
- **字段名**: `balance_recharge_multiplier`
- **默认值**: `1.0`（1元人民币 = 1余额）
- **说明**: 控制用户充值时获得的余额比例

**示例**:
```
充值金额: 100 元
倍率: 1.0  → 获得余额: 100
倍率: 1.2  → 获得余额: 120（赠送20%）
倍率: 0.8  → 获得余额: 80（扣除20%手续费效果）
```

#### 2. **充值手续费率** (Recharge Fee Rate)
- **字段名**: `recharge_fee_rate`
- **默认值**: `0`（无手续费）
- **单位**: 百分比（例如：0.03 = 3%）
- **说明**: 从充值金额中扣除的手续费

**示例**:
```
充值金额: 100 元
手续费率: 0    → 实际到账: 100（无手续费）
手续费率: 0.03 → 实际到账: 97（扣除3%手续费）
```

### 计算公式

```javascript
// 最终用户获得的余额
最终余额 = (充值金额 × (1 - 手续费率)) × 充值倍率

// 示例 1: 充值 100 元，手续费 3%，倍率 1.0
最终余额 = (100 × 0.97) × 1.0 = 97

// 示例 2: 充值 100 元，无手续费，倍率 1.2（赠送20%）
最终余额 = (100 × 1.0) × 1.2 = 120
```

---

## 🔧 如何设置充值比例

### 方法 1: Web 管理后台（推荐）

1. 登录管理员账户
2. 进入 **设置** 页面
3. 找到 **支付设置** 部分
4. 找到以下字段：
   - **充值倍率** (`balance_recharge_multiplier`)
   - **充值手续费率** (`recharge_fee_rate`)
5. 修改数值并保存

### 方法 2: 直接调用 API

```bash
curl -X PUT http://localhost:3000/api/v1/admin/settings/payment \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "balance_recharge_multiplier": 1.2,
    "recharge_fee_rate": 0.03
  }'
```

### 方法 3: 数据库直接修改

```sql
-- 设置充值倍率为 1.2（赠送20%）
UPDATE settings 
SET value = '1.2' 
WHERE key = 'BALANCE_RECHARGE_MULTIPLIER';

-- 设置手续费率为 3%
UPDATE settings 
SET value = '0.03' 
WHERE key = 'RECHARGE_FEE_RATE';
```

---

## 💡 实际应用场景

### 场景 1: 促销活动（赠送余额）
```json
{
  "balance_recharge_multiplier": 1.2,
  "recharge_fee_rate": 0
}
```
- 用户充值 100 元 → 获得 120 余额
- **适用**: 拉新活动、节日促销

### 场景 2: 收取手续费
```json
{
  "balance_recharge_multiplier": 1.0,
  "recharge_fee_rate": 0.03
}
```
- 用户充值 100 元 → 获得 97 余额
- **适用**: 覆盖支付渠道手续费

### 场景 3: 组合使用
```json
{
  "balance_recharge_multiplier": 1.1,
  "recharge_fee_rate": 0.02
}
```
- 用户充值 100 元 → 获得 107.8 余额
  - 扣手续费: 100 × 0.98 = 98
  - 加倍率: 98 × 1.1 = 107.8
- **适用**: 平衡运营成本和用户激励

---

## 📊 USDT 充值的特殊性

### USDT 充值金额生成

USDT 充值使用 **HMAC 加密唯一金额**，而不是固定金额：

```
用户请求充值 10 USDT
↓
系统生成唯一金额: 10.234567 USDT
↓
用户转账 10.234567 USDT 到平台地址
↓
监控服务检测到匹配金额
↓
自动完成订单并充值余额
```

### 充值比例应用时机

```
用户充值 10.234567 USDT
↓
折算人民币: 10.234567 × 7.2 = 73.69 CNY（假设汇率为7.2）
↓
应用手续费率: 73.69 × (1 - 0.03) = 71.48 CNY
↓
应用充值倍率: 71.48 × 1.2 = 85.78
↓
最终用户余额增加: 85.78
```

**注意**: 
- USDT 到 CNY 的汇率由 `subscription_usd_to_cny_rate` 设置控制
- 默认汇率: 7.2（可在管理后台修改）

---

## 🔍 前端显示逻辑

### 充值预览

在充值页面，用户可以看到：

```vue
<template>
  <!-- 充值金额输入 -->
  <input v-model="amount" /> <!-- 例如：100 -->
  
  <!-- 实际到账预览 -->
  <div class="preview">
    <span>充值金额: ¥{{ amount }}</span>
    <span>手续费({{ feeRate * 100 }}%): -¥{{ fee }}</span>
    <span>赠送({{ (multiplier - 1) * 100 }}%): +¥{{ bonus }}</span>
    <span class="total">实际到账: ¥{{ finalBalance }}</span>
  </div>
</template>

<script>
const feeRate = 0.03 // 3% 手续费
const multiplier = 1.2 // 1.2 倍率

const fee = amount * feeRate // 3
const bonus = (amount - fee) * (multiplier - 1) // 19.4
const finalBalance = (amount - fee) * multiplier // 116.4
</script>
```

### 代码位置

- **充值预览计算**: `frontend/src/views/user/PaymentView.vue` (第 66-80 行)
- **手续费显示**: 如果 `recharge_fee_rate > 0`，则显示手续费提示
- **倍率提示**: 如果 `balance_recharge_multiplier != 1.0`，则显示赠送/扣除比例

---

## 📱 USDT 充值完整流程（用户视角）

1. **进入充值页面**: `/payment`
2. **输入金额**: 例如 100 元
3. **选择支付方式**: 点击 **USDT** 卡片
4. **选择网络**: TRC20（推荐，手续费低）或 ERC20
5. **看到支付信息**:
   - 收款地址: `TXxx...xxx`
   - 应付金额: `100.234567 USDT`（唯一金额）
   - QR 码
6. **扫码转账**: 使用钱包扫码或复制地址
7. **等待确认**: 
   - 方式1: 页面自动刷新（监控服务自动检测）
   - 方式2: 用户点击"手动检查"按钮
8. **充值成功**: 
   - 订单状态变为 `PAID`
   - 余额自动增加

---

## 🎯 管理员常见操作

### 1. 设置促销活动（充100送20）

```json
{
  "balance_recharge_multiplier": 1.2
}
```

### 2. 取消促销（恢复默认）

```json
{
  "balance_recharge_multiplier": 1.0
}
```

### 3. 添加手续费（收取3%）

```json
{
  "recharge_fee_rate": 0.03
}
```

### 4. 查看当前设置

```bash
curl http://localhost:3000/api/v1/payment/config \
  -H "Authorization: Bearer USER_TOKEN"
```

响应：
```json
{
  "balance_recharge_multiplier": 1.2,
  "recharge_fee_rate": 0.03,
  "subscription_usd_to_cny_rate": 7.2
}
```

---

## 🛠️ 开发者信息

### 后端实现

- **服务**: `backend/internal/service/payment_config_service.go`
- **计算逻辑**: `backend/internal/service/payment_amounts.go`
- **常量定义**:
  ```go
  const defaultBalanceRechargeMultiplier = 1.0
  SettingBalanceRechargeMult = "BALANCE_RECHARGE_MULTIPLIER"
  SettingRechargeFeeRate = "RECHARGE_FEE_RATE"
  ```

### 前端实现

- **充值页面**: `frontend/src/views/user/PaymentView.vue`
- **支付方法选择器**: `frontend/src/components/payment/PaymentMethodSelector.vue`
- **USDT 组件**:
  - `USDTPaymentFlow.vue` - 完整流程
  - `USDTPaymentPanel.vue` - 支付面板
  - `USDTNetworkSelector.vue` - 网络选择

### 数据库表

```sql
-- 配置存储
CREATE TABLE settings (
  key VARCHAR(255) PRIMARY KEY,
  value TEXT
);

-- 充值订单
CREATE TABLE payment_orders (
  id BIGINT PRIMARY KEY,
  user_id BIGINT,
  amount DECIMAL(10,2),
  payment_type VARCHAR(50),
  crypto_amount_usd DECIMAL(18,6), -- USDT 金额（6位小数）
  status VARCHAR(20)
);
```

---

## 🎉 总结

### USDT 充值 UI
- **主页面**: `/payment`
- **支付方式**: PaymentMethodSelector 中的 USDT 卡片
- **流程组件**: USDTPaymentFlow、USDTPaymentPanel

### 充值比例设置
- **管理后台**: 设置 → 支付设置
- **关键字段**:
  - `balance_recharge_multiplier` - 充值倍率（默认 1.0）
  - `recharge_fee_rate` - 手续费率（默认 0）
- **计算公式**: `最终余额 = (充值金额 × (1 - 手续费率)) × 充值倍率`

### 监控方式
- **自动**: 后台区块链监控服务（15-30秒轮询）
- **手动**: 用户点击"手动检查"按钮

需要我帮你：
1. 修改默认充值比例？
2. 添加管理后台快捷设置按钮？
3. 优化 USDT 充值 UI 显示？

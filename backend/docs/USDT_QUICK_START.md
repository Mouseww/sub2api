# USDT 支付快速启动指南

## 🎯 三步开启 USDT 支付

### 第 1 步: 管理后台启用 ✅

```
登录管理后台 → 设置 → 支付系统设置 → 启用的支付方式
```

点击这两个按钮使其变为**蓝色**：
- **USDT (TRC20)** 
- **USDT (ERC20)** *(可选)*

然后点击底部的 **保存所有设置**

---

### 第 2 步: 配置环境变量 ⚙️

创建 `backend/.env` 文件：

```env
# 必需配置
USDT_TRC20_ENABLED=true
USDT_TRC20_DEPOSIT_ADDRESS=TYourWalletAddress123456789
USDT_TRC20_API_KEY=your-trongrid-api-key
USDT_TRC20_HMAC_SECRET=your-64-char-hex-secret
USDT_TRC20_CONFIRMATIONS=19
USDT_TRC20_POLL_INTERVAL=15
```

**如何获取**：
- **收款地址**: TronLink 钱包或交易所 USDT TRC20 地址
- **API 密钥**: https://www.trongrid.io/ 注册免费获取
- **HMAC 密钥**: 运行 `openssl rand -hex 32` 生成

---

### 第 3 步: 启动服务 🚀

```bash
cd backend
source .env
go build -o bin/server.exe ./cmd/server
./bin/server.exe
```

**验证成功**：日志显示
```
INFO  Blockchain monitor service started
INFO  Started monitoring network  network=TRC20 interval=15s
```

---

## 📍 用户界面位置

### 用户充值页面
- **路径**: `/payment`
- **操作**: 输入金额 → 点击 USDT 卡片 → 选择网络 → 扫码支付

### 管理后台设置
- **路径**: `/admin` → 设置 → 支付系统设置
- **可配置**: 
  - 启用/禁用 USDT 支付类型
  - 充值倍率（例如：1.2 = 充 100 送 20）
  - 充值手续费率（例如：0.03 = 3% 手续费）

---

## 💰 充值比例在哪里设置

**位置**: 管理后台 → 设置 → 支付设置

**两个字段**：

1. **充值倍率** (`payment_balance_recharge_multiplier`)
   - 默认: `1.0` （充 100 得 100）
   - 促销: `1.2` （充 100 得 120）

2. **充值手续费率** (`payment_recharge_fee_rate`)
   - 默认: `0` （无手续费）
   - 示例: `0.03` （3% 手续费）

**计算公式**：
```
最终余额 = (充值金额 × (1 - 手续费率)) × 充值倍率
```

---

## ✅ 检查清单

- [ ] 管理后台 USDT 按钮已启用（蓝色）
- [ ] `.env` 文件已创建并填写
- [ ] TronGrid API 密钥已获取
- [ ] HMAC 密钥已生成（64 字符）
- [ ] 收款地址已准备（T 开头）
- [ ] 服务已启动并看到监控日志
- [ ] 测试订单创建成功

---

## 🔧 故障排查

### 问题 1: 看不到 USDT 选项
- ✅ 确认管理后台已启用并保存
- ✅ 刷新前端页面（Ctrl+F5）

### 问题 2: 订单创建失败
- ✅ 检查 `.env` 文件是否加载：`echo $USDT_TRC20_ENABLED`
- ✅ 验证收款地址格式（TRC20 必须 T 开头）

### 问题 3: 支付后余额没增加
- ✅ 检查监控服务日志：`journalctl -u sub2api | grep "Blockchain monitor"`
- ✅ 验证 API 密钥有效性
- ✅ 等待足够确认数（TRC20: 19 个确认 ≈ 57 秒）

---

## 📚 完整文档

详细信息请查看：
- **完整启用指南**: `USDT_ENABLE_GUIDE.md`
- **集成指南**: `USDT_PAYMENT_INTEGRATION_GUIDE.md`
- **部署清单**: `USDT_DEPLOYMENT_CHECKLIST.md`

---

**🎉 三步即可启用 USDT 自动充值！**

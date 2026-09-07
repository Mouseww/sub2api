# Sub2API USDT 支付配置完成指南

## ✅ 已完成的配置

1. **环境变量配置已添加**
   - 收款地址: `TZAoME7F2Zwa9cBS8aTCQPTeYJdhWSvtwH` (TRC20)
   - HMAC 密钥: 已自动生成
   - 轮询间隔: 15秒
   - 确认数: 19个区块

2. **配置文件位置**: `/root/sub2api/deploy/.env`

## ⚠️ 需要完成的步骤

### 第 1 步: 获取 TronGrid API Key（必需）

1. 访问 https://www.trongrid.io/
2. 注册账户（免费）
3. 创建 API 密钥
4. 复制 API Key

免费额度：1000 请求/天（足够使用）

### 第 2 步: 更新配置文件

在服务器上执行：

```bash
ssh root@186.241.91.102

cd /root/sub2api/deploy

# 编辑 .env 文件，找到这一行：
# USDT_TRC20_API_KEY=请在这里填写你的TronGrid_API_Key

# 替换为你的实际 API Key：
sed -i 's/USDT_TRC20_API_KEY=请在这里填写你的TronGrid_API_Key/USDT_TRC20_API_KEY=你的实际API_KEY/' .env
```

或者手动编辑：

```bash
nano .env

# 找到并修改这一行：
USDT_TRC20_API_KEY=你从TronGrid获取的API_Key
```

### 第 3 步: 重启 Docker 容器

```bash
cd /root/sub2api/deploy
docker-compose down
docker-compose up -d
```

### 第 4 步: 验证服务启动

查看日志确认区块链监控服务已启动：

```bash
docker-compose logs -f sub2api | grep -i "blockchain monitor"
```

预期看到：
```
INFO  Blockchain monitor service started
INFO  Started monitoring network  network=TRC20 interval=15s
```

如果看不到或有错误，检查：
```bash
# 查看完整日志
docker-compose logs sub2api | tail -100

# 确认环境变量已加载
docker exec sub2api env | grep USDT
```

### 第 5 步: 在管理后台启用 USDT 支付类型

1. **登录管理后台**
   - 访问: `http://186.241.91.102:8080/admin`
   - 使用管理员账号登录

2. **进入设置页面**
   - 点击左侧菜单 → **设置** (Settings)

3. **启用 USDT 支付**
   - 滚动到 **支付系统设置** (Payment System Settings)
   - 找到 **启用的支付方式** (Enabled Payment Types)
   - 点击 **USDT (TRC20)** 按钮，使其变为**蓝色**
   - 滚动到底部
   - 点击 **保存所有设置** (Save All Settings)

⚠️ **重要**: 必须在管理后台点击启用并保存，否则前端不会显示 USDT 选项！

### 第 6 步: 测试充值功能

1. **访问用户充值页面**
   - 访问: `http://186.241.91.102:8080/payment`
   - 或登录后点击 **充值** 菜单

2. **测试流程**
   - 输入充值金额（例如: 100 元）
   - 应该能看到 **USDT** 支付方式卡片
   - 点击 USDT 卡片
   - 选择 **TRC20** 网络
   - 会显示：
     - 唯一金额（例如: `10.234567 USDT`）
     - 收款地址和二维码
     - 15分钟倒计时

## 🔍 故障排查

### 问题 1: 前端看不到 USDT 选项

**原因**: 管理后台未启用

**解决方法**:
1. 登录管理后台
2. 设置 → 支付系统设置
3. 点击 USDT (TRC20) 按钮（变蓝）
4. 保存设置
5. 清空浏览器缓存（Ctrl+Shift+Delete）
6. 刷新页面（Ctrl+F5）

### 问题 2: 服务启动失败

**检查日志**:
```bash
docker-compose logs sub2api | grep -i error
docker-compose logs sub2api | tail -50
```

**常见原因**:
- API Key 未填写或格式错误
- 收款地址格式错误（必须 T 开头，34 个字符）
- HMAC 密钥格式错误

**验证配置**:
```bash
# 检查环境变量
docker exec sub2api env | grep USDT_TRC20
```

### 问题 3: 订单创建成功但看不到金额/二维码

**可能原因**: TronGrid API Key 未配置或无效

**测试 API Key**:
```bash
curl "https://api.trongrid.io/v1/accounts/TZAoME7F2Zwa9cBS8aTCQPTeYJdhWSvtwH/transactions" \
  -H "TRON-PRO-API-KEY: 你的API_KEY"
```

预期返回 JSON 数据（不是错误信息）

### 问题 4: 支付后余额没有自动增加

**检查区块链监控服务**:
```bash
docker-compose logs sub2api | grep -i "blockchain monitor"
docker-compose logs sub2api | grep -i "payment received"
```

**确认交易**:
1. TRC20 需要 **19 个确认**（约 57 秒）
2. 在 TronScan 查看交易: https://tronscan.org/#/address/TZAoME7F2Zwa9cBS8aTCQPTeYJdhWSvtwH

## 📊 配置充值比例（可选）

**位置**: 管理后台 → 设置 → 支付设置

**可配置参数**:

1. **充值倍率** (`payment_balance_recharge_multiplier`)
   - 默认: `1.0` （充 100 得 100）
   - 促销: `1.2` （充 100 得 120，赠送 20%）

2. **充值手续费率** (`payment_recharge_fee_rate`)
   - 默认: `0` （无手续费）
   - 示例: `0.03` （3% 手续费）

**计算公式**:
```
最终余额 = (充值金额 × (1 - 手续费率)) × 充值倍率
```

示例: 充 100 元，手续费 3%，倍率 1.2
```
最终余额 = (100 × 0.97) × 1.2 = 116.4
```

## 📁 配置文件说明

### 已配置的参数

```env
# 启用 TRC20
USDT_TRC20_ENABLED=true

# 你的收款地址（TRC20）
USDT_TRC20_DEPOSIT_ADDRESS=TZAoME7F2Zwa9cBS8aTCQPTeYJdhWSvtwH

# USDT 官方 TRC20 合约地址（不要修改）
USDT_TRC20_CONTRACT=TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t

# TronGrid API 端点
USDT_TRC20_API_URL=https://api.trongrid.io

# 需要你填写的 API Key
USDT_TRC20_API_KEY=请在这里填写你的TronGrid_API_Key

# 已自动生成的 HMAC 密钥（用于生成唯一金额）
USDT_TRC20_HMAC_SECRET=d68294169c72531c4490b03489000706b1b841a2587e32c52260f46b7f76ddc1

# 轮询间隔（15秒检查一次新交易）
USDT_TRC20_POLL_INTERVAL=15

# 确认数（19个区块 ≈ 57秒）
USDT_TRC20_CONFIRMATIONS=19
```

## 🚀 快速执行命令

```bash
# 一键完成第2-4步（假设你已经有 API Key）
ssh root@186.241.91.102 << 'EOFCMD'
cd /root/sub2api/deploy

# 替换 API Key（将 YOUR_API_KEY_HERE 替换为你的实际 API Key）
sed -i 's/USDT_TRC20_API_KEY=请在这里填写你的TronGrid_API_Key/USDT_TRC20_API_KEY=YOUR_API_KEY_HERE/' .env

# 重启容器
docker-compose down && docker-compose up -d

# 等待5秒
sleep 5

# 查看日志
docker-compose logs sub2api | grep -i "blockchain monitor"
EOFCMD
```

## 📝 检查清单

完成以下所有步骤后，USDT 支付即可正常使用：

- [ ] 已从 https://www.trongrid.io/ 获取 API Key
- [ ] 已更新 .env 文件中的 `USDT_TRC20_API_KEY`
- [ ] 已重启 Docker 容器
- [ ] 日志显示 "Blockchain monitor service started"
- [ ] 日志显示 "Started monitoring network  network=TRC20"
- [ ] 已登录管理后台
- [ ] 已在 设置 → 支付系统设置 中启用 USDT (TRC20)
- [ ] 已点击 "保存所有设置"
- [ ] 前端充值页面可以看到 USDT 选项
- [ ] 测试创建订单成功（返回唯一金额和二维码）

## 🎉 配置完成

完成上述所有步骤后：

1. ✅ 用户可以在充值页面选择 USDT 支付
2. ✅ 系统会生成唯一金额和二维码
3. ✅ 用户转账后约 1 分钟自动到账
4. ✅ 余额自动增加

## 📚 相关文档

- **完整启用指南**: `backend/docs/USDT_ENABLE_GUIDE.md`
- **快速启动指南**: `backend/docs/USDT_QUICK_START.md`
- **集成指南**: `backend/docs/USDT_PAYMENT_INTEGRATION_GUIDE.md`
- **部署清单**: `backend/docs/USDT_DEPLOYMENT_CHECKLIST.md`

## 💬 需要帮助？

如果遇到问题：

1. 查看日志: `docker-compose logs sub2api | tail -100`
2. 检查配置: `docker exec sub2api env | grep USDT`
3. 查看故障排查部分
4. 参考完整文档

---

**配置生成时间**: 2026-09-06 16:00

**收款地址**: TZAoME7F2Zwa9cBS8aTCQPTeYJdhWSvtwH (TRC20)

**状态**: ⏳ 等待 TronGrid API Key 配置

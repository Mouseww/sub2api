#!/bin/bash

# USDT TRC20 配置脚本
# 用法: ./configure_usdt.sh

set -e

echo "=== Sub2API USDT TRC20 配置向导 ==="
echo ""

# 检查是否在 deploy 目录
if [ ! -f ".env" ]; then
    echo "错误: 未找到 .env 文件，请确保在 deploy 目录下运行此脚本"
    exit 1
fi

# 配置参数
DEPOSIT_ADDRESS="TZAoME7F2Zwa9cBS8aTCQPTeYJdhWSvtwH"
NETWORK="TRC20"

echo "配置信息:"
echo "  收款地址: $DEPOSIT_ADDRESS"
echo "  网络: $NETWORK"
echo ""

# 生成 HMAC 密钥
echo "正在生成 HMAC 密钥..."
HMAC_SECRET=$(openssl rand -hex 32)
echo "  HMAC 密钥: $HMAC_SECRET"
echo ""

# 提示用户输入 TronGrid API Key
echo "请输入 TronGrid API Key (从 https://www.trongrid.io/ 获取):"
read -p "API Key: " API_KEY

if [ -z "$API_KEY" ]; then
    echo "错误: API Key 不能为空"
    exit 1
fi

echo ""
echo "=== 正在更新配置文件 ==="

# 备份原始配置
cp .env .env.backup.$(date +%Y%m%d_%H%M%S)
echo "  已备份原配置到 .env.backup.*"

# 添加 USDT 配置到 .env 文件
cat >> .env << EOF

# -----------------------------------------------------------------------------
# USDT Payment Configuration (TRC20)
# USDT 支付配置 (TRC20)
# -----------------------------------------------------------------------------
# Enable USDT TRC20 payment
USDT_TRC20_ENABLED=true

# TRC20 deposit address (your receiving address)
# TRC20 收款地址
USDT_TRC20_DEPOSIT_ADDRESS=$DEPOSIT_ADDRESS

# TRC20 USDT contract address (official USDT TRC20 contract)
# TRC20 USDT 合约地址（官方合约）
USDT_TRC20_CONTRACT=TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t

# TronGrid API configuration
# TronGrid API 配置
USDT_TRC20_API_URL=https://api.trongrid.io
USDT_TRC20_API_KEY=$API_KEY

# HMAC secret for generating unique amounts (auto-generated)
# HMAC 密钥用于生成唯一金额（自动生成）
USDT_TRC20_HMAC_SECRET=$HMAC_SECRET

# Polling interval in seconds
# 轮询间隔（秒）
USDT_TRC20_POLL_INTERVAL=15

# Required confirmations
# 所需确认数
USDT_TRC20_CONFIRMATIONS=19
EOF

echo "  已添加 USDT TRC20 配置"
echo ""

echo "=== 配置完成 ==="
echo ""
echo "下一步："
echo "1. 重启 Docker 容器以应用配置:"
echo "   docker-compose down && docker-compose up -d"
echo ""
echo "2. 查看日志确认服务启动:"
echo "   docker-compose logs -f sub2api | grep -i 'blockchain monitor'"
echo ""
echo "3. 登录管理后台启用 USDT 支付类型:"
echo "   - 访问: http://YOUR_IP:8080/admin"
echo "   - 进入: 设置 → 支付系统设置"
echo "   - 点击: USDT (TRC20) 按钮使其变蓝"
echo "   - 保存: 点击底部 '保存所有设置'"
echo ""
echo "4. 测试充值功能:"
echo "   - 访问: http://YOUR_IP:8080/payment"
echo "   - 输入充值金额"
echo "   - 选择 USDT 支付方式"
echo ""

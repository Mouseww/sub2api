#!/bin/bash

# USDT 支付快速启动脚本
# 用法: ./start-usdt.sh

echo "🚀 启动 Sub2API USDT 支付服务..."
echo ""

# 检查环境变量
check_env() {
    local var_name=$1
    local var_value=$(eval echo \$$var_name)
    
    if [ -z "$var_value" ]; then
        echo "❌ 环境变量 $var_name 未设置"
        return 1
    else
        echo "✅ $var_name: ${var_value:0:10}..."
        return 0
    fi
}

echo "📋 检查环境变量配置..."
echo ""

all_ok=true

# 检查必需的 TRC20 配置
if check_env "USDT_TRC20_DEPOSIT_ADDRESS"; then :; else all_ok=false; fi
if check_env "USDT_TRC20_CONTRACT"; then :; else all_ok=false; fi
if check_env "USDT_TRC20_API_URL"; then :; else all_ok=false; fi
if check_env "USDT_TRC20_API_KEY"; then :; else all_ok=false; fi
if check_env "USDT_TRC20_HMAC_SECRET"; then :; else all_ok=false; fi

echo ""

# 检查可选的 ERC20 配置
if [ ! -z "$USDT_ERC20_DEPOSIT_ADDRESS" ]; then
    echo "📌 检测到 ERC20 配置..."
    if check_env "USDT_ERC20_CONTRACT"; then :; fi
    if check_env "USDT_ERC20_API_URL"; then :; fi
    if check_env "USDT_ERC20_API_KEY"; then :; fi
    if check_env "USDT_ERC20_HMAC_SECRET"; then :; fi
    echo ""
fi

if [ "$all_ok" = false ]; then
    echo ""
    echo "❌ 环境变量配置不完整！"
    echo ""
    echo "请参考 .env.usdt.example 配置文件："
    echo "  cp .env.usdt.example .env"
    echo "  vim .env"
    echo "  source .env"
    echo ""
    exit 1
fi

echo "✅ 环境变量检查通过"
echo ""

# 检查监控服务状态
if [ "$BLOCKCHAIN_MONITOR_DISABLED" = "1" ]; then
    echo "⚠️  自动监控已禁用（BLOCKCHAIN_MONITOR_DISABLED=1）"
    echo "   用户需要手动轮询订单状态"
else
    echo "✅ 自动监控已启用"
    echo "   TRC20 轮询间隔: ${USDT_TRC20_POLL_INTERVAL:-15} 秒"
    if [ ! -z "$USDT_ERC20_DEPOSIT_ADDRESS" ]; then
        echo "   ERC20 轮询间隔: ${USDT_ERC20_POLL_INTERVAL:-30} 秒"
    fi
fi

echo ""
echo "🔧 配置摘要:"
echo "  - TRC20 地址: ${USDT_TRC20_DEPOSIT_ADDRESS:0:10}...${USDT_TRC20_DEPOSIT_ADDRESS: -4}"
if [ ! -z "$USDT_ERC20_DEPOSIT_ADDRESS" ]; then
    echo "  - ERC20 地址: ${USDT_ERC20_DEPOSIT_ADDRESS:0:10}...${USDT_ERC20_DEPOSIT_ADDRESS: -4}"
fi
echo "  - TRC20 确认数: ${USDT_TRC20_CONFIRMATIONS:-19}"
if [ ! -z "$USDT_ERC20_CONFIRMATIONS" ]; then
    echo "  - ERC20 确认数: ${USDT_ERC20_CONFIRMATIONS:-12}"
fi

echo ""
echo "🎯 启动服务..."
echo ""

# 启动服务
./bin/server.exe

exit 0

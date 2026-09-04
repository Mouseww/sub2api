@echo off
REM USDT 支付快速启动脚本
REM 用法: start-usdt.bat

echo 🚀 启动 Sub2API USDT 支付服务...
echo.

echo 📋 检查环境变量配置...
echo.

set ALL_OK=1

REM 检查必需的 TRC20 配置
if "%USDT_TRC20_DEPOSIT_ADDRESS%"=="" (
    echo ❌ 环境变量 USDT_TRC20_DEPOSIT_ADDRESS 未设置
    set ALL_OK=0
) else (
    echo ✅ USDT_TRC20_DEPOSIT_ADDRESS: %USDT_TRC20_DEPOSIT_ADDRESS:~0,10%...
)

if "%USDT_TRC20_CONTRACT%"=="" (
    echo ❌ 环境变量 USDT_TRC20_CONTRACT 未设置
    set ALL_OK=0
) else (
    echo ✅ USDT_TRC20_CONTRACT: %USDT_TRC20_CONTRACT:~0,10%...
)

if "%USDT_TRC20_API_URL%"=="" (
    echo ❌ 环境变量 USDT_TRC20_API_URL 未设置
    set ALL_OK=0
) else (
    echo ✅ USDT_TRC20_API_URL: %USDT_TRC20_API_URL%
)

if "%USDT_TRC20_API_KEY%"=="" (
    echo ❌ 环境变量 USDT_TRC20_API_KEY 未设置
    set ALL_OK=0
) else (
    echo ✅ USDT_TRC20_API_KEY: %USDT_TRC20_API_KEY:~0,10%...
)

if "%USDT_TRC20_HMAC_SECRET%"=="" (
    echo ❌ 环境变量 USDT_TRC20_HMAC_SECRET 未设置
    set ALL_OK=0
) else (
    echo ✅ USDT_TRC20_HMAC_SECRET: %USDT_TRC20_HMAC_SECRET:~0,10%...
)

echo.

REM 检查可选的 ERC20 配置
if not "%USDT_ERC20_DEPOSIT_ADDRESS%"=="" (
    echo 📌 检测到 ERC20 配置...
    echo ✅ USDT_ERC20_DEPOSIT_ADDRESS: %USDT_ERC20_DEPOSIT_ADDRESS:~0,10%...
    echo.
)

if %ALL_OK%==0 (
    echo.
    echo ❌ 环境变量配置不完整！
    echo.
    echo 请参考 .env.usdt.example 配置文件：
    echo   copy .env.usdt.example .env
    echo   notepad .env
    echo   然后在 PowerShell 中运行:
    echo   Get-Content .env ^| ForEach-Object { if ($_ -match '^\s*([^#][^=]+)=(.*)$') { [Environment]::SetEnvironmentVariable($matches[1].Trim(), $matches[2].Trim(), 'Process') } }
    echo.
    pause
    exit /b 1
)

echo ✅ 环境变量检查通过
echo.

REM 检查监控服务状态
if "%BLOCKCHAIN_MONITOR_DISABLED%"=="1" (
    echo ⚠️  自动监控已禁用 ^(BLOCKCHAIN_MONITOR_DISABLED=1^)
    echo    用户需要手动轮询订单状态
) else (
    echo ✅ 自动监控已启用
    if "%USDT_TRC20_POLL_INTERVAL%"=="" (
        echo    TRC20 轮询间隔: 15 秒
    ) else (
        echo    TRC20 轮询间隔: %USDT_TRC20_POLL_INTERVAL% 秒
    )
    if not "%USDT_ERC20_DEPOSIT_ADDRESS%"=="" (
        if "%USDT_ERC20_POLL_INTERVAL%"=="" (
            echo    ERC20 轮询间隔: 30 秒
        ) else (
            echo    ERC20 轮询间隔: %USDT_ERC20_POLL_INTERVAL% 秒
        )
    )
)

echo.
echo 🔧 配置摘要:
echo   - TRC20 地址: %USDT_TRC20_DEPOSIT_ADDRESS:~0,10%...%USDT_TRC20_DEPOSIT_ADDRESS:~-4%
if not "%USDT_ERC20_DEPOSIT_ADDRESS%"=="" (
    echo   - ERC20 地址: %USDT_ERC20_DEPOSIT_ADDRESS:~0,10%...%USDT_ERC20_DEPOSIT_ADDRESS:~-4%
)
if "%USDT_TRC20_CONFIRMATIONS%"=="" (
    echo   - TRC20 确认数: 19
) else (
    echo   - TRC20 确认数: %USDT_TRC20_CONFIRMATIONS%
)
if not "%USDT_ERC20_CONFIRMATIONS%"=="" (
    echo   - ERC20 确认数: %USDT_ERC20_CONFIRMATIONS%
)

echo.
echo 🎯 启动服务...
echo.

REM 启动服务
bin\server.exe

exit /b 0

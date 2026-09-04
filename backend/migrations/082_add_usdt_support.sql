-- 添加加密货币充值地址表
CREATE TABLE IF NOT EXISTS crypto_deposit_addresses (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    currency VARCHAR(20) NOT NULL,
    network VARCHAR(20) NOT NULL,
    address VARCHAR(128) NOT NULL,
    provider_instance_id VARCHAR(64),
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_crypto_deposit_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (user_id, currency, network)
);

CREATE INDEX IF NOT EXISTS idx_crypto_deposit_user_id ON crypto_deposit_addresses(user_id);
CREATE INDEX IF NOT EXISTS idx_crypto_deposit_address ON crypto_deposit_addresses(address);
CREATE INDEX IF NOT EXISTS idx_crypto_deposit_status ON crypto_deposit_addresses(status);

-- 扩展 payment_orders 表，新增加密货币相关字段
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS crypto_currency VARCHAR(20);
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS crypto_network VARCHAR(20);
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS crypto_address VARCHAR(128);
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS crypto_tx_hash VARCHAR(128);
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS crypto_confirmations INT DEFAULT 0;
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS crypto_required_confirmations INT;
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS crypto_amount_usd DECIMAL(20,2);

-- 为加密货币字段添加索引
CREATE INDEX IF NOT EXISTS idx_payment_orders_crypto_tx_hash ON payment_orders(crypto_tx_hash);
CREATE INDEX IF NOT EXISTS idx_payment_orders_crypto_address ON payment_orders(crypto_address);

-- 添加注释
COMMENT ON TABLE crypto_deposit_addresses IS '加密货币充值地址表';
COMMENT ON COLUMN crypto_deposit_addresses.user_id IS '用户 ID';
COMMENT ON COLUMN crypto_deposit_addresses.currency IS '币种，如 USDT, BTC, ETH';
COMMENT ON COLUMN crypto_deposit_addresses.network IS '区块链网络，如 TRC20, ERC20, BEP20';
COMMENT ON COLUMN crypto_deposit_addresses.address IS '充值地址';
COMMENT ON COLUMN crypto_deposit_addresses.provider_instance_id IS '支付服务商实例 ID';
COMMENT ON COLUMN crypto_deposit_addresses.status IS '地址状态：active（活跃）、inactive（停用）、expired（过期）';

COMMENT ON COLUMN payment_orders.crypto_currency IS '加密货币类型，如 USDT';
COMMENT ON COLUMN payment_orders.crypto_network IS '区块链网络，如 TRC20, ERC20, BEP20';
COMMENT ON COLUMN payment_orders.crypto_address IS '充值地址';
COMMENT ON COLUMN payment_orders.crypto_tx_hash IS '区块链交易哈希';
COMMENT ON COLUMN payment_orders.crypto_confirmations IS '当前确认数';
COMMENT ON COLUMN payment_orders.crypto_required_confirmations IS '所需确认数';
COMMENT ON COLUMN payment_orders.crypto_amount_usd IS '美元等值金额（用于汇率锁定）';

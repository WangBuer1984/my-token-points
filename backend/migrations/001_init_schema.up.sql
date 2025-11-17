-- ==========================================
-- 初始化数据库 Schema
-- ==========================================

-- 1. 用户余额表 (当前状态)
CREATE TABLE IF NOT EXISTS user_balances (
    id BIGSERIAL PRIMARY KEY,
    chain_name VARCHAR(50) NOT NULL,
    user_address VARCHAR(42) NOT NULL,
    balance NUMERIC(78, 0) NOT NULL DEFAULT 0,
    last_update_block BIGINT,
    last_update_time TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_user_balances_chain_address UNIQUE (chain_name, user_address)
);

-- 索引
CREATE INDEX idx_user_balances_chain ON user_balances(chain_name);
CREATE INDEX idx_user_balances_address ON user_balances(user_address);
CREATE INDEX idx_user_balances_updated_at ON user_balances(updated_at);

COMMENT ON TABLE user_balances IS '用户余额表 - 存储用户在各链上的当前余额';
COMMENT ON COLUMN user_balances.chain_name IS '链名称 (sepolia, base_sepolia)';
COMMENT ON COLUMN user_balances.user_address IS '用户地址 (以太坊地址格式)';
COMMENT ON COLUMN user_balances.balance IS '当前余额 (wei单位, NUMERIC(78,0)存储uint256)';
COMMENT ON COLUMN user_balances.last_update_block IS '最后更新的区块号';
COMMENT ON COLUMN user_balances.last_update_time IS '最后更新时间';

-- ==========================================

-- 2. 余额变动历史表 (事件记录)
CREATE TABLE IF NOT EXISTS balance_changes (
    id BIGSERIAL PRIMARY KEY,
    chain_name VARCHAR(50) NOT NULL,
    user_address VARCHAR(42) NOT NULL,
    event_type VARCHAR(20) NOT NULL,
    amount_delta NUMERIC(78, 0) NOT NULL,
    balance_before NUMERIC(78, 0) NOT NULL,
    balance_after NUMERIC(78, 0) NOT NULL,
    tx_hash VARCHAR(66) NOT NULL,
    block_number BIGINT NOT NULL,
    block_time TIMESTAMP NOT NULL,
    confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_balance_changes_tx_user UNIQUE (chain_name, tx_hash, user_address),
    CONSTRAINT ck_event_type CHECK (event_type IN ('transfer_in', 'transfer_out', 'mint', 'burn'))
);

-- 索引
CREATE INDEX idx_balance_changes_user ON balance_changes(chain_name, user_address, block_time);
CREATE INDEX idx_balance_changes_block ON balance_changes(chain_name, block_number);
CREATE INDEX idx_balance_changes_confirmed ON balance_changes(confirmed);
CREATE INDEX idx_balance_changes_timestamp ON balance_changes(chain_name, block_time);
CREATE INDEX idx_balance_changes_tx ON balance_changes(tx_hash);

COMMENT ON TABLE balance_changes IS '余额变动历史表 - 记录所有余额变化事件';
COMMENT ON COLUMN balance_changes.event_type IS '事件类型: transfer_in(转入), transfer_out(转出), mint(铸造), burn(销毁)';
COMMENT ON COLUMN balance_changes.amount_delta IS '余额变动量 (正数=增加, 负数=减少)';
COMMENT ON COLUMN balance_changes.confirmed IS '是否已确认 (6区块延迟确认机制)';

-- ==========================================

-- 3. 用户积分表 (当前状态)
CREATE TABLE IF NOT EXISTS user_points (
    id BIGSERIAL PRIMARY KEY,
    chain_name VARCHAR(50) NOT NULL,
    user_address VARCHAR(42) NOT NULL,
    total_points NUMERIC(20, 10) NOT NULL DEFAULT 0,
    last_calc_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_user_points_chain_address UNIQUE (chain_name, user_address)
);

-- 索引
CREATE INDEX idx_user_points_chain ON user_points(chain_name);
CREATE INDEX idx_user_points_last_calc ON user_points(last_calc_at);
CREATE INDEX idx_user_points_total ON user_points(total_points DESC);

COMMENT ON TABLE user_points IS '用户积分表 - 存储用户累计积分';
COMMENT ON COLUMN user_points.total_points IS '累计积分总数';
COMMENT ON COLUMN user_points.last_calc_at IS '最后一次积分计算时间 (用于增量计算)';

-- ==========================================

-- 4. 积分计算历史表 (审计记录)
CREATE TABLE IF NOT EXISTS points_history (
    id BIGSERIAL PRIMARY KEY,
    chain_name VARCHAR(50) NOT NULL,
    user_address VARCHAR(42) NOT NULL,
    calc_period_start TIMESTAMP NOT NULL,
    calc_period_end TIMESTAMP NOT NULL,
    balance_snapshot JSONB NOT NULL,
    points_earned NUMERIC(20, 10) NOT NULL,
    calculation_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_calculation_type CHECK (calculation_type IN ('normal', 'backfill'))
);

-- 索引
CREATE INDEX idx_points_history_user ON points_history(chain_name, user_address, calc_period_end);
CREATE INDEX idx_points_history_period ON points_history(calc_period_start, calc_period_end);
CREATE INDEX idx_points_history_calc_type ON points_history(calculation_type);

COMMENT ON TABLE points_history IS '积分计算历史表 - 记录每次积分计算的详细信息';
COMMENT ON COLUMN points_history.balance_snapshot IS '分段余额快照 (JSONB格式): [{balance, start_time, end_time}]';
COMMENT ON COLUMN points_history.calculation_type IS '计算类型: normal(正常), backfill(回溯)';

-- ==========================================

-- 5. 同步状态表 (系统状态)
CREATE TABLE IF NOT EXISTS sync_state (
    id SERIAL PRIMARY KEY,
    chain_name VARCHAR(50) NOT NULL UNIQUE,
    last_synced_block BIGINT NOT NULL DEFAULT 0,
    last_confirmed_block BIGINT NOT NULL DEFAULT 0,
    last_sync_at TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(20) NOT NULL DEFAULT 'stopped',
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_status CHECK (status IN ('running', 'stopped', 'error'))
);

-- 索引
CREATE INDEX idx_sync_state_status ON sync_state(status);
CREATE INDEX idx_sync_state_updated ON sync_state(updated_at);

COMMENT ON TABLE sync_state IS '同步状态表 - 记录各链的区块同步状态';
COMMENT ON COLUMN sync_state.last_synced_block IS '最后同步的区块号';
COMMENT ON COLUMN sync_state.last_confirmed_block IS '最后确认的区块号 (last_synced_block - 6)';
COMMENT ON COLUMN sync_state.status IS '同步状态: running(运行中), stopped(已停止), error(错误)';

-- ==========================================
-- 说明
-- ==========================================
-- 本数据库设计遵循以下原则：
-- 1. 不使用触发器 (Trigger) - updated_at 字段由应用层手动更新
-- 2. 不使用外键 (Foreign Key) - 关联关系由应用层维护
-- 3. 使用 CHECK 约束保证数据完整性
-- 4. 使用 UNIQUE 约束防止重复数据
-- 5. 使用索引优化查询性能

-- ==========================================
-- 完成
-- ==========================================

-- 显示创建的表
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
  AND tablename IN ('user_balances', 'balance_changes', 'user_points', 'points_history', 'sync_state')
ORDER BY tablename;


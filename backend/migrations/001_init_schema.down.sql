-- ==========================================
-- 回滚数据库 Schema
-- ==========================================
-- 本迁移不使用触发器和外键，直接删除表即可

-- 删除表 (按依赖关系逆序删除)
DROP TABLE IF EXISTS points_history;
DROP TABLE IF EXISTS user_points;
DROP TABLE IF EXISTS balance_changes;
DROP TABLE IF EXISTS user_balances;
DROP TABLE IF EXISTS sync_state;

-- 完成
SELECT 'Schema rolled back successfully' AS status;


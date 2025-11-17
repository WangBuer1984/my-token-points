package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"my-token-points/internal/model"
)

// BalanceRepository 余额数据访问接口
type BalanceRepository interface {
	// 查询用户余额
	GetUserBalance(ctx context.Context, chainName, userAddress string) (*model.UserBalance, error)
	
	// 批量查询用户余额
	GetUserBalances(ctx context.Context, chainName string, offset, limit int) ([]*model.UserBalance, error)
	
	// 更新或创建用户余额
	UpsertUserBalance(ctx context.Context, balance *model.UserBalance) error
	
	// 记录余额变动
	RecordBalanceChange(ctx context.Context, change *model.BalanceChange) error
	
	// 查询余额变动历史
	GetBalanceChanges(ctx context.Context, chainName, userAddress string, startTime, endTime time.Time) ([]*model.BalanceChange, error)
	
	// 查询待确认的余额变动
	GetUnconfirmedChanges(ctx context.Context, chainName string, beforeBlock int64) ([]*model.BalanceChange, error)
	
	// 确认余额变动
	ConfirmBalanceChange(ctx context.Context, chainName, txHash string) error
	
	// 查询某个区块之后的所有余额变动（用于余额重建）
	GetChangesFromBlock(ctx context.Context, chainName string, fromBlock int64) ([]*model.BalanceChange, error)
}

// balanceRepo 余额数据访问实现
type balanceRepo struct {
	db *sqlx.DB
}

// NewBalanceRepository 创建余额仓储实例
func NewBalanceRepository(db *sqlx.DB) BalanceRepository {
	return &balanceRepo{db: db}
}

// GetUserBalance 查询用户余额
func (r *balanceRepo) GetUserBalance(ctx context.Context, chainName, userAddress string) (*model.UserBalance, error) {
	query := `
		SELECT id, chain_name, user_address, balance, last_update_block, last_update_time, created_at, updated_at
		FROM user_balances
		WHERE chain_name = $1 AND user_address = $2
	`
	
	var balance model.UserBalance
	err := r.db.GetContext(ctx, &balance, query, chainName, userAddress)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	return &balance, nil
}

// GetUserBalances 批量查询用户余额
func (r *balanceRepo) GetUserBalances(ctx context.Context, chainName string, offset, limit int) ([]*model.UserBalance, error) {
	query := `
		SELECT id, chain_name, user_address, balance, last_update_block, last_update_time, created_at, updated_at
		FROM user_balances
		WHERE chain_name = $1
		ORDER BY balance DESC, user_address ASC
		LIMIT $2 OFFSET $3
	`
	
	var balances []*model.UserBalance
	err := r.db.SelectContext(ctx, &balances, query, chainName, limit, offset)
	if err != nil {
		return nil, err
	}
	
	return balances, nil
}

// UpsertUserBalance 更新或创建用户余额
func (r *balanceRepo) UpsertUserBalance(ctx context.Context, balance *model.UserBalance) error {
	query := `
		INSERT INTO user_balances (chain_name, user_address, balance, last_update_block, last_update_time)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (chain_name, user_address)
		DO UPDATE SET
			balance = EXCLUDED.balance,
			last_update_block = EXCLUDED.last_update_block,
			last_update_time = EXCLUDED.last_update_time,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`
	
	return r.db.QueryRowContext(
		ctx, query,
		balance.ChainName, balance.UserAddress, balance.Balance,
		balance.LastUpdateBlock, balance.LastUpdateTime,
	).Scan(&balance.ID, &balance.CreatedAt, &balance.UpdatedAt)
}

// RecordBalanceChange 记录余额变动
func (r *balanceRepo) RecordBalanceChange(ctx context.Context, change *model.BalanceChange) error {
	query := `
		INSERT INTO balance_changes (
			chain_name, user_address, tx_hash, block_number, block_time,
			event_type, amount_delta, balance_before, balance_after, confirmed
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`
	
	return r.db.QueryRowContext(
		ctx, query,
		change.ChainName, change.UserAddress, change.TxHash,
		change.BlockNumber, change.BlockTime, change.EventType,
		change.AmountDelta, change.BalanceBefore, change.BalanceAfter, change.Confirmed,
	).Scan(&change.ID, &change.CreatedAt)
}

// GetBalanceChanges 查询余额变动历史
func (r *balanceRepo) GetBalanceChanges(ctx context.Context, chainName, userAddress string, startTime, endTime time.Time) ([]*model.BalanceChange, error) {
	query := `
		SELECT id, chain_name, user_address, tx_hash, block_number, block_time,
			   event_type, amount_delta, balance_before, balance_after, confirmed, created_at
		FROM balance_changes
		WHERE chain_name = $1 AND user_address = $2 
		  AND block_time >= $3 AND block_time < $4
		  AND confirmed = true
		ORDER BY block_number ASC, id ASC
	`
	
	var changes []*model.BalanceChange
	err := r.db.SelectContext(ctx, &changes, query, chainName, userAddress, startTime, endTime)
	if err != nil {
		return nil, err
	}
	
	return changes, nil
}

// GetUnconfirmedChanges 查询待确认的余额变动
func (r *balanceRepo) GetUnconfirmedChanges(ctx context.Context, chainName string, beforeBlock int64) ([]*model.BalanceChange, error) {
	query := `
		SELECT id, chain_name, user_address, tx_hash, block_number, block_time,
			   event_type, amount_delta, balance_before, balance_after, confirmed, created_at
		FROM balance_changes
		WHERE chain_name = $1 AND confirmed = false AND block_number <= $2
		ORDER BY block_number ASC, id ASC
	`
	
	var changes []*model.BalanceChange
	err := r.db.SelectContext(ctx, &changes, query, chainName, beforeBlock)
	if err != nil {
		return nil, err
	}
	
	return changes, nil
}

// ConfirmBalanceChange 确认余额变动
func (r *balanceRepo) ConfirmBalanceChange(ctx context.Context, chainName, txHash string) error {
	query := `
		UPDATE balance_changes
		SET confirmed = true
		WHERE chain_name = $1 AND tx_hash = $2
	`
	
	_, err := r.db.ExecContext(ctx, query, chainName, txHash)
	return err
}

// GetChangesFromBlock 查询某个区块之后的所有余额变动
func (r *balanceRepo) GetChangesFromBlock(ctx context.Context, chainName string, fromBlock int64) ([]*model.BalanceChange, error) {
	query := `
		SELECT id, chain_name, user_address, tx_hash, block_number, block_time,
			   event_type, amount_delta, balance_before, balance_after, confirmed, created_at
		FROM balance_changes
		WHERE chain_name = $1 AND block_number >= $2
		ORDER BY block_number ASC, id ASC
	`
	
	var changes []*model.BalanceChange
	err := r.db.SelectContext(ctx, &changes, query, chainName, fromBlock)
	if err != nil {
		return nil, err
	}
	
	return changes, nil
}


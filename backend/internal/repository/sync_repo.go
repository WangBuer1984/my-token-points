package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"my-token-points/internal/model"
)

// SyncRepository 同步状态数据访问接口
type SyncRepository interface {
	// 获取链的同步状态
	GetSyncState(ctx context.Context, chainName string) (*model.SyncState, error)
	
	// 更新同步状态
	UpdateSyncState(ctx context.Context, state *model.SyncState) error
	
	// 初始化同步状态
	InitSyncState(ctx context.Context, chainName string, startBlock int64) error
}

// syncRepo 同步状态数据访问实现
type syncRepo struct {
	db *sqlx.DB
}

// NewSyncRepository 创建同步状态仓储实例
func NewSyncRepository(db *sqlx.DB) SyncRepository {
	return &syncRepo{db: db}
}

// GetSyncState 获取链的同步状态
func (r *syncRepo) GetSyncState(ctx context.Context, chainName string) (*model.SyncState, error) {
	query := `
		SELECT id, chain_name, last_synced_block, last_confirmed_block, last_sync_at, status, error_message, created_at, updated_at
		FROM sync_state
		WHERE chain_name = $1
	`
	
	var state model.SyncState
	err := r.db.GetContext(ctx, &state, query, chainName)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	return &state, nil
}

// UpdateSyncState 更新同步状态
func (r *syncRepo) UpdateSyncState(ctx context.Context, state *model.SyncState) error {
	query := `
		UPDATE sync_state
		SET last_synced_block = $1,
			last_confirmed_block = $2,
			last_sync_at = $3,
			status = $4,
			error_message = $5,
			updated_at = NOW()
		WHERE chain_name = $6
		RETURNING updated_at
	`
	
	return r.db.QueryRowContext(
		ctx, query,
		state.LastSyncedBlock, state.LastConfirmedBlock, state.LastSyncAt, 
		state.Status, state.ErrorMessage, state.ChainName,
	).Scan(&state.UpdatedAt)
}

// InitSyncState 初始化同步状态
func (r *syncRepo) InitSyncState(ctx context.Context, chainName string, startBlock int64) error {
	query := `
		INSERT INTO sync_state (chain_name, last_synced_block, last_confirmed_block, last_sync_at, status)
		VALUES ($1, $2, $3, NOW(), $4)
		ON CONFLICT (chain_name) DO NOTHING
	`
	
	_, err := r.db.ExecContext(ctx, query, chainName, startBlock-1, startBlock-1, model.StatusRunning)
	return err
}


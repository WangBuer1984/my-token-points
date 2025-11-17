package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"my-token-points/internal/model"
)

// PointsRepository 积分数据访问接口
type PointsRepository interface {
	// 查询用户积分
	GetUserPoints(ctx context.Context, chainName, userAddress string) (*model.UserPoints, error)
	
	// 批量查询用户积分
	GetUserPointsList(ctx context.Context, chainName string, offset, limit int) ([]*model.UserPoints, error)
	
	// 更新或创建用户积分
	UpsertUserPoints(ctx context.Context, points *model.UserPoints) error
	
	// 记录积分计算历史
	RecordPointsHistory(ctx context.Context, history *model.PointsHistory) error
	
	// 查询积分历史
	GetPointsHistory(ctx context.Context, chainName, userAddress string, startTime, endTime time.Time) ([]*model.PointsHistory, error)
	
	// 获取最后一次计算的时间
	GetLastCalculationTime(ctx context.Context, chainName string) (*time.Time, error)
	
	// 查询需要计算积分的时间区间（未计算的小时）
	GetUncalculatedPeriods(ctx context.Context, chainName string, fromTime, toTime time.Time) ([]time.Time, error)
}

// pointsRepo 积分数据访问实现
type pointsRepo struct {
	db *sqlx.DB
}

// NewPointsRepository 创建积分仓储实例
func NewPointsRepository(db *sqlx.DB) PointsRepository {
	return &pointsRepo{db: db}
}

// GetUserPoints 查询用户积分
func (r *pointsRepo) GetUserPoints(ctx context.Context, chainName, userAddress string) (*model.UserPoints, error) {
	query := `
		SELECT id, chain_name, user_address, total_points, last_calc_at, created_at, updated_at
		FROM user_points
		WHERE chain_name = $1 AND user_address = $2
	`
	
	var points model.UserPoints
	err := r.db.GetContext(ctx, &points, query, chainName, userAddress)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	return &points, nil
}

// GetUserPointsList 批量查询用户积分
func (r *pointsRepo) GetUserPointsList(ctx context.Context, chainName string, offset, limit int) ([]*model.UserPoints, error) {
	query := `
		SELECT id, chain_name, user_address, total_points, last_calc_at, created_at, updated_at
		FROM user_points
		WHERE chain_name = $1
		ORDER BY total_points DESC, user_address ASC
		LIMIT $2 OFFSET $3
	`
	
	var pointsList []*model.UserPoints
	err := r.db.SelectContext(ctx, &pointsList, query, chainName, limit, offset)
	if err != nil {
		return nil, err
	}
	
	return pointsList, nil
}

// UpsertUserPoints 更新或创建用户积分
func (r *pointsRepo) UpsertUserPoints(ctx context.Context, points *model.UserPoints) error {
	query := `
		INSERT INTO user_points (chain_name, user_address, total_points, last_calc_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (chain_name, user_address)
		DO UPDATE SET
			total_points = EXCLUDED.total_points,
			last_calc_at = EXCLUDED.last_calc_at,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`
	
	return r.db.QueryRowContext(
		ctx, query,
		points.ChainName, points.UserAddress, points.TotalPoints, points.LastCalcAt,
	).Scan(&points.ID, &points.CreatedAt, &points.UpdatedAt)
}

// RecordPointsHistory 记录积分计算历史
func (r *pointsRepo) RecordPointsHistory(ctx context.Context, history *model.PointsHistory) error {
	query := `
		INSERT INTO points_history (
			chain_name, user_address, calc_period_start, calc_period_end,
			balance_snapshot, points_earned, calculation_type
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	
	return r.db.QueryRowContext(
		ctx, query,
		history.ChainName, history.UserAddress, history.CalcPeriodStart, history.CalcPeriodEnd,
		history.BalanceSnapshot, history.PointsEarned, history.CalculationType,
	).Scan(&history.ID, &history.CreatedAt)
}

// GetPointsHistory 查询积分历史
func (r *pointsRepo) GetPointsHistory(ctx context.Context, chainName, userAddress string, startTime, endTime time.Time) ([]*model.PointsHistory, error) {
	query := `
		SELECT id, chain_name, user_address, calc_period_start, calc_period_end,
			   balance_snapshot, points_earned, calculation_type, created_at
		FROM points_history
		WHERE chain_name = $1 AND user_address = $2
		  AND calc_period_start >= $3 AND calc_period_end <= $4
		ORDER BY calc_period_start ASC
	`
	
	var history []*model.PointsHistory
	err := r.db.SelectContext(ctx, &history, query, chainName, userAddress, startTime, endTime)
	if err != nil {
		return nil, err
	}
	
	return history, nil
}

// GetLastCalculationTime 获取最后一次计算的时间
func (r *pointsRepo) GetLastCalculationTime(ctx context.Context, chainName string) (*time.Time, error) {
	query := `
		SELECT MAX(calc_period_end) as last_time
		FROM points_history
		WHERE chain_name = $1
	`
	
	var lastTime sql.NullTime
	err := r.db.QueryRowContext(ctx, query, chainName).Scan(&lastTime)
	if err != nil {
		return nil, err
	}
	
	if !lastTime.Valid {
		return nil, nil
	}
	
	return &lastTime.Time, nil
}

// GetUncalculatedPeriods 查询需要计算积分的时间区间
func (r *pointsRepo) GetUncalculatedPeriods(ctx context.Context, chainName string, fromTime, toTime time.Time) ([]time.Time, error) {
	// 获取已计算的小时
	query := `
		SELECT DISTINCT calc_period_start
		FROM points_history
		WHERE chain_name = $1
		  AND calc_period_start >= $2
		  AND calc_period_start < $3
		ORDER BY calc_period_start
	`
	
	var calculated []time.Time
	err := r.db.SelectContext(ctx, &calculated, query, chainName, fromTime, toTime)
	if err != nil {
		return nil, err
	}
	
	// 生成所有应该计算的小时
	calculatedMap := make(map[int64]bool)
	for _, t := range calculated {
		calculatedMap[t.Unix()] = true
	}
	
	var uncalculated []time.Time
	current := fromTime.Truncate(time.Hour)
	end := toTime.Truncate(time.Hour)
	
	for current.Before(end) {
		if !calculatedMap[current.Unix()] {
			uncalculated = append(uncalculated, current)
		}
		current = current.Add(time.Hour)
	}
	
	return uncalculated, nil
}


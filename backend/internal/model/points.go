package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// UserPoints 用户积分模型
type UserPoints struct {
	ID          int64     `db:"id" json:"id"`
	ChainName   string    `db:"chain_name" json:"chain_name"`
	UserAddress string    `db:"user_address" json:"user_address"`
	TotalPoints float64   `db:"total_points" json:"total_points"`
	LastCalcAt  *time.Time `db:"last_calc_at" json:"last_calc_at"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// BalanceSnapshot 余额快照
type BalanceSnapshot struct {
	Balance   string    `json:"balance"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// BalanceSnapshots 余额快照数组 (用于JSONB)
type BalanceSnapshots []BalanceSnapshot

// Value 实现 driver.Valuer 接口
func (b BalanceSnapshots) Value() (driver.Value, error) {
	return json.Marshal(b)
}

// Scan 实现 sql.Scanner 接口
func (b *BalanceSnapshots) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, b)
}

// PointsHistory 积分计算历史模型
type PointsHistory struct {
	ID              int64            `db:"id" json:"id"`
	ChainName       string           `db:"chain_name" json:"chain_name"`
	UserAddress     string           `db:"user_address" json:"user_address"`
	CalcPeriodStart time.Time        `db:"calc_period_start" json:"calc_period_start"`
	CalcPeriodEnd   time.Time        `db:"calc_period_end" json:"calc_period_end"`
	BalanceSnapshot BalanceSnapshots `db:"balance_snapshot" json:"balance_snapshot"`
	PointsEarned    float64          `db:"points_earned" json:"points_earned"`
	CalculationType string           `db:"calculation_type" json:"calculation_type"` // normal, backfill
	CreatedAt       time.Time        `db:"created_at" json:"created_at"`
}

// CalculationType 计算类型常量
const (
	CalcTypeNormal   = "normal"
	CalcTypeBackfill = "backfill"
)


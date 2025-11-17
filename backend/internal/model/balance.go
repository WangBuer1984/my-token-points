package model

import (
	"time"
)

// UserBalance 用户余额模型
type UserBalance struct {
	ID              int64     `db:"id" json:"id"`
	ChainName       string    `db:"chain_name" json:"chain_name"`
	UserAddress     string    `db:"user_address" json:"user_address"`
	Balance         string    `db:"balance" json:"balance"` // 使用string存储大数
	LastUpdateBlock int64     `db:"last_update_block" json:"last_update_block"`
	LastUpdateTime  time.Time `db:"last_update_time" json:"last_update_time"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

// BalanceChange 余额变动模型
type BalanceChange struct {
	ID            int64     `db:"id" json:"id"`
	ChainName     string    `db:"chain_name" json:"chain_name"`
	UserAddress   string    `db:"user_address" json:"user_address"`
	TxHash        string    `db:"tx_hash" json:"tx_hash"`
	BlockNumber   int64     `db:"block_number" json:"block_number"`
	BlockTime     time.Time `db:"block_time" json:"block_time"`
	EventType     EventType `db:"event_type" json:"event_type"` // mint, burn, transfer_in, transfer_out
	AmountDelta   string    `db:"amount_delta" json:"amount_delta"`
	BalanceBefore string    `db:"balance_before" json:"balance_before"`
	BalanceAfter  string    `db:"balance_after" json:"balance_after"`
	Confirmed     bool      `db:"confirmed" json:"confirmed"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

// EventType 事件类型
type EventType string

// EventType 常量
const (
	EventTypeMint        EventType = "mint"
	EventTypeBurn        EventType = "burn"
	EventTypeTransferIn  EventType = "transfer_in"
	EventTypeTransferOut EventType = "transfer_out"
)


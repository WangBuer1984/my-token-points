package model

import (
	"time"
)

// SyncState 同步状态模型
type SyncState struct {
	ID                 int       `db:"id" json:"id"`
	ChainName          string    `db:"chain_name" json:"chain_name"`
	LastSyncedBlock    int64     `db:"last_synced_block" json:"last_synced_block"`
	LastConfirmedBlock int64     `db:"last_confirmed_block" json:"last_confirmed_block"`
	LastSyncAt         time.Time `db:"last_sync_at" json:"last_sync_at"`
	Status             string    `db:"status" json:"status"` // running, stopped, error
	ErrorMessage       *string   `db:"error_message" json:"error_message,omitempty"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time `db:"updated_at" json:"updated_at"`
}

// SyncStatus 同步状态常量
const (
	StatusRunning = "running"
	StatusStopped = "stopped"
	StatusError   = "error"
)


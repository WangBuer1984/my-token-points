package balance

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"my-token-points/internal/model"
	"my-token-points/internal/repository"
)

// BalanceUpdate 余额更新请求
type BalanceUpdate struct {
	ChainName   string
	UserAddress string
	TxHash      string
	BlockNumber int64
	BlockTime   time.Time
	EventType   model.EventType
	AmountDelta string // 可以是正数或负数（string格式的big.Int）
}

// BalanceService 余额服务
type BalanceService struct {
	balanceRepo repository.BalanceRepository
	logger      *logrus.Logger
}

// NewBalanceService 创建余额服务
func NewBalanceService(
	balanceRepo repository.BalanceRepository,
	logger *logrus.Logger,
) *BalanceService {
	return &BalanceService{
		balanceRepo: balanceRepo,
		logger:      logger,
	}
}

// UpdateBalance 更新用户余额
func (s *BalanceService) UpdateBalance(ctx context.Context, update *BalanceUpdate) error {
	// 标准化地址（转为小写）
	userAddress := strings.ToLower(update.UserAddress)

	// 解析变动金额
	amountDelta := new(big.Int)
	if _, ok := amountDelta.SetString(update.AmountDelta, 10); !ok {
		return fmt.Errorf("invalid amount delta: %s", update.AmountDelta)
	}

	// 获取当前余额
	currentBalance, err := s.balanceRepo.GetUserBalance(ctx, update.ChainName, userAddress)
	if err != nil {
		return fmt.Errorf("failed to get user balance: %w", err)
	}

	// 计算新余额
	var balanceBefore, balanceAfter *big.Int

	if currentBalance == nil {
		// 新用户
		balanceBefore = big.NewInt(0)
	} else {
		balanceBefore = new(big.Int)
		if _, ok := balanceBefore.SetString(currentBalance.Balance, 10); !ok {
			return fmt.Errorf("invalid current balance: %s", currentBalance.Balance)
		}
	}

	balanceAfter = new(big.Int).Add(balanceBefore, amountDelta)

	// 余额不能为负
	if balanceAfter.Sign() < 0 {
		s.logger.Warnf("Negative balance detected for user %s on %s: before=%s, delta=%s, after=%s",
			userAddress, update.ChainName, balanceBefore.String(), amountDelta.String(), balanceAfter.String())
		// 根据业务需求，可以选择：
		// 1. 返回错误
		// 2. 将余额设置为 0
		// 这里我们将余额设置为 0 并记录警告
		balanceAfter = big.NewInt(0)
	}

	// 记录余额变动（初始状态为未确认）
	change := &model.BalanceChange{
		ChainName:     update.ChainName,
		UserAddress:   userAddress,
		TxHash:        update.TxHash,
		BlockNumber:   update.BlockNumber,
		BlockTime:     update.BlockTime,
		EventType:     update.EventType,
		AmountDelta:   amountDelta.String(),
		BalanceBefore: balanceBefore.String(),
		BalanceAfter:  balanceAfter.String(),
		Confirmed:     true, // 直接标记为已确认，因为事件监听已经有6区块延迟
	}

	if err := s.balanceRepo.RecordBalanceChange(ctx, change); err != nil {
		return fmt.Errorf("failed to record balance change: %w", err)
	}

	// 更新用户余额
	newBalance := &model.UserBalance{
		ChainName:       update.ChainName,
		UserAddress:     userAddress,
		Balance:         balanceAfter.String(),
		LastUpdateBlock: update.BlockNumber,
		LastUpdateTime:  update.BlockTime,
	}

	if err := s.balanceRepo.UpsertUserBalance(ctx, newBalance); err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	s.logger.Debugf("Updated balance for %s on %s: %s -> %s (delta: %s)",
		userAddress, update.ChainName, balanceBefore.String(), balanceAfter.String(), amountDelta.String())

	return nil
}

// GetUserBalance 查询用户余额
func (s *BalanceService) GetUserBalance(ctx context.Context, chainName, userAddress string) (*model.UserBalance, error) {
	userAddress = strings.ToLower(userAddress)
	return s.balanceRepo.GetUserBalance(ctx, chainName, userAddress)
}

// GetBalanceChanges 查询余额变动历史
func (s *BalanceService) GetBalanceChanges(ctx context.Context, chainName, userAddress string, startTime, endTime time.Time) ([]*model.BalanceChange, error) {
	userAddress = strings.ToLower(userAddress)
	return s.balanceRepo.GetBalanceChanges(ctx, chainName, userAddress, startTime, endTime)
}

// RebuildBalance 重建用户余额（从某个区块开始重新计算）
func (s *BalanceService) RebuildBalance(ctx context.Context, chainName, userAddress string, fromBlock int64) error {
	userAddress = strings.ToLower(userAddress)

	// 获取该区块之后的所有变动
	changes, err := s.balanceRepo.GetChangesFromBlock(ctx, chainName, fromBlock)
	if err != nil {
		return fmt.Errorf("failed to get changes: %w", err)
	}

	// 过滤出该用户的变动
	var userChanges []*model.BalanceChange
	for _, change := range changes {
		if strings.EqualFold(change.UserAddress, userAddress) {
			userChanges = append(userChanges, change)
		}
	}

	if len(userChanges) == 0 {
		s.logger.Infof("No changes found for user %s on %s from block %d", userAddress, chainName, fromBlock)
		return nil
	}

	// 重新计算余额
	balance := big.NewInt(0)
	var lastBlock int64
	var lastTime time.Time

	for _, change := range userChanges {
		delta := new(big.Int)
		if _, ok := delta.SetString(change.AmountDelta, 10); !ok {
			return fmt.Errorf("invalid amount delta in change %d: %s", change.ID, change.AmountDelta)
		}

		balance.Add(balance, delta)
		lastBlock = change.BlockNumber
		lastTime = change.BlockTime

		s.logger.Debugf("Rebuild: block=%d, delta=%s, balance=%s", change.BlockNumber, delta.String(), balance.String())
	}

	// 更新余额
	newBalance := &model.UserBalance{
		ChainName:       chainName,
		UserAddress:     userAddress,
		Balance:         balance.String(),
		LastUpdateBlock: lastBlock,
		LastUpdateTime:  lastTime,
	}

	if err := s.balanceRepo.UpsertUserBalance(ctx, newBalance); err != nil {
		return fmt.Errorf("failed to update rebuilt balance: %w", err)
	}

	s.logger.Infof("Rebuilt balance for %s on %s: %s (processed %d changes)",
		userAddress, chainName, balance.String(), len(userChanges))

	return nil
}


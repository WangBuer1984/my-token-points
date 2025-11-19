package points

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

// PointsConfig 积分配置
type PointsConfig struct {
	// 积分利率（每小时每token的积分）
	HourlyRate float64
	// 计算间隔（小时）
	CalcInterval time.Duration
	// 是否启用回溯计算
	EnableBackfill bool
	// 回溯开始时间
	BackfillStartTime *time.Time
}

// PointsService 积分服务
type PointsService struct {
	pointsRepo  repository.PointsRepository
	balanceRepo repository.BalanceRepository
	logger      *logrus.Logger
	config      *PointsConfig
}

// NewPointsService 创建积分服务
func NewPointsService(
	pointsRepo repository.PointsRepository,
	balanceRepo repository.BalanceRepository,
	logger *logrus.Logger,
	config *PointsConfig,
) *PointsService {
	// 设置默认值
	if config.HourlyRate == 0 {
		config.HourlyRate = 0.05 // 默认每小时 5% 的积分率
	}
	if config.CalcInterval == 0 {
		config.CalcInterval = time.Hour // 默认每小时计算一次
	}

	return &PointsService{
		pointsRepo:  pointsRepo,
		balanceRepo: balanceRepo,
		logger:      logger,
		config:      config,
	}
}

// CalculatePointsForPeriod 计算指定时间段的积分
func (s *PointsService) CalculatePointsForPeriod(
	ctx context.Context,
	chainName string,
	userAddress string,
	periodStart time.Time,
	periodEnd time.Time,
	calculationType string,
) (float64, error) {
	userAddress = strings.ToLower(userAddress)

	// 获取该时间段内的所有余额变动
	changes, err := s.balanceRepo.GetBalanceChanges(ctx, chainName, userAddress, periodStart, periodEnd)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance changes: %w", err)
	}

	// 如果没有变动，获取该时间段开始前的最后余额
	if len(changes) == 0 {
		// 尝试获取更早的变动来确定初始余额
		earlierChanges, err := s.balanceRepo.GetBalanceChanges(
			ctx, chainName, userAddress,
			periodStart.Add(-24*time.Hour), periodStart,
		)
		if err == nil && len(earlierChanges) > 0 {
			// 使用最后一个变动的 BalanceAfter 作为初始余额
			lastChange := earlierChanges[len(earlierChanges)-1]
			balance := new(big.Int)
			if _, ok := balance.SetString(lastChange.BalanceAfter, 10); ok {
				// 整个周期保持这个余额
				points := s.calculatePointsForBalance(balance, periodStart, periodEnd)
				
				// 记录积分历史
				snapshot := model.BalanceSnapshots{
					{
						Balance:   lastChange.BalanceAfter,
						StartTime: periodStart,
						EndTime:   periodEnd,
					},
				}
				if err := s.recordPointsHistory(ctx, chainName, userAddress, periodStart, periodEnd, snapshot, points, calculationType); err != nil {
					s.logger.Warnf("Failed to record points history: %v", err)
				}
				
				return points, nil
			}
		}
		
		// 如果找不到历史余额，说明该时间段内余额为0
		return 0, nil
	}

	// 计算时间加权积分
	var totalPoints float64
	var snapshots model.BalanceSnapshots

	// 初始余额（period开始时的余额）
	var currentBalance *big.Int
	currentTime := periodStart

	// 如果第一个变动不是在period开始时，需要找到period开始前的余额
	if changes[0].BlockTime.After(periodStart) {
		// 使用第一个变动的 BalanceBefore
		currentBalance = new(big.Int)
		if _, ok := currentBalance.SetString(changes[0].BalanceBefore, 10); !ok {
			return 0, fmt.Errorf("invalid balance before: %s", changes[0].BalanceBefore)
		}
	} else {
		currentBalance = big.NewInt(0)
	}

	// 遍历所有变动，计算每个时间段的积分
	for _, change := range changes {
		changeTime := change.BlockTime

		// 确保变动时间在计算范围内
		if changeTime.Before(periodStart) {
			// 更新初始余额
			currentBalance = new(big.Int)
			if _, ok := currentBalance.SetString(change.BalanceAfter, 10); !ok {
				return 0, fmt.Errorf("invalid balance after: %s", change.BalanceAfter)
			}
			continue
		}

		if changeTime.After(periodEnd) {
			break
		}

		// 计算从 currentTime 到 changeTime 期间的积分
		if changeTime.After(currentTime) {
			points := s.calculatePointsForBalance(currentBalance, currentTime, changeTime)
			totalPoints += points

			// 记录快照
			if currentBalance.Sign() > 0 {
				snapshots = append(snapshots, model.BalanceSnapshot{
					Balance:   currentBalance.String(),
					StartTime: currentTime,
					EndTime:   changeTime,
				})
			}

			s.logger.Debugf("Period: %s to %s, Balance: %s, Points: %.6f",
				currentTime.Format(time.RFC3339), changeTime.Format(time.RFC3339),
				currentBalance.String(), points)
		}

		// 更新余额和时间
		currentBalance = new(big.Int)
		if _, ok := currentBalance.SetString(change.BalanceAfter, 10); !ok {
			return 0, fmt.Errorf("invalid balance after: %s", change.BalanceAfter)
		}
		currentTime = changeTime
	}

	// 处理最后一个时间段（从最后一个变动到period结束）
	if currentTime.Before(periodEnd) && currentBalance.Sign() > 0 {
		points := s.calculatePointsForBalance(currentBalance, currentTime, periodEnd)
		totalPoints += points

		snapshots = append(snapshots, model.BalanceSnapshot{
			Balance:   currentBalance.String(),
			StartTime: currentTime,
			EndTime:   periodEnd,
		})

		s.logger.Debugf("Final period: %s to %s, Balance: %s, Points: %.6f",
			currentTime.Format(time.RFC3339), periodEnd.Format(time.RFC3339),
			currentBalance.String(), points)
	}

	// 记录积分历史
	if err := s.recordPointsHistory(ctx, chainName, userAddress, periodStart, periodEnd, snapshots, totalPoints, calculationType); err != nil {
		s.logger.Warnf("Failed to record points history: %v", err)
	}

	s.logger.Infof("Calculated points for %s on %s (%s to %s): %.6f",
		userAddress, chainName, periodStart.Format(time.RFC3339), periodEnd.Format(time.RFC3339), totalPoints)

	return totalPoints, nil
}

// calculatePointsForBalance 计算单个余额在指定时间段的积分
func (s *PointsService) calculatePointsForBalance(balance *big.Int, startTime, endTime time.Time) float64 {
	if balance.Sign() <= 0 {
		return 0
	}

	// 计算持有时间（小时）
	duration := endTime.Sub(startTime)
	hours := duration.Hours()

	// 转换余额为 float64 (考虑到 ERC20 的 18 位小数)
	balanceFloat := new(big.Float).SetInt(balance)
	divisor := new(big.Float).SetFloat64(1e18) // 18位小数
	balanceInTokens, _ := new(big.Float).Quo(balanceFloat, divisor).Float64()

	// 计算积分 = 余额 × 利率 × 持有时间
	points := balanceInTokens * s.config.HourlyRate * hours

	return points
}

// CalculatePointsForAllUsers 计算所有用户在指定时间段的积分
func (s *PointsService) CalculatePointsForAllUsers(
	ctx context.Context,
	chainName string,
	periodStart time.Time,
	periodEnd time.Time,
	calculationType string,
) error {
	// 获取所有有余额的用户
	balances, err := s.balanceRepo.GetUserBalances(ctx, chainName, 0, 10000) // TODO: 分页处理
	if err != nil {
		return fmt.Errorf("failed to get user balances: %w", err)
	}

	s.logger.Infof("Calculating points for %d users on %s (period: %s to %s)",
		len(balances), chainName, periodStart.Format(time.RFC3339), periodEnd.Format(time.RFC3339))

	successCount := 0
	errorCount := 0

	for _, balance := range balances {
		// 计算该用户的积分
		earnedPoints, err := s.CalculatePointsForPeriod(
			ctx, chainName, balance.UserAddress,
			periodStart, periodEnd, calculationType,
		)
		if err != nil {
			s.logger.Errorf("Failed to calculate points for user %s: %v", balance.UserAddress, err)
			errorCount++
			continue
		}

		// 更新用户总积分
		if err := s.updateUserTotalPoints(ctx, chainName, balance.UserAddress, earnedPoints, periodEnd); err != nil {
			s.logger.Errorf("Failed to update total points for user %s: %v", balance.UserAddress, err)
			errorCount++
			continue
		}

		successCount++
	}

	s.logger.Infof("Points calculation completed: %d succeeded, %d failed", successCount, errorCount)

	if errorCount > 0 && successCount == 0 {
		return fmt.Errorf("all user points calculations failed")
	}

	return nil
}

// updateUserTotalPoints 更新用户总积分
func (s *PointsService) updateUserTotalPoints(
	ctx context.Context,
	chainName string,
	userAddress string,
	earnedPoints float64,
	calcTime time.Time,
) error {
	// 获取当前积分
	currentPoints, err := s.pointsRepo.GetUserPoints(ctx, chainName, userAddress)
	if err != nil {
		return err
	}

	var newTotalPoints float64
	if currentPoints != nil {
		newTotalPoints = currentPoints.TotalPoints + earnedPoints
	} else {
		newTotalPoints = earnedPoints
	}

	// 更新积分
	updatedPoints := &model.UserPoints{
		ChainName:   chainName,
		UserAddress: userAddress,
		TotalPoints: newTotalPoints,
		LastCalcAt:  &calcTime,
	}

	return s.pointsRepo.UpsertUserPoints(ctx, updatedPoints)
}

// recordPointsHistory 记录积分历史
func (s *PointsService) recordPointsHistory(
	ctx context.Context,
	chainName string,
	userAddress string,
	periodStart time.Time,
	periodEnd time.Time,
	snapshots model.BalanceSnapshots,
	pointsEarned float64,
	calculationType string,
) error {
	history := &model.PointsHistory{
		ChainName:       chainName,
		UserAddress:     userAddress,
		CalcPeriodStart: periodStart,
		CalcPeriodEnd:   periodEnd,
		BalanceSnapshot: snapshots,
		PointsEarned:    pointsEarned,
		CalculationType: calculationType,
	}

	return s.pointsRepo.RecordPointsHistory(ctx, history)
}

// GetUserPoints 查询用户积分
func (s *PointsService) GetUserPoints(ctx context.Context, chainName, userAddress string) (*model.UserPoints, error) {
	userAddress = strings.ToLower(userAddress)
	return s.pointsRepo.GetUserPoints(ctx, chainName, userAddress)
}

// GetUserPointsHistory 查询用户积分历史
func (s *PointsService) GetUserPointsHistory(
	ctx context.Context,
	chainName, userAddress string,
	startTime, endTime time.Time,
) ([]*model.PointsHistory, error) {
	userAddress = strings.ToLower(userAddress)
	return s.pointsRepo.GetPointsHistory(ctx, chainName, userAddress, startTime, endTime)
}

// GetTopUsers 获取积分排行榜
func (s *PointsService) GetTopUsers(ctx context.Context, chainName string, limit int) ([]*model.UserPoints, error) {
	return s.pointsRepo.GetUserPointsList(ctx, chainName, 0, limit)
}

// BackfillPoints 回溯计算积分
func (s *PointsService) BackfillPoints(
	ctx context.Context,
	chainName string,
	startTime time.Time,
	endTime time.Time,
) error {
	s.logger.Infof("Starting points backfill for %s from %s to %s",
		chainName, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

	// 获取未计算的时间段
	uncalculatedPeriods, err := s.pointsRepo.GetUncalculatedPeriods(ctx, chainName, startTime, endTime)
	if err != nil {
		return fmt.Errorf("failed to get uncalculated periods: %w", err)
	}

	if len(uncalculatedPeriods) == 0 {
		s.logger.Info("No uncalculated periods found")
		return nil
	}

	s.logger.Infof("Found %d uncalculated periods", len(uncalculatedPeriods))

	// 逐个计算每个小时的积分
	for _, periodStart := range uncalculatedPeriods {
		periodEnd := periodStart.Add(time.Hour)

		// 确保不超过结束时间
		if periodEnd.After(endTime) {
			periodEnd = endTime
		}

		s.logger.Infof("Backfilling period: %s to %s", periodStart.Format(time.RFC3339), periodEnd.Format(time.RFC3339))

		// 计算所有用户的积分
		if err := s.CalculatePointsForAllUsers(ctx, chainName, periodStart, periodEnd, model.CalcTypeBackfill); err != nil {
			s.logger.Errorf("Failed to calculate points for period %s: %v", periodStart.Format(time.RFC3339), err)
			// 继续处理下一个时间段
			continue
		}
	}

	s.logger.Info("Points backfill completed")
	return nil
}


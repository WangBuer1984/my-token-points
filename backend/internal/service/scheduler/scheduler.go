package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"my-token-points/internal/model"
	"my-token-points/internal/service/points"
)

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	// 启用定时积分计算
	EnableCalculation bool
	// Cron 表达式 (例如: "0 * * * *" 表示每小时执行一次)
	CronExpression string
	// 支持的链配置
	Chains []ChainConfig
}

// ChainConfig 链配置
type ChainConfig struct {
	Name string
	// 是否启用该链的积分计算
	Enabled bool
}

// Scheduler 定时任务调度器
type Scheduler struct {
	cron          *cron.Cron
	pointsService *points.PointsService
	config        *SchedulerConfig
	logger        *logrus.Logger
	
	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
}

// NewScheduler 创建调度器
func NewScheduler(
	pointsService *points.PointsService,
	config *SchedulerConfig,
	logger *logrus.Logger,
) *Scheduler {
	// 设置默认值
	if config.CronExpression == "" {
		config.CronExpression = "0 * * * *" // 默认每小时执行一次
	}

	return &Scheduler{
		cron:          cron.New(cron.WithSeconds()), // 支持秒级精度
		pointsService: pointsService,
		config:        config,
		logger:        logger,
		stopCh:        make(chan struct{}),
	}
}

// Start 启动调度器
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler already running")
	}

	if !s.config.EnableCalculation {
		s.logger.Info("Points calculation scheduler is disabled")
		return nil
	}

	s.logger.Infof("Starting points calculation scheduler with cron: %s", s.config.CronExpression)

	// 添加定时任务
	_, err := s.cron.AddFunc(s.config.CronExpression, func() {
		s.runPointsCalculation()
	})
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	// 启动 cron
	s.cron.Start()
	s.running = true

	s.logger.Info("Points calculation scheduler started successfully")

	// 可选：启动时立即执行一次
	// go s.runPointsCalculation()

	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.logger.Info("Stopping points calculation scheduler...")

	// 停止 cron
	ctx := s.cron.Stop()
	<-ctx.Done()

	close(s.stopCh)
	s.running = false

	s.logger.Info("Points calculation scheduler stopped")
	return nil
}

// runPointsCalculation 执行积分计算
func (s *Scheduler) runPointsCalculation() {
	s.logger.Info("Starting scheduled points calculation")

	ctx := context.Background()

	// 计算上一个小时的积分
	now := time.Now()
	periodEnd := now.Truncate(time.Hour) // 当前小时的开始时间
	periodStart := periodEnd.Add(-time.Hour) // 上一个小时的开始时间

	s.logger.Infof("Calculating points for period: %s to %s",
		periodStart.Format(time.RFC3339), periodEnd.Format(time.RFC3339))

	// 为每条链计算积分
	var wg sync.WaitGroup
	for _, chainConfig := range s.config.Chains {
		if !chainConfig.Enabled {
			continue
		}

		wg.Add(1)
		go func(chainName string) {
			defer wg.Done()

			s.logger.Infof("Calculating points for chain: %s", chainName)

			err := s.pointsService.CalculatePointsForAllUsers(
				ctx,
				chainName,
				periodStart,
				periodEnd,
				model.CalcTypeNormal,
			)
			if err != nil {
				s.logger.Errorf("Failed to calculate points for chain %s: %v", chainName, err)
				return
			}

			s.logger.Infof("Successfully calculated points for chain: %s", chainName)
		}(chainConfig.Name)
	}

	wg.Wait()

	s.logger.Info("Scheduled points calculation completed")
}

// RunBackfill 执行回溯计算
func (s *Scheduler) RunBackfill(ctx context.Context, chainName string, startTime, endTime time.Time) error {
	s.logger.Infof("Starting backfill for chain %s from %s to %s",
		chainName, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

	return s.pointsService.BackfillPoints(ctx, chainName, startTime, endTime)
}

// TriggerCalculation 手动触发积分计算
func (s *Scheduler) TriggerCalculation(ctx context.Context, chainName string) error {
	s.logger.Infof("Manually triggering points calculation for chain: %s", chainName)

	// 计算上一个小时的积分
	now := time.Now()
	periodEnd := now.Truncate(time.Hour)
	periodStart := periodEnd.Add(-time.Hour)

	return s.pointsService.CalculatePointsForAllUsers(
		ctx,
		chainName,
		periodStart,
		periodEnd,
		model.CalcTypeNormal,
	)
}

// IsRunning 检查调度器是否运行中
func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}


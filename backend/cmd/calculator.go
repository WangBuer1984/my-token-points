package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"my-token-points/config"
	"my-token-points/internal/pkg/database"
	"my-token-points/internal/pkg/logger"
	"my-token-points/internal/repository"
	"my-token-points/internal/service/points"
	"my-token-points/internal/service/scheduler"
)

// calculatorCmd 积分计算命令
var calculatorCmd = &cobra.Command{
	Use:   "calculator",
	Short: "启动积分计算服务",
	Long:  "启动积分计算调度器，定时计算用户积分",
	Run: func(cmd *cobra.Command, args []string) {
		runCalculator()
	},
}

func init() {
	rootCmd.AddCommand(calculatorCmd)
}

func runCalculator() {
	fmt.Println("正在启动积分计算服务...")

	// 1. 加载配置
	cfg, err := config.LoadConfig(cfgFile, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	log := logger.InitLogger(cfg.App.LogLevel)
	log.Infof("启动积分计算服务，环境: %s", cfg.App.Env)

	// 3. 初始化数据库
	db, err := database.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()
	log.Info("✅ 数据库连接成功")

	// 4. 创建 Repository 实例
	balanceRepo := repository.NewBalanceRepository(db)
	pointsRepo := repository.NewPointsRepository(db)

	// 5. 创建积分服务
	pointsConfig := &points.PointsConfig{
		HourlyRate:     cfg.Points.HourlyRate,
		CalcInterval:   cfg.Points.CalcInterval,
		EnableBackfill: cfg.Points.EnableBackfill,
	}
	pointsService := points.NewPointsService(pointsRepo, balanceRepo, log, pointsConfig)

	// 6. 创建调度器配置
	schedulerConfig := &scheduler.SchedulerConfig{
		EnableCalculation: true,
		CronExpression:    cfg.Points.CronExpression,
		Chains:            []scheduler.ChainConfig{},
	}

	// 添加链配置
	for _, chain := range cfg.Chains {
		schedulerConfig.Chains = append(schedulerConfig.Chains, scheduler.ChainConfig{
			Name:    chain.Name,
			Enabled: true,
		})
	}

	// 7. 创建调度器
	schedulerService := scheduler.NewScheduler(pointsService, schedulerConfig, log)

	// 8. 启动调度器
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := schedulerService.Start(ctx); err != nil {
		log.Fatalf("启动调度器失败: %v", err)
	}

	log.Info("✅ 积分计算服务启动完成")

	// 9. 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("收到关闭信号，正在停止...")
	if err := schedulerService.Stop(); err != nil {
		log.Errorf("停止调度器失败: %v", err)
	}
	log.Info("✅ 积分计算服务已停止")
}


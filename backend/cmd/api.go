package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"my-token-points/config"
	"my-token-points/internal/api"
	"my-token-points/internal/pkg/database"
	"my-token-points/internal/pkg/logger"
	"my-token-points/internal/repository"
	"my-token-points/internal/service/balance"
	"my-token-points/internal/service/points"
	"my-token-points/internal/service/scheduler"
)

// apiCmd API服务命令
var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "启动API服务",
	Long:  "启动HTTP API服务，提供余额和积分查询接口",
	Run: func(cmd *cobra.Command, args []string) {
		runAPI()
	},
}

func init() {
	rootCmd.AddCommand(apiCmd)
}

func runAPI() {
	fmt.Println("正在启动API服务...")

	// 1. 加载配置
	cfg, err := config.LoadConfig(cfgFile, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	log := logger.InitLogger(cfg.App.LogLevel)
	log.Infof("启动API服务，环境: %s", cfg.App.Env)

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

	// 5. 创建服务实例
	balanceService := balance.NewBalanceService(balanceRepo, log)

	pointsConfig := &points.PointsConfig{
		HourlyRate:   cfg.Points.HourlyRate,
		CalcInterval: cfg.Points.CalcInterval,
	}
	pointsService := points.NewPointsService(pointsRepo, balanceRepo, log, pointsConfig)

	// 6. 创建调度器（用于手动触发计算）
	schedulerConfig := &scheduler.SchedulerConfig{
		EnableCalculation: false, // API服务中不启动自动调度
		Chains:            []scheduler.ChainConfig{},
	}
	for _, chain := range cfg.Chains {
		schedulerConfig.Chains = append(schedulerConfig.Chains, scheduler.ChainConfig{
			Name:    chain.Name,
			Enabled: true,
		})
	}
	schedulerService := scheduler.NewScheduler(pointsService, schedulerConfig, log)

	// 7. 创建API服务器
	serverConfig := &api.ServerConfig{
		Host: cfg.API.Host,
		Port: cfg.API.Port,
		Mode: cfg.API.Mode,
	}
	apiServer := api.NewServer(serverConfig, balanceService, pointsService, schedulerService, log)

	// 8. 启动API服务器
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Fatalf("API服务器启动失败: %v", err)
		}
	}()

	log.Infof("✅ API服务启动完成 (http://%s:%d)", cfg.API.Host, cfg.API.Port)
	log.Infof("API文档: http://%s:%d/api/v1", cfg.API.Host, cfg.API.Port)

	// 9. 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("收到关闭信号，正在停止...")

	// 停止API服务器
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := apiServer.Stop(ctx); err != nil {
		log.Errorf("停止API服务器失败: %v", err)
	}

	log.Info("✅ API服务已停止")
}

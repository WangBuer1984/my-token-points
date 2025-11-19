package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"my-token-points/config"
	"my-token-points/internal/api"
	"my-token-points/internal/pkg/database"
	"my-token-points/internal/pkg/logger"
	"my-token-points/internal/repository"
	"my-token-points/internal/service/balance"
	"my-token-points/internal/service/listener"
	"my-token-points/internal/service/points"
	"my-token-points/internal/service/scheduler"
)

// startCmd 启动所有服务
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "启动所有服务",
	Long:  "启动事件监听、积分计算和API服务",
	Run: func(cmd *cobra.Command, args []string) {
		runStart()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func runStart() {
	fmt.Println("正在启动服务...")

	// 1. 加载配置
	cfg, err := config.LoadConfig(cfgFile, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	log := logger.InitLogger(cfg.App.LogLevel)
	log.Infof("启动 %s 服务，环境: %s", cfg.App.Name, cfg.App.Env)

	// 3. 初始化数据库
	db, err := database.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()
	log.Info("✅ 数据库连接成功")

	// 4. 创建 Repository 实例
	syncRepo := repository.NewSyncRepository(db)
	balanceRepo := repository.NewBalanceRepository(db)
	pointsRepo := repository.NewPointsRepository(db)

	// 5. 创建 Service 实例
	balanceService := balance.NewBalanceService(balanceRepo, log)

	pointsConfig := &points.PointsConfig{
		HourlyRate:     cfg.Points.HourlyRate,
		CalcInterval:   cfg.Points.CalcInterval,
		EnableBackfill: cfg.Points.EnableBackfill,
	}
	pointsService := points.NewPointsService(pointsRepo, balanceRepo, log, pointsConfig)

	// 6. 创建调度器
	schedulerConfig := &scheduler.SchedulerConfig{
		EnableCalculation: cfg.Points.Enabled,
		CronExpression:    cfg.Points.CronExpression,
		Chains:            []scheduler.ChainConfig{},
	}
	for _, chain := range cfg.Chains {
		schedulerConfig.Chains = append(schedulerConfig.Chains, scheduler.ChainConfig{
			Name:    chain.Name,
			Enabled: true,
		})
	}
	schedulerService := scheduler.NewScheduler(pointsService, schedulerConfig, log)

	// 7. 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// 8. 启动事件监听服务
	log.Info("启动事件监听服务...")
	for _, chainCfg := range cfg.Chains {
		wg.Add(1)
		go func(chain config.ChainConfig) {
			defer wg.Done()
			log.Infof("启动 %s 链的事件监听...", chain.Name)

			// 创建事件监听器
			eventListener, err := listener.NewEventListener(
				chain.Name,
				&chain,
				int(cfg.Confirmation.Blocks),
				syncRepo,
				balanceService,
				log,
			)
			if err != nil {
				log.Errorf("创建 %s 监听器失败: %v", chain.Name, err)
				return
			}

			// 启动监听
			if err := eventListener.Start(ctx); err != nil {
				log.Errorf("启动 %s 监听器失败: %v", chain.Name, err)
				return
			}

			<-ctx.Done()
			eventListener.Stop()
			log.Infof("%s 监听器已停止", chain.Name)
		}(chainCfg)
	}

	// 9. 启动积分计算服务
	if cfg.Points.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Info("启动积分计算调度器...")

			if err := schedulerService.Start(ctx); err != nil {
				log.Errorf("启动调度器失败: %v", err)
				return
			}

			<-ctx.Done()
			if err := schedulerService.Stop(); err != nil {
				log.Errorf("停止调度器失败: %v", err)
			}
			log.Info("积分计算调度器已停止")
		}()
	}

	// 10. 启动API服务
	var apiServer *api.Server
	if cfg.API.Enabled {
		// 创建API服务器
		serverConfig := &api.ServerConfig{
			Host: cfg.API.Host,
			Port: cfg.API.Port,
			Mode: cfg.API.Mode,
		}
		apiServer = api.NewServer(serverConfig, balanceService, pointsService, schedulerService, log)

		// 在单独的 goroutine 中启动服务器
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Infof("启动API服务 (http://%s:%d)...", cfg.API.Host, cfg.API.Port)

			// 启动服务器（这是阻塞的）
			if err := apiServer.Start(); err != nil {
				log.Errorf("API服务器启动失败: %v", err)
			}
		}()

		// 在另一个 goroutine 中等待关闭信号
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ctx.Done()

			// 收到关闭信号，优雅关闭 API 服务器
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()

			if err := apiServer.Stop(shutdownCtx); err != nil {
				log.Errorf("停止API服务器失败: %v", err)
			} else {
				log.Info("API服务器已停止")
			}
		}()
	}

	// 等待所有服务启动
	time.Sleep(2 * time.Second)
	log.Info("✅ 所有服务启动完成")

	if cfg.API.Enabled {
		log.Infof("📊 API服务地址: http://%s:%d", cfg.API.Host, cfg.API.Port)
		log.Infof("📚 健康检查: http://%s:%d/health", cfg.API.Host, cfg.API.Port)
	}
	if cfg.Points.Enabled {
		log.Info("⏰ 积分计算调度器已启动")
	}

	// 11. 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("收到关闭信号，正在优雅关闭...")
	cancel()
	wg.Wait()
	log.Info("✅ 服务已停止")
}

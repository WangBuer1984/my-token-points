package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"my-token-points/config"
	"my-token-points/internal/pkg/database"
	"my-token-points/internal/pkg/logger"
	"my-token-points/internal/repository"
	"my-token-points/internal/service/balance"
	"my-token-points/internal/service/listener"
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
	
	// 5. 创建 Service 实例
	balanceService := balance.NewBalanceService(balanceRepo, log)
	
	// 6. 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// 7. 启动事件监听服务
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

	// 8. 启动积分计算服务 (第二阶段)
	if cfg.Points.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Info("启动积分计算服务...")
			// TODO (第二阶段): 实现积分计算服务
			// pointsRepo := repository.NewPointsRepository(db)
			// pointsSvc := points.NewPointsService(balanceRepo, pointsRepo, log)
			// pointsSvc.Start(ctx)
			<-ctx.Done()
		}()
	}

	// 9. 启动API服务 (第二阶段)
	if cfg.API.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Infof("启动API服务 (http://%s:%d)...", cfg.API.Host, cfg.API.Port)
			// TODO (第二阶段): 实现API服务
			// pointsRepo := repository.NewPointsRepository(db)
			// apiServer := api.NewServer(cfg, balanceRepo, pointsRepo, syncRepo, log)
			// apiServer.Start(ctx)
			<-ctx.Done()
		}()
	}

	log.Info("✅ 所有服务启动完成")

	// 10. 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("收到关闭信号，正在优雅关闭...")
	cancel()
	wg.Wait()
	log.Info("✅ 服务已停止")
}


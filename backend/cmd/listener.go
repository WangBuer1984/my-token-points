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

var (
	chainName string
)

// listenerCmd 事件监听命令
var listenerCmd = &cobra.Command{
	Use:   "listener",
	Short: "启动事件监听服务",
	Long:  "启动区块链事件监听服务，追踪代币转账、mint和burn事件",
	Run: func(cmd *cobra.Command, args []string) {
		runListener()
	},
}

func init() {
	listenerCmd.Flags().StringVar(&chainName, "chain", "", "指定监听的链（不指定则监听所有链）")
	rootCmd.AddCommand(listenerCmd)
}

func runListener() {
	fmt.Println("正在启动事件监听服务...")

	// 1. 加载配置
	cfg, err := config.LoadConfig(cfgFile, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	log := logger.InitLogger(cfg.App.LogLevel)
	log.Infof("启动事件监听服务，环境: %s", cfg.App.Env)

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

	// 7. 过滤要监听的链
	chainsToListen := cfg.Chains
	if chainName != "" {
		chainsToListen = []config.ChainConfig{}
		for _, chain := range cfg.Chains {
			if chain.Name == chainName {
				chainsToListen = append(chainsToListen, chain)
				break
			}
		}
		if len(chainsToListen) == 0 {
			log.Fatalf("未找到链配置: %s", chainName)
		}
	}

	// 8. 启动事件监听服务
	for _, chainCfg := range chainsToListen {
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

	log.Info("✅ 事件监听服务启动完成")

	// 9. 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("收到关闭信号，正在停止...")
	cancel()
	wg.Wait()
	log.Info("✅ 事件监听服务已停止")
}


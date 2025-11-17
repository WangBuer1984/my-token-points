package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	env     string
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "my-token-points",
	Short: "多链ERC20代币积分系统",
	Long: `
多链 ERC20 代币事件追踪和积分计算系统

功能:
  - 追踪代币事件 (Transfer, Mint, Burn)
  - 重建用户余额
  - 基于持有时间计算积分
  - 支持多链部署 (Sepolia, Base Sepolia)
  - 支持积分回溯计算

使用示例:
  # 启动所有服务
  my-token-points start --env dev

  # 仅启动事件监听
  my-token-points listener --env dev

  # 仅启动积分计算
  my-token-points calculator --env dev

  # 数据库迁移
  my-token-points migrate --env dev
`,
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// 全局标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认: ./config)")
	rootCmd.PersistentFlags().StringVar(&env, "env", "dev", "环境 (dev/prod)")
}


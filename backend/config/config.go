package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 全局配置结构
type Config struct {
	App          AppConfig          `mapstructure:"app"`
	Database     DatabaseConfig     `mapstructure:"database"`
	API          APIConfig          `mapstructure:"api"`
	Chains       []ChainConfig      `mapstructure:"chains"`
	Confirmation ConfirmationConfig `mapstructure:"confirmation"`
	Points       PointsConfig       `mapstructure:"points"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name     string `mapstructure:"name"`
	Env      string `mapstructure:"env"`
	LogLevel string `mapstructure:"log_level"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // 秒
}

// APIConfig API服务配置
type APIConfig struct {
	Enabled bool       `mapstructure:"enabled"`
	Host    string     `mapstructure:"host"`
	Port    int        `mapstructure:"port"`
	Mode    string     `mapstructure:"mode"` // debug, release, test
	CORS    CORSConfig `mapstructure:"cors"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	Enabled        bool     `mapstructure:"enabled"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

// ChainConfig 区块链配置
type ChainConfig struct {
	Name            string `mapstructure:"name"`
	ChainID         int64  `mapstructure:"chain_id"`
	RPCURL          string `mapstructure:"rpc_url"`
	ContractAddress string `mapstructure:"contract_address"`
	StartBlock      uint64 `mapstructure:"start_block"`
	ScanInterval    int    `mapstructure:"scan_interval"` // 秒
	BatchSize       uint64 `mapstructure:"batch_size"`
	ExplorerURL     string `mapstructure:"explorer_url"`      // 区块浏览器 URL
	ExplorerAPIURL  string `mapstructure:"explorer_api_url"`  // 区块浏览器 API URL
}

// ConfirmationConfig 确认机制配置
type ConfirmationConfig struct {
	Blocks uint64 `mapstructure:"blocks"`
}

// PointsConfig 积分计算配置
type PointsConfig struct {
	Enabled           bool          `mapstructure:"enabled"`
	CronExpression    string        `mapstructure:"cron_expression"`    // Cron表达式
	HourlyRate        float64       `mapstructure:"hourly_rate"`        // 小时积分利率
	CalcInterval      time.Duration `mapstructure:"calc_interval"`      // 计算间隔
	EnableBackfill    bool          `mapstructure:"enable_backfill"`    // 启用回溯计算
	BackfillOnStartup bool          `mapstructure:"backfill_on_startup"` // 启动时自动回溯
	BackfillMaxDays   int           `mapstructure:"backfill_max_days"`   // 最多回溯天数
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string, env string) (*Config, error) {
	v := viper.New()

	// 设置配置文件名和类型
	v.SetConfigName(env) // dev.yaml 或 prod.yaml
	v.SetConfigType("yaml")

	// 添加配置文件搜索路径
	if configPath != "" {
		v.AddConfigPath(configPath)
	}
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	// 支持环境变量
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析配置到结构体
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// validateConfig 验证配置有效性
func validateConfig(config *Config) error {
	// 验证数据库配置
	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if config.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}

	// 验证链配置
	if len(config.Chains) == 0 {
		return fmt.Errorf("at least one chain configuration is required")
	}

	for _, chain := range config.Chains {
		if chain.Name == "" {
			return fmt.Errorf("chain name is required")
		}
		if chain.RPCURL == "" {
			return fmt.Errorf("rpc_url is required for chain %s", chain.Name)
		}
		if chain.ContractAddress == "" {
			return fmt.Errorf("contract_address is required for chain %s", chain.Name)
		}
		if chain.ChainID == 0 {
			return fmt.Errorf("chain_id is required for chain %s", chain.Name)
		}
	}

	// 验证确认区块数
	if config.Confirmation.Blocks == 0 {
		config.Confirmation.Blocks = 6 // 默认值
	}

	// 验证积分配置
	if config.Points.Enabled {
		if config.Points.CronExpression == "" {
			config.Points.CronExpression = "0 * * * *" // 默认每小时
		}
		if config.Points.HourlyRate == 0 {
			config.Points.HourlyRate = 0.05 // 默认5%
		}
		if config.Points.CalcInterval == 0 {
			config.Points.CalcInterval = time.Hour // 默认1小时
		}
	}

	// 设置API默认模式
	if config.API.Mode == "" {
		if config.App.Env == "dev" {
			config.API.Mode = "debug"
		} else {
			config.API.Mode = "release"
		}
	}

	return nil
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}


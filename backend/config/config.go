package config

import (
	"fmt"
	"strings"

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
	Enabled           bool    `mapstructure:"enabled"`
	CronSchedule      string  `mapstructure:"cron_schedule"`
	AnnualRate        float64 `mapstructure:"annual_rate"`
	BackfillOnStartup bool    `mapstructure:"backfill_on_startup"`
	BackfillMaxDays   int     `mapstructure:"backfill_max_days"`
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
		if config.Points.CronSchedule == "" {
			return fmt.Errorf("points cron_schedule is required when points is enabled")
		}
		if config.Points.AnnualRate <= 0 {
			return fmt.Errorf("points annual_rate must be positive")
		}
	}

	return nil
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}


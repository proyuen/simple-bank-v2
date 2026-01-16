// Package config 负责加载和管理应用程序配置
// 使用 Viper 库从环境变量和 .env 文件中读取配置
package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config 存储应用程序的所有配置
// 这些值从环境变量中读取，使用 mapstructure 标签进行映射
type Config struct {
	// ========== 数据库配置 ==========
	DBHost     string `mapstructure:"DB_HOST"`     // 数据库主机地址
	DBPort     string `mapstructure:"DB_PORT"`     // 数据库端口
	DBUser     string `mapstructure:"DB_USER"`     // 数据库用户名
	DBPassword string `mapstructure:"DB_PASSWORD"` // 数据库密码
	DBName     string `mapstructure:"DB_NAME"`     // 数据库名称

	// ========== 服务器配置 ==========
	ServerAddress string `mapstructure:"SERVER_ADDRESS"` // 服务器监听地址 (例如: 0.0.0.0:8080)

	// ========== JWT 配置 ==========
	TokenSecretKey       string        `mapstructure:"TOKEN_SECRET_KEY"`        // JWT 签名密钥
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`   // Access Token 有效期
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`  // Refresh Token 有效期
}

// LoadConfig 从指定路径加载配置
//
// 加载顺序：
// 1. 读取 .env 文件（如果存在）
// 2. 读取系统环境变量（会覆盖 .env 中的值）
//
// 参数:
//   - path: 配置文件所在目录路径（例如 "." 表示当前目录）
//
// 返回:
//   - config: 加载完成的配置结构体
//   - err: 如果加载失败则返回错误
//
// 使用示例:
//
//	cfg, err := config.LoadConfig(".")
//	if err != nil {
//	    log.Fatal("无法加载配置:", err)
//	}
//	fmt.Println("服务器地址:", cfg.ServerAddress)
func LoadConfig(path string) (config Config, err error) {
	// 告诉 Viper 在哪个目录查找配置文件
	viper.AddConfigPath(path)

	// 设置配置文件名（不包含扩展名）
	viper.SetConfigName(".env")

	// 设置配置文件类型为环境变量格式
	viper.SetConfigType("env")

	// 自动读取系统环境变量
	// 这允许环境变量覆盖 .env 文件中的值
	// 这在 Docker/Kubernetes 部署时非常有用
	viper.AutomaticEnv()

	// 尝试读取配置文件
	err = viper.ReadInConfig()
	if err != nil {
		// 如果是文件不存在错误，我们可以继续（依赖环境变量）
		// 如果是其他错误，则返回
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return
		}
		// 文件不存在时清除错误，继续使用环境变量
		err = nil
	}

	// 将配置值解析到 Config 结构体
	// mapstructure 标签指定了环境变量名与结构体字段的映射关系
	err = viper.Unmarshal(&config)
	return
}

// DBSource 返回 PostgreSQL 连接字符串 (DSN)
//
// DSN 格式: host=xxx port=xxx user=xxx password=xxx dbname=xxx sslmode=disable
//
// 使用示例:
//
//	dsn := cfg.DBSource()
//	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
func (c *Config) DBSource() string {
	return "host=" + c.DBHost +
		" port=" + c.DBPort +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=disable"
}

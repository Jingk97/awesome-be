// Package config 提供配置管理功能
//
// 配置模块是 GoFast 框架的基础设施核心，负责管理应用的所有配置信息。
// 基于 Viper 实现，提供了强大的配置管理能力。
//
// 核心特性：
// - YAML 格式配置文件
// - 多环境支持（dev、test、prod）
// - 环境变量覆盖（前缀 GOFAST_）
// - 命令行参数支持
// - 配置验证
//
// 使用示例：
//
//	cfg := config.Load()
//	fmt.Println(cfg.Server.HTTP.Port)
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config 是应用的完整配置结构
//
// 架构思路：
// - 使用嵌套结构体组织配置，清晰易读
// - 使用 mapstructure tag 支持 Viper 解析
// - 使用指针类型表示可选配置（如 Slaves）
//
// 初级工程师学习要点：
// - 理解结构体标签（struct tag）的作用
// - 掌握嵌套结构体的使用
// - 了解指针和值类型的区别
type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Server     ServerConfig     `mapstructure:"server"`
	Databases  []DatabaseConfig `mapstructure:"databases"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Logger     LoggerConfig     `mapstructure:"logger"`
	Health     HealthConfig     `mapstructure:"health"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Middleware MiddlewareConfig `mapstructure:"middleware"`
}

// AppConfig 应用基础配置
type AppConfig struct {
	Name string `mapstructure:"name"` // 应用名称
	Env  string `mapstructure:"env"`  // 运行环境: dev, test, prod
}

// ServerConfig 服务器配置
type ServerConfig struct {
	HTTP HTTPConfig `mapstructure:"http"`
	GRPC GRPCConfig `mapstructure:"grpc"`
}

// HTTPConfig HTTP 服务配置
type HTTPConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	MaxHeaderBytes int           `mapstructure:"max_header_bytes"`
}

// GRPCConfig gRPC 服务配置
type GRPCConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	MaxRecvMsgSize int    `mapstructure:"max_recv_msg_size"`
	MaxSendMsgSize int    `mapstructure:"max_send_msg_size"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Name            string             `mapstructure:"name"`
	Type            string             `mapstructure:"type"`
	MaxIdleConns    int                `mapstructure:"max_idle_conns"`
	MaxOpenConns    int                `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration      `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration      `mapstructure:"conn_max_idle_time"`
	DialTimeout     time.Duration      `mapstructure:"dial_timeout"`
	ReadTimeout     time.Duration      `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration      `mapstructure:"write_timeout"`
	LogLevel        string             `mapstructure:"log_level"`
	SlowThreshold   time.Duration      `mapstructure:"slow_threshold"`
	Reload          ReloadConfig       `mapstructure:"reload"`
	HealthCheck     HealthCheckConfig  `mapstructure:"health_check"`
	Master          DBInstanceConfig   `mapstructure:"master"`
	Slaves          []DBInstanceConfig `mapstructure:"slaves"`
}

// DBInstanceConfig 数据库实例配置
type DBInstanceConfig struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	Database  string `mapstructure:"database"`
	Charset   string `mapstructure:"charset"`
	ParseTime bool   `mapstructure:"parse_time"`
	Loc       string `mapstructure:"loc"`
	SSLMode   string `mapstructure:"sslmode"` // PostgreSQL 专用
}

// ReloadConfig 热更新配置
type ReloadConfig struct {
	GracePeriod   time.Duration `mapstructure:"grace_period"`
	ForceClose    bool          `mapstructure:"force_close"`
	CheckInterval time.Duration `mapstructure:"check_interval"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Enabled  bool          `mapstructure:"enabled"`
	Interval time.Duration `mapstructure:"interval"`
	Timeout  time.Duration `mapstructure:"timeout"`
	Retries  int           `mapstructure:"retries"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Mode               string            `mapstructure:"mode"`
	Addr               string            `mapstructure:"addr"`
	Password           string            `mapstructure:"password"`
	DB                 int               `mapstructure:"db"`
	PoolSize           int               `mapstructure:"pool_size"`
	MinIdleConns       int               `mapstructure:"min_idle_conns"`
	MaxRetries         int               `mapstructure:"max_retries"`
	DialTimeout        time.Duration     `mapstructure:"dial_timeout"`
	ReadTimeout        time.Duration     `mapstructure:"read_timeout"`
	WriteTimeout       time.Duration     `mapstructure:"write_timeout"`
	PoolTimeout        time.Duration     `mapstructure:"pool_timeout"`
	IdleTimeout        time.Duration     `mapstructure:"idle_timeout"`
	IdleCheckFrequency time.Duration     `mapstructure:"idle_check_frequency"`
	Reload             ReloadConfig      `mapstructure:"reload"`
	HealthCheck        HealthCheckConfig `mapstructure:"health_check"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level            string           `mapstructure:"level"`
	Format           string           `mapstructure:"format"`
	Output           string           `mapstructure:"output"`
	File             LoggerFileConfig `mapstructure:"file"`
	EnableCaller     bool             `mapstructure:"enable_caller"`
	EnableStacktrace bool             `mapstructure:"enable_stacktrace"`
}

// LoggerFileConfig 日志文件配置
type LoggerFileConfig struct {
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// HealthConfig 健康检查模块配置
type HealthConfig struct {
	Timeout  time.Duration `mapstructure:"timeout"`
	Detailed bool          `mapstructure:"detailed"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret        string `mapstructure:"secret"`
	Expire        int    `mapstructure:"expire"`
	RefreshExpire int    `mapstructure:"refresh_expire"`
	Issuer        string `mapstructure:"issuer"`
}

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
	CORS      CORSConfig      `mapstructure:"cors"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Trace     TraceConfig     `mapstructure:"trace"`
}

// CORSConfig CORS 配置
type CORSConfig struct {
	Enabled       bool     `mapstructure:"enabled"`
	AllowOrigins  []string `mapstructure:"allow_origins"`
	AllowMethods  []string `mapstructure:"allow_methods"`
	AllowHeaders  []string `mapstructure:"allow_headers"`
	ExposeHeaders []string `mapstructure:"expose_headers"`
	MaxAge        int      `mapstructure:"max_age"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled  bool          `mapstructure:"enabled"`
	Requests int           `mapstructure:"requests"`
	Window   time.Duration `mapstructure:"window"`
}

// TraceConfig 链路追踪配置
type TraceConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Header  string `mapstructure:"header"`
}

// Load 加载配置
//
// 架构思路：
// 1. 设置默认值（保证即使没有配置文件也能运行）
// 2. 读取配置文件
// 3. 读取环境变量（覆盖配置文件）
// 4. 读取命令行参数（最高优先级）
// 5. 验证配置
//
// 初级工程师学习要点：
// - 理解配置优先级的重要性
// - 掌握 Viper 的基本用法
// - 学习错误处理的最佳实践
//
// 高级工程师思考：
// - 如何支持配置热更新？
// - 如何处理敏感信息（密码、密钥）？
// - 如何支持配置中心（如 Consul、etcd）？
func Load() (*Config, error) {
	v := viper.New()

	// 第一步：设置默认值
	setDefaults(v)

	// 第二步：设置配置文件
	v.SetConfigName("config")       // 配置文件名（不含扩展名）
	v.SetConfigType("yaml")         // 配置文件类型
	v.AddConfigPath(".")            // 当前目录
	v.AddConfigPath("./config")     // config 目录
	v.AddConfigPath("/etc/gofast/") // 系统配置目录

	// 第三步：读取配置文件
	if err := v.ReadInConfig(); err != nil {
		// 配置文件不存在不是致命错误，使用默认值即可
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// 第四步：设置环境变量
	// 环境变量前缀为 GOFAST_
	// 例如：GOFAST_SERVER_HTTP_PORT=8080
	v.SetEnvPrefix("GOFAST")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 第五步：绑定命令行参数
	bindFlags(v)

	// 第六步：解析配置到结构体
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 第七步：验证配置
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// setDefaults 设置默认配置
//
// 架构思路：
// - 提供合理的默认值，确保应用能够开箱即用
// - 默认值应该适合开发环境，生产环境通过配置文件覆盖
//
// 初级工程师学习要点：
// - 理解默认值的重要性（降低配置复杂度）
// - 掌握 Viper 的 SetDefault 方法
func setDefaults(v *viper.Viper) {
	// 应用配置
	v.SetDefault("app.name", "gofast")
	v.SetDefault("app.env", "dev")
	// 注意：不再设置 app.debug，已移除该配置项

	// HTTP 服务配置
	v.SetDefault("server.http.host", "0.0.0.0")
	v.SetDefault("server.http.port", 8080)
	v.SetDefault("server.http.read_timeout", "60s")
	v.SetDefault("server.http.write_timeout", "60s")
	v.SetDefault("server.http.max_header_bytes", 1048576)

	// gRPC 服务配置
	v.SetDefault("server.grpc.host", "0.0.0.0")
	v.SetDefault("server.grpc.port", 9090)
	v.SetDefault("server.grpc.max_recv_msg_size", 4194304)
	v.SetDefault("server.grpc.max_send_msg_size", 4194304)

	// 日志配置
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
	v.SetDefault("logger.output", "stdout")
	v.SetDefault("logger.enable_caller", true)
	v.SetDefault("logger.enable_stacktrace", false)

	// 健康检查配置
	v.SetDefault("health.timeout", "5s")
	v.SetDefault("health.detailed", true)

	// JWT 配置
	v.SetDefault("jwt.expire", 7200)
	v.SetDefault("jwt.refresh_expire", 604800)
	v.SetDefault("jwt.issuer", "gofast")

	// CORS 配置
	v.SetDefault("middleware.cors.enabled", true)
	v.SetDefault("middleware.cors.allow_origins", []string{"*"})
	v.SetDefault("middleware.cors.allow_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("middleware.cors.allow_headers", []string{"*"})
	v.SetDefault("middleware.cors.max_age", 86400)

	// 链路追踪配置
	v.SetDefault("middleware.trace.enabled", true)
	v.SetDefault("middleware.trace.header", "X-Trace-ID")
}

// bindFlags 绑定命令行参数
//
// 架构思路：
// - 支持常用配置的命令行参数覆盖
// - 命令行参数优先级最高，方便调试和临时修改
//
// 初级工程师学习要点：
// - 理解命令行参数的使用场景
// - 掌握 pflag 库的基本用法
func bindFlags(v *viper.Viper) {
	pflag.String("config", "", "配置文件路径")
	pflag.String("env", "", "运行环境 (dev/test/prod)")
	pflag.Int("port", 0, "HTTP 服务端口")

	pflag.Parse()
	v.BindPFlags(pflag.CommandLine)
}

// validate 验证配置
//
// 架构思路：
// - 在应用启动时验证配置，快速失败（Fail Fast）
// - 避免运行时才发现配置错误
//
// 初级工程师学习要点：
// - 理解配置验证的重要性
// - 掌握基本的数据验证方法
//
// 高级工程师思考：
// - 如何使用验证库（如 validator）简化验证逻辑？
// - 如何提供更友好的错误提示？
func validate(cfg *Config) error {
	// 验证应用配置
	if cfg.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}

	if cfg.App.Env != "dev" && cfg.App.Env != "test" && cfg.App.Env != "prod" {
		return fmt.Errorf("app.env must be one of: dev, test, prod")
	}

	// 验证 HTTP 服务配置
	if cfg.Server.HTTP.Port <= 0 || cfg.Server.HTTP.Port > 65535 {
		return fmt.Errorf("server.http.port must be between 1 and 65535")
	}

	// TODO: 添加更多配置验证
	// - 数据库配置验证
	// - Redis 配置验证
	// - JWT 密钥验证

	return nil
}

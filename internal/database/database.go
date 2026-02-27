// Package database 提供数据库访问功能
//
// 数据库模块基于 GORM 封装，提供统一的数据库访问接口。
//
// 初级工程师学习要点：
// - 理解读写分离的概念和好处
// - 掌握连接池的配置和管理
// - 学习如何使用 Context 传递请求信息
package database

import (
	"context"
	"fmt"
	"sync/atomic"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/jingpc/awesome-be/internal/config"
	"github.com/jingpc/awesome-be/internal/health"
	"github.com/jingpc/awesome-be/internal/logger"
)

// Database 数据库实例
//
// 初级工程师学习要点：
// - Database 封装了 GORM 的 DB 实例
// - 支持读写分离：master 用于写操作，slaves 用于读操作
// - 使用 atomic.Uint32 实现轮询负载均衡（线程安全）
type Database struct {
	name       string
	config     config.DatabaseConfig
	master     *gorm.DB      // 主库（写操作）
	slaves     []*gorm.DB    // 从库列表（读操作）
	slaveIndex atomic.Uint32 // 从库轮询索引
}

// New 创建数据库实例
//
// 初级工程师学习要点：
// - 这是工厂函数，根据配置创建数据库连接
// - 支持多种数据库类型（MySQL、PostgreSQL、SQLite）
// - 自动注册到健康检查管理器
// - 接收 logger 参数，将 GORM 日志集成到统一日志系统
func New(cfg config.DatabaseConfig, log *logger.Logger, healthMgr *health.Manager) (*Database, error) {
	// 1. 创建主库连接
	master, err := connect(cfg, cfg.Master, log)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master database: %w", err)
	}

	// 2. 配置连接池
	sqlDB, err := master.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// 3. 创建从库连接（如果配置了）
	var slaves []*gorm.DB
	for i, slaveCfg := range cfg.Slaves {
		slave, err := connect(cfg, slaveCfg, log)
		if err != nil {
			// 从库连接失败不是致命错误，记录日志并继续
			// 后续会降级到主库
			log.Warn(fmt.Sprintf("failed to connect to slave database %d: %v", i, err))
			continue
		}

		// 配置从库连接池
		slaveSQLDB, _ := slave.DB()
		slaveSQLDB.SetMaxIdleConns(cfg.MaxIdleConns)
		slaveSQLDB.SetMaxOpenConns(cfg.MaxOpenConns)
		slaveSQLDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
		slaveSQLDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

		slaves = append(slaves, slave)
	}

	db := &Database{
		name:   cfg.Name,
		config: cfg,
		master: master,
		slaves: slaves,
	}

	// 4. 验证数据库连接（执行 SELECT 1）
	ctx := context.Background()
	if err := master.WithContext(ctx).Exec("SELECT 1").Error; err != nil {
		return nil, fmt.Errorf("failed to verify database connection: %w", err)
	}

	// 5. 注册健康检查（如果提供了 healthMgr）
	if healthMgr != nil {
		checker := &DatabaseHealthChecker{
			name: cfg.Name,
			db:   master,
		}
		healthMgr.Register(checker)
	}

	return db, nil
}

// connect 创建数据库连接
//
// 初级工程师学习要点：
// - 根据数据库类型选择不同的驱动
// - 构建 DSN（Data Source Name）连接字符串
// - 使用自定义 GORM 日志适配器集成到统一日志系统
func connect(cfg config.DatabaseConfig, instance config.DBInstanceConfig, log *logger.Logger) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Type {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
			instance.Username,
			instance.Password,
			instance.Host,
			instance.Port,
			instance.Database,
			instance.Charset,
			instance.ParseTime,
			instance.Loc,
		)
		dialector = mysql.Open(dsn)

	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			instance.Host,
			instance.Port,
			instance.Username,
			instance.Password,
			instance.Database,
			instance.SSLMode,
		)
		dialector = postgres.Open(dsn)

	case "sqlite":
		dialector = sqlite.Open(instance.Database)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	// 配置 GORM
	gormConfig := &gorm.Config{
		Logger: NewGormLogger(log, cfg.LogLevel, cfg.SlowThreshold),
	}

	// 创建连接
	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Master 获取主库连接（用于写操作）
//
// 初级工程师学习要点：
// - 所有写操作（INSERT、UPDATE、DELETE）都应该使用主库
// - 返回的是 GORM 的 DB 实例，可以直接进行数据库操作
func (d *Database) Master(ctx context.Context) *gorm.DB {
	return d.master.WithContext(ctx)
}

// Slave 获取从库连接（用于读操作）
//
// 初级工程师学习要点：
// - 所有读操作（SELECT）都应该使用从库
// - 如果没有从库，自动降级到主库
// - 使用轮询算法在多个从库之间负载均衡
func (d *Database) Slave(ctx context.Context) *gorm.DB {
	// 如果没有从库，使用主库
	if len(d.slaves) == 0 {
		return d.master.WithContext(ctx)
	}

	// 轮询选择从库
	// 使用 atomic 操作保证并发安全
	index := d.slaveIndex.Add(1) % uint32(len(d.slaves))
	return d.slaves[index].WithContext(ctx)
}

// Close 关闭数据库连接
//
// 初级工程师学习要点：
// - 应用退出时应该关闭数据库连接，释放资源
// - 需要关闭主库和所有从库
func (d *Database) Close() error {
	// 关闭主库
	if sqlDB, err := d.master.DB(); err == nil {
		sqlDB.Close()
	}

	// 关闭所有从库
	for _, slave := range d.slaves {
		if sqlDB, err := slave.DB(); err == nil {
			sqlDB.Close()
		}
	}

	return nil
}

// Name 返回数据库实例名称
func (d *Database) Name() string {
	return d.name
}

// DatabaseHealthChecker 数据库健康检查器
//
// 初级工程师学习要点：
// - 实现 health.HealthChecker 接口
// - 通过 Ping 检查数据库连接是否正常
type DatabaseHealthChecker struct {
	name string
	db   *gorm.DB
}

// Name 返回检查器名称
func (c *DatabaseHealthChecker) Name() string {
	return c.name
}

// Ping 执行轻量级健康检查（用于 Liveness）
//
// 初级工程师学习要点：
// - Ping 只检查数据库连接是否存活
// - 用于 Kubernetes Liveness Probe
func (c *DatabaseHealthChecker) Ping(ctx context.Context) error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Ping 数据库
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// Check 执行完整健康检查（用于 Readiness）
//
// 初级工程师学习要点：
// - Check 执行完整的数据库检查（Ping + SELECT 1）
// - 用于 Kubernetes Readiness Probe
func (c *DatabaseHealthChecker) Check(ctx context.Context) error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Ping 数据库
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// 执行 SELECT 1 验证数据库功能
	if err := c.db.WithContext(ctx).Exec("SELECT 1").Error; err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}

	return nil
}

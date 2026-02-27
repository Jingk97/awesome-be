package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jingpc/awesome-be/internal/config"
	"github.com/jingpc/awesome-be/internal/database"
	"github.com/jingpc/awesome-be/internal/health"
	"github.com/jingpc/awesome-be/internal/logger"
	"github.com/jingpc/awesome-be/internal/redis"
	"github.com/jingpc/awesome-be/internal/router"
	"github.com/jingpc/awesome-be/pkg/errors"
	"github.com/jingpc/awesome-be/pkg/middleware"
	"github.com/jingpc/awesome-be/pkg/response"
)

func main() {
	// ==================== 第一阶段：初始化配置 ====================
	// 配置是整个应用的基础，必须最先加载 ✅
	cfg, err := config.Load()
	if err != nil {
		// 系统启动错误，直接退出
		fmt.Fprintf(os.Stderr, "[FATAL] %v\n", errors.ErrConfigLoadFailed.WithError(err))
		os.Exit(1)
	}

	// ==================== 第二阶段：初始化日志 ====================
	// 日志模块依赖配置，用于记录应用运行状态
	appLogger, err := logger.New(cfg.Logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[FATAL] Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer appLogger.Sync() // 确保日志缓冲区刷新

	// 记录应用启动日志
	appLogger.Info("application starting", "name", cfg.App.Name, "env", cfg.App.Env, "version", "1.0.0")

	// ==================== 第三阶段：初始化健康检查管理器 ====================
	// 健康检查管理器需要在基础设施模块之前初始化
	// 这样数据库、Redis 等模块可以在初始化时自动注册健康检查
	healthMgr := health.NewManager(cfg.Health)
	appLogger.Info("health check manager initialized")

	// ==================== 第四阶段：初始化基础设施 ====================
	// 基础设施模块包括：数据库、Redis、消息队列等
	// 这些模块会自动注册到健康检查管理器

	// 4.1 初始化数据库（如果配置了）
	var dbMgr *database.Manager
	if len(cfg.Databases) > 0 {
		var err error
		dbMgr, err = database.NewManager(cfg.Databases, appLogger, healthMgr)
		if err != nil {
			appLogger.Fatal("failed to initialize database", "error", errors.ErrDBConnectFailed.WithError(err))
		}
		defer dbMgr.Close()
		appLogger.Info("database initialized", "count", len(cfg.Databases))
	}

	// 4.2 初始化 Redis（如果配置了）
	var rdb *redis.Redis
	if cfg.Redis.Mode != "" {
		var err error
		rdb, err = redis.New(cfg.Redis, healthMgr)
		if err != nil {
			appLogger.Fatal("failed to initialize redis", "error", errors.ErrRedisConnectFailed.WithError(err))
		}
		defer rdb.Close()
		appLogger.Info("redis initialized", "mode", cfg.Redis.Mode)
	}

	// ==================== 第五阶段：初始化 HTTP 服务器 ====================
	// 设置 Gin 模式（根据环境决定）
	if cfg.App.Env == "dev" {
		gin.SetMode(gin.DebugMode)
		// 在 dev 模式下，将 Gin 的输出重定向到我们的日志系统
		gin.DefaultWriter = logger.NewGinWriter(appLogger)
		gin.DefaultErrorWriter = logger.NewGinWriter(appLogger)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建 Gin 引擎（不使用默认中间件）
	engine := gin.New()

	// 注册自定义中间件（替换 Gin 默认中间件）
	engine.Use(response.Recovery(appLogger))         // Panic 恢复（统一错误响应）
	engine.Use(logger.GinLogger(appLogger))          // 请求日志
	engine.Use(middleware.CORS(cfg.Middleware.CORS)) // CORS 跨域
	// TODO: 实现其他中间件 (pkg/middleware)
	// engine.Use(middleware.RateLimit(cfg.Middleware.RateLimit))  // 限流
	// engine.Use(middleware.Trace(cfg.Middleware.Trace))  // 链路追踪

	// 注册所有路由（使用新的路由注册方式）
	router.Setup(engine, &router.RouterConfig{
		Logger: appLogger,
		DB:     dbMgr,
		Redis:  rdb,
	})

	// ==================== 第六阶段：启动 HTTP 服务器 ====================
	// 使用优雅关闭机制
	srv := startHTTPServer(engine, cfg.Server.HTTP.Port, appLogger)

	// ==================== 第七阶段：等待退出信号 ====================
	// 监听系统信号，实现优雅关闭
	waitForShutdown(srv, appLogger)

	appLogger.Info("GoFast application stopped")
}

func startHTTPServer(router *gin.Engine, port int, log *logger.Logger) *gin.Engine {
	addr := fmt.Sprintf(":%d", port)
	log.Info("HTTP server starting", "addr", addr)

	// 在 goroutine 中启动服务器
	go func() {
		if err := router.Run(addr); err != nil {
			log.Fatal("failed to start HTTP server", "error", errors.ErrServerStartFailed.WithError(err))
		}
	}()

	return router
}

func waitForShutdown(router *gin.Engine, log *logger.Logger) {
	// 创建信号通道
	quit := make(chan os.Signal, 1)

	// 监听 SIGINT 和 SIGTERM 信号
	// SIGINT: Ctrl+C
	// SIGTERM: kill 命令（Kubernetes 默认使用）
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞等待信号
	sig := <-quit
	log.Info("received shutdown signal, shutting down gracefully", "signal", sig.String())

	// 设置优雅关闭超时时间
	// 这个时间应该：
	// 1. 大于最长的请求处理时间
	// 2. 小于 Kubernetes 的 terminationGracePeriodSeconds（默认 30 秒）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// TODO: 关闭各个模块
	// 关闭顺序应该与初始化顺序相反：
	// 1. 停止接收新请求（HTTP 服务器）
	// 2. 等待正在处理的请求完成
	// 3. 关闭数据库连接
	// 4. 关闭 Redis 连接
	// 5. 刷新日志缓冲区

	// 等待上下文超时或所有资源释放完成
	<-ctx.Done()

	if ctx.Err() == context.DeadlineExceeded {
		log.Warn("shutdown timeout, forcing exit")
	} else {
		log.Info("shutdown completed")
	}
}

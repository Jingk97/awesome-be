package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jingpc/gofast/internal/config"
	"github.com/jingpc/gofast/internal/database"
	"github.com/jingpc/gofast/internal/health"
	"github.com/jingpc/gofast/internal/logger"
	"github.com/jingpc/gofast/internal/redis"
	"github.com/jingpc/gofast/internal/router"
	"github.com/jingpc/gofast/pkg/errors"
	"github.com/jingpc/gofast/pkg/middleware"
	"github.com/jingpc/gofast/pkg/response"
)

// main 是应用程序的入口点
//
// 架构思路：
// 1. 按照依赖顺序初始化各个模块（配置 -> 日志 -> 基础设施 -> 业务逻辑）
// 2. 使用优雅关闭机制，确保服务停止时正确释放资源
// 3. 统一的错误处理，区分系统启动错误和运行时错误
//
// 初级工程师学习要点：
// - 理解应用启动的生命周期
// - 掌握依赖注入的思想（先初始化被依赖的模块）
// - 学习优雅关闭的重要性（避免数据丢失）
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

// startHTTPServer 启动 HTTP 服务器
//
// 架构思路：
// - 使用 goroutine 启动服务器，避免阻塞主流程
// - 返回 *http.Server 用于后续的优雅关闭
//
// 初级工程师学习要点：
// - 理解 goroutine 的使用场景
// - 掌握 HTTP 服务器的启动方式
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

// waitForShutdown 等待退出信号并执行优雅关闭
//
// 架构思路：
// 1. 监听 SIGINT (Ctrl+C) 和 SIGTERM (kill) 信号
// 2. 收到信号后，给服务器一定时间完成正在处理的请求
// 3. 超时后强制关闭
//
// 初级工程师学习要点：
// - 理解信号处理的重要性（避免数据丢失）
// - 掌握 context.WithTimeout 的使用
// - 学习优雅关闭的最佳实践
//
// 高级工程师思考：
// - 如何处理长连接（WebSocket、SSE）？
// - 如何确保数据库事务完成？
// - 如何通知上游服务（负载均衡器）？
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

// ==================== 工程协作说明 ====================
//
// 【TODO 标记规范】
// 在团队协作中，使用 TODO 标记未完成的功能：
//
// 1. 简单 TODO：
//    // TODO: 实现用户登录功能
//
// 2. 带责任人的 TODO：
//    // TODO(张三): 实现用户登录功能
//
// 3. 带优先级的 TODO：
//    // TODO(P0): 修复数据库连接泄漏问题
//    // TODO(P1): 优化查询性能
//    // TODO(P2): 添加缓存
//
// 4. 带截止日期的 TODO：
//    // TODO(2024-03-01): 升级到 Go 1.22
//
// 【注释规范】
// 1. 包注释：每个包都应该有包注释，说明包的用途
// 2. 函数注释：公开函数必须有注释，说明功能、参数、返回值
// 3. 复杂逻辑注释：复杂的业务逻辑应该添加注释解释
// 4. 架构思路注释：关键的架构决策应该记录在注释中
//
// 【代码审查要点】
// 初级 -> 中级：
// - 代码规范：命名、格式、注释
// - 错误处理：是否正确处理所有错误
// - 资源管理：是否正确关闭文件、连接等资源
//
// 中级 -> 高级：
// - 架构设计：模块划分是否合理
// - 性能优化：是否有性能瓶颈
// - 并发安全：是否有竞态条件
//
// 高级 -> 资深：
// - 系统设计：整体架构是否可扩展
// - 可维护性：代码是否易于维护和测试
// - 最佳实践：是否遵循行业最佳实践

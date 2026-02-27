// Package health 健康检查 Handler
//
// 核心功能：
// - 提供存活探针接口
// - 提供就绪探针接口
// - 检查基础设施状态
//
// 初级工程师学习要点：
// - 理解 Kubernetes 健康检查机制
// - 掌握如何检查依赖服务状态
// - 学习如何设计健康检查接口
package health

import (
	"github.com/gin-gonic/gin"
	"github.com/jingpc/gofast/internal/database"
	"github.com/jingpc/gofast/internal/logger"
	"github.com/jingpc/gofast/internal/redis"
	"github.com/jingpc/gofast/pkg/response"
)

// Handler 健康检查处理器
type Handler struct {
	logger *logger.Logger
	db     *database.Manager
	redis  *redis.Redis
}

// NewHandler 创建健康检查处理器
func NewHandler(logger *logger.Logger, db *database.Manager, redis *redis.Redis) *Handler {
	return &Handler{
		logger: logger,
		db:     db,
		redis:  redis,
	}
}

// Liveness 存活探针
//
// 架构思路：
// - 只检查应用本身是否还在运行
// - 不检查依赖服务（数据库、Redis 等）
// - 快速返回，避免超时
//
// 初级工程师学习要点：
// - Liveness 失败会导致容器重启
// - 不应该检查外部依赖
// - 应该尽可能简单和快速
func (h *Handler) Liveness(c *gin.Context) {
	response.Success(c, gin.H{
		"status": "ok",
	})
}

// Readiness 就绪探针
//
// 架构思路：
// - 检查应用是否准备好接收流量
// - 检查依赖服务（数据库、Redis 等）
// - 可以稍微慢一些，但不应该太慢
//
// 初级工程师学习要点：
// - Readiness 失败会从负载均衡器移除
// - 应该检查所有关键依赖
// - 依赖恢复后应该自动恢复就绪状态
//
// 高级工程师思考：
// - 如何处理部分依赖失败？
// - 如何避免雪崩效应？
// - 如何设置合理的超时时间？
func (h *Handler) Readiness(c *gin.Context) {
	checks := make(map[string]string)

	// 检查数据库连接
	if h.db != nil {
		dbInstance := h.db.Get("default")
		if dbInstance != nil {
			// 使用 Master 获取主库连接进行健康检查
			sqlDB, err := dbInstance.Master(c.Request.Context()).DB()
			if err != nil || sqlDB.Ping() != nil {
				checks["database"] = "unhealthy"
			} else {
				checks["database"] = "healthy"
			}
		} else {
			checks["database"] = "not_configured"
		}
	} else {
		checks["database"] = "not_configured"
	}

	// 检查 Redis 连接
	if h.redis != nil {
		if err := h.redis.Client().Ping(c.Request.Context()).Err(); err != nil {
			checks["redis"] = "unhealthy"
		} else {
			checks["redis"] = "healthy"
		}
	} else {
		checks["redis"] = "not_configured"
	}

	// 判断整体状态
	allHealthy := true
	for _, status := range checks {
		if status == "unhealthy" {
			allHealthy = false
			break
		}
	}

	if allHealthy {
		response.Success(c, gin.H{
			"status": "ready",
			"checks": checks,
		})
	} else {
		// 返回 503 Service Unavailable
		c.JSON(503, gin.H{
			"status": "not_ready",
			"checks": checks,
		})
	}
}

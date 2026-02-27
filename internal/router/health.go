// Package router 健康检查路由
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jingpc/awesome-be/internal/handler/health"
)

// SetupHealthRoutes 设置健康检查路由
//
// 架构思路：
// - 健康检查路由独立，不在 API 版本下
// - 用于 Kubernetes 的 liveness 和 readiness 探针
// - 不需要认证
//
// 初级工程师学习要点：
// - 理解健康检查的重要性
// - 掌握 Kubernetes 探针的使用
// - 学习如何设计健康检查接口
func SetupHealthRoutes(engine *gin.Engine, cfg *RouterConfig) {
	// 创建健康检查 Handler
	handler := health.NewHandler(cfg.Logger, cfg.DB, cfg.Redis)

	// 健康检查路由组
	healthGroup := engine.Group("/health")
	{
		// 存活探针 (Liveness Probe)
		// 用于检测应用是否还在运行
		healthGroup.GET("/live", handler.Liveness)

		// 就绪探针 (Readiness Probe)
		// 用于检测应用是否准备好接收流量
		healthGroup.GET("/ready", handler.Readiness)
	}
}

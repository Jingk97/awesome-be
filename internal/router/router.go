// Package router 提供路由注册功能
//
// 核心功能：
// - 统一的路由注册入口
// - 按业务模块分组
// - 依赖注入配置
//
// 初级工程师学习要点：
// - 理解路由分组的概念
// - 掌握依赖注入的使用
// - 学习如何组织路由结构
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jingpc/gofast/internal/database"
	"github.com/jingpc/gofast/internal/logger"
	"github.com/jingpc/gofast/internal/redis"
)

// RouterConfig 路由配置
//
// 架构思路：
// - 通过配置结构体传递依赖
// - 避免全局变量
// - 便于测试和解耦
type RouterConfig struct {
	Logger *logger.Logger    // 日志管理器
	DB     *database.Manager // 数据库管理器
	Redis  *redis.Redis      // Redis 客户端
}

// Setup 设置所有路由
//
// 架构思路：
// 1. 健康检查路由独立，不在 API 版本下
// 2. 业务路由按版本分组 (/api/v1, /api/v2)
// 3. 每个业务模块独立的路由文件
//
// 初级工程师学习要点：
// - 理解路由分组的层次结构
// - 掌握如何传递配置给子路由
// - 学习版本化 API 的设计
func Setup(engine *gin.Engine, cfg *RouterConfig) {
	// 健康检查路由 (不需要认证，不在 API 版本下)
	SetupHealthRoutes(engine, cfg)

	// API v1 路由组
	v1 := engine.Group("/api/v1")
	{
		// 示例路由 (演示错误处理)
		SetupExampleRoutes(v1, cfg)

		// TODO: 其他业务路由
		// SetupUserRoutes(v1, cfg)
		// SetupOrderRoutes(v1, cfg)
	}

	// TODO: API v2 路由组 (未来版本)
	// v2 := engine.Group("/api/v2")
	// {
	//     SetupUserRoutesV2(v2, cfg)
	// }
}

// Package router 示例路由
package router

import (
	"github.com/gin-gonic/gin"
	exampleHandler "github.com/jingpc/awesome-be/internal/handler/example"
	exampleService "github.com/jingpc/awesome-be/internal/service/example"
)

// SetupExampleRoutes 设置示例路由
//
// 架构思路：
// - 演示完整的三层架构 (Handler -> Service -> Repository)
// - 演示各种错误处理场景
// - 演示依赖注入的使用
//
// 初级工程师学习要点：
// - 理解三层架构的依赖关系
// - 掌握如何初始化 Handler 和 Service
// - 学习路由分组的使用
func SetupExampleRoutes(group *gin.RouterGroup, cfg *RouterConfig) {
	// 初始化 Service (业务逻辑层)
	svc := exampleService.NewService(cfg.Logger, cfg.DB, cfg.Redis)

	// 初始化 Handler (HTTP 处理层)
	handler := exampleHandler.NewHandler(cfg.Logger, svc)

	// 示例路由组
	exampleGroup := group.Group("/examples")
	{
		// 成功响应示例
		exampleGroup.GET("/ping", handler.Ping)

		// 错误响应示例
		exampleGroup.GET("/error", handler.Error)

		// Panic 恢复示例
		exampleGroup.GET("/panic", handler.Panic)

		// 数据库错误示例
		exampleGroup.GET("/db-error", handler.DBError)

		// 404 错误示例
		exampleGroup.GET("/not-found", handler.NotFound)
	}
}

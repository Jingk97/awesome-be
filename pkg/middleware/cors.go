// Package middleware 提供 HTTP 中间件
//
// 中间件是 Web 框架的核心组件，用于处理横切关注点（Cross-Cutting Concerns）。
//
// 初级工程师学习要点：
// - 中间件是一个函数，接收 HTTP 请求，处理后传递给下一个中间件
// - 中间件可以在请求前后执行逻辑（如日志、认证、CORS）
// - 中间件的执行顺序很重要
package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/jingpc/gofast/internal/config"
)

// CORS 返回 CORS 中间件
//
// 初级工程师学习要点：
// - CORS (Cross-Origin Resource Sharing) 跨域资源共享
// - 浏览器的同源策略限制跨域请求
// - 通过 CORS 响应头允许特定的跨域请求
//
// 使用示例：
//
//	router.Use(middleware.CORS(cfg.Middleware.CORS))
//
// 架构思路：
// - 不重复造轮子，直接使用 gin-contrib/cors 官方库
// - 我们只做配置映射和封装
// - 如果未启用，返回空中间件（不影响性能）
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	// 如果未启用，返回空中间件
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// 映射配置到 gin-contrib/cors
	corsConfig := cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
		AllowWildcard:    cfg.AllowWildcard,
	}

	// 返回 gin-contrib/cors 中间件
	return cors.New(corsConfig)
}

// Package health 提供健康检查功能
//
// 健康检查模块为 GoFast 框架提供标准的 HTTP 健康检查端点，
// 支持 Kubernetes 的存活探针（Liveness Probe）和就绪探针（Readiness Probe）。
//
// 初级工程师学习要点：
// - 理解健康检查在微服务中的重要性
// - 掌握 Liveness 和 Readiness 的区别
// - 学习接口设计和自动注册模式
package health

import (
	"context"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jingpc/awesome-be/internal/config"
)

// HealthChecker 定义健康检查器接口
//
// 初级工程师学习要点：
// - 接口定义了一组方法，任何实现这些方法的类型都满足该接口
// - 这样可以让不同的组件（Database、Redis）实现统一的健康检查
type HealthChecker interface {
	// Name 返回检查器名称（使用配置中的 name 字段）
	Name() string

	// Ping 执行轻量级检查（用于 Liveness）
	// 只检查连接是否存活，不检查功能是否正常
	Ping(ctx context.Context) error

	// Check 执行完整检查（用于 Readiness）
	// 检查服务是否完全就绪，可以处理请求
	Check(ctx context.Context) error
}

// Manager 管理所有健康检查器
//
// 初级工程师学习要点：
// - Manager 使用 map 存储所有注册的健康检查器
// - 使用 sync.RWMutex 保证并发安全（多个 goroutine 可以同时访问）
type Manager struct {
	checkers map[string]HealthChecker
	mu       sync.RWMutex
	config   config.HealthConfig
}

// NewManager 创建健康检查管理器
func NewManager(cfg config.HealthConfig) *Manager {
	return &Manager{
		checkers: make(map[string]HealthChecker),
		config:   cfg,
	}
}

// Register 注册健康检查器
//
// 初级工程师学习要点：
// - 这个方法会被 Database、Redis 等模块调用，自动注册健康检查
// - 使用写锁（Lock）保证并发安全
func (m *Manager) Register(checker HealthChecker) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.checkers[checker.Name()] = checker
	return nil
}

// CheckResult 单个检查器的检查结果
type CheckResult struct {
	Status  string `json:"status"`            // "ok" 或 "error"
	Message string `json:"message,omitempty"` // 错误信息（如果有）
}

// HealthStatus 整体健康状态
type HealthStatus struct {
	Status    string                 `json:"status"`           // "ok" 或 "error"
	Timestamp string                 `json:"timestamp"`        // ISO8601 时间戳
	Checks    map[string]CheckResult `json:"checks,omitempty"` // 各组件的检查结果
}

// Ping 执行轻量级检查（用于 Liveness）
//
// 初级工程师学习要点：
// - Ping 只检查连接是否存活
// - 用于 Kubernetes Liveness Probe
// - 失败时 Kubernetes 会重启 Pod
func (m *Manager) Ping(ctx context.Context) *HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 创建带超时的 Context
	checkCtx, cancel := context.WithTimeout(ctx, m.config.Timeout)
	defer cancel()

	status := &HealthStatus{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
		Checks:    make(map[string]CheckResult),
	}

	// 如果没有注册任何检查器，直接返回 ok
	if len(m.checkers) == 0 {
		return status
	}

	// 并发执行所有 Ping 检查
	var wg sync.WaitGroup
	var mu sync.Mutex

	for name, checker := range m.checkers {
		wg.Add(1)
		go func(name string, checker HealthChecker) {
			defer wg.Done()

			// 执行 Ping 检查
			err := checker.Ping(checkCtx)

			// 记录结果
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				status.Checks[name] = CheckResult{
					Status:  "error",
					Message: err.Error(),
				}
				status.Status = "error"
			} else {
				status.Checks[name] = CheckResult{
					Status:  "ok",
					Message: "",
				}
			}
		}(name, checker)
	}

	wg.Wait()

	return status
}

// Check 执行完整检查（用于 Readiness）
//
// 初级工程师学习要点：
// - 使用读锁（RLock）允许多个 goroutine 同时读取
// - 使用 Context 控制超时，避免检查时间过长
// - 使用 WaitGroup 等待所有检查完成
func (m *Manager) Check(ctx context.Context) *HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 创建带超时的 Context
	checkCtx, cancel := context.WithTimeout(ctx, m.config.Timeout)
	defer cancel()

	status := &HealthStatus{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
		Checks:    make(map[string]CheckResult),
	}

	// 如果没有注册任何检查器，直接返回 ok
	if len(m.checkers) == 0 {
		return status
	}

	// 并发执行所有检查
	var wg sync.WaitGroup
	var mu sync.Mutex // 保护 status.Checks 的并发写入

	for name, checker := range m.checkers {
		wg.Add(1)
		go func(name string, checker HealthChecker) {
			defer wg.Done()

			// 执行检查
			err := checker.Check(checkCtx)

			// 记录结果
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				status.Checks[name] = CheckResult{
					Status:  "error",
					Message: err.Error(),
				}
				status.Status = "error" // 任何一个检查失败，整体状态为 error
			} else {
				status.Checks[name] = CheckResult{
					Status:  "ok",
					Message: "",
				}
			}
		}(name, checker)
	}

	wg.Wait()

	return status
}

// LivenessHandler 存活检查 HTTP 处理函数
//
// 初级工程师学习要点：
// - Liveness 检查使用 Ping 方法（轻量级检查）
// - 只检查连接是否存活，不检查功能是否完整
// - 如果失败，Kubernetes 会重启 Pod
func (m *Manager) LivenessHandler(c *gin.Context) {
	status := m.Ping(c.Request.Context())

	// 根据配置决定是否返回详细信息
	if !m.config.Detailed {
		// 简化模式：只返回整体状态
		c.JSON(getStatusCode(status.Status), gin.H{
			"status":    status.Status,
			"timestamp": status.Timestamp,
		})
		return
	}

	// 详细模式：返回所有组件的状态
	c.JSON(getStatusCode(status.Status), status)
}

// ReadinessHandler 就绪检查 HTTP 处理函数
//
// 初级工程师学习要点：
// - Readiness 检查会检查所有依赖服务
// - 如果失败，Kubernetes 会将 Pod 从 Service 中移除（不再接收流量）
// - 但不会重启 Pod
func (m *Manager) ReadinessHandler(c *gin.Context) {
	status := m.Check(c.Request.Context())

	// 根据配置决定是否返回详细信息
	if !m.config.Detailed {
		// 简化模式：只返回整体状态
		c.JSON(getStatusCode(status.Status), gin.H{
			"status":    status.Status,
			"timestamp": status.Timestamp,
		})
		return
	}

	// 详细模式：返回所有组件的状态
	c.JSON(getStatusCode(status.Status), status)
}

// getStatusCode 根据状态返回 HTTP 状态码
//
// 初级工程师学习要点：
// - 健康返回 200，不健康返回 503（Service Unavailable）
// - Kubernetes 根据状态码判断 Pod 是否就绪
func getStatusCode(status string) int {
	if status == "ok" {
		return 200
	}
	return 503
}

# 健康检查模块设计

## 1. 概述

健康检查模块为 GoFast 框架提供标准的 HTTP 健康检查端点，支持 Kubernetes 的存活探针（Liveness Probe）和就绪探针（Readiness Probe）。该模块能够自动识别已启用的基础中间件（MySQL、PostgreSQL、Redis 等），并提供统一的健康状态报告。

### 1.1 与后台健康检查的区别

GoFast 框架中有两种健康检查机制：

| 类型 | HTTP 健康检查端点（本模块） | 后台健康检查（Database/Redis 模块） |
|------|---------------------------|----------------------------------|
| **触发方式** | 按需触发（HTTP 请求） | 定期自动执行（如每 30 秒） |
| **用途** | Kubernetes 探针、负载均衡器 | 内部监控、日志记录、告警 |
| **响应方式** | 返回 HTTP 响应（JSON） | 记录日志 |
| **对外暴露** | 是（HTTP 端点） | 否（内部机制） |
| **配置位置** | `health` 配置段 | `databases[].health_check` 或 `redis.health_check` |

**示例**：
- **HTTP 健康检查**：Kubernetes 每 5 秒请求 `/health/ready` 端点，检查应用是否就绪
- **后台健康检查**：应用每 30 秒在后台执行 `PING` 检查 Redis 连接，记录日志

## 2. 设计目标

- **自动注册**：各基础组件在初始化时自动注册健康检查器
- **标准接口**：提供符合 Kubernetes 规范的 `/health/live` 和 `/health/ready` 端点
- **详细报告**：返回每个组件的健康状态和错误信息
- **高性能**：支持并发检查，避免阻塞
- **易扩展**：新组件可轻松集成健康检查

## 3. 核心接口

### 3.1 健康检查器接口

```go
// HealthChecker 定义健康检查器接口
type HealthChecker interface {
    // Name 返回检查器名称（如 "mysql", "redis"）
    Name() string

    // Check 执行健康检查，返回 nil 表示健康
    Check(ctx context.Context) error
}
```

### 3.2 健康检查管理器

```go
// Manager 管理所有健康检查器
type Manager struct {
    checkers map[string]HealthChecker
    mu       sync.RWMutex
}

// Register 注册健康检查器
func (m *Manager) Register(checker HealthChecker) error

// Check 执行所有健康检查
func (m *Manager) Check(ctx context.Context) *HealthStatus

// LivenessHandler 存活检查 HTTP 处理函数
func (m *Manager) LivenessHandler(c *gin.Context)

// ReadinessHandler 就绪检查 HTTP 处理函数
func (m *Manager) ReadinessHandler(c *gin.Context)
```

## 4. 健康检查类型

### 4.1 存活检查（Liveness）

**端点**：`GET /health/live`

**用途**：检查应用进程是否还在运行，如果失败 Kubernetes 会重启 Pod

**检查内容**：
- 应用进程是否响应
- 不检查依赖服务（避免级联重启）

**响应示例**：
```json
{
  "status": "ok",
  "timestamp": "2026-02-15T10:30:00Z"
}
```

### 4.2 就绪检查（Readiness）

**端点**：`GET /health/ready`

**用途**：检查应用是否准备好接收流量，如果失败 Kubernetes 会将 Pod 从 Service 中移除

**检查内容**：
- 数据库连接是否正常
- Redis 连接是否正常
- 其他依赖服务是否可用

**响应示例**：
```json
{
  "status": "ok",
  "timestamp": "2026-02-15T10:30:00Z",
  "checks": {
    "mysql": {
      "status": "ok",
      "message": ""
    },
    "redis": {
      "status": "ok",
      "message": ""
    }
  }
}
```

**失败响应示例**：
```json
{
  "status": "error",
  "timestamp": "2026-02-15T10:30:00Z",
  "checks": {
    "mysql": {
      "status": "ok",
      "message": ""
    },
    "redis": {
      "status": "error",
      "message": "dial tcp 127.0.0.1:6379: connect: connection refused"
    }
  }
}
```

## 5. 自动注册机制

### 5.1 Database 模块集成

```go
// internal/database/health.go
type DatabaseHealthChecker struct {
    db *gorm.DB
}

func (c *DatabaseHealthChecker) Name() string {
    return "database"
}

func (c *DatabaseHealthChecker) Check(ctx context.Context) error {
    sqlDB, err := c.db.DB()
    if err != nil {
        return err
    }
    return sqlDB.PingContext(ctx)
}

// 在 database 初始化时自动注册
func New(cfg *Config, healthMgr *health.Manager) (*Database, error) {
    // ... 初始化数据库连接

    // 自动注册健康检查
    if healthMgr != nil {
        checker := &DatabaseHealthChecker{db: db}
        healthMgr.Register(checker)
    }

    return &Database{db: db}, nil
}
```

### 5.2 Redis 模块集成

```go
// internal/redis/health.go
type RedisHealthChecker struct {
    client redis.UniversalClient
}

func (c *RedisHealthChecker) Name() string {
    return "redis"
}

func (c *RedisHealthChecker) Check(ctx context.Context) error {
    return c.client.Ping(ctx).Err()
}

// 在 redis 初始化时自动注册
func New(cfg *Config, healthMgr *health.Manager) (*Redis, error) {
    // ... 初始化 Redis 连接

    // 自动注册健康检查
    if healthMgr != nil {
        checker := &RedisHealthChecker{client: client}
        healthMgr.Register(checker)
    }

    return &Redis{client: client}, nil
}
```

## 6. 配置项

```yaml
# config.mini.yaml
health:
  # 健康检查超时时间
  timeout: 5s

  # 是否启用详细模式（返回每个组件的状态）
  detailed: true
```

## 7. 使用示例

### 7.1 初始化顺序

```go
// cmd/server/main.go
func main() {
    // 1. 加载配置
    cfg := config.Load()

    // 2. 初始化健康检查管理器
    healthMgr := health.NewManager()

    // 3. 初始化各组件（自动注册健康检查）
    db := database.New(cfg.Database, healthMgr)
    rdb := redis.New(cfg.Redis, healthMgr)

    // 4. 注册健康检查路由
    router := gin.Default()
    router.GET("/health/live", healthMgr.LivenessHandler)
    router.GET("/health/ready", healthMgr.ReadinessHandler)

    // 5. 启动服务器
    router.Run(":8080")
}
```

### 7.2 Kubernetes 配置

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: gofast-app
spec:
  containers:
  - name: app
    image: gofast:latest
    ports:
    - containerPort: 8080
    livenessProbe:
      httpGet:
        path: /health/live
        port: 8080
      initialDelaySeconds: 10
      periodSeconds: 10
      timeoutSeconds: 5
      failureThreshold: 3
    readinessProbe:
      httpGet:
        path: /health/ready
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 5
      timeoutSeconds: 5
      failureThreshold: 3
```

## 8. 扩展自定义检查器

如果需要添加自定义健康检查（如外部 API、消息队列等），只需实现 `HealthChecker` 接口：

```go
// 自定义检查器示例
type CustomServiceChecker struct {
    serviceURL string
}

func (c *CustomServiceChecker) Name() string {
    return "custom-service"
}

func (c *CustomServiceChecker) Check(ctx context.Context) error {
    // 实现自定义检查逻辑
    resp, err := http.Get(c.serviceURL)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("service returned status %d", resp.StatusCode)
    }
    return nil
}

// 手动注册
healthMgr.Register(&CustomServiceChecker{
    serviceURL: "http://external-service/health",
})
```

## 9. 注意事项

### 9.1 存活检查 vs 就绪检查

- **存活检查**应该非常轻量，只检查应用本身是否还在运行
- **就绪检查**可以检查依赖服务，但要设置合理的超时时间
- 避免在存活检查中检查依赖服务，否则可能导致级联重启

### 9.2 超时设置

- 健康检查应该快速返回（建议 < 5 秒）
- 使用 `context.WithTimeout` 控制每个检查的超时时间
- 并发执行多个检查以提高性能

### 9.3 错误处理

- 健康检查失败时应返回明确的错误信息
- 不要在健康检查中记录过多日志（避免日志洪水）
- 考虑添加重试机制（可选）

## 10. 实现优先级

1. **P0**：实现基础的健康检查管理器和接口定义
2. **P0**：实现存活检查和就绪检查 HTTP 处理函数
3. **P0**：Database 和 Redis 模块集成健康检查
4. **P1**：添加配置项支持
5. **P1**：添加详细模式和错误信息
6. **P2**：添加指标统计（检查次数、失败次数等）
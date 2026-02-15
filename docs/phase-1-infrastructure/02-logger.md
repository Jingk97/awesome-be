# 日志模块 (Logger)

## 概述

日志模块是 GoFast 框架的核心基础设施，基于 Uber 的 Zap 库封装，提供高性能、结构化的日志功能。

### 核心特性

- ✅ **高性能** - 基于 Zap，零内存分配
- ✅ **结构化日志** - JSON 格式，便于日志分析
- ✅ **统一格式** - 规范的日志字段和输出格式
- ✅ **Gin 日志替换** - 替换 Gin 默认日志
- ✅ **链路追踪支持** - 预留 TraceID 字段
- ✅ **Panic 捕获** - 全局 Recovery，不影响服务
- ✅ **日志分级** - 支持多级别日志输出
- ✅ **热更新** - 支持运行时动态调整日志级别

## 日志级别

### 级别说明

| 级别 | 使用场景 | 示例 |
|------|---------|------|
| **Debug** | 开发调试信息，生产环境不输出 | 函数参数、中间变量、详细流程 |
| **Info** | 正常业务流程信息 | 用户登录、订单创建、服务启动 |
| **Warn** | 警告信息，不影响主流程 | 配置缺失使用默认值、重试操作 |
| **Error** | 错误信息，影响当前请求 | 数据库查询失败、API 调用失败 |
| **Fatal** | 致命错误，服务无法继续运行 | 数据库连接失败、配置加载失败 |

### 日志记录原则

为了避免日志重复，各层的日志记录职责如下：

| 层级 | 日志记录职责 |
|------|-------------|
| **Repository 层** | 不记录日志，只返回错误 |
| **Service 层** | 只记录关键业务操作（Info），不记录错误 |
| **Handler 层** | 记录所有错误（Error） |
| **Middleware 层** | 记录请求/响应（Info）、Panic（Error） |

**示例**：

```go
// ❌ 不推荐：多层重复记录
func (r *Repository) FindByID(id int64) error {
    err := r.db.First(&user, id).Error
    if err != nil {
        logger.Error("repo: find user failed", "error", err)  // 第一次
        return err
    }
}

func (s *Service) GetUser(id int64) error {
    err := s.repo.FindByID(id)
    if err != nil {
        logger.Error("service: get user failed", "error", err)  // 第二次，重复！
        return err
    }
}

// ✅ 推荐：只在最外层记录
func (r *Repository) FindByID(id int64) error {
    err := r.db.First(&user, id).Error
    if err != nil {
        return fmt.Errorf("find user failed: %w", err)  // 只包装错误
    }
}

func (s *Service) GetUser(id int64) error {
    err := r.repo.FindByID(id)
    if err != nil {
        return err  // 继续向上传递
    }
}

func (h *Handler) GetUser(c *gin.Context) {
    err := h.service.GetUser(id)
    if err != nil {
        logger.Error("failed to get user", "error", err)  // 只在这里记录
    }
}
```

## 日志格式

### 统一字段规范

GoFast 的日志采用结构化格式，包含以下字段：

#### 核心字段（必填，始终存在）

```go
type LogEntry struct {
    Timestamp string `json:"timestamp"`  // ISO8601 时间戳
    Level     string `json:"level"`      // 日志级别（小写）
    Msg       string `json:"msg"`        // 日志消息
}
```

#### 可选字段（为空时不输出）

```go
type OptionalFields struct {
    TraceID string `json:"trace_id,omitempty"`  // 链路追踪 ID
    Caller  string `json:"caller,omitempty"`    // 调用位置
    Service string `json:"service,omitempty"`   // 服务名称
    Env     string `json:"env,omitempty"`       // 运行环境
}
```

### 什么是 `omitempty`？（初级程序员必读）

`omitempty` 是 Go 语言 JSON 标签的一个选项，用于控制 JSON 编码时的行为。

#### 基础概念

在 Go 中，当我们把结构体转换为 JSON 时，需要使用 `json` 标签来指定字段名：

```go
type User struct {
    Name string `json:"name"`  // JSON 中的字段名是 "name"
    Age  int    `json:"age"`   // JSON 中的字段名是 "age"
}
```

#### 默认行为（不使用 omitempty）

**不使用 `omitempty` 时**，所有字段都会输出到 JSON，即使值为空：

```go
type User struct {
    Name    string `json:"name"`     // 没有 omitempty
    Email   string `json:"email"`    // 没有 omitempty
    Phone   string `json:"phone"`    // 没有 omitempty
}

user := User{
    Name: "John",
    // Email 和 Phone 为空字符串
}

// 转换为 JSON
jsonData, _ := json.Marshal(user)
fmt.Println(string(jsonData))

// 输出：所有字段都存在，即使为空
// {"name":"John","email":"","phone":""}
```

#### 使用 omitempty 的行为

**使用 `omitempty` 时**，如果字段值为"零值"，该字段不会出现在 JSON 中：

```go
type User struct {
    Name    string `json:"name"`              // 没有 omitempty，始终输出
    Email   string `json:"email,omitempty"`   // 有 omitempty，为空时不输出
    Phone   string `json:"phone,omitempty"`   // 有 omitempty，为空时不输出
}

user := User{
    Name: "John",
    // Email 和 Phone 为空字符串
}

// 转换为 JSON
jsonData, _ := json.Marshal(user)
fmt.Println(string(jsonData))

// 输出：空字段不存在
// {"name":"John"}
// 注意：email 和 phone 字段完全不存在
```

#### 什么是"零值"？

Go 语言中，每种类型都有一个默认的"零值"：

| 类型 | 零值 | 示例 |
|------|------|------|
| string | `""` (空字符串) | `var s string` → `s == ""` |
| int, int64 | `0` | `var i int` → `i == 0` |
| float64 | `0.0` | `var f float64` → `f == 0.0` |
| bool | `false` | `var b bool` → `b == false` |
| pointer | `nil` | `var p *int` → `p == nil` |
| slice | `nil` | `var s []int` → `s == nil` |
| map | `nil` | `var m map[string]int` → `m == nil` |

#### 完整示例对比

```go
package main

import (
    "encoding/json"
    "fmt"
)

// 不使用 omitempty
type LogWithoutOmit struct {
    Timestamp string `json:"timestamp"`
    Level     string `json:"level"`
    Msg       string `json:"msg"`
    TraceID   string `json:"trace_id"`   // 没有 omitempty
    UserID    int64  `json:"user_id"`    // 没有 omitempty
}

// 使用 omitempty
type LogWithOmit struct {
    Timestamp string `json:"timestamp"`
    Level     string `json:"level"`
    Msg       string `json:"msg"`
    TraceID   string `json:"trace_id,omitempty"`  // 有 omitempty
    UserID    int64  `json:"user_id,omitempty"`   // 有 omitempty
}

func main() {
    // 场景1：TraceID 为空
    log1 := LogWithoutOmit{
        Timestamp: "2026-02-14T10:00:00Z",
        Level:     "info",
        Msg:       "user login",
        TraceID:   "",  // 空字符串
        UserID:    0,   // 零值
    }

    json1, _ := json.Marshal(log1)
    fmt.Println("不使用 omitempty:")
    fmt.Println(string(json1))
    // 输出：{"timestamp":"2026-02-14T10:00:00Z","level":"info","msg":"user login","trace_id":"","user_id":0}
    // 注意：trace_id 和 user_id 都存在，值为空

    log2 := LogWithOmit{
        Timestamp: "2026-02-14T10:00:00Z",
        Level:     "info",
        Msg:       "user login",
        TraceID:   "",  // 空字符串
        UserID:    0,   // 零值
    }

    json2, _ := json.Marshal(log2)
    fmt.Println("\n使用 omitempty:")
    fmt.Println(string(json2))
    // 输出：{"timestamp":"2026-02-14T10:00:00Z","level":"info","msg":"user login"}
    // 注意：trace_id 和 user_id 字段完全不存在

    // 场景2：TraceID 有值
    log3 := LogWithOmit{
        Timestamp: "2026-02-14T10:00:00Z",
        Level:     "info",
        Msg:       "user login",
        TraceID:   "abc123",  // 有值
        UserID:    456,       // 有值
    }

    json3, _ := json.Marshal(log3)
    fmt.Println("\n使用 omitempty（有值时）:")
    fmt.Println(string(json3))
    // 输出：{"timestamp":"2026-02-14T10:00:00Z","level":"info","msg":"user login","trace_id":"abc123","user_id":456}
    // 注意：有值时，字段正常输出
}
```

#### 为什么在日志中使用 omitempty？

1. **减少日志大小**：空字段不输出，节省存储空间
2. **日志更简洁**：只显示有意义的信息
3. **符合最佳实践**：业界标准做法

**示例**：

```go
// 没有 TraceID 的日志（更简洁）
{"timestamp":"2026-02-14T10:00:00Z","level":"info","msg":"user login","user_id":123}

// 有 TraceID 的日志（自动包含）
{"timestamp":"2026-02-14T10:00:00Z","level":"info","msg":"user login","trace_id":"abc123","user_id":123}
```

#### 注意事项

1. **核心字段不使用 omitempty**：
   ```go
   type LogEntry struct {
       Timestamp string `json:"timestamp"`  // 不用 omitempty，始终输出
       Level     string `json:"level"`      // 不用 omitempty，始终输出
       Msg       string `json:"msg"`        // 不用 omitempty，始终输出
   }
   ```

2. **可选字段使用 omitempty**：
   ```go
   type OptionalFields struct {
       TraceID string `json:"trace_id,omitempty"`  // 用 omitempty，为空时不输出
       UserID  int64  `json:"user_id,omitempty"`   // 用 omitempty，为 0 时不输出
   }
   ```

### 输出格式

#### JSON 格式（生产环境推荐）

```json
{
  "timestamp": "2026-02-14T10:15:23.123Z",
  "level": "info",
  "msg": "user created",
  "trace_id": "abc123",
  "user_id": 456,
  "service": "gofast",
  "env": "prod"
}
```

#### Console 格式（开发环境推荐）

```
2026-02-14T10:15:23.123Z  INFO  user created  trace_id=abc123  user_id=456  caller=service/user.go:45
```

### 不同场景的日志格式

#### 1. HTTP 访问日志

```json
{
  "timestamp": "2026-02-14T10:15:23.123Z",
  "level": "info",
  "msg": "http request",
  "trace_id": "abc123",
  "method": "GET",
  "path": "/api/v1/users/123",
  "status": 200,
  "duration": 45.2,
  "ip": "192.168.1.100",
  "user_agent": "Mozilla/5.0..."
}
```

#### 2. 错误日志

```json
{
  "timestamp": "2026-02-14T10:15:23.123Z",
  "level": "error",
  "msg": "database query failed",
  "trace_id": "abc123",
  "error": "connection timeout",
  "caller": "repository/user.go:45",
  "user_id": 123
}
```

#### 3. Panic 日志

```json
{
  "timestamp": "2026-02-14T10:15:23.123Z",
  "level": "error",
  "msg": "panic recovered",
  "trace_id": "abc123",
  "error": "runtime error: invalid memory address",
  "stack": "goroutine 1 [running]:\n...",
  "method": "POST",
  "path": "/api/v1/users"
}
```

## 配置说明

### 完整配置示例

```yaml
logger:
  # 基础配置
  level: "info"                    # 日志级别: debug, info, warn, error, fatal
  format: "json"                   # 日志格式: json, console

  # 输出配置
  outputs:
    # 标准输出（开发环境）
    - type: "stdout"
      level: "debug"
      format: "console"

    # 文件输出（生产环境）
    - type: "file"
      level: "info"
      format: "json"
      filename: "logs/app.log"
      max_size: 100              # MB
      max_backups: 10            # 保留文件数
      max_age: 30                # 保留天数
      compress: true             # 是否压缩

    # 错误日志单独输出
    - type: "file"
      level: "error"
      format: "json"
      filename: "logs/error.log"
      max_size: 100
      max_backups: 30
      max_age: 90
      compress: true

  # 功能开关
  enable_caller: true              # 是否显示调用位置
  enable_stacktrace: true          # 是否显示堆栈（Error 级别以上）

  # 固定字段
  fields:
    service: "gofast"              # 服务名称
    env: "prod"                    # 运行环境
```

### 配置项说明

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| level | string | 日志级别 | info |
| format | string | 日志格式（json/console） | json |
| outputs | array | 输出配置列表 | - |
| enable_caller | bool | 是否显示调用位置 | true |
| enable_stacktrace | bool | 是否显示堆栈 | true |
| fields | map | 固定字段 | - |

## 使用示例

### 基础使用

```go
package main

import "gofast/pkg/logger"

func main() {
    // 简单日志
    logger.Info("server started", "port", 8080)

    // 带多个字段
    logger.Info("user created",
        "user_id", 123,
        "username", "john",
        "email", "john@example.com",
    )

    // 带错误
    logger.Error("database query failed",
        "error", err,
        "query", "SELECT * FROM users",
    )
}
```

### 带 Context 使用（自动提取 TraceID）

```go
func (h *UserHandler) GetUser(c *gin.Context) {
    ctx := c.Request.Context()

    // 自动从 Context 提取 TraceID
    logger.InfoCtx(ctx, "get user request", "user_id", 123)

    user, err := h.service.GetUser(ctx, 123)
    if err != nil {
        logger.ErrorCtx(ctx, "failed to get user",
            "error", err,
            "user_id", 123,
        )
        return
    }

    logger.InfoCtx(ctx, "get user success", "user_id", user.ID)
}

// 输出（如果有 TraceID）：
// {"timestamp":"...","level":"info","msg":"get user request","trace_id":"abc123","user_id":123}

// 输出（如果没有 TraceID）：
// {"timestamp":"...","level":"info","msg":"get user request","user_id":123}
```

### 预设字段 Logger

```go
// 创建带预设字段的 Logger
requestLogger := logger.With(
    "request_id", requestID,
    "user_id", userID,
)

// 后续日志自动带上这些字段
requestLogger.Info("processing order")
// 输出: {"timestamp":"...","level":"info","msg":"processing order","request_id":"req123","user_id":456}

requestLogger.Info("order completed")
// 输出: {"timestamp":"...","level":"info","msg":"order completed","request_id":"req123","user_id":456}
```

## Gin 日志替换

### 替换 Gin 默认日志

```go
package main

import (
    "github.com/gin-gonic/gin"
    "gofast/internal/middleware"
    "gofast/pkg/logger"
)

func main() {
    // 1. 禁用 Gin 默认日志
    gin.SetMode(gin.ReleaseMode)

    // 2. 使用 gin.New() 而不是 gin.Default()
    router := gin.New()

    // 3. 使用自定义中间件
    router.Use(
        middleware.Logger(),    // 自定义日志中间件
        middleware.Recovery(),  // 自定义 Recovery 中间件
    )

    // 4. 服务启动日志
    logger.Info("HTTP server starting",
        "host", "0.0.0.0",
        "port", 8080,
    )

    router.Run(":8080")
}
```

### 自定义日志中间件

```go
// internal/middleware/logger.go
package middleware

import (
    "time"
    "github.com/gin-gonic/gin"
    "gofast/pkg/logger"
)

func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        method := c.Request.Method

        // 处理请求
        c.Next()

        // 记录日志
        duration := time.Since(start).Milliseconds()
        status := c.Writer.Status()

        logger.InfoCtx(c.Request.Context(), "http request",
            "method", method,
            "path", path,
            "status", status,
            "duration", duration,
            "ip", c.ClientIP(),
        )
    }
}
```

### 自定义 Recovery 中间件

```go
// internal/middleware/recovery.go
package middleware

import (
    "runtime/debug"
    "github.com/gin-gonic/gin"
    "gofast/pkg/logger"
    "gofast/pkg/response"
    "gofast/pkg/errors"
)

func Recovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // 获取堆栈信息
                stack := debug.Stack()

                // 记录 Panic 日志
                logger.ErrorCtx(c.Request.Context(),
                    "panic recovered",
                    "error", err,
                    "stack", string(stack),
                    "method", c.Request.Method,
                    "path", c.Request.URL.Path,
                    "ip", c.ClientIP(),
                )

                // 返回统一错误响应
                response.Error(c, errors.ErrInternalError)

                // 不中断服务
                c.Abort()
            }
        }()

        c.Next()
    }
}
```

## 链路追踪

### TraceID 处理

```go
// internal/middleware/trace.go
package middleware

import (
    "context"
    "github.com/gin-gonic/gin"
)

func Trace() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从 Header 获取 TraceID
        traceID := c.GetHeader("X-Trace-ID")

        // 如果存在，注入到 Context
        if traceID != "" {
            ctx := context.WithValue(c.Request.Context(), "trace_id", traceID)
            c.Request = c.Request.WithContext(ctx)

            // 回写到响应 Header
            c.Header("X-Trace-ID", traceID)
        }

        c.Next()
    }
}
```

### 日志自动提取 TraceID

```go
// pkg/logger/logger.go
func InfoCtx(ctx context.Context, msg string, fields ...interface{}) {
    // 从 Context 提取 TraceID
    if traceID, ok := ctx.Value("trace_id").(string); ok && traceID != "" {
        // 自动添加 trace_id 字段
        fields = append(fields, "trace_id", traceID)
    }

    // 记录日志
    zapLogger.Info(msg, toZapFields(fields)...)
}
```

## 日志轮转

使用 `lumberjack` 进行日志轮转：

```yaml
logger:
  outputs:
    - type: "file"
      filename: "logs/app.log"
      max_size: 100        # MB，单个文件最大大小
      max_backups: 10      # 保留旧文件的最大个数
      max_age: 30          # 保留旧文件的最大天数
      compress: true       # 是否压缩旧文件（gzip）
```

**轮转策略**：
- 按大小轮转：文件达到 100MB 自动创建新文件
- 按时间清理：超过 30 天的日志自动删除
- 压缩归档：旧日志自动压缩节省空间

**文件命名**：
```
logs/
├── app.log              # 当前日志
├── app-2026-02-13.log   # 昨天的日志
├── app-2026-02-12.log.gz # 前天的日志（已压缩）
└── app-2026-02-11.log.gz
```

## 最佳实践

### 1. 日志内容规范

```go
// ✅ 推荐：结构化日志
logger.Info("user login",
    "user_id", 123,
    "username", "john",
    "ip", "192.168.1.100",
    "duration", 45.2,
)

// ❌ 不推荐：字符串拼接
logger.Info(fmt.Sprintf("user %d login from %s", 123, "192.168.1.100"))
```

### 2. 敏感信息脱敏

```go
// ✅ 推荐：脱敏敏感信息
logger.Info("user created",
    "username", "john",
    "password", "[REDACTED]",  // 脱敏
    "email", maskEmail("john@example.com"),  // john***@example.com
)

// ❌ 不推荐：记录敏感信息
logger.Info("user created",
    "username", "john",
    "password", "my-password",  // 不要记录密码！
)
```

### 3. 错误日志包含上下文

```go
// ✅ 推荐：包含足够的上下文
logger.Error("failed to create order",
    "user_id", 123,
    "product_id", 456,
    "quantity", 2,
    "error", err,
)

// ❌ 不推荐：缺少上下文
logger.Error("create order failed", "error", err)
```

### 4. 避免日志泄露

```go
// ❌ 不要记录完整的请求体（可能包含敏感信息）
logger.Debug("request body", "body", string(body))

// ✅ 只记录必要的字段
logger.Debug("request received",
    "method", c.Request.Method,
    "path", c.Request.URL.Path,
    "content_length", c.Request.ContentLength,
)
```

### 5. 合理使用日志级别

```go
// Debug - 开发调试
logger.Debug("processing user data", "user_id", 123, "step", "validation")

// Info - 正常业务流程
logger.Info("user login success", "user_id", 123)

// Warn - 警告但不影响主流程
logger.Warn("cache miss, using database", "key", "user:123")

// Error - 错误，影响当前请求
logger.Error("failed to send email", "error", err, "user_id", 123)

// Fatal - 致命错误，服务无法继续
logger.Fatal("failed to connect database", "error", err)
```

## 常见问题

### Q1: 为什么要替换 Gin 的默认日志？

**A**: Gin 的默认日志：
- 格式不统一，难以解析
- 无法集成到日志系统
- 缺少结构化字段
- 无法控制日志级别

使用自定义日志后：
- 统一的 JSON 格式
- 便于日志收集和分析
- 支持链路追踪
- 灵活的日志级别控制

### Q2: 如何在生产环境减少日志量？

**A**:
1. 设置合适的日志级别（info 或 warn）
2. 避免在循环中记录日志
3. 使用采样（高频日志只记录部分）
4. 定期清理旧日志

```yaml
logger:
  level: "info"  # 生产环境使用 info
  outputs:
    - type: "file"
      max_age: 7  # 只保留 7 天
```

### Q3: 如何调试日志问题？

**A**: 启用调试模式：

```yaml
logger:
  level: "debug"
  format: "console"  # 使用 console 格式更易读
  enable_caller: true
```

### Q4: 日志文件太大怎么办？

**A**: 调整轮转配置：

```yaml
logger:
  outputs:
    - type: "file"
      max_size: 50      # 减小单文件大小
      max_backups: 5    # 减少保留文件数
      max_age: 7        # 减少保留天数
      compress: true    # 启用压缩
```

### Q5: 如何查看特定 TraceID 的所有日志？

**A**: 使用 `grep` 或日志分析工具：

```bash
# 查看特定 TraceID 的日志
grep "abc123" logs/app.log

# 使用 jq 解析 JSON 日志
cat logs/app.log | jq 'select(.trace_id=="abc123")'
```

## 下一步

- 📖 阅读 [数据库模块文档](./03-database.md)
- 📖 阅读 [Redis 模块文档](./04-redis.md)
- 💻 查看 [完整示例代码](../examples/logger-example.md)

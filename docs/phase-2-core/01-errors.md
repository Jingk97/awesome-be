# 错误处理模块 (Errors)

## 概述

错误处理模块是 GoFast 框架的核心功能，提供统一的错误定义、错误码管理和错误响应机制。

### 核心特性

- ✅ **统一错误码** - 清晰的错误码分类体系
- ✅ **错误和日志解耦** - 错误传递不依赖日志系统
- ✅ **错误包装** - 支持错误链和上下文信息
- ✅ **系统错误处理** - 统一处理启动和运行时错误
- ✅ **统一响应格式** - 标准的 JSON 错误响应
- ✅ **国际化支持** - 错误消息多语言支持（预留）

## 错误码分类体系

GoFast 采用分层的错误码体系：

```
错误码范围：
├── 0         - 成功
├── 1xxx      - 系统错误（启动、配置、依赖）
├── 4xxx      - 客户端错误（参数、认证、权限）
└── 5xxx      - 服务器错误（内部、数据库、第三方）
```

### 错误码到 HTTP 状态码的映射

| 错误码范围 | HTTP 状态码 | 说明 |
|-----------|------------|------|
| 0 | 200 | 成功 |
| 1xxx | 500 | 系统错误 |
| 400x | 400 | 参数错误 |
| 401x | 401 | 认证错误 |
| 403x | 403 | 权限错误 |
| 404x | 404 | 资源不存在 |
| 409x | 409 | 冲突错误 |
| 429x | 429 | 限流错误 |
| 5xxx | 500 | 服务器错误 |

## 错误码定义

### 系统错误（1xxx）

系统错误用于应用启动和运行时的系统级错误。

#### 配置错误（10xx）

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 1001 | CodeConfigLoadFailed | 配置加载失败 |
| 1002 | CodeConfigParseFailed | 配置解析失败 |
| 1003 | CodeConfigValidateFailed | 配置验证失败 |

#### 数据库错误（11xx）

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 1101 | CodeDBConnectFailed | 数据库连接失败 |
| 1102 | CodeDBPingFailed | 数据库 PING 失败 |
| 1103 | CodeDBMigrateFailed | 数据库迁移失败 |

#### Redis 错误（12xx）

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 1201 | CodeRedisConnectFailed | Redis 连接失败 |
| 1202 | CodeRedisPingFailed | Redis PING 失败 |

#### 依赖服务错误（13xx）

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 1301 | CodeServiceUnavailable | 依赖服务不可用 |
| 1302 | CodeServiceTimeout | 依赖服务超时 |

#### 启动错误（14xx）

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 1401 | CodePortBindFailed | 端口绑定失败 |
| 1402 | CodeServerStartFailed | 服务启动失败 |

### 客户端错误（4xxx）

客户端错误用于请求参数、认证、权限等客户端相关错误。

#### 参数错误（400x）

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 4001 | CodeInvalidParams | 参数错误 |
| 4002 | CodeMissingParams | 缺少参数 |
| 4003 | CodeInvalidFormat | 格式错误 |

#### 认证错误（401x）

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 4011 | CodeUnauthorized | 未认证 |
| 4012 | CodeTokenExpired | Token 过期 |
| 4013 | CodeTokenInvalid | Token 无效 |

#### 权限错误（403x）

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 4031 | CodeForbidden | 无权限 |
| 4032 | CodeAccessDenied | 访问被拒绝 |

#### 资源错误（404x）

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 4041 | CodeNotFound | 资源不存在 |
| 4042 | CodeUserNotFound | 用户不存在 |
| 4043 | CodeOrderNotFound | 订单不存在 |

#### 冲突错误（409x）

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 4091 | CodeConflict | 资源冲突 |
| 4092 | CodeDuplicate | 资源重复 |

#### 限流错误（429x）

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 4291 | CodeTooManyRequests | 请求过多 |
| 4292 | CodeRateLimitExceeded | 超过限流 |

### 服务器错误（5xxx）

服务器错误用于内部错误、数据库错误、第三方服务错误等。

#### 内部错误（500x）

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 5001 | CodeInternalError | 内部错误 |
| 5002 | CodePanic | Panic 错误 |

#### 数据库错误（501x）

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 5011 | CodeDBError | 数据库错误 |
| 5012 | CodeDBQueryFailed | 查询失败 |
| 5013 | CodeDBTxFailed | 事务失败 |

#### 缓存错误（502x）

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 5021 | CodeCacheError | 缓存错误 |
| 5022 | CodeCacheGetFailed | 缓存获取失败 |
| 5023 | CodeCacheSetFailed | 缓存设置失败 |

#### RPC 错误（503x）

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 5031 | CodeRPCError | RPC 调用错误 |
| 5032 | CodeRPCTimeout | RPC 超时 |

#### 第三方服务错误（504x）

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 5041 | CodeThirdPartyError | 第三方服务错误 |
| 5042 | CodePaymentFailed | 支付失败 |
| 5043 | CodeSMSFailed | 短信发送失败 |

## 错误处理原则

### 核心原则：错误和日志解耦

**错误传递不应该依赖日志系统**

```
┌─────────────────────────────────────────────────────┐
│ Handler 层                                           │
│ - 捕获所有错误                                       │
│ - 记录错误日志（唯一记录日志的地方）                │
│ - 转换为 HTTP 响应                                   │
└─────────────────────────────────────────────────────┘
                        ↑
                    返回错误
                        │
┌─────────────────────────────────────────────────────┐
│ Service 层                                           │
│ - 包装错误，添加业务上下文                          │
│ - 不记录日志                                         │
│ - 向上传递错误                                       │
└─────────────────────────────────────────────────────┘
                        ↑
                    返回错误
                        │
┌─────────────────────────────────────────────────────┐
│ Repository 层                                        │
│ - 返回原始错误                                       │
│ - 不记录日志                                         │
│ - 不包装错误                                         │
└─────────────────────────────────────────────────────┘
```

### 各层职责

| 层级 | 错误处理职责 | 是否记录日志 |
|------|-------------|-------------|
| Repository | 返回原始错误 | ❌ 否 |
| Service | 包装错误，添加上下文 | ❌ 否 |
| Handler | 记录日志，返回响应 | ✅ 是 |
| Middleware | 捕获 Panic，记录日志 | ✅ 是 |

## 使用示例

### 系统启动错误处理

```go
// cmd/http/main.go
package main

import (
    "context"
    "fmt"
    "os"
    "gofast/pkg/config"
    "gofast/pkg/database"
    "gofast/pkg/logger"
    "gofast/pkg/errors"
)

func main() {
    // 1. 加载配置
    cfg, err := config.Load("./config/config.yaml")
    if err != nil {
        // 系统启动错误，直接退出
        fmt.Fprintf(os.Stderr, "[FATAL] %v\n",
            errors.ErrConfigLoadFailed.WithError(err))
        os.Exit(1)
    }

    // 2. 初始化日志
    if err := logger.Init(cfg.Logger); err != nil {
        fmt.Fprintf(os.Stderr, "[FATAL] Failed to initialize logger: %v\n", err)
        os.Exit(1)
    }

    // 3. 初始化数据库
    db, err := database.New(cfg.Database)
    if err != nil {
        logger.Fatal("failed to initialize database",
            "error", errors.ErrDBConnectFailed.WithError(err),
        )
    }

    // 4. 测试数据库连接
    if err := db.Ping(context.Background()); err != nil {
        logger.Fatal("failed to ping database",
            "error", errors.ErrDBPingFailed.WithError(err),
        )
    }

    // 5. 初始化 Redis
    cache, err := cache.New(cfg.Redis)
    if err != nil {
        logger.Fatal("failed to initialize redis",
            "error", errors.ErrRedisConnectFailed.WithError(err),
        )
    }

    // 6. 启动 HTTP 服务
    addr := fmt.Sprintf("%s:%d", cfg.Server.HTTP.Host, cfg.Server.HTTP.Port)
    logger.Info("starting HTTP server", "addr", addr)

    if err := router.Run(addr); err != nil {
        logger.Fatal("failed to start HTTP server",
            "error", errors.ErrServerStartFailed.WithError(err),
            "addr", addr,
        )
    }
}
```

### Repository 层错误处理

```go
// internal/repository/user_repository.go
package repository

import (
    "context"
    "gofast/internal/model"
    "gofast/pkg/database"
)

type UserRepository struct {
    db database.Database
}

// FindByID 根据 ID 查询用户
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
    var user model.User
    err := r.db.Slave(ctx).First(&user, id).Error
    if err != nil {
        // 只返回错误，不记录日志
        return nil, err
    }
    return &user, nil
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
    // 只返回错误，不记录日志
    return r.db.Master(ctx).Create(user).Error
}
```

### Service 层错误处理

```go
// internal/service/user_service.go
package service

import (
    "context"
    "fmt"
    "gofast/internal/model"
    "gofast/internal/repository"
    "gofast/pkg/errors"
)

type UserService struct {
    userRepo repository.UserRepository
}

// GetUser 获取用户
func (s *UserService) GetUser(ctx context.Context, id int64) (*model.User, error) {
    user, err := s.userRepo.FindByID(ctx, id)
    if err != nil {
        // 包装错误，添加业务上下文
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.ErrUserNotFound.WithError(err)
        }
        return nil, fmt.Errorf("failed to get user %d: %w", id, err)
    }

    // 业务逻辑...

    return user, nil
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*model.User, error) {
    // 检查用户名是否存在
    exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
    if err != nil {
        return nil, fmt.Errorf("check username exists: %w", err)
    }

    if exists {
        // 返回业务错误
        return nil, errors.ErrDuplicate.WithDetail("username already exists")
    }

    // 创建用户
    user := &model.User{
        Username: req.Username,
        Email:    req.Email,
    }

    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, fmt.Errorf("create user: %w", err)
    }

    return user, nil
}
```

### Handler 层错误处理

```go
// internal/handler/http/user_handler.go
package http

import (
    "strconv"
    "github.com/gin-gonic/gin"
    "gofast/internal/service"
    "gofast/pkg/errors"
    "gofast/pkg/logger"
    "gofast/pkg/response"
)

type UserHandler struct {
    userService *service.UserService
}

// GetUser 获取用户
func (h *UserHandler) GetUser(c *gin.Context) {
    // 1. 参数验证
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        response.Error(c, errors.ErrInvalidParams.WithDetail("invalid user id"))
        return
    }

    // 2. 调用 Service
    user, err := h.userService.GetUser(c.Request.Context(), id)
    if err != nil {
        // 记录错误日志（唯一记录日志的地方）
        logger.ErrorCtx(c.Request.Context(), "failed to get user",
            "error", err,
            "user_id", id,
            "path", c.Request.URL.Path,
            "method", c.Request.Method,
        )

        // 转换为 HTTP 响应
        response.Error(c, errors.FromError(err))
        return
    }

    // 3. 返回成功响应
    response.Success(c, user)
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
    // 1. 参数绑定
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, errors.ErrInvalidParams.WithDetail(err.Error()))
        return
    }

    // 2. 参数验证
    if err := req.Validate(); err != nil {
        response.Error(c, errors.ErrInvalidParams.WithDetail(err.Error()))
        return
    }

    // 3. 调用 Service
    user, err := h.userService.CreateUser(c.Request.Context(), &req)
    if err != nil {
        // 记录错误日志
        logger.ErrorCtx(c.Request.Context(), "failed to create user",
            "error", err,
            "username", req.Username,
        )

        response.Error(c, errors.FromError(err))
        return
    }

    // 4. 返回成功响应
    response.Success(c, user)
}
```

### Middleware 错误处理

```go
// internal/middleware/recovery.go
package middleware

import (
    "runtime/debug"
    "github.com/gin-gonic/gin"
    "gofast/pkg/errors"
    "gofast/pkg/logger"
    "gofast/pkg/response"
)

// Recovery Panic 恢复中间件
func Recovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // 获取堆栈信息
                stack := string(debug.Stack())

                // 记录 Panic 日志
                logger.ErrorCtx(c.Request.Context(), "panic recovered",
                    "error", err,
                    "stack", stack,
                    "method", c.Request.Method,
                    "path", c.Request.URL.Path,
                    "ip", c.ClientIP(),
                )

                // 返回统一错误响应
                response.Error(c, errors.ErrPanic.WithDetail("internal server error"))

                // 中断请求
                c.Abort()
            }
        }()

        c.Next()
    }
}
```

## 错误转换

### 标准错误转换

```go
// pkg/errors/convert.go
package errors

import (
    "errors"
    "gorm.io/gorm"
)

// FromError 从标准错误转换为业务错误
func FromError(err error) *Error {
    if err == nil {
        return nil
    }

    // 如果已经是 Error 类型，直接返回
    var e *Error
    if errors.As(err, &e) {
        return e
    }

    // GORM 错误转换
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return ErrNotFound.WithError(err)
    }
    if errors.Is(err, gorm.ErrDuplicatedKey) {
        return ErrDuplicate.WithError(err)
    }

    // Redis 错误转换
    if isRedisError(err) {
        return ErrCacheError.WithError(err)
    }

    // 默认返回内部错误
    return ErrInternalError.WithError(err)
}

func isRedisError(err error) bool {
    // 检查是否是 Redis 错误
    return strings.Contains(err.Error(), "redis")
}
```

## 统一响应格式

### 响应结构

```go
// pkg/response/response.go
package response

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "gofast/pkg/errors"
)

// Response 统一响应结构
type Response struct {
    Code    int         `json:"code"`              // 业务错误码
    Message string      `json:"message"`           // 错误消息
    Data    interface{} `json:"data,omitempty"`    // 数据（成功时）
    TraceID string      `json:"trace_id,omitempty"` // 链路追踪 ID
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Response{
        Code:    0,
        Message: "success",
        Data:    data,
        TraceID: getTraceID(c),
    })
}

// Error 错误响应
func Error(c *gin.Context, err *errors.Error) {
    c.JSON(err.Code.HTTPStatus(), Response{
        Code:    int(err.Code),
        Message: err.Message,
        TraceID: getTraceID(c),
    })
}

// getTraceID 从 Context 获取 TraceID
func getTraceID(c *gin.Context) string {
    if traceID, ok := c.Request.Context().Value("trace_id").(string); ok {
        return traceID
    }
    return ""
}
```

### 响应示例

**成功响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 123,
    "username": "john",
    "email": "john@example.com"
  },
  "trace_id": "abc123def456"
}
```

**错误响应**：
```json
{
  "code": 4042,
  "message": "User not found",
  "trace_id": "abc123def456"
}
```

**参数错误响应**：
```json
{
  "code": 4001,
  "message": "Invalid parameters",
  "trace_id": "abc123def456"
}
```

## 错误包装

### 使用 fmt.Errorf 包装错误

```go
// Service 层包装错误
func (s *UserService) GetUser(ctx context.Context, id int64) (*User, error) {
    user, err := s.userRepo.FindByID(ctx, id)
    if err != nil {
        // 使用 %w 包装错误，保留错误链
        return nil, fmt.Errorf("failed to get user %d: %w", id, err)
    }
    return user, nil
}

// 错误链示例：
// failed to get user 123: record not found
```

### 使用 errors.Is 和 errors.As

```go
// 检查错误类型
if errors.Is(err, gorm.ErrRecordNotFound) {
    return errors.ErrNotFound
}

// 提取错误信息
var e *errors.Error
if errors.As(err, &e) {
    fmt.Printf("Error code: %d\n", e.Code)
}
```

## 最佳实践

### 1. 错误传递不记录日志

```go
// ✅ 推荐：Repository 和 Service 层不记录日志
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*User, error) {
    var user User
    err := r.db.First(&user, id).Error
    if err != nil {
        return nil, err  // 只返回错误
    }
    return &user, nil
}

// ❌ 不推荐：在底层记录日志
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*User, error) {
    var user User
    err := r.db.First(&user, id).Error
    if err != nil {
        logger.Error("database error", "error", err)  // 不要这样做
        return nil, err
    }
    return &user, nil
}
```

### 2. 使用预定义错误

```go
// ✅ 推荐：使用预定义错误
if user == nil {
    return errors.ErrUserNotFound
}

// ❌ 不推荐：每次创建新错误
if user == nil {
    return &errors.Error{
        Code:    4042,
        Message: "User not found",
    }
}
```

### 3. 添加错误上下文

```go
// ✅ 推荐：添加业务上下文
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) error {
    if err := s.userRepo.Create(ctx, user); err != nil {
        return fmt.Errorf("create user %s: %w", req.Username, err)
    }
    return nil
}

// ❌ 不推荐：直接返回原始错误
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) error {
    return s.userRepo.Create(ctx, user)
}
```

### 4. 系统错误立即退出

```go
// ✅ 推荐：系统启动错误立即退出
func main() {
    cfg, err := config.Load("config.yaml")
    if err != nil {
        fmt.Fprintf(os.Stderr, "[FATAL] %v\n", errors.ErrConfigLoadFailed.WithError(err))
        os.Exit(1)
    }
}

// ❌ 不推荐：系统错误继续运行
func main() {
    cfg, err := config.Load("config.yaml")
    if err != nil {
        logger.Error("config load failed", "error", err)
        // 继续运行，可能导致更多错误
    }
}
```

### 5. 错误响应包含 TraceID

```go
// ✅ 推荐：错误响应包含 TraceID
func Error(c *gin.Context, err *errors.Error) {
    c.JSON(err.Code.HTTPStatus(), Response{
        Code:    int(err.Code),
        Message: err.Message,
        TraceID: getTraceID(c),  // 包含 TraceID
    })
}
```

## 常见问题

### Q1: 为什么错误和日志要解耦？

**A**:
1. **避免日志重复**：如果每层都记录，会有多次重复日志
2. **职责清晰**：Repository 只负责数据访问，不负责日志
3. **易于测试**：不需要 mock logger
4. **统一管理**：在 Handler 层统一记录，便于添加上下文信息

### Q2: 什么时候使用系统错误？

**A**:
- 应用启动时的错误（配置加载、数据库连接）
- 依赖服务不可用
- 端口绑定失败
- 任何导致应用无法正常运行的错误

### Q3: 如何自定义错误码？

**A**:
在对应的错误码文件中添加：

```go
// pkg/errors/client.go
const (
    // 添加自定义错误码
    CodeCustomError Code = 4099  // 自定义错误
)

var (
    ErrCustomError = &Error{
        Code:    CodeCustomError,
        Message: "Custom error message",
    }
)
```

### Q4: 如何处理第三方库的错误？

**A**:
使用 `FromError` 函数转换：

```go
// 自动转换 GORM 错误
err := db.First(&user, id).Error
if err != nil {
    return errors.FromError(err)  // 自动转换为 ErrNotFound
}
```

### Q5: 错误消息如何国际化？

**A**:
预留国际化接口：

```go
// 配置文件
errors:
  i18n:
    enabled: true
    default_lang: "zh-CN"

// 使用
err := errors.ErrUserNotFound.WithLang("en-US")
```

## 下一步

- 📖 阅读 [事务管理文档](./02-transaction.md)
- 📖 阅读 [JWT 认证文档](./03-jwt.md)
- 💻 查看 [完整示例代码](../examples/errors-example.md)
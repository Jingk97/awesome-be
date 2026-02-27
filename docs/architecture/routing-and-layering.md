# GoFast 路由与业务分层架构设计

## 文档概述

本文档汇总了 GoFast 框架中路由管理和业务分层的架构设计方案，包括路由分组策略、依赖注入模式、以及三层架构的实现细节。

## 一、路由架构设计

### 1.1 路由分组策略

GoFast 采用**模块化路由分组**策略，将路由按业务模块和版本进行组织。

#### 路由层次结构

```
/
├── /health                    # 健康检查路由（独立，不在API版本下）
│   ├── /live                 # 存活探针
│   └── /ready                # 就绪探针
│
└── /api
    └── /v1                   # API v1 版本
        ├── /examples         # 示例模块
        │   ├── /ping
        │   ├── /error
        │   ├── /panic
        │   ├── /db-error
        │   └── /not-found
        │
        └── /users            # 用户模块（待实现）
            ├── GET    /
            ├── POST   /
            ├── GET    /:id
            ├── PUT    /:id
            └── DELETE /:id
```

#### 设计原则

1. **健康检查独立**：健康检查路由不在 API 版本下，便于 Kubernetes 等容器编排工具访问
2. **版本化管理**：业务路由按版本分组（/api/v1, /api/v2），便于 API 演进
3. **模块化组织**：每个业务模块独立的路由文件，便于维护和扩展
4. **统一前缀**：所有业务 API 使用 `/api` 前缀，便于网关路由和权限控制

### 1.2 路由配置结构

#### RouterConfig 设计

```go
// RouterConfig 路由配置
type RouterConfig struct {
    Logger *logger.Logger    // 日志管理器
    DB     *database.Manager // 数据库管理器
    Redis  *redis.Redis      // Redis 客户端
}
```

**设计思路**：
- 通过配置结构体传递依赖，避免全局变量
- 便于单元测试（可以注入 Mock 对象）
- 支持依赖注入，降低耦合

### 1.3 路由注册流程

#### 主路由注册器

```go
// internal/router/router.go
func Setup(engine *gin.Engine, cfg *RouterConfig) {
    // 1. 健康检查路由（不需要认证）
    SetupHealthRoutes(engine, cfg)

    // 2. API v1 路由组
    v1 := engine.Group("/api/v1")
    {
        SetupExampleRoutes(v1, cfg)
        // SetupUserRoutes(v1, cfg)
        // SetupOrderRoutes(v1, cfg)
    }

    // 3. API v2 路由组（未来版本）
    // v2 := engine.Group("/api/v2")
}
```

#### 模块路由注册器

```go
// internal/router/example.go
func SetupExampleRoutes(group *gin.RouterGroup, cfg *RouterConfig) {
    // 1. 初始化 Service（业务逻辑层）
    svc := exampleService.NewService(cfg.Logger, cfg.DB, cfg.Redis)

    // 2. 初始化 Handler（HTTP 处理层）
    handler := exampleHandler.NewHandler(cfg.Logger, svc)

    // 3. 注册路由
    exampleGroup := group.Group("/examples")
    {
        exampleGroup.GET("/ping", handler.Ping)
        exampleGroup.GET("/error", handler.Error)
        exampleGroup.GET("/panic", handler.Panic)
        exampleGroup.GET("/db-error", handler.DBError)
        exampleGroup.GET("/not-found", handler.NotFound)
    }
}
```

**关键特性**：
- 每个模块独立的路由注册函数
- 在路由注册时完成依赖注入
- 支持路由分组和嵌套

### 1.4 路由扩展指南

#### 添加新业务模块的步骤

1. **创建 Handler**：`internal/handler/user/user.go`
2. **创建 Service**：`internal/service/user/user.go`
3. **创建路由文件**：`internal/router/user.go`
4. **注册路由**：在 `router.go` 中调用 `SetupUserRoutes(v1, cfg)`

示例：

```go
// internal/router/user.go
func SetupUserRoutes(group *gin.RouterGroup, cfg *RouterConfig) {
    // 初始化依赖
    userRepo := repository.NewUserRepository(cfg.DB, cfg.Redis)
    userSvc := userService.NewService(cfg.Logger, userRepo)
    handler := userHandler.NewHandler(cfg.Logger, userSvc)

    // 注册路由
    userGroup := group.Group("/users")
    {
        userGroup.GET("", handler.List)           // 列表
        userGroup.POST("", handler.Create)        // 创建
        userGroup.GET("/:id", handler.GetByID)    // 详情
        userGroup.PUT("/:id", handler.Update)     // 更新
        userGroup.DELETE("/:id", handler.Delete)  // 删除
    }
}
```

## 二、业务分层架构

### 2.1 三层架构概述

GoFast 采用经典的**三层架构**（Three-tier Architecture）：

```
┌─────────────────────────────────────────┐
│         Handler 层（HTTP 处理层）        │
│   - 参数验证                             │
│   - 调用 Service                         │
│   - 构造响应                             │
│   - 记录日志                             │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│         Service 层（业务逻辑层）         │
│   - 业务逻辑                             │
│   - 事务管理                             │
│   - 调用 Repository                      │
│   - 返回错误                             │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│       Repository 层（数据访问层）        │
│   - 数据库操作                           │
│   - 缓存管理                             │
│   - 读写分离                             │
└─────────────────────────────────────────┘
```

### 2.2 Handler 层设计

#### 职责定义

- ✅ 接收和验证 HTTP 请求参数
- ✅ 调用 Service 层处理业务逻辑
- ✅ 构造和返回 HTTP 响应
- ✅ 记录请求日志
- ❌ **不包含业务逻辑**
- ❌ **不直接访问数据库**

#### 实现示例

```go
// internal/handler/example/example.go
type Handler struct {
    logger  *logger.Logger
    service *example.Service
}

func NewHandler(logger *logger.Logger, service *example.Service) *Handler {
    return &Handler{
        logger:  logger,
        service: service,
    }
}

// DBError 数据库错误示例
func (h *Handler) DBError(c *gin.Context) {
    // 1. 调用 Service 层
    err := h.service.GetUser(c.Request.Context(), 999)

    // 2. Handler 层记录日志
    if err != nil {
        h.logger.Error("failed to get user",
            zap.Error(err),
            zap.Int("user_id", 999),
            zap.String("path", c.Request.URL.Path),
        )
        response.Error(c, err)
        return
    }

    // 3. 返回成功响应
    response.Success(c, gin.H{"message": "user found"})
}
```

#### 设计要点

1. **依赖注入**：通过构造函数注入 Logger 和 Service
2. **日志记录**：Handler 层负责记录业务日志
3. **错误处理**：使用统一的 `response.Error()` 处理错误
4. **上下文传递**：使用 `c.Request.Context()` 传递请求上下文

### 2.3 Service 层设计

#### 职责定义

- ✅ 实现所有业务逻辑
- ✅ 调用 Repository 层进行数据操作
- ✅ 返回错误给 Handler 层
- ✅ 管理事务（如果需要）
- ❌ **不记录日志**（由 Handler 层记录）
- ❌ **不直接操作数据库**（通过 Repository）

#### 实现示例

```go
// internal/service/example/example.go
type Service struct {
    logger *logger.Logger
    db     *database.Manager
    redis  *redis.Redis
}

func NewService(logger *logger.Logger, db *database.Manager, redis *redis.Redis) *Service {
    return &Service{
        logger: logger,
        db:     db,
        redis:  redis,
    }
}

// GetUser 获取用户（演示数据库错误）
func (s *Service) GetUser(ctx context.Context, id int) error {
    // 1. 获取数据库连接
    dbInstance := s.db.Get("default")
    if dbInstance == nil {
        return errors.ErrDBError.WithDetail("default database not found")
    }

    // 2. 定义数据结构
    var user struct {
        ID   int
        Name string
    }

    // 3. 查询数据（读操作使用从库）
    err := dbInstance.Slave(ctx).
        Table("users").
        Where("id = ?", id).
        First(&user).Error

    return err
}
```

#### 设计要点

1. **业务逻辑集中**：所有业务规则都在 Service 层实现
2. **不记录日志**：Service 只返回错误，日志由 Handler 记录
3. **读写分离**：读操作使用 `Slave()`，写操作使用 `Master()`
4. **错误转换**：GORM 错误自动转换为业务错误

### 2.4 Repository 层设计

#### 职责定义（未来实现）

- ✅ 封装数据库操作
- ✅ 实现读写分离
- ✅ 管理缓存
- ✅ 数据转换
- ❌ **不包含业务逻辑**

#### 设计示例

```go
// internal/repository/user_repository.go (待实现)
type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    FindByID(ctx context.Context, id int64) (*model.User, error)
    Update(ctx context.Context, user *model.User) error
    Delete(ctx context.Context, id int64) error
}

type userRepository struct {
    db    *database.Manager
    cache *redis.Redis
}

func (r *userRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
    // 1. 查询缓存
    cacheKey := fmt.Sprintf("user:%d", id)
    var user model.User
    if err := r.cache.Get(ctx, cacheKey, &user); err == nil {
        return &user, nil
    }

    // 2. 查询数据库（从库）
    err := r.db.Slave(ctx).First(&user, id).Error
    if err != nil {
        return nil, err
    }

    // 3. 写入缓存
    r.cache.Set(ctx, cacheKey, &user, 5*time.Minute)

    return &user, nil
}
```

## 三、依赖注入模式

### 3.1 依赖注入流程

```
main.go
  ↓ 初始化基础设施
  ├── Logger
  ├── Database
  └── Redis
  ↓ 创建 RouterConfig
  ↓ 调用 router.Setup()
  ↓
router.Setup()
  ↓ 调用各模块路由注册
  ↓
SetupExampleRoutes()
  ↓ 初始化 Service
  ├── exampleService.NewService(logger, db, redis)
  ↓ 初始化 Handler
  ├── exampleHandler.NewHandler(logger, service)
  ↓ 注册路由
  └── group.GET("/ping", handler.Ping)
```

### 3.2 依赖注入优势

1. **解耦**：各层之间通过接口通信，降低耦合
2. **可测试**：可以注入 Mock 对象进行单元测试
3. **灵活**：可以轻松替换实现（如切换数据库）
4. **清晰**：依赖关系一目了然

### 3.3 构造函数注入示例

```go
// Handler 构造函数
func NewHandler(logger *logger.Logger, service *example.Service) *Handler {
    return &Handler{
        logger:  logger,
        service: service,
    }
}

// Service 构造函数
func NewService(logger *logger.Logger, db *database.Manager, redis *redis.Redis) *Service {
    return &Service{
        logger: logger,
        db:     db,
        redis:  redis,
    }
}
```

## 四、错误处理机制

### 4.1 错误处理流程

```
Service 层
  ↓ 返回错误（不记录日志）
  ↓
Handler 层
  ↓ 记录错误日志
  ↓ 调用 response.Error()
  ↓
统一响应格式
  {
    "code": 40001,
    "message": "用户不存在",
    "detail": "user id: 123"
  }
```

### 4.2 GORM 错误自动转换

框架自动将 GORM 错误转换为业务错误：

- `gorm.ErrRecordNotFound` → `errors.ErrNotFound`
- `gorm.ErrDuplicatedKey` → `errors.ErrDuplicate`
- 其他错误 → `errors.ErrDBError`

### 4.3 错误处理示例

```go
// Service 层：只返回错误
func (s *Service) GetUser(ctx context.Context, id int) error {
    err := dbInstance.Slave(ctx).
        Table("users").
        Where("id = ?", id).
        First(&user).Error
    return err  // 不记录日志
}

// Handler 层：记录日志并返回响应
func (h *Handler) DBError(c *gin.Context) {
    err := h.service.GetUser(c.Request.Context(), 999)
    if err != nil {
        // Handler 层记录日志
        h.logger.Error("failed to get user",
            zap.Error(err),
            zap.Int("user_id", 999),
        )
        // 返回错误响应
        response.Error(c, err)
        return
    }
    response.Success(c, gin.H{"message": "user found"})
}
```

## 五、最佳实践

### 5.1 路由组织最佳实践

#### ✅ 推荐做法

1. **按业务模块拆分路由文件**
   ```
   internal/router/
   ├── router.go      # 主路由注册器
   ├── health.go      # 健康检查路由
   ├── example.go     # 示例路由
   └── user.go        # 用户路由
   ```

2. **使用路由分组**
   ```go
   v1 := engine.Group("/api/v1")
   {
       userGroup := v1.Group("/users")
       {
           userGroup.GET("", handler.List)
           userGroup.POST("", handler.Create)
       }
   }
   ```

3. **在路由注册时完成依赖注入**
   ```go
   func SetupUserRoutes(group *gin.RouterGroup, cfg *RouterConfig) {
       svc := userService.NewService(cfg.Logger, cfg.DB)
       handler := userHandler.NewHandler(cfg.Logger, svc)
       // 注册路由...
   }
   ```

#### ❌ 不推荐做法

1. **所有路由写在一个文件**（难以维护）
2. **使用全局变量传递依赖**（难以测试）
3. **在 Handler 中初始化 Service**（耦合度高）

### 5.2 分层架构最佳实践

#### Handler 层

✅ **只做参数验证和响应构造**
```go
func (h *Handler) GetUser(c *gin.Context) {
    id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    user, err := h.service.GetByID(c.Request.Context(), id)
    if err != nil {
        response.Error(c, err)
        return
    }
    response.Success(c, user)
}
```

❌ **不要在 Handler 中写业务逻辑**
```go
func (h *Handler) GetUser(c *gin.Context) {
    // ❌ 不要在 Handler 中直接操作数据库
    var user model.User
    h.db.First(&user, id)

    // ❌ 不要在 Handler 中写业务逻辑
    if user.Status == model.UserStatusBanned {
        response.Error(c, errors.ErrUserBanned)
        return
    }
}
```

#### Service 层

✅ **实现业务逻辑，通过 Repository 操作数据**
```go
func (s *Service) Create(ctx context.Context, req *CreateUserRequest) error {
    // 业务规则验证
    exists, _ := s.userRepo.ExistsByUsername(ctx, req.Username)
    if exists {
        return errors.ErrUserAlreadyExists
    }

    // 业务逻辑处理
    user := &model.User{Username: req.Username}
    return s.userRepo.Create(ctx, user)
}
```

❌ **不要在 Service 中记录日志**
```go
func (s *Service) Create(ctx context.Context, req *CreateUserRequest) error {
    // ❌ Service 不记录日志，由 Handler 记录
    s.logger.Info("creating user", zap.String("username", req.Username))
    // ...
}
```

### 5.3 日志记录最佳实践

**原则**：Handler 层记录日志，Service 层只返回错误

```go
// Handler 层
func (h *Handler) DBError(c *gin.Context) {
    err := h.service.GetUser(c.Request.Context(), 999)
    if err != nil {
        // ✅ Handler 层记录日志
        h.logger.Error("failed to get user",
            zap.Error(err),
            zap.Int("user_id", 999),
            zap.String("path", c.Request.URL.Path),
        )
        response.Error(c, err)
        return
    }
}

// Service 层
func (s *Service) GetUser(ctx context.Context, id int) error {
    // ✅ Service 层只返回错误，不记录日志
    return dbInstance.Slave(ctx).First(&user, id).Error
}
```

### 5.4 读写分离最佳实践

```go
// 读操作使用从库
func (s *Service) GetUser(ctx context.Context, id int) error {
    return s.db.Slave(ctx).First(&user, id).Error
}

// 写操作使用主库
func (s *Service) CreateUser(ctx context.Context, user *model.User) error {
    return s.db.Master(ctx).Create(user).Error
}
```

## 六、完整示例

### 6.1 添加用户模块的完整流程

#### 步骤 1：创建 Handler

```go
// internal/handler/user/user.go
package user

import (
    "github.com/gin-gonic/gin"
    "github.com/jingpc/gofast/internal/logger"
    "github.com/jingpc/gofast/internal/service/user"
    "github.com/jingpc/gofast/pkg/response"
)

type Handler struct {
    logger  *logger.Logger
    service *user.Service
}

func NewHandler(logger *logger.Logger, service *user.Service) *Handler {
    return &Handler{
        logger:  logger,
        service: service,
    }
}

func (h *Handler) List(c *gin.Context) {
    users, err := h.service.List(c.Request.Context())
    if err != nil {
        h.logger.Error("failed to list users", zap.Error(err))
        response.Error(c, err)
        return
    }
    response.Success(c, users)
}
```

#### 步骤 2：创建 Service

```go
// internal/service/user/user.go
package user

import (
    "context"
    "github.com/jingpc/gofast/internal/database"
    "github.com/jingpc/gofast/internal/logger"
)

type Service struct {
    logger *logger.Logger
    db     *database.Manager
}

func NewService(logger *logger.Logger, db *database.Manager) *Service {
    return &Service{
        logger: logger,
        db:     db,
    }
}

func (s *Service) List(ctx context.Context) ([]User, error) {
    var users []User
    err := s.db.Slave(ctx).Find(&users).Error
    return users, err
}
```

#### 步骤 3：创建路由文件

```go
// internal/router/user.go
package router

import (
    "github.com/gin-gonic/gin"
    userHandler "github.com/jingpc/gofast/internal/handler/user"
    userService "github.com/jingpc/gofast/internal/service/user"
)

func SetupUserRoutes(group *gin.RouterGroup, cfg *RouterConfig) {
    // 初始化 Service
    svc := userService.NewService(cfg.Logger, cfg.DB)

    // 初始化 Handler
    handler := userHandler.NewHandler(cfg.Logger, svc)

    // 注册路由
    userGroup := group.Group("/users")
    {
        userGroup.GET("", handler.List)
        userGroup.POST("", handler.Create)
        userGroup.GET("/:id", handler.GetByID)
        userGroup.PUT("/:id", handler.Update)
        userGroup.DELETE("/:id", handler.Delete)
    }
}
```

#### 步骤 4：注册路由

```go
// internal/router/router.go
func Setup(engine *gin.Engine, cfg *RouterConfig) {
    SetupHealthRoutes(engine, cfg)

    v1 := engine.Group("/api/v1")
    {
        SetupExampleRoutes(v1, cfg)
        SetupUserRoutes(v1, cfg)  // 添加这行
    }
}
```

## 七、架构演进规划

### 7.1 当前架构（Phase 1）

- ✅ 路由分组和模块化
- ✅ Handler + Service 两层架构
- ✅ 依赖注入模式
- ✅ 统一错误处理
- ✅ 读写分离

### 7.2 未来演进（Phase 2）

- ⏳ 引入 Repository 层
- ⏳ 实现缓存层
- ⏳ 事务管理器
- ⏳ DTO 转换层
- ⏳ 中间件扩展（认证、权限、限流）

### 7.3 长期规划（Phase 3）

- 📋 领域驱动设计（DDD）
- 📋 CQRS 模式
- 📋 事件驱动架构
- 📋 微服务拆分

## 八、总结

### 核心设计原则

1. **单向依赖**：Handler → Service → Repository
2. **职责分离**：每层只做自己的事情
3. **依赖注入**：通过构造函数注入依赖
4. **统一错误处理**：Service 返回错误，Handler 记录日志
5. **读写分离**：读操作用从库，写操作用主库

### 关键要点

- **Handler 层**：参数验证 + 调用 Service + 构造响应 + 记录日志
- **Service 层**：业务逻辑 + 调用 Repository + 返回错误
- **Repository 层**：数据访问 + 缓存管理 + 读写分离

### 学习路径

1. 📖 理解三层架构的职责划分
2. 💻 查看 `internal/router/example.go` 示例代码
3. 💻 查看 `internal/handler/example/example.go` 示例代码
4. 💻 查看 `internal/service/example/example.go` 示例代码
5. 🚀 尝试添加自己的业务模块

## 相关文档

- [分层架构详解](./layers.md)
- [错误处理机制](../phase-2-core/error-handling.md)
- [数据库管理](../phase-1-infrastructure/03-database.md)
- [日志系统](../phase-1-infrastructure/02-logger.md)

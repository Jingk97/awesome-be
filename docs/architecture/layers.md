# GoFast 框架分层设计

## 分层架构概述

GoFast 采用经典的**分层架构**（Layered Architecture），将应用分为多个层次，每层有明确的职责和边界。这种设计模式在大型项目中被广泛采用，具有良好的可维护性和可扩展性。

### 分层结构图

```
┌─────────────────────────────────────────────────────────┐
│                    传输层 (Transport)                    │
│              HTTP (Gin) / gRPC (Protocol Buffers)       │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                   中间件层 (Middleware)                  │
│        日志、恢复、CORS、认证、限流、链路追踪            │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                 控制器层 (Handler/Controller)            │
│          参数验证、调用 Service、构造响应                │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                    服务层 (Service)                      │
│          业务逻辑、事务管理、调用 Repository             │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                 数据访问层 (Repository)                  │
│          数据库操作、缓存操作、读写分离                  │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                  模型层 (Model/Entity)                   │
│              数据结构定义、业务规则                      │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                 基础设施层 (Infrastructure)              │
│          MySQL、PostgreSQL、Redis、MQ                   │
└─────────────────────────────────────────────────────────┘
```

### 分层原则

1. **单向依赖**：上层可以依赖下层，下层不能依赖上层
2. **职责单一**：每层只负责自己的职责，不越界
3. **接口隔离**：层与层之间通过接口通信，降低耦合
4. **依赖注入**：通过依赖注入管理层之间的依赖关系

## 各层详解

### 1. 传输层 (Transport Layer)

**职责**：接收外部请求，路由到对应的 Handler

**包含组件**：
- HTTP Server (Gin)
- gRPC Server
- 路由注册

**代码位置**：
- `cmd/http/main.go` - HTTP 服务入口
- `cmd/grpc/main.go` - gRPC 服务入口
- `api/http/router.go` - HTTP 路由定义
- `api/grpc/*.proto` - gRPC 服务定义

**示例代码**：

```go
// api/http/router.go
package http

import (
    "github.com/gin-gonic/gin"
    "gofast/internal/handler/http"
    "gofast/internal/middleware"
)

func NewRouter(
    userHandler *http.UserHandler,
    authMiddleware *middleware.AuthMiddleware,
) *gin.Engine {
    r := gin.New()

    // 全局中间件
    r.Use(
        middleware.Logger(),
        middleware.Recovery(),
        middleware.CORS(),
        middleware.Trace(),
    )

    // 公开路由
    public := r.Group("/api/v1")
    {
        public.POST("/login", userHandler.Login)
        public.POST("/register", userHandler.Register)
    }

    // 需要认证的路由
    auth := r.Group("/api/v1")
    auth.Use(authMiddleware.Auth())
    {
        auth.GET("/users/:id", userHandler.GetUser)
        auth.PUT("/users/:id", userHandler.UpdateUser)
        auth.DELETE("/users/:id", userHandler.DeleteUser)
    }

    return r
}
```

**关键点**：
- 只负责路由匹配和请求分发
- 不包含业务逻辑
- 注册全局中间件

---

### 2. 中间件层 (Middleware Layer)

**职责**：处理横切关注点（Cross-cutting Concerns）

**包含组件**：
- 日志记录
- Panic 恢复
- CORS 处理
- JWT 认证
- 权限检查
- 限流
- 链路追踪

**代码位置**：`internal/middleware/`

**示例代码**：

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
        duration := time.Since(start)
        status := c.Writer.Status()

        logger.Info("HTTP Request",
            "method", method,
            "path", path,
            "status", status,
            "duration", duration,
            "ip", c.ClientIP(),
        )
    }
}
```

```go
// internal/middleware/auth.go
package middleware

import (
    "github.com/gin-gonic/gin"
    "gofast/pkg/jwt"
    "gofast/pkg/response"
    "gofast/pkg/errors"
)

type AuthMiddleware struct {
    jwtManager *jwt.Manager
}

func NewAuthMiddleware(jwtManager *jwt.Manager) *AuthMiddleware {
    return &AuthMiddleware{
        jwtManager: jwtManager,
    }
}

func (m *AuthMiddleware) Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从 Header 获取 Token
        token := c.GetHeader("Authorization")
        if token == "" {
            response.Error(c, errors.ErrUnauthorized)
            c.Abort()
            return
        }

        // 验证 Token
        claims, err := m.jwtManager.ParseToken(token)
        if err != nil {
            response.Error(c, errors.ErrUnauthorized)
            c.Abort()
            return
        }

        // 将用户信息注入到 Context
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Next()
    }
}
```

**关键点**：
- 不包含业务逻辑
- 可复用、可组合
- 通过 `c.Next()` 调用下一个中间件或 Handler

---

### 3. 控制器层 (Handler/Controller Layer)

**职责**：
- 接收和验证请求参数
- 调用 Service 层处理业务逻辑
- 构造和返回响应

**代码位置**：`internal/handler/`

**示例代码**：

```go
// internal/handler/http/user_handler.go
package http

import (
    "strconv"
    "github.com/gin-gonic/gin"
    "gofast/internal/service"
    "gofast/pkg/response"
    "gofast/pkg/errors"
)

type UserHandler struct {
    userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
    return &UserHandler{
        userService: userService,
    }
}

// GetUser 获取用户信息
func (h *UserHandler) GetUser(c *gin.Context) {
    // 1. 参数绑定和验证
    idStr := c.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        response.Error(c, errors.ErrInvalidParams)
        return
    }

    // 2. 调用 Service 层
    user, err := h.userService.GetByID(c.Request.Context(), id)
    if err != nil {
        response.Error(c, err)
        return
    }

    // 3. 构造响应
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

    // 3. 调用 Service 层
    user, err := h.userService.Create(c.Request.Context(), &req)
    if err != nil {
        response.Error(c, err)
        return
    }

    // 4. 构造响应
    response.Success(c, user)
}

// 请求结构体
type CreateUserRequest struct {
    Username string `json:"username" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

func (r *CreateUserRequest) Validate() error {
    // 自定义验证逻辑
    if len(r.Username) < 3 {
        return errors.New("username must be at least 3 characters")
    }
    return nil
}
```

**关键点**：
- **不包含业务逻辑**（这是初学者最容易犯的错误）
- 只做参数验证和响应构造
- 所有业务逻辑都在 Service 层

---

### 4. 服务层 (Service Layer)

**职责**：
- 实现业务逻辑
- 管理事务
- 调用 Repository 层进行数据操作
- 协调多个 Repository

**代码位置**：`internal/service/`

**示例代码**：

```go
// internal/service/user_service.go
package service

import (
    "context"
    "gofast/internal/model"
    "gofast/internal/repository"
    "gofast/pkg/errors"
    "gofast/pkg/transaction"
    "golang.org/x/crypto/bcrypt"
)

type UserService struct {
    userRepo  repository.UserRepository
    txManager *transaction.Manager
}

func NewUserService(
    userRepo repository.UserRepository,
    txManager *transaction.Manager,
) *UserService {
    return &UserService{
        userRepo:  userRepo,
        txManager: txManager,
    }
}

// GetByID 获取用户信息
func (s *UserService) GetByID(ctx context.Context, id int64) (*model.User, error) {
    // 参数验证
    if id <= 0 {
        return nil, errors.ErrInvalidParams
    }

    // 调用 Repository
    user, err := s.userRepo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }

    if user == nil {
        return nil, errors.ErrNotFound
    }

    return user, nil
}

// Create 创建用户
func (s *UserService) Create(ctx context.Context, req *CreateUserRequest) (*model.User, error) {
    // 1. 业务规则验证
    exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, errors.ErrUserAlreadyExists
    }

    // 2. 密码加密
    hashedPassword, err := bcrypt.GenerateFromPassword(
        []byte(req.Password),
        bcrypt.DefaultCost,
    )
    if err != nil {
        return nil, errors.ErrInternalError
    }

    // 3. 构造用户对象
    user := &model.User{
        Username: req.Username,
        Email:    req.Email,
        Password: string(hashedPassword),
        Status:   model.UserStatusActive,
    }

    // 4. 保存到数据库
    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }

    return user, nil
}

// UpdateProfile 更新用户资料（带事务）
func (s *UserService) UpdateProfile(
    ctx context.Context,
    userID int64,
    req *UpdateProfileRequest,
) error {
    // 使用事务管理器
    return s.txManager.Transaction(ctx, func(ctx context.Context) error {
        // 1. 查询用户
        user, err := s.userRepo.FindByID(ctx, userID)
        if err != nil {
            return err
        }
        if user == nil {
            return errors.ErrNotFound
        }

        // 2. 更新用户信息
        user.Nickname = req.Nickname
        user.Avatar = req.Avatar

        if err := s.userRepo.Update(ctx, user); err != nil {
            return err
        }

        // 3. 记录操作日志（假设有日志表）
        // log := &model.UserLog{...}
        // if err := s.logRepo.Create(ctx, log); err != nil {
        //     return err  // 事务会自动回滚
        // }

        return nil  // 事务自动提交
    })
}
```

**关键点**：
- **包含所有业务逻辑**
- 不直接操作数据库，通过 Repository
- 使用事务管理器处理复杂操作
- 可以调用多个 Repository

---

### 5. 数据访问层 (Repository Layer)

**职责**：
- 封装数据库操作
- 实现读写分离
- 缓存管理
- 数据转换

**代码位置**：`internal/repository/`

**示例代码**：

```go
// internal/repository/user_repository.go
package repository

import (
    "context"
    "gofast/internal/model"
    "gofast/pkg/cache"
    "gofast/pkg/database"
    "gorm.io/gorm"
)

// UserRepository 用户仓储接口
type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    FindByID(ctx context.Context, id int64) (*model.User, error)
    FindByUsername(ctx context.Context, username string) (*model.User, error)
    Update(ctx context.Context, user *model.User) error
    Delete(ctx context.Context, id int64) error
    ExistsByUsername(ctx context.Context, username string) (bool, error)
    List(ctx context.Context, page, pageSize int) ([]*model.User, int64, error)
}

// userRepository 用户仓储实现
type userRepository struct {
    db    *database.Manager
    cache cache.Cache
}

func NewUserRepository(
    db *database.Manager,
    cache cache.Cache,
) UserRepository {
    return &userRepository{
        db:    db,
        cache: cache,
    }
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
    // 写操作 → 主库
    return r.db.Master(ctx).Create(user).Error
}

// FindByID 根据 ID 查询用户
func (r *userRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
    // 1. 先查缓存
    cacheKey := fmt.Sprintf("user:%d", id)
    var user model.User
    if err := r.cache.Get(ctx, cacheKey, &user); err == nil {
        return &user, nil
    }

    // 2. 查数据库（读操作 → 从库）
    err := r.db.Slave(ctx).First(&user, id).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, nil
        }
        return nil, err
    }

    // 3. 写入缓存
    r.cache.Set(ctx, cacheKey, &user, 5*time.Minute)

    return &user, nil
}

// FindByUsername 根据用户名查询
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
    var user model.User
    err := r.db.Slave(ctx).
        Where("username = ?", username).
        First(&user).Error

    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, nil
        }
        return nil, err
    }

    return &user, nil
}

// Update 更新用户
func (r *userRepository) Update(ctx context.Context, user *model.User) error {
    // 写操作 → 主库
    err := r.db.Master(ctx).Save(user).Error
    if err != nil {
        return err
    }

    // 删除缓存
    cacheKey := fmt.Sprintf("user:%d", user.ID)
    r.cache.Delete(ctx, cacheKey)

    return nil
}

// Delete 删除用户
func (r *userRepository) Delete(ctx context.Context, id int64) error {
    // 软删除
    err := r.db.Master(ctx).Delete(&model.User{}, id).Error
    if err != nil {
        return err
    }

    // 删除缓存
    cacheKey := fmt.Sprintf("user:%d", id)
    r.cache.Delete(ctx, cacheKey)

    return nil
}

// ExistsByUsername 检查用户名是否存在
func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
    var count int64
    err := r.db.Slave(ctx).
        Model(&model.User{}).
        Where("username = ?", username).
        Count(&count).Error

    return count > 0, err
}

// List 分页查询用户列表
func (r *userRepository) List(ctx context.Context, page, pageSize int) ([]*model.User, int64, error) {
    var users []*model.User
    var total int64

    // 查询总数
    if err := r.db.Slave(ctx).Model(&model.User{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // 分页查询
    offset := (page - 1) * pageSize
    err := r.db.Slave(ctx).
        Offset(offset).
        Limit(pageSize).
        Find(&users).Error

    return users, total, err
}
```

**关键点**：
- 定义接口，便于测试和替换实现
- 读操作使用从库，写操作使用主库
- 集成缓存层
- 不包含业务逻辑

---

### 6. 模型层 (Model/Entity Layer)

**职责**：
- 定义数据结构
- 定义业务规则
- 数据验证

**代码位置**：`internal/model/`

**示例代码**：

```go
// internal/model/user.go
package model

import (
    "time"
)

// User 用户模型
type User struct {
    ID        int64     `json:"id" gorm:"primaryKey"`
    Username  string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
    Email     string    `json:"email" gorm:"uniqueIndex;size:100;not null"`
    Password  string    `json:"-" gorm:"size:255;not null"`  // 不返回给前端
    Nickname  string    `json:"nickname" gorm:"size:50"`
    Avatar    string    `json:"avatar" gorm:"size:255"`
    Status    UserStatus `json:"status" gorm:"type:tinyint;default:1"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    DeletedAt *time.Time `json:"-" gorm:"index"`  // 软删除
}

// UserStatus 用户状态
type UserStatus int

const (
    UserStatusInactive UserStatus = 0  // 未激活
    UserStatusActive   UserStatus = 1  // 正常
    UserStatusBanned   UserStatus = 2  // 封禁
)

// TableName 指定表名
func (User) TableName() string {
    return "users"
}

// IsActive 是否激活
func (u *User) IsActive() bool {
    return u.Status == UserStatusActive
}

// Validate 验证用户数据
func (u *User) Validate() error {
    if u.Username == "" {
        return errors.New("username is required")
    }
    if u.Email == "" {
        return errors.New("email is required")
    }
    return nil
}
```

**关键点**：
- 使用 GORM 标签定义数据库映射
- 使用 JSON 标签定义 API 响应格式
- 敏感字段（如密码）使用 `json:"-"` 不返回给前端
- 可以包含简单的业务规则方法

---

## 层与层之间的交互

### 完整请求流程示例

假设用户请求：`GET /api/v1/users/123`

```
1. 传输层 (Gin Router)
   ↓ 路由匹配到 userHandler.GetUser

2. 中间件层
   ↓ Logger: 记录请求开始
   ↓ Recovery: 捕获 Panic
   ↓ Auth: 验证 JWT Token
   ↓ 提取用户信息注入到 Context

3. Handler 层 (userHandler.GetUser)
   ↓ 解析参数: id = 123
   ↓ 调用 Service: userService.GetByID(ctx, 123)

4. Service 层 (userService.GetByID)
   ↓ 参数验证: id > 0
   ↓ 调用 Repository: userRepo.FindByID(ctx, 123)

5. Repository 层 (userRepo.FindByID)
   ↓ 查询缓存: cache.Get("user:123")
   ↓ 缓存未命中
   ↓ 查询数据库: db.Slave().First(&user, 123)
   ↓ 写入缓存: cache.Set("user:123", user)
   ↓ 返回 user

6. Service 层
   ↓ 接收 user
   ↓ 业务逻辑处理（如果需要）
   ↓ 返回 user

7. Handler 层
   ↓ 接收 user
   ↓ 构造响应: response.Success(c, user)

8. 中间件层
   ↓ Logger: 记录请求结束

9. 传输层
   ↓ 返回 JSON 响应给客户端
```

### 依赖关系图

```
Handler ──依赖──> Service ──依赖──> Repository ──依赖──> Model
   │                │                  │
   │                │                  └──依赖──> Database/Cache
   │                │
   │                └──依赖──> Transaction Manager
   │
   └──依赖──> Response/Errors
```

### 数据流向

```
请求数据流:
Client → Handler (DTO) → Service (DTO) → Repository (Model) → Database

响应数据流:
Database → Repository (Model) → Service (Model/DTO) → Handler (DTO) → Client
```

**说明**：
- **DTO (Data Transfer Object)**：用于传输层和业务层之间的数据传输
- **Model/Entity**：用于业务层和数据层之间的数据传输

---

## 最佳实践

### 1. Handler 层最佳实践

**✅ 推荐**：Handler 只做参数验证和响应构造

```go
func (h *UserHandler) GetUser(c *gin.Context) {
    // 1. 参数验证
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        response.Error(c, errors.ErrInvalidParams)
        return
    }

    // 2. 调用 Service
    user, err := h.userService.GetByID(c.Request.Context(), id)
    if err != nil {
        response.Error(c, err)
        return
    }

    // 3. 构造响应
    response.Success(c, user)
}
```

**❌ 不推荐**：Handler 包含业务逻辑

```go
func (h *UserHandler) GetUser(c *gin.Context) {
    id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

    // ❌ 不要在 Handler 中直接操作数据库
    var user model.User
    h.db.First(&user, id)

    // ❌ 不要在 Handler 中写业务逻辑
    if user.Status == model.UserStatusBanned {
        response.Error(c, errors.ErrUserBanned)
        return
    }

    response.Success(c, user)
}
```

### 2. Service 层最佳实践

**✅ 推荐**：Service 包含业务逻辑，通过 Repository 操作数据

```go
func (s *UserService) Create(ctx context.Context, req *CreateUserRequest) (*model.User, error) {
    // 业务规则验证
    exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, errors.ErrUserAlreadyExists
    }

    // 密码加密
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

    // 构造对象
    user := &model.User{
        Username: req.Username,
        Password: string(hashedPassword),
    }

    // 调用 Repository
    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }

    return user, nil
}
```

**❌ 不推荐**：Service 直接操作数据库

```go
func (s *UserService) Create(ctx context.Context, req *CreateUserRequest) (*model.User, error) {
    // ❌ 不要在 Service 中直接使用 GORM
    var count int64
    s.db.Model(&model.User{}).Where("username = ?", req.Username).Count(&count)
    if count > 0 {
        return nil, errors.ErrUserAlreadyExists
    }

    user := &model.User{Username: req.Username}
    s.db.Create(user)  // ❌ 应该通过 Repository

    return user, nil
}
```

### 3. Repository 层最佳实践

**✅ 推荐**：定义接口，实现读写分离和缓存

```go
// 定义接口
type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    FindByID(ctx context.Context, id int64) (*model.User, error)
}

// 实现接口
type userRepository struct {
    db    *database.Manager
    cache cache.Cache
}

func (r *userRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
    // 先查缓存
    var user model.User
    if err := r.cache.Get(ctx, cacheKey, &user); err == nil {
        return &user, nil
    }

    // 查数据库（从库）
    err := r.db.Slave(ctx).First(&user, id).Error
    if err != nil {
        return nil, err
    }

    // 写入缓存
    r.cache.Set(ctx, cacheKey, &user, 5*time.Minute)

    return &user, nil
}
```

**❌ 不推荐**：Repository 包含业务逻辑

```go
func (r *userRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
    var user model.User
    r.db.First(&user, id)

    // ❌ 不要在 Repository 中写业务逻辑
    if user.Status == model.UserStatusBanned {
        return nil, errors.ErrUserBanned
    }

    return &user, nil
}
```

### 4. 事务管理最佳实践

**✅ 推荐**：在 Service 层使用事务管理器

```go
func (s *UserService) Transfer(ctx context.Context, fromID, toID int64, amount float64) error {
    return s.txManager.Transaction(ctx, func(ctx context.Context) error {
        // 扣款
        if err := s.accountRepo.Deduct(ctx, fromID, amount); err != nil {
            return err  // 自动回滚
        }

        // 加款
        if err := s.accountRepo.Add(ctx, toID, amount); err != nil {
            return err  // 自动回滚
        }

        // 记录日志
        if err := s.logRepo.Create(ctx, log); err != nil {
            return err  // 自动回滚
        }

        return nil  // 自动提交
    })
}
```

**❌ 不推荐**：手动管理事务

```go
func (s *UserService) Transfer(ctx context.Context, fromID, toID int64, amount float64) error {
    // ❌ 手动管理事务容易出错
    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    if err := s.accountRepo.Deduct(ctx, fromID, amount); err != nil {
        tx.Rollback()
        return err
    }

    if err := s.accountRepo.Add(ctx, toID, amount); err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit().Error
}
```

---

## 单元测试

### Handler 层测试

```go
func TestUserHandler_GetUser(t *testing.T) {
    // Mock Service
    mockService := &MockUserService{
        GetByIDFunc: func(ctx context.Context, id int64) (*model.User, error) {
            return &model.User{
                ID:       id,
                Username: "testuser",
            }, nil
        },
    }

    handler := NewUserHandler(mockService)

    // 创建测试请求
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Params = gin.Params{{Key: "id", Value: "123"}}

    // 执行
    handler.GetUser(c)

    // 断言
    assert.Equal(t, 200, w.Code)
}
```

### Service 层测试

```go
func TestUserService_Create(t *testing.T) {
    // Mock Repository
    mockRepo := &MockUserRepository{
        ExistsByUsernameFunc: func(ctx context.Context, username string) (bool, error) {
            return false, nil
        },
        CreateFunc: func(ctx context.Context, user *model.User) error {
            user.ID = 1
            return nil
        },
    }

    service := NewUserService(mockRepo, nil)

    // 执行
    user, err := service.Create(context.Background(), &CreateUserRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "password123",
    })

    // 断言
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, int64(1), user.ID)
}
```

### Repository 层测试

```go
func TestUserRepository_FindByID(t *testing.T) {
    // 使用测试数据库
    db := setupTestDB()
    defer db.Close()

    repo := NewUserRepository(db, nil)

    // 准备测试数据
    user := &model.User{
        Username: "testuser",
        Email:    "test@example.com",
    }
    db.Create(user)

    // 执行
    found, err := repo.FindByID(context.Background(), user.ID)

    // 断言
    assert.NoError(t, err)
    assert.NotNil(t, found)
    assert.Equal(t, user.Username, found.Username)
}
```

---

## 常见问题

### Q1: 为什么要分这么多层？

**A**: 分层架构的优势：
- **职责清晰**：每层只负责自己的事情
- **易于维护**：修改某一层不影响其他层
- **便于测试**：可以 Mock 依赖进行单元测试
- **可复用**：Service 可以被 HTTP 和 gRPC 共用
- **团队协作**：不同层可以并行开发

### Q2: Handler 和 Service 的区别是什么？

**A**:
- **Handler**：处理 HTTP 请求，负责参数验证和响应构造，不包含业务逻辑
- **Service**：实现业务逻辑，不关心请求来自 HTTP 还是 gRPC

### Q3: 什么时候需要 DTO？

**A**:
- 当 API 请求/响应结构与数据库模型不一致时
- 需要隐藏某些字段时（如密码）
- 需要组合多个模型时

```go
// DTO 示例
type UserResponse struct {
    ID       int64  `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    // 不包含 Password 字段
}

func ToUserResponse(user *model.User) *UserResponse {
    return &UserResponse{
        ID:       user.ID,
        Username: user.Username,
        Email:    user.Email,
    }
}
```

### Q4: Repository 可以调用其他 Repository 吗？

**A**:
- **不推荐**：Repository 之间不应该相互调用
- **推荐**：在 Service 层协调多个 Repository

```go
// ✅ 推荐：在 Service 层协调
func (s *OrderService) Create(ctx context.Context, req *CreateOrderRequest) error {
    // 调用多个 Repository
    user, _ := s.userRepo.FindByID(ctx, req.UserID)
    product, _ := s.productRepo.FindByID(ctx, req.ProductID)

    // 业务逻辑
    order := &model.Order{...}
    return s.orderRepo.Create(ctx, order)
}
```

### Q5: 如何处理跨层的数据传递？

**A**: 使用 Context 传递请求级别的数据

```go
// 在中间件中注入用户信息
func (m *AuthMiddleware) Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        claims, _ := m.jwtManager.ParseToken(token)
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}

// 在 Handler 中获取
func (h *UserHandler) GetProfile(c *gin.Context) {
    userID := c.GetInt64("user_id")
    user, _ := h.userService.GetByID(c.Request.Context(), userID)
    response.Success(c, user)
}
```

---

## 总结

GoFast 的分层架构遵循以下原则：

1. **Handler 层**：只做参数验证和响应构造
2. **Service 层**：实现所有业务逻辑
3. **Repository 层**：封装数据访问，实现读写分离和缓存
4. **Model 层**：定义数据结构和简单的业务规则

**记住**：
- 上层依赖下层，下层不依赖上层
- 每层只做自己的事情，不越界
- 通过接口和依赖注入降低耦合
- 业务逻辑永远在 Service 层

## 下一步

- 📖 阅读 [配置模块文档](../phase-1-infrastructure/01-config.md)
- 📖 阅读 [日志模块文档](../phase-1-infrastructure/02-logger.md)
- 💻 查看 [完整 CRUD 示例](../examples/crud-example.md)
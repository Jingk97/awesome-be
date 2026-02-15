# 数据库模块 (Database)

## 概述

数据库模块是 GoFast 框架的核心基础设施，基于 GORM 封装，提供统一的数据库访问接口和连接池管理。

### 核心特性

- ✅ **多数据库支持** - MySQL、PostgreSQL、SQLite
- ✅ **读写分离** - 主库写、从库读，自动负载均衡
- ✅ **连接池管理** - 完整的连接池配置和监控
- ✅ **多实例支持** - 同时连接多个数据库
- ✅ **配置热更新** - 滚动更新，优雅关闭
- ✅ **健康检查** - 定期检查数据库连接状态
- ✅ **事务管理** - 声明式事务，自动回滚/提交

## 支持的数据库类型

| 数据库 | 类型标识 | 说明 |
|--------|---------|------|
| MySQL | `mysql` | 最常用的关系型数据库 |
| PostgreSQL | `postgres` | 功能强大的开源数据库 |
| SQLite | `sqlite` | 轻量级嵌入式数据库 |

## 配置说明

### 完整配置示例

```yaml
databases:
  # 主数据库实例
  - name: "main"                  # 数据库实例名称（唯一标识）
    type: "mysql"                 # 数据库类型

    # 连接池配置
    max_idle_conns: 10            # 最大空闲连接数
    max_open_conns: 100           # 最大打开连接数
    conn_max_lifetime: 3600s      # 连接最大生命周期（1小时）
    conn_max_idle_time: 600s      # 连接最大空闲时间（10分钟）

    # 超时配置
    dial_timeout: 10s             # 连接超时
    read_timeout: 30s             # 读取超时
    write_timeout: 30s            # 写入超时

    # 日志配置
    log_level: "info"             # 日志级别: silent, error, warn, info
    slow_threshold: 1s            # 慢查询阈值

    # 热更新配置
    reload:
      grace_period: 30s           # 优雅关闭等待时间
      force_close: true           # 超时后是否强制关闭
      check_interval: 1s          # 检查间隔

    # 健康检查
    health_check:
      enabled: true               # 是否启用健康检查
      interval: 30s               # 检查间隔
      timeout: 5s                 # 超时时间
      retries: 3                  # 重试次数

    # 主库配置（写操作）
    master:
      host: "127.0.0.1"
      port: 3306
      username: "root"
      password: ""                # 建议通过环境变量设置
      database: "gofast"
      charset: "utf8mb4"
      parse_time: true
      loc: "Local"

    # 从库配置（读操作）- 可选
    slaves:
      - host: "127.0.0.1"
        port: 3307
        username: "root"
        password: ""
        database: "gofast"
        charset: "utf8mb4"
        parse_time: true
        loc: "Local"

      - host: "127.0.0.1"
        port: 3308
        username: "root"
        password: ""
        database: "gofast"
        charset: "utf8mb4"
        parse_time: true
        loc: "Local"

  # 日志数据库实例（PostgreSQL）
  - name: "log"
    type: "postgres"
    max_idle_conns: 5
    max_open_conns: 50
    master:
      host: "127.0.0.1"
      port: 5432
      username: "postgres"
      password: ""
      database: "logdb"
      sslmode: "disable"
```

### 配置项说明

#### 连接池配置

| 配置项 | 类型 | 说明 | 推荐值 |
|--------|------|------|--------|
| max_idle_conns | int | 最大空闲连接数 | CPU 核心数 * 2 |
| max_open_conns | int | 最大打开连接数 | CPU 核心数 * 10 |
| conn_max_lifetime | duration | 连接最大生命周期 | 1 小时 |
| conn_max_idle_time | duration | 连接最大空闲时间 | 10 分钟 |

**推荐值说明**（基于 8 核 CPU）：
- `max_idle_conns`: 16（8 * 2）
- `max_open_conns`: 80（8 * 10）
- `conn_max_lifetime`: 3600s（防止连接泄漏）
- `conn_max_idle_time`: 600s（释放空闲连接）

#### 超时配置

| 配置项 | 类型 | 说明 | 推荐值 |
|--------|------|------|--------|
| dial_timeout | duration | 连接超时 | 10s |
| read_timeout | duration | 读取超时 | 30s |
| write_timeout | duration | 写入超时 | 30s |

#### 主从配置

| 配置项 | 类型 | 说明 | 必填 |
|--------|------|------|------|
| master | object | 主库配置（写操作） | ✅ |
| slaves | array | 从库配置（读操作） | ❌ |

**注意**：
- 如果不配置从库，读操作会自动使用主库
- 从库支持多个实例，自动负载均衡（轮询）

## 读写分离

### 自动路由规则

```go
// 写操作 → 主库
db.Master(ctx).Create(&user)
db.Master(ctx).Save(&user)
db.Master(ctx).Delete(&user)
db.Master(ctx).Exec("UPDATE users SET ...")

// 读操作 → 从库（自动负载均衡）
db.Slave(ctx).First(&user, id)
db.Slave(ctx).Find(&users)
db.Slave(ctx).Where("status = ?", 1).Find(&users)
db.Slave(ctx).Count(&count)
```

### 负载均衡策略

从库使用**轮询（Round-Robin）**算法进行负载均衡：

```
请求1 → 从库1
请求2 → 从库2
请求3 → 从库1
请求4 → 从库2
...
```

### 从库故障处理

如果所有从库都不可用，自动降级到主库：

```go
// 从库不可用时，自动使用主库
db.Slave(ctx).First(&user, id)  // 实际使用主库
```

## 使用示例

### 基础使用

```go
package main

import (
    "context"
    "gofast/pkg/database"
    "gofast/internal/model"
)

func main() {
    // 1. 获取数据库实例
    db := database.Get("main")

    ctx := context.Background()

    // 2. 写操作（使用主库）
    user := &model.User{
        Username: "john",
        Email:    "john@example.com",
    }
    db.Master(ctx).Create(user)

    // 3. 读操作（使用从库）
    var users []model.User
    db.Slave(ctx).Where("status = ?", 1).Find(&users)

    // 4. 查询单条记录
    var user model.User
    db.Slave(ctx).First(&user, 123)
}
```

### Repository 层使用

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

func NewUserRepository(db database.Database) *UserRepository {
    return &UserRepository{db: db}
}

// Create 创建用户（写操作 → 主库）
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
    return r.db.Master(ctx).Create(user).Error
}

// FindByID 根据 ID 查询（读操作 → 从库）
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
    var user model.User
    err := r.db.Slave(ctx).First(&user, id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// Update 更新用户（写操作 → 主库）
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
    return r.db.Master(ctx).Save(user).Error
}

// Delete 删除用户（写操作 → 主库）
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
    return r.db.Master(ctx).Delete(&model.User{}, id).Error
}

// List 分页查询（读操作 → 从库）
func (r *UserRepository) List(ctx context.Context, page, pageSize int) ([]*model.User, int64, error) {
    var users []*model.User
    var total int64

    // 查询总数
    r.db.Slave(ctx).Model(&model.User{}).Count(&total)

    // 分页查询
    offset := (page - 1) * pageSize
    err := r.db.Slave(ctx).
        Offset(offset).
        Limit(pageSize).
        Find(&users).Error

    return users, total, err
}
```

### 事务使用

```go
// internal/service/user_service.go
package service

import (
    "context"
    "gofast/internal/repository"
    "gofast/pkg/transaction"
)

type UserService struct {
    userRepo repository.UserRepository
    logRepo  repository.LogRepository
    txMgr    *transaction.Manager
}

// CreateUser 创建用户（带事务）
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) error {
    return s.txMgr.Transaction(ctx, func(ctx context.Context) error {
        // 1. 创建用户
        user := &model.User{
            Username: req.Username,
            Email:    req.Email,
        }
        if err := s.userRepo.Create(ctx, user); err != nil {
            return err  // 自动回滚
        }

        // 2. 记录日志
        log := &model.UserLog{
            UserID: user.ID,
            Action: "create",
        }
        if err := s.logRepo.Create(ctx, log); err != nil {
            return err  // 自动回滚
        }

        return nil  // 自动提交
    })
}
```

## 配置热更新

### 什么是配置热更新？

配置热更新允许在应用运行时修改数据库配置，应用会自动检测变化并重新加载配置，**无需重启服务**。

### 支持热更新的配置

以下配置支持热更新（修改后立即生效）：

| 配置项 | 说明 |
|--------|------|
| max_idle_conns | 最大空闲连接数 |
| max_open_conns | 最大打开连接数 |
| conn_max_lifetime | 连接最大生命周期 |
| conn_max_idle_time | 连接最大空闲时间 |
| read_timeout | 读取超时 |
| write_timeout | 写入超时 |

### 不支持热更新的配置

以下配置**不支持**热更新（修改后需要滚动更新）：

| 配置项 | 原因 |
|--------|------|
| host | 需要重建连接 |
| port | 需要重建连接 |
| username | 需要重建连接 |
| password | 需要重建连接 |
| database | 需要重建连接 |

### 滚动更新流程

当修改需要重建连接的配置时，系统会执行滚动更新：

```
1. 读取新配置
2. 创建新连接池
3. 测试新连接（PING）
4. 标记旧连接池为"待关闭"
5. 新请求使用新连接池
6. 等待旧连接池的活跃连接完成（最多等待 grace_period）
7. 超时后强制关闭旧连接池
8. 完成切换
```

**配置示例**：
```yaml
databases:
  - name: "main"
    reload:
      grace_period: 30s      # 优雅关闭等待时间
      force_close: true      # 超时后是否强制关闭
      check_interval: 1s     # 检查间隔
```

### 热更新使用示例

**场景**：数据库主库故障，需要切换到备用主库

1. 修改配置文件：
```yaml
databases:
  - name: "main"
    master:
      host: "backup-db.example.com"  # 从 127.0.0.1 改为备用主库
      port: 3306
```

2. 保存文件后，应用会自动检测到变化并输出日志：
```
[INFO] Database config changed, reloading...
[INFO] Creating new database connection pool
[INFO] Testing new connection: PING OK
[INFO] Switching to new connection pool
[INFO] Waiting for old connections to finish (grace_period: 30s)
[DEBUG] Active connections: 5
[DEBUG] Active connections: 2
[DEBUG] Active connections: 0
[INFO] Gracefully closed old connection pool
[INFO] Database reload completed
```

3. 无需重启，数据库连接已切换到新主库

## 健康检查

### 自动健康检查

系统会定期检查数据库连接状态：

```yaml
databases:
  - name: "main"
    health_check:
      enabled: true          # 是否启用
      interval: 30s          # 检查间隔
      timeout: 5s            # 超时时间
      retries: 3             # 重试次数
```

### 健康检查内容

1. **连接检查**：执行 `SELECT 1` 测试连接
2. **连接池统计**：监控连接池状态
3. **慢查询检测**：记录慢查询日志

### 健康检查日志

```
[INFO] Database health check passed
[DEBUG] Connection pool stats: max=100, open=45, in_use=12, idle=33

[ERROR] Database health check failed: connection timeout
[WARN] Attempting to reconnect (retry 1/3)
```

## 连接池监控

### 获取连接池统计

```go
// 获取连接池统计信息
stats := db.Stats()

fmt.Printf("Max Open Connections: %d\n", stats.MaxOpenConnections)
fmt.Printf("Open Connections: %d\n", stats.OpenConnections)
fmt.Printf("In Use: %d\n", stats.InUse)
fmt.Printf("Idle: %d\n", stats.Idle)
```

### 连接池指标

| 指标 | 说明 |
|------|------|
| MaxOpenConnections | 最大打开连接数 |
| OpenConnections | 当前打开连接数 |
| InUse | 正在使用的连接数 |
| Idle | 空闲连接数 |

### 监控告警

建议监控以下指标：

1. **连接池使用率** = InUse / MaxOpenConnections
   - 超过 80% 时告警，考虑增加连接池大小

2. **空闲连接数**
   - 长期为 0 时告警，说明连接池不足

3. **连接等待时间**
   - 超过阈值时告警，说明连接池压力大

## 多数据库实例

### 配置多个数据库

```yaml
databases:
  # 主业务数据库
  - name: "main"
    type: "mysql"
    master:
      host: "127.0.0.1"
      database: "gofast"

  # 日志数据库
  - name: "log"
    type: "postgres"
    master:
      host: "127.0.0.1"
      database: "logdb"

  # 分析数据库
  - name: "analytics"
    type: "mysql"
    master:
      host: "127.0.0.1"
      database: "analytics"
```

### 使用多个数据库

```go
// 获取不同的数据库实例
mainDB := database.Get("main")
logDB := database.Get("log")
analyticsDB := database.Get("analytics")

// 使用不同的数据库
mainDB.Master(ctx).Create(&user)
logDB.Master(ctx).Create(&log)
analyticsDB.Slave(ctx).Find(&reports)
```

## 最佳实践

### 1. 连接池大小设置

```yaml
# ✅ 推荐：根据 CPU 核心数设置
databases:
  - name: "main"
    max_idle_conns: 16      # CPU 核心数 * 2（8核 = 16）
    max_open_conns: 80      # CPU 核心数 * 10（8核 = 80）

# ❌ 不推荐：设置过大或过小
databases:
  - name: "main"
    max_idle_conns: 1000    # 太大，浪费资源
    max_open_conns: 5       # 太小，性能瓶颈
```

### 2. 读写分离使用

```go
// ✅ 推荐：明确区分读写操作
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
    var user model.User
    err := r.db.Slave(ctx).First(&user, id).Error  // 读操作使用从库
    return &user, err
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
    return r.db.Master(ctx).Save(user).Error  // 写操作使用主库
}

// ❌ 不推荐：不区分读写
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
    var user model.User
    err := r.db.Master(ctx).First(&user, id).Error  // 读操作也用主库，浪费资源
    return &user, err
}
```

### 3. 事务使用

```go
// ✅ 推荐：使用事务管理器
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) error {
    return s.txMgr.Transaction(ctx, func(ctx context.Context) error {
        // 所有操作在同一个事务中
        if err := s.userRepo.Create(ctx, user); err != nil {
            return err  // 自动回滚
        }
        return nil  // 自动提交
    })
}

// ❌ 不推荐：手动管理事务
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) error {
    tx := s.db.Master(ctx).Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    if err := tx.Create(user).Error; err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit().Error
}
```

### 4. Context 使用

```go
// ✅ 推荐：始终传递 Context
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
    var user model.User
    err := r.db.Slave(ctx).First(&user, id).Error  // 传递 ctx
    return &user, err
}

// ❌ 不推荐：不传递 Context
func (r *UserRepository) FindByID(id int64) (*model.User, error) {
    var user model.User
    err := r.db.Slave(context.Background()).First(&user, id).Error  // 使用 Background
    return &user, err
}
```

### 5. 慢查询优化

```yaml
# 配置慢查询阈值
databases:
  - name: "main"
    slow_threshold: 1s      # 超过 1 秒的查询会被记录
```

```go
// 优化慢查询
// ❌ 不推荐：N+1 查询
for _, order := range orders {
    user, _ := userRepo.FindByID(ctx, order.UserID)  // 每次都查询数据库
}

// ✅ 推荐：批量查询
userIDs := extractUserIDs(orders)
users, _ := userRepo.FindByIDs(ctx, userIDs)  // 一次查询所有用户
userMap := toMap(users)
for _, order := range orders {
    user := userMap[order.UserID]
}
```

## 常见问题

### Q1: 如何选择数据库类型？

**A**: 根据业务需求选择：
- **MySQL**: 最常用，生态完善，适合大多数场景
- **PostgreSQL**: 功能强大，支持 JSON、全文搜索等高级特性
- **SQLite**: 轻量级，适合嵌入式或测试环境

### Q2: 读写分离后，如何保证数据一致性？

**A**:
1. **主从延迟**：MySQL 主从复制通常有几毫秒到几秒的延迟
2. **解决方案**：
   - 写后立即读：使用主库
   - 对一致性要求不高的读：使用从库
   - 关键业务：使用主库

```go
// 写后立即读，使用主库
user, _ := userRepo.Create(ctx, user)
user, _ = userRepo.FindByID(ctx, user.ID)  // 使用主库读取

// 普通查询，使用从库
users, _ := userRepo.List(ctx, page, pageSize)  // 使用从库
```

### Q3: 连接池满了怎么办？

**A**:
1. **临时方案**：增加 `max_open_conns`
2. **长期方案**：
   - 优化慢查询
   - 检查是否有连接泄漏
   - 使用连接池监控

### Q4: 如何处理数据库连接失败？

**A**:
1. **自动重试**：配置 `max_retries`
2. **健康检查**：启用自动健康检查
3. **告警通知**：集成监控系统

### Q5: 事务中可以使用从库吗？

**A**:
不推荐。事务中的所有操作应该使用主库，确保数据一致性：

```go
// ✅ 推荐：事务中使用主库
s.txMgr.Transaction(ctx, func(ctx context.Context) error {
    user, _ := s.userRepo.FindByID(ctx, id)  // 使用主库
    user.Status = 1
    s.userRepo.Update(ctx, user)  // 使用主库
    return nil
})
```

## 下一步

- 📖 阅读 [Redis 模块文档](./04-redis.md)
- 📖 阅读 [事务管理文档](../phase-2-core/02-transaction.md)
- 💻 查看 [完整示例代码](../examples/database-example.md)

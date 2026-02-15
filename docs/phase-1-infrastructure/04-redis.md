# Redis 模块 (Cache)

## 概述

Redis 模块是 GoFast 框架的缓存基础设施，基于 go-redis 封装，提供统一的缓存访问接口和连接池管理。

### 核心特性

- ✅ **多模式支持** - 单机、哨兵、集群模式
- ✅ **连接池管理** - 完整的连接池配置和监控
- ✅ **配置热更新** - 滚动更新，优雅关闭
- ✅ **健康检查** - 定期检查 Redis 连接状态
- ✅ **分布式锁** - 基于 Redlock 算法
- ✅ **多实例支持** - 同时连接多个 Redis 实例
- ✅ **缓存模式** - Cache-Aside、Read-Through、Write-Through

## 支持的 Redis 模式

| 模式 | 类型标识 | 说明 | 适用场景 |
|------|---------|------|---------|
| 单机模式 | `standalone` | 单个 Redis 实例 | 开发环境、小型应用 |
| 哨兵模式 | `sentinel` | 主从 + 哨兵，自动故障转移 | 生产环境、高可用 |
| 集群模式 | `cluster` | Redis Cluster，数据分片 | 大规模应用、海量数据 |

## 配置说明

### 单机模式配置

```yaml
redis:
  mode: "standalone"          # 单机模式
  addr: "127.0.0.1:6379"      # Redis 地址
  password: ""                 # Redis 密码（建议通过环境变量设置）
  db: 0                        # 数据库编号（0-15）

  # 连接池配置
  pool_size: 10                # 连接池大小（最大活跃连接数）
  min_idle_conns: 5            # 最小空闲连接数
  max_retries: 3               # 最大重试次数

  # 超时配置
  dial_timeout: 5s             # 连接超时
  read_timeout: 3s             # 读取超时
  write_timeout: 3s            # 写入超时
  pool_timeout: 4s             # 从连接池获取连接的超时
  idle_timeout: 300s           # 空闲连接超时（5分钟）

  # 连接检查
  idle_check_frequency: 60s    # 空闲连接检查频率

  # 热更新配置
  reload:
    grace_period: 30s          # 优雅关闭等待时间
    force_close: true          # 超时后是否强制关闭
    check_interval: 1s         # 检查间隔

  # 健康检查
  health_check:
    enabled: true              # 是否启用健康检查
    interval: 30s              # 检查间隔
    timeout: 5s                # 超时时间
```

### 哨兵模式配置

```yaml
redis:
  mode: "sentinel"             # 哨兵模式
  master_name: "mymaster"      # 主节点名称
  sentinel_addrs:              # 哨兵地址列表
    - "127.0.0.1:26379"
    - "127.0.0.1:26380"
    - "127.0.0.1:26381"
  password: ""                 # Redis 密码
  db: 0

  # 哨兵配置
  sentinel_password: ""        # 哨兵密码（如果有）
  route_by_latency: true       # 按延迟路由到从节点
  route_randomly: false        # 随机路由到从节点

  # 连接池配置（同单机模式）
  pool_size: 10
  min_idle_conns: 5
```

### 集群模式配置

```yaml
redis:
  mode: "cluster"              # 集群模式
  cluster_addrs:               # 集群节点地址列表
    - "127.0.0.1:7000"
    - "127.0.0.1:7001"
    - "127.0.0.1:7002"
    - "127.0.0.1:7003"
    - "127.0.0.1:7004"
    - "127.0.0.1:7005"
  password: ""

  # 集群配置
  max_redirects: 3             # 最大重定向次数
  read_only: false             # 是否允许从从节点读取
  route_by_latency: true       # 按延迟路由

  # 连接池配置（同单机模式）
  pool_size: 10
  min_idle_conns: 5
```

### 配置项说明

#### 连接池配置

| 配置项 | 类型 | 说明 | 推荐值 |
|--------|------|------|--------|
| pool_size | int | 连接池大小 | CPU 核心数 * 2 |
| min_idle_conns | int | 最小空闲连接数 | pool_size / 2 |
| max_retries | int | 最大重试次数 | 3 |
| idle_timeout | duration | 空闲连接超时 | 5 分钟 |

**推荐值说明**（基于 8 核 CPU）：
- `pool_size`: 16（8 * 2）
- `min_idle_conns`: 8（16 / 2）
- `idle_timeout`: 300s（释放长时间空闲的连接）

#### 超时配置

| 配置项 | 类型 | 说明 | 推荐值 |
|--------|------|------|--------|
| dial_timeout | duration | 连接超时 | 5s |
| read_timeout | duration | 读取超时 | 3s |
| write_timeout | duration | 写入超时 | 3s |
| pool_timeout | duration | 获取连接超时 | 4s |

## 使用示例

### 基础操作

```go
package main

import (
    "context"
    "time"
    "gofast/pkg/cache"
)

func main() {
    // 1. 获取 Redis 实例
    redis := cache.Get("default")

    ctx := context.Background()

    // 2. 字符串操作
    redis.Set(ctx, "key", "value", 10*time.Minute)
    value, _ := redis.Get(ctx, "key")
    redis.Del(ctx, "key")

    // 3. Hash 操作
    redis.HSet(ctx, "user:123", "name", "John", "age", 30)
    name, _ := redis.HGet(ctx, "user:123", "name")
    userMap, _ := redis.HGetAll(ctx, "user:123")

    // 4. List 操作
    redis.LPush(ctx, "queue", "task1", "task2")
    task, _ := redis.RPop(ctx, "queue")

    // 5. Set 操作
    redis.SAdd(ctx, "tags", "go", "redis", "cache")
    members, _ := redis.SMembers(ctx, "tags")

    // 6. Sorted Set 操作
    redis.ZAdd(ctx, "leaderboard", 100, "player1", 200, "player2")
    topPlayers, _ := redis.ZRange(ctx, "leaderboard", 0, 9)
}
```

### Repository 层使用

```go
// internal/repository/user_repository.go
package repository

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    "gofast/internal/model"
    "gofast/pkg/cache"
    "gofast/pkg/database"
)

type UserRepository struct {
    db    database.Database
    cache cache.Cache
}

func NewUserRepository(db database.Database, cache cache.Cache) *UserRepository {
    return &UserRepository{
        db:    db,
        cache: cache,
    }
}

// FindByID 根据 ID 查询用户（带缓存）
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
    // 1. 查缓存
    cacheKey := fmt.Sprintf("user:%d", id)
    cached, err := r.cache.Get(ctx, cacheKey)
    if err == nil {
        var user model.User
        if err := json.Unmarshal([]byte(cached), &user); err == nil {
            return &user, nil
        }
    }

    // 2. 查数据库
    var user model.User
    if err := r.db.Slave(ctx).First(&user, id).Error; err != nil {
        return nil, err
    }

    // 3. 写入缓存
    data, _ := json.Marshal(user)
    r.cache.Set(ctx, cacheKey, string(data), 5*time.Minute)

    return &user, nil
}

// Update 更新用户（更新缓存）
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
    // 1. 更新数据库
    if err := r.db.Master(ctx).Save(user).Error; err != nil {
        return err
    }

    // 2. 更新缓存
    cacheKey := fmt.Sprintf("user:%d", user.ID)
    data, _ := json.Marshal(user)
    r.cache.Set(ctx, cacheKey, string(data), 5*time.Minute)

    return nil
}

// Delete 删除用户（删除缓存）
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
    // 1. 删除数据库记录
    if err := r.db.Master(ctx).Delete(&model.User{}, id).Error; err != nil {
        return err
    }

    // 2. 删除缓存
    cacheKey := fmt.Sprintf("user:%d", id)
    r.cache.Del(ctx, cacheKey)

    return nil
}
```

### 分布式锁使用

```go
// internal/service/order_service.go
package service

import (
    "context"
    "errors"
    "time"
    "gofast/pkg/cache"
)

type OrderService struct {
    cache cache.Cache
}

// CreateOrder 创建订单（使用分布式锁）
func (s *OrderService) CreateOrder(ctx context.Context, userID int64, productID int64) error {
    // 1. 获取分布式锁
    lockKey := fmt.Sprintf("lock:order:user:%d", userID)
    locked, err := s.cache.Lock(ctx, lockKey, 10*time.Second)
    if err != nil {
        return err
    }
    if !locked {
        return errors.New("failed to acquire lock, please try again")
    }
    defer s.cache.Unlock(ctx, lockKey)

    // 2. 执行业务逻辑（防止重复下单）
    // 检查用户是否已有未支付订单
    // 创建订单
    // ...

    return nil
}
```

### 缓存模式使用

#### Cache-Aside（旁路缓存）

```go
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
    // 1. 查缓存
    cacheKey := fmt.Sprintf("user:%d", id)
    if cached, err := r.cache.Get(ctx, cacheKey); err == nil {
        var user model.User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }

    // 2. 缓存未命中，查数据库
    var user model.User
    if err := r.db.Slave(ctx).First(&user, id).Error; err != nil {
        return nil, err
    }

    // 3. 写入缓存
    data, _ := json.Marshal(user)
    r.cache.Set(ctx, cacheKey, string(data), 5*time.Minute)

    return &user, nil
}
```

#### Read-Through（读穿透）

```go
// 封装的缓存助手
func GetOrSet(ctx context.Context, cache cache.Cache, key string, expiration time.Duration, fn func() (interface{}, error)) (interface{}, error) {
    // 先查缓存
    if cached, err := cache.Get(ctx, key); err == nil {
        return cached, nil
    }

    // 缓存未命中，执行函数获取数据
    value, err := fn()
    if err != nil {
        return nil, err
    }

    // 写入缓存
    cache.Set(ctx, key, value, expiration)

    return value, nil
}

// 使用
user, err := GetOrSet(ctx, cache, "user:123", 5*time.Minute, func() (interface{}, error) {
    return userRepo.FindByIDFromDB(ctx, 123)
})
```

## 配置热更新

### 滚动更新流程

当修改 Redis 配置时，系统会执行滚动更新：

```
1. 读取新配置
2. 创建新 Redis 客户端
3. 测试新连接（PING）
4. 标记旧客户端为"待关闭"
5. 新请求使用新客户端
6. 等待旧客户端的活跃连接完成（最多等待 grace_period）
7. 超时后强制关闭旧客户端
8. 完成切换
```

### 热更新使用示例

**场景**：Redis 主节点故障，需要切换到新节点

1. 修改配置文件：
```yaml
redis:
  addr: "new-redis.example.com:6379"  # 从 127.0.0.1:6379 改为新节点
```

2. 保存文件后，应用会自动检测到变化并输出日志：
```
[INFO] Redis config changed, reloading...
[INFO] Creating new redis client
[INFO] Testing new connection: PONG
[INFO] Switching to new redis client
[INFO] Waiting for old connections to finish (grace_period: 30s)
[DEBUG] Active connections: 3
[DEBUG] Active connections: 1
[DEBUG] Active connections: 0
[INFO] Gracefully closed old redis client
[INFO] Redis reload completed
```

## 健康检查

### 自动健康检查

系统会定期检查 Redis 连接状态：

```yaml
redis:
  health_check:
    enabled: true          # 是否启用
    interval: 30s          # 检查间隔
    timeout: 5s            # 超时时间
```

### 健康检查内容

1. **连接检查**：执行 `PING` 测试连接
2. **连接池统计**：监控连接池状态
3. **性能指标**：命中率、超时次数等

### 健康检查日志

```
[INFO] Redis health check passed
[DEBUG] Pool stats: total=10, idle=7, stale=0, hits=1000, misses=50, timeouts=0

[ERROR] Redis health check failed: connection timeout
[WARN] Redis pool has too many timeouts: 100
```

## 连接池监控

### 获取连接池统计

```go
// 获取连接池统计信息
stats := redis.Stats()

fmt.Printf("Total Connections: %d\n", stats.TotalConns)
fmt.Printf("Idle Connections: %d\n", stats.IdleConns)
fmt.Printf("Stale Connections: %d\n", stats.StaleConns)
fmt.Printf("Hits: %d\n", stats.Hits)
fmt.Printf("Misses: %d\n", stats.Misses)
fmt.Printf("Timeouts: %d\n", stats.Timeouts)
```

### 连接池指标

| 指标 | 说明 |
|------|------|
| TotalConns | 总连接数 |
| IdleConns | 空闲连接数 |
| StaleConns | 过期连接数 |
| Hits | 命中次数 |
| Misses | 未命中次数 |
| Timeouts | 超时次数 |

## 缓存问题解决方案

### 1. 缓存穿透（查询不存在的数据）

**问题**：大量请求查询不存在的数据，导致请求直接打到数据库

**解决方案1：缓存空值**

```go
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
    cacheKey := fmt.Sprintf("user:%d", id)

    // 查缓存
    cached, err := r.cache.Get(ctx, cacheKey)
    if err == nil {
        if cached == "null" {
            return nil, errors.New("user not found")
        }
        var user model.User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }

    // 查数据库
    var user model.User
    err = r.db.Slave(ctx).First(&user, id).Error
    if err == gorm.ErrRecordNotFound {
        // 缓存空值，防止穿透
        r.cache.Set(ctx, cacheKey, "null", 1*time.Minute)
        return nil, errors.New("user not found")
    }

    // 缓存正常值
    data, _ := json.Marshal(user)
    r.cache.Set(ctx, cacheKey, string(data), 5*time.Minute)

    return &user, nil
}
```

**解决方案2：布隆过滤器**

```go
// 使用布隆过滤器判断数据是否存在
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
    // 先用布隆过滤器判断
    if !r.bloomFilter.MightContain(ctx, fmt.Sprintf("user:%d", id)) {
        return nil, errors.New("user not found")
    }

    // 继续查询缓存和数据库...
}
```

### 2. 缓存击穿（热点数据过期）

**问题**：热点数据过期时，大量请求同时查询数据库

**解决方案：使用互斥锁**

```go
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
    cacheKey := fmt.Sprintf("user:%d", id)
    lockKey := fmt.Sprintf("lock:user:%d", id)

    // 查缓存
    if cached, err := r.cache.Get(ctx, cacheKey); err == nil {
        var user model.User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }

    // 获取锁
    locked, _ := r.cache.Lock(ctx, lockKey, 10*time.Second)
    if !locked {
        // 没获取到锁，等待后重试
        time.Sleep(100 * time.Millisecond)
        return r.FindByID(ctx, id)
    }
    defer r.cache.Unlock(ctx, lockKey)

    // 双重检查
    if cached, err := r.cache.Get(ctx, cacheKey); err == nil {
        var user model.User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }

    // 查数据库
    var user model.User
    r.db.Slave(ctx).First(&user, id)

    // 写入缓存
    data, _ := json.Marshal(user)
    r.cache.Set(ctx, cacheKey, string(data), 5*time.Minute)

    return &user, nil
}
```

### 3. 缓存雪崩（大量缓存同时过期）

**问题**：大量缓存同时过期，导致请求同时打到数据库

**解决方案1：随机过期时间**

```go
func (r *UserRepository) SetCache(ctx context.Context, key string, value interface{}) error {
    // 基础过期时间 + 随机时间（0-60秒）
    baseExpiration := 5 * time.Minute
    randomExpiration := time.Duration(rand.Intn(60)) * time.Second
    expiration := baseExpiration + randomExpiration

    return r.cache.Set(ctx, key, value, expiration)
}
```

**解决方案2：永不过期 + 异步更新**

```go
func (r *UserRepository) SetCacheWithRefresh(ctx context.Context, key string, value interface{}) error {
    // 缓存永不过期
    r.cache.Set(ctx, key, value, 0)

    // 异步刷新
    go func() {
        time.Sleep(4 * time.Minute)  // 4分钟后刷新
        // 重新加载数据并更新缓存
        newValue, _ := r.loadFromDB(ctx, key)
        r.cache.Set(ctx, key, newValue, 0)
    }()

    return nil
}
```

## 多 Redis 实例

### 配置多个 Redis 实例

```yaml
redis_instances:
  # 会话缓存
  - name: "session"
    mode: "standalone"
    addr: "127.0.0.1:6379"
    db: 0
    pool_size: 20

  # 数据缓存
  - name: "data"
    mode: "sentinel"
    master_name: "mymaster"
    sentinel_addrs:
      - "127.0.0.1:26379"
    pool_size: 50

  # 分布式锁
  - name: "lock"
    mode: "cluster"
    cluster_addrs:
      - "127.0.0.1:7000"
      - "127.0.0.1:7001"
    pool_size: 30
```

### 使用多个 Redis 实例

```go
// 获取不同的 Redis 实例
sessionCache := cache.Get("session")
dataCache := cache.Get("data")
lockCache := cache.Get("lock")

// 使用不同的缓存
sessionCache.Set(ctx, "session:123", userData, 30*time.Minute)
dataCache.Set(ctx, "user:123", user, 5*time.Minute)
lockCache.Lock(ctx, "order:123", 10*time.Second)
```

## 最佳实践

### 1. 缓存键命名规范

```go
// ✅ 推荐：使用清晰的命名规范
cacheKey := fmt.Sprintf("user:%d", userID)
cacheKey := fmt.Sprintf("order:%d:items", orderID)
cacheKey := fmt.Sprintf("session:%s", sessionID)

// ❌ 不推荐：命名不清晰
cacheKey := fmt.Sprintf("u%d", userID)
cacheKey := "data"
```

### 2. 设置合理的过期时间

```go
// ✅ 推荐：根据数据特性设置过期时间
cache.Set(ctx, "user:123", user, 5*time.Minute)      // 用户数据：5分钟
cache.Set(ctx, "session:abc", session, 30*time.Minute) // 会话：30分钟
cache.Set(ctx, "config", config, 1*time.Hour)        // 配置：1小时

// ❌ 不推荐：所有数据使用相同的过期时间
cache.Set(ctx, key, value, 10*time.Minute)  // 不区分数据类型
```

### 3. 缓存更新策略

```go
// ✅ 推荐：更新数据库后立即更新缓存
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
    // 1. 更新数据库
    if err := r.db.Master(ctx).Save(user).Error; err != nil {
        return err
    }

    // 2. 更新缓存
    cacheKey := fmt.Sprintf("user:%d", user.ID)
    data, _ := json.Marshal(user)
    r.cache.Set(ctx, cacheKey, string(data), 5*time.Minute)

    return nil
}

// ❌ 不推荐：只更新数据库，不更新缓存
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
    return r.db.Master(ctx).Save(user).Error  // 缓存会过期才更新
}
```

### 4. 批量操作

```go
// ✅ 推荐：使用批量操作
keys := []string{"user:1", "user:2", "user:3"}
values, _ := cache.MGet(ctx, keys...)

// ❌ 不推荐：循环单个操作
for _, key := range keys {
    value, _ := cache.Get(ctx, key)  // N 次网络请求
}
```

### 5. 错误处理

```go
// ✅ 推荐：缓存失败不影响主流程
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
    // 查缓存（失败不影响）
    cacheKey := fmt.Sprintf("user:%d", id)
    if cached, err := r.cache.Get(ctx, cacheKey); err == nil {
        var user model.User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }

    // 查数据库（主流程）
    var user model.User
    if err := r.db.Slave(ctx).First(&user, id).Error; err != nil {
        return nil, err
    }

    // 写入缓存（失败不影响）
    data, _ := json.Marshal(user)
    r.cache.Set(ctx, cacheKey, string(data), 5*time.Minute)

    return &user, nil
}
```

## 常见问题

### Q1: 如何选择 Redis 模式？

**A**: 根据业务需求选择：
- **单机模式**: 开发环境、小型应用、数据量小
- **哨兵模式**: 生产环境、需要高可用、自动故障转移
- **集群模式**: 大规模应用、海量数据、需要数据分片

### Q2: 缓存过期时间如何设置？

**A**: 根据数据特性设置：
- **热点数据**: 5-10 分钟
- **会话数据**: 30 分钟 - 2 小时
- **配置数据**: 1 小时 - 1 天
- **统计数据**: 1 天 - 7 天

### Q3: 如何避免缓存和数据库数据不一致？

**A**:
1. **更新数据库后立即更新缓存**
2. **设置合理的过期时间**
3. **使用消息队列异步更新缓存**
4. **关键业务不使用缓存**

### Q4: 分布式锁如何使用？

**A**:
```go
// 获取锁
locked, _ := cache.Lock(ctx, "lock:order:123", 10*time.Second)
if !locked {
    return errors.New("failed to acquire lock")
}
defer cache.Unlock(ctx, "lock:order:123")

// 执行业务逻辑...
```

### Q5: Redis 连接池满了怎么办？

**A**:
1. **临时方案**: 增加 `pool_size`
2. **长期方案**:
   - 检查是否有连接泄漏
   - 优化慢操作
   - 使用连接池监控

## 下一步

- 📖 阅读 [错误处理文档](../phase-2-core/01-errors.md)
- 📖 阅读 [事务管理文档](../phase-2-core/02-transaction.md)
- 💻 查看 [完整示例代码](../examples/redis-example.md)
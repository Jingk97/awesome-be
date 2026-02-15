# 配置模块 (Config)

## 概述

配置模块是 GoFast 框架的基础设施核心，负责管理应用的所有配置信息。基于 Viper 实现，提供了强大的配置管理能力。

### 核心特性

- ✅ **YAML 格式配置** - 清晰易读的配置文件格式
- ✅ **多环境支持** - 支持 dev、test、prod 等多环境配置
- ✅ **环境变量覆盖** - 敏感数据通过环境变量管理
- ✅ **命令行参数** - 灵活的启动参数配置
- ✅ **配置热更新** - 运行时动态更新配置（部分配置）
- ✅ **配置验证** - 启动时自动验证配置完整性
- ✅ **优先级机制** - 清晰的配置覆盖优先级

## 配置优先级

GoFast 使用 Viper 的配置优先级机制，从高到低依次为：

```
1. 命令行参数 (Flag)          优先级最高
2. 环境变量 (Environment)
3. 配置文件 (Config File)
4. 默认值 (Default)           优先级最低
```

### 优先级示例

假设我们有以下配置：

**配置文件 (config.yaml)**
```yaml
server:
  http:
    port: 8080
```

**环境变量**
```bash
export GOFAST_SERVER_HTTP_PORT=8081
```

**命令行参数**
```bash
./gofast --server.http.port=8082
```

**最终结果**：应用会监听 `8082` 端口（命令行参数优先级最高）

## 配置文件结构

### 完整配置示例

```yaml
# config.yaml - 完整配置文件示例

# ==================== 应用基础配置 ====================
app:
  name: "gofast"           # 应用名称
  env: "dev"               # 运行环境: dev, test, prod
  debug: true              # 是否开启调试模式

# ==================== 服务器配置 ====================
server:
  # HTTP 服务配置
  http:
    host: "0.0.0.0"              # 监听地址
    port: 8080                    # 监听端口
    read_timeout: 60s             # 读取超时时间
    write_timeout: 60s            # 写入超时时间
    max_header_bytes: 1048576     # 最大请求头大小 (1MB)

  # gRPC 服务配置
  grpc:
    host: "0.0.0.0"               # 监听地址
    port: 9090                    # 监听端口
    max_recv_msg_size: 4194304    # 最大接收消息大小 (4MB)
    max_send_msg_size: 4194304    # 最大发送消息大小 (4MB)

# ==================== 数据库配置 ====================
databases:
  # 主数据库实例
  - name: "main"                  # 数据库实例名称（唯一标识）
    type: "mysql"                 # 数据库类型: mysql, postgres, sqlite

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

    # 主库配置（用于写操作）
    master:
      host: "127.0.0.1"
      port: 3306
      username: "root"
      password: ""                # 建议通过环境变量设置
      database: "gofast"
      charset: "utf8mb4"
      parse_time: true
      loc: "Local"

    # 从库配置（用于读操作）- 可选
    slaves:
      - host: "127.0.0.1"
        port: 3307
        username: "root"
        password: ""
        database: "gofast"
        charset: "utf8mb4"
        parse_time: true
        loc: "Local"

  # 日志数据库实例（示例：使用 PostgreSQL）
  - name: "log"
    type: "postgres"
    max_idle_conns: 5
    max_open_conns: 50
    conn_max_lifetime: 3600s
    conn_max_idle_time: 600s
    dial_timeout: 10s
    read_timeout: 30s
    write_timeout: 30s
    log_level: "info"
    slow_threshold: 1s

    reload:
      grace_period: 30s
      force_close: true
      check_interval: 1s

    health_check:
      enabled: true
      interval: 30s
      timeout: 5s
      retries: 3

    master:
      host: "127.0.0.1"
      port: 5432
      username: "postgres"
      password: ""
      database: "logdb"
      sslmode: "disable"

# ==================== Redis 配置 ====================
redis:
  mode: "standalone"              # Redis 模式: standalone, sentinel, cluster
  addr: "127.0.0.1:6379"          # Redis 地址（单机模式）
  password: ""                     # Redis 密码（建议通过环境变量设置）
  db: 0                            # 数据库编号（0-15）

  # 连接池配置
  pool_size: 10                    # 连接池大小（最大活跃连接数）
  min_idle_conns: 5                # 最小空闲连接数
  max_retries: 3                   # 最大重试次数

  # 超时配置
  dial_timeout: 5s                 # 连接超时
  read_timeout: 3s                 # 读取超时
  write_timeout: 3s                # 写入超时
  pool_timeout: 4s                 # 从连接池获取连接的超时
  idle_timeout: 300s               # 空闲连接超时（5分钟）

  # 连接检查
  idle_check_frequency: 60s        # 空闲连接检查频率

  # 热更新配置
  reload:
    grace_period: 30s              # 优雅关闭等待时间
    force_close: true              # 超时后是否强制关闭
    check_interval: 1s             # 检查间隔

  # 健康检查
  health_check:
    enabled: true                  # 是否启用健康检查
    interval: 30s                  # 检查间隔
    timeout: 5s                    # 超时时间

# ==================== 日志配置 ====================
logger:
  level: "info"                    # 日志级别: debug, info, warn, error, fatal
  format: "json"                   # 日志格式: json, console
  output: "stdout"                 # 输出位置: stdout, stderr, file

  # 文件输出配置（当 output 为 file 时生效）
  file:
    filename: "logs/app.log"       # 日志文件路径
    max_size: 100                  # 单个文件最大大小 (MB)
    max_backups: 10                # 保留旧文件的最大个数
    max_age: 30                    # 保留旧文件的最大天数
    compress: true                 # 是否压缩旧文件

  enable_caller: true              # 是否显示调用位置（文件名和行号）
  enable_stacktrace: false         # 是否显示堆栈信息（仅 error 级别以上）

# ==================== 健康检查配置 ====================
health:
  timeout: 5s                      # 健康检查超时时间
  detailed: true                   # 是否返回详细信息（包含各组件状态）

# ==================== JWT 配置 ====================
jwt:
  secret: ""                       # JWT 密钥（必须通过环境变量设置）
  expire: 7200                     # 访问令牌过期时间（秒，默认 2 小时）
  refresh_expire: 604800           # 刷新令牌过期时间（秒，默认 7 天）
  issuer: "gofast"                 # 签发者

# ==================== 中间件配置 ====================
middleware:
  # CORS 跨域配置
  cors:
    enabled: true                  # 是否启用 CORS
    allow_origins: ["*"]           # 允许的源
    allow_methods:                 # 允许的 HTTP 方法
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
      - "OPTIONS"
    allow_headers: ["*"]           # 允许的请求头
    expose_headers: []             # 暴露的响应头
    max_age: 86400                 # 预检请求缓存时间（秒）

  # 限流配置（预留，当前不实现）
  rate_limit:
    enabled: false                 # 是否启用限流
    requests: 100                  # 时间窗口内允许的请求数
    window: 60s                    # 时间窗口大小

  # 链路追踪配置
  trace:
    enabled: true                  # 是否启用链路追踪
    header: "X-Trace-ID"           # 追踪 ID 的 Header 名称

# ==================== 扩展配置（预留）====================
extensions:
  # Elasticsearch 配置
  elasticsearch:
    enabled: false
    urls: ["http://localhost:9200"]

  # 消息队列配置
  message_queue:
    enabled: false
    type: "rabbitmq"               # rabbitmq, kafka, redis
    url: ""
```

### 配置项说明

#### 应用配置 (app)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| name | string | 应用名称 | gofast |
| env | string | 运行环境（dev/test/prod） | dev |
| debug | bool | 是否开启调试模式 | false |

#### 服务器配置 (server)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| http.host | string | HTTP 监听地址 | 0.0.0.0 |
| http.port | int | HTTP 监听端口 | 8080 |
| http.read_timeout | duration | 读取超时时间 | 60s |
| http.write_timeout | duration | 写入超时时间 | 60s |
| grpc.host | string | gRPC 监听地址 | 0.0.0.0 |
| grpc.port | int | gRPC 监听端口 | 9090 |

#### 数据库配置 (databases)

| 配置项 | 类型 | 说明 | 必填 |
|--------|------|------|------|
| name | string | 数据库实例名称（唯一标识） | ✅ |
| type | string | 数据库类型（mysql/postgres/sqlite） | ✅ |
| max_idle_conns | int | 最大空闲连接数 | ❌ |
| max_open_conns | int | 最大打开连接数 | ❌ |
| master.host | string | 主库地址 | ✅ |
| master.port | int | 主库端口 | ✅ |
| master.username | string | 用户名 | ✅ |
| master.password | string | 密码（建议环境变量） | ✅ |
| slaves | array | 从库配置（可选） | ❌ |

## 环境变量使用

### 环境变量命名规则

环境变量使用 `GOFAST_` 作为前缀，配置路径中的 `.` 替换为 `_`，全部大写。

**映射规则**：
```
配置项: server.http.port
环境变量: GOFAST_SERVER_HTTP_PORT

配置项: databases.0.master.password
环境变量: GOFAST_DATABASES_0_MASTER_PASSWORD
```

### 常用环境变量示例

```bash
# 服务器配置
export GOFAST_SERVER_HTTP_PORT=8080
export GOFAST_SERVER_GRPC_PORT=9090

# 数据库配置（敏感信息）
export GOFAST_DATABASES_0_MASTER_HOST=prod-db.example.com
export GOFAST_DATABASES_0_MASTER_PASSWORD=secure-password

# Redis 配置
export GOFAST_REDIS_ADDR=redis.example.com:6379
export GOFAST_REDIS_PASSWORD=redis-password

# JWT 配置（敏感信息）
export GOFAST_JWT_SECRET=super-secret-key-change-in-production

# 日志配置
export GOFAST_LOGGER_LEVEL=info
export GOFAST_LOGGER_OUTPUT=file
```

### 敏感数据最佳实践

**❌ 不推荐**：在配置文件中直接写入敏感数据
```yaml
database:
  master:
    password: "my-real-password"  # 不要这样做！
jwt:
  secret: "my-secret-key"         # 不要这样做！
```

**✅ 推荐**：配置文件留空，通过环境变量设置
```yaml
database:
  master:
    password: ""  # 留空，通过环境变量设置
jwt:
  secret: ""      # 留空，通过环境变量设置
```

```bash
# .env 文件（不要提交到 Git）
GOFAST_DATABASES_0_MASTER_PASSWORD=real-password
GOFAST_JWT_SECRET=real-secret-key
```

## 命令行参数

### 支持的命令行参数

```bash
# 指定配置文件路径
./gofast --config=/path/to/config.yaml
./gofast -c /path/to/config.yaml

# 指定运行环境
./gofast --env=prod

# 覆盖特定配置项
./gofast --server.http.port=8081
./gofast --logger.level=debug
./gofast --app.debug=false

# 组合使用
./gofast -c config.prod.yaml \
  --server.http.port=8081 \
  --logger.level=info \
  --app.debug=false
```

### 配置文件查找顺序

如果不指定 `--config` 参数，系统会按以下顺序查找配置文件：

1. `./config/config.yaml`
2. `./config.yaml`
3. `/etc/gofast/config.yaml`

### 多环境配置

支持通过 `--env` 参数指定环境，系统会自动加载对应的配置文件：

```bash
# 开发环境
./gofast --env=dev    # 加载 config.dev.yaml

# 测试环境
./gofast --env=test   # 加载 config.test.yaml

# 生产环境
./gofast --env=prod   # 加载 config.prod.yaml
```

## 配置热更新

### 什么是配置热更新？

配置热更新允许在应用运行时修改配置文件，应用会自动检测变化并重新加载配置，**无需重启服务**。

### 支持热更新的配置

以下配置支持热更新（修改后立即生效）：

| 配置项 | 说明 |
|--------|------|
| logger.level | 日志级别 |
| logger.format | 日志格式 |
| databases[].max_idle_conns | 数据库最大空闲连接数 |
| databases[].max_open_conns | 数据库最大打开连接数 |
| redis.pool_size | Redis 连接池大小 |
| middleware.cors.* | CORS 配置 |
| middleware.rate_limit.* | 限流配置 |

### 不支持热更新的配置

以下配置**不支持**热更新（修改后需要重启服务）：

| 配置项 | 原因 |
|--------|------|
| server.http.port | 端口已绑定，无法动态修改 |
| server.grpc.port | 端口已绑定，无法动态修改 |
| databases[].master.* | 数据库连接信息，涉及连接池重建 |
| jwt.secret | 密钥变更会导致现有 Token 失效 |

### 热更新使用示例

**场景**：线上应用日志太多，想临时调整日志级别

1. 修改配置文件：
```yaml
logger:
  level: "warn"  # 从 info 改为 warn
```

2. 保存文件后，应用会自动检测到变化并输出日志：
```
[INFO] Config file changed, reloading...
[INFO] Logger level updated: info -> warn
```

3. 无需重启，日志级别立即生效

## 使用示例

### 基础使用

```go
package main

import (
    "fmt"
    "log"

    "gofast/pkg/config"
)

func main() {
    // 1. 加载配置（启动时调用一次）
    cfg, err := config.Load("./config/config.yaml")
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }

    // 2. 获取配置值
    appName := cfg.App.Name
    httpPort := cfg.Server.HTTP.Port
    dbHost := cfg.Databases[0].Master.Host

    fmt.Printf("App: %s, Port: %d, DB: %s\n", appName, httpPort, dbHost)

    // 3. 使用辅助方法获取配置
    logLevel := config.GetString("logger.level")
    redisAddr := config.GetString("redis.addr")
    jwtExpire := config.GetInt("jwt.expire")

    fmt.Printf("Log: %s, Redis: %s, JWT: %d\n", logLevel, redisAddr, jwtExpire)
}
```

### 监听配置变化

```go
package main

import (
    "gofast/pkg/config"
    "gofast/pkg/logger"
)

func main() {
    // 加载配置
    cfg, _ := config.Load("./config/config.yaml")

    // 初始化日志
    log := logger.New(cfg.Logger)

    // 注册配置变化回调
    config.OnChange("logger", func(oldCfg, newCfg *config.Config) {
        // 当日志配置变化时，更新日志级别
        if oldCfg.Logger.Level != newCfg.Logger.Level {
            log.SetLevel(newCfg.Logger.Level)
            log.Info("Logger level updated",
                "old", oldCfg.Logger.Level,
                "new", newCfg.Logger.Level,
            )
        }
    })

    // 启动应用...
}
```

### 获取数据库配置

```go
package main

import (
    "gofast/pkg/config"
    "gofast/pkg/database"
)

func main() {
    cfg, _ := config.Load("./config/config.yaml")

    // 获取指定名称的数据库配置
    mainDB := cfg.GetDatabase("main")
    if mainDB == nil {
        panic("Database 'main' not found")
    }

    // 初始化数据库连接
    db, err := database.New(mainDB)
    if err != nil {
        panic(err)
    }

    // 使用数据库...
}
```

## 配置验证

### 启动时自动验证

应用启动时会自动验证配置的完整性和正确性：

```go
// 验证规则示例
- 必填项检查：数据库连接信息、JWT 密钥等
- 格式验证：端口号范围（1-65535）、超时时间格式等
- 逻辑验证：如果配置了从库，必须有主库
```

### 验证失败示例

```bash
$ ./gofast -c config.yaml

[FATAL] Config validation failed:
  - databases[0].master.password is required
  - jwt.secret is required
  - server.http.port must be between 1 and 65535

Please check your configuration file or environment variables.
```

## 最佳实践

### 1. 敏感数据管理

```yaml
# ✅ 推荐：配置文件不包含敏感数据
database:
  master:
    host: "localhost"
    username: "root"
    password: ""  # 通过环境变量设置

# ❌ 不推荐：直接写入敏感数据
database:
  master:
    password: "my-password"  # 不要这样做
```

### 2. 多环境配置

```bash
# 项目结构
config/
├── config.yaml          # 基础配置（开发环境）
├── config.dev.yaml      # 开发环境
├── config.test.yaml     # 测试环境
├── config.prod.yaml     # 生产环境
└── config.example.yaml  # 配置模板（提交到 Git）

# .gitignore
config/config.yaml
config/config.*.yaml
!config/config.example.yaml
```

### 3. 配置文件注释

```yaml
# ✅ 推荐：添加清晰的注释
server:
  http:
    port: 8080              # HTTP 服务端口
    read_timeout: 60s       # 读取超时（防止慢速攻击）

# ❌ 不推荐：没有注释
server:
  http:
    port: 8080
    read_timeout: 60s
```

### 4. 配置分组

```yaml
# ✅ 推荐：按功能分组，使用注释分隔
# ==================== 服务器配置 ====================
server:
  http:
    port: 8080

# ==================== 数据库配置 ====================
databases:
  - name: "main"
```

### 5. 默认值设置

```go
// ✅ 推荐：在代码中设置合理的默认值
func LoadConfig() *Config {
    viper.SetDefault("server.http.port", 8080)
    viper.SetDefault("logger.level", "info")
    viper.SetDefault("redis.pool_size", 10)
    // ...
}
```

## 常见问题

### Q1: 配置文件找不到怎么办？

**A**: 检查配置文件路径，或使用 `--config` 参数明确指定：
```bash
./gofast --config=/absolute/path/to/config.yaml
```

### Q2: 环境变量没有生效？

**A**: 检查环境变量命名是否正确：
- 必须以 `GOFAST_` 开头
- 使用 `_` 替代 `.`
- 全部大写

```bash
# ✅ 正确
export GOFAST_SERVER_HTTP_PORT=8080

# ❌ 错误
export SERVER_HTTP_PORT=8080  # 缺少前缀
export gofast_server_http_port=8080  # 未大写
```

### Q3: 如何知道哪些配置支持热更新？

**A**: 参考本文档的"配置热更新"章节，或查看日志输出：
```bash
[INFO] Config file changed
[INFO] Hot-reloadable configs: logger.level, databases.*.max_idle_conns
[WARN] Non-reloadable configs changed (requires restart): server.http.port
```

### Q4: 生产环境如何管理配置？

**A**: 推荐方案：
1. 配置文件只包含非敏感的默认值
2. 敏感数据通过环境变量或密钥管理系统（如 Vault）注入
3. 使用 Kubernetes ConfigMap/Secret 管理配置

### Q5: 如何调试配置加载问题？

**A**: 启用调试模式：
```bash
./gofast --app.debug=true --logger.level=debug
```

查看日志输出：
```
[DEBUG] Loading config from: ./config/config.yaml
[DEBUG] Config loaded successfully
[DEBUG] Environment variables applied: 3
[DEBUG] Final config: {...}
```

## 下一步

- 📖 阅读 [日志模块文档](./02-logger.md)
- 📖 阅读 [数据库模块文档](./03-database.md)
- 💻 查看 [完整示例代码](../examples/config-example.md)

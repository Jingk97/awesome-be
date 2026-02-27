# GoFast - 通用 Go Web 框架

基于 Gin 的企业级 Web 应用框架，提供完整的基础设施和最佳实践。

## 架构速读（快速了解项目）

- 定位：提供可复用的 Web 服务骨架，内置配置、日志、错误处理、统一响应、健康检查、数据库与 Redis 基础设施。
- 入口与启动流程：`cmd/server/main.go` 以“七阶段”启动顺序组织（配置 → 日志 → 健康管理器 → 基础设施 → Gin/中间件 → 路由 → HTTP 启动与优雅退出）。
- 请求流与分层：`router.Setup` 统一注册路由；示例业务采用 `Handler -> Service -> (Repository/DB/Redis)` 的分层思路。Handler 负责记录日志并返回统一响应，Service 仅返回错误，不直接记录日志。
- 依赖注入：`internal/router.RouterConfig` 通过构造函数传递 Logger/DB/Redis，避免全局变量，便于测试与解耦。
- 配置体系：`internal/config` 基于 Viper，支持默认值、`config/config.yaml` 配置文件、环境变量（前缀 `GOFAST_`）与命令行参数（`--config`/`--env`/`--port`）覆盖。
- 日志与追踪：`internal/logger` 封装 Zap；Gin 请求日志与 GORM SQL 日志统一进入日志系统；`X-Trace-ID` 写入 Context，并在响应 `trace_id` 字段返回。
- 健康检查：`/health/live` 与 `/health/ready` 提供 K8s 探针接口；当前 Handler 直接检测 DB/Redis。`internal/health.Manager` 支持组件注册与并发检查，DB/Redis 已注册检查器，可用于后续统一健康路由输出。
- 数据库：`internal/database` 基于 GORM，支持主从读写分离、轮询读取、连接池配置与健康检查。
- Redis：`internal/redis` 基于 go-redis UniversalClient，支持 standalone/sentinel/cluster 模式与健康检查，并提供常用操作封装。
- 错误与响应：`pkg/errors` 提供错误码体系与错误转换；`pkg/response` 统一响应结构并自动映射 HTTP 状态码。

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 运行应用

```bash
go run cmd/server/main.go
```

### 3. 测试健康检查

```bash
# 存活检查
curl http://localhost:8080/health/live

# 就绪检查
curl http://localhost:8080/health/ready

# 测试接口
curl http://localhost:8080/ping
```

## 项目结构

```
gofast/
├── cmd/                    # 应用入口
│   └── server/            # HTTP 服务器
│       └── main.go        # 主启动文件
├── internal/              # 内部包（不对外暴露）
│   ├── config/           # 配置模块
│   ├── logger/           # 日志模块
│   ├── database/         # 数据库模块
│   ├── redis/            # Redis 模块
│   ├── health/           # 健康检查模块
│   ├── router/           # 路由注册
│   ├── handler/          # HTTP 处理层
│   └── service/          # 业务逻辑层
├── pkg/                   # 公共包（可对外暴露）
│   ├── errors/           # 统一错误码与错误转换
│   ├── response/         # 统一响应格式
│   └── middleware/       # 中间件
├── config/                # 配置文件
├── go.mod                 # Go 模块定义
└── README.md             # 项目说明
```

## 开发规范

### 代码规范

1. **命名规范**
   - 包名：小写，单数，简短（如 `config`, `logger`）
   - 文件名：小写，下划线分隔（如 `user_service.go`）
   - 变量名：驼峰命名（如 `userName`, `userID`）
   - 常量名：大写，下划线分隔（如 `MAX_RETRY_COUNT`）

2. **注释规范**
   - 包注释：说明包的用途
   - 函数注释：说明功能、参数、返回值
   - 复杂逻辑：添加注释解释

3. **错误处理**
   - 不要忽略错误
   - 使用 `errors.Wrap` 添加上下文
   - 在合适的层级记录日志

### Git 工作流

1. **分支命名**
   - 功能分支：`feature/用户登录`
   - 修复分支：`fix/修复数据库连接泄漏`
   - 重构分支：`refactor/优化查询性能`

2. **提交信息**
   ```
   类型(范围): 简短描述
   
   详细描述（可选）
   
   关联 Issue（可选）
   ```

   示例：
   ```
   feat(auth): 实现用户登录功能
   
   - 添加 JWT 认证
   - 实现登录接口
   - 添加单元测试
   
   Closes #123
   ```

3. **提交类型**
   - `feat`: 新功能
   - `fix`: 修复 Bug
   - `docs`: 文档更新
   - `style`: 代码格式调整
   - `refactor`: 重构
   - `test`: 测试相关
   - `chore`: 构建/工具相关

### TODO 标记规范

```go
// TODO: 简单描述
// TODO(张三): 带责任人
// TODO(P0): 带优先级（P0=紧急, P1=重要, P2=普通）
// TODO(2024-03-01): 带截止日期
```

## 常见问题

### Q: 为什么使用 internal 目录？
A: `internal` 目录是 Go 的特殊目录，其中的包只能被同一模块内的代码导入，不能被外部模块使用。这样可以：
- 明确哪些是内部实现，哪些是公共 API
- 避免内部实现被外部依赖
- 提高代码的可维护性

### Q: 为什么要使用依赖注入？
A: 依赖注入的好处：
- 降低模块间的耦合
- 提高代码的可测试性
- 便于替换实现（如测试时使用 Mock）

### Q: 如何进行单元测试？
A: 参考 `*_test.go` 文件，使用 Go 的 testing 包：
```go
func TestUserService_Create(t *testing.T) {
    // 准备测试数据
    // 执行测试
    // 验证结果
}
```


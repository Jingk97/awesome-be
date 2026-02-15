# GoFast 框架文档

GoFast 是一个基于 Gin 框架的通用 Web 应用框架，提供了完整的基础设施和最佳实践，帮助你快速构建高质量的 Go 应用。

## 技术栈

- **HTTP 框架**: Gin
- **ORM**: GORM
- **配置管理**: Viper
- **日志**: Zap
- **依赖注入**: Wire
- **缓存**: Redis

## 文档结构

### 架构设计
- [整体架构](./architecture/overview.md) - 框架整体架构设计
- [分层设计](./architecture/layers.md) - 各层职责和交互关系

### 阶段一：基础设施层
- [配置模块](./phase-1-infrastructure/01-config.md) - 配置管理、热更新、环境变量
- [日志模块](./phase-1-infrastructure/02-logger.md) - 日志封装、分级、输出
- [数据库模块](./phase-1-infrastructure/03-database.md) - 多数据库、读写分离、连接池
- [Redis 模块](./phase-1-infrastructure/04-redis.md) - Redis 封装、连接池

### 阶段二：核心功能层
- [错误处理](./phase-2-core/01-errors.md) - 错误码、统一响应
- [事务管理](./phase-2-core/02-transaction.md) - 声明式事务
- [JWT 认证](./phase-2-core/03-jwt.md) - 认证、授权
- [中间件](./phase-2-core/04-middleware.md) - 日志、恢复、CORS、认证

### 阶段三：业务层框架
- [Repository 层](./phase-3-business/01-repository.md) - 数据访问层
- [Service 层](./phase-3-business/02-service.md) - 业务逻辑层
- [Handler 层](./phase-3-business/03-handler.md) - 请求处理层
- [依赖注入](./phase-3-business/04-wire.md) - Wire 使用指南

### 示例
- [完整 CRUD 示例](./examples/crud-example.md)
- [微服务示例](./examples/microservice-example.md)

## 快速开始

```bash
# 克隆项目
git clone https://github.com/yourusername/gofast.git

# 安装依赖
go mod download

# 复制配置文件
cp config/config.example.yaml config/config.yaml

# 运行项目
go run cmd/http/main.go
```

## 适用场景

- ✅ 中小型 Web 应用
- ✅ RESTful API 服务
- ✅ 单体应用（可拆分为微服务）
- ✅ 支持 HTTP 和 gRPC

## 设计原则

1. **简单易用** - 降低学习成本，快速上手
2. **模块化** - 各模块独立，便于维护和扩展
3. **可扩展** - 预留扩展接口，支持自定义
4. **最佳实践** - 遵循 Go 社区最佳实践
5. **生产就绪** - 考虑性能、安全、可维护性

## 贡献指南

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

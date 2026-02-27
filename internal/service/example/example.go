// Package example 示例 Service
//
// 核心功能：
// - 演示业务逻辑层的实现
// - 演示数据库访问
// - 演示错误处理
//
// 初级工程师学习要点：
// - Service 层负责业务逻辑
// - Service 层不记录日志（由 Handler 层记录）
// - Service 层只返回错误
// - Service 层可以调用 Repository 层
package example

import (
	"context"

	"github.com/jingpc/gofast/internal/database"
	"github.com/jingpc/gofast/internal/logger"
	"github.com/jingpc/gofast/internal/redis"
	"github.com/jingpc/gofast/pkg/errors"
)

// Service 示例服务
type Service struct {
	logger *logger.Logger
	db     *database.Manager
	redis  *redis.Redis
}

// NewService 创建示例服务
//
// 架构思路：
// - 通过构造函数注入依赖
// - 便于单元测试
// - 便于替换实现
func NewService(logger *logger.Logger, db *database.Manager, redis *redis.Redis) *Service {
	return &Service{
		logger: logger,
		db:     db,
		redis:  redis,
	}
}

// GetUser 获取用户 (演示数据库错误)
//
// 架构思路：
// - Service 层不记录日志，只返回错误
// - 日志由 Handler 层统一记录
// - GORM 错误会被自动转换为业务错误
//
// 初级工程师学习要点：
// - 使用 WithContext 传递上下文
// - 检查数据库连接是否存在
// - 返回错误而不是记录日志
// - 读操作使用 Slave() 方法
//
// 高级工程师思考：
// - 如何处理事务？
// - 如何优化查询性能？
// - 如何实现缓存？
func (s *Service) GetUser(ctx context.Context, id int) error {
	// 获取数据库连接
	dbInstance := s.db.Get("default")
	if dbInstance == nil {
		return errors.ErrDBError.WithDetail("default database not found")
	}

	// 定义用户结构
	var user struct {
		ID   int
		Name string
	}

	// 查询用户（读操作使用从库）
	// GORM 错误会被自动转换为业务错误：
	// - gorm.ErrRecordNotFound -> errors.ErrNotFound
	// - gorm.ErrDuplicatedKey -> errors.ErrDuplicate
	// - 其他错误 -> errors.ErrDBError
	err := dbInstance.Slave(ctx).
		Table("users").
		Where("id = ?", id).
		First(&user).Error

	return err
}

// TODO: 其他业务方法示例
//
// CreateUser 创建用户
// func (s *Service) CreateUser(ctx context.Context, name string) error {
//     // 业务逻辑
//     return nil
// }
//
// UpdateUser 更新用户
// func (s *Service) UpdateUser(ctx context.Context, id int, name string) error {
//     // 业务逻辑
//     return nil
// }
//
// DeleteUser 删除用户
// func (s *Service) DeleteUser(ctx context.Context, id int) error {
//     // 业务逻辑
//     return nil
// }

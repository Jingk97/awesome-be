// Package redis 提供 Redis 缓存功能
//
// Redis 模块基于 go-redis 封装，提供统一的缓存访问接口。
//
// 初级工程师学习要点：
// - 理解 Redis 的三种模式（standalone/sentinel/cluster）
// - 掌握连接池的配置和管理
// - 学习如何使用 Context 传递请求信息
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/jingpc/gofast/internal/config"
	"github.com/jingpc/gofast/internal/health"
)

// Redis Redis 客户端
//
// 初级工程师学习要点：
// - Redis 封装了 go-redis 的 UniversalClient
// - UniversalClient 可以自动适配三种模式（standalone/sentinel/cluster）
type Redis struct {
	name   string
	client redis.UniversalClient
	config config.RedisConfig
}

// New 创建 Redis 实例
//
// 初级工程师学习要点：
// - 根据配置的 mode 创建不同类型的 Redis 客户端
// - UniversalClient 是一个接口，可以统一处理三种模式
// - 自动注册到健康检查管理器
func New(cfg config.RedisConfig, healthMgr *health.Manager) (*Redis, error) {
	// 创建 Redis 客户端
	client := redis.NewUniversalClient(&redis.UniversalOptions{
		// 根据 mode 自动选择客户端类型
		Addrs:      getAddrs(cfg),
		MasterName: cfg.MasterName,

		// 认证
		Password: cfg.Password,
		DB:       cfg.DB,

		// 连接池配置
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,

		// 超时配置
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolTimeout:  cfg.PoolTimeout,

		// 连接检查
		ConnMaxIdleTime: cfg.IdleCheckFrequency,
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	r := &Redis{
		name:   cfg.Name,
		client: client,
		config: cfg,
	}

	// 注册健康检查（如果提供了 healthMgr）
	if healthMgr != nil {
		checker := &RedisHealthChecker{
			name:   cfg.Name,
			client: client,
		}
		healthMgr.Register(checker)
	}

	return r, nil
}

// getAddrs 根据配置获取 Redis 地址列表
//
// 初级工程师学习要点：
// - standalone 模式：单个地址
// - sentinel 模式：哨兵地址列表
// - cluster 模式：集群节点地址列表
func getAddrs(cfg config.RedisConfig) []string {
	switch cfg.Mode {
	case "sentinel":
		return cfg.SentinelAddrs
	case "cluster":
		return cfg.ClusterAddrs
	default: // standalone
		return []string{cfg.Addr}
	}
}

// Client 获取 Redis 客户端
//
// 初级工程师学习要点：
// - 返回原始的 go-redis 客户端
// - 可以使用 go-redis 的所有方法
func (r *Redis) Client() redis.UniversalClient {
	return r.client
}

// Close 关闭 Redis 连接
func (r *Redis) Close() error {
	return r.client.Close()
}

// Name 返回 Redis 实例名称
func (r *Redis) Name() string {
	return r.name
}

// ==================== 常用操作封装 ====================
// 以下是一些常用的 Redis 操作封装，方便使用

// Get 获取字符串值
//
// 初级工程师学习要点：
// - 使用 Context 传递请求信息
// - 如果 key 不存在，返回 redis.Nil 错误
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Set 设置字符串值
//
// 初级工程师学习要点：
// - expiration 为 0 表示永不过期
// - 使用 time.Duration 类型表示过期时间
func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Del 删除 key
func (r *Redis) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists 检查 key 是否存在
func (r *Redis) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

// Expire 设置 key 的过期时间
func (r *Redis) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取 key 的剩余过期时间
func (r *Redis) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// Incr 自增
func (r *Redis) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// Decr 自减
func (r *Redis) Decr(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
}

// HGet 获取 Hash 字段值
func (r *Redis) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HSet 设置 Hash 字段值
func (r *Redis) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.client.HSet(ctx, key, values...).Err()
}

// HGetAll 获取 Hash 所有字段
func (r *Redis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// HDel 删除 Hash 字段
func (r *Redis) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

// LPush 从列表左侧插入
func (r *Redis) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

// RPush 从列表右侧插入
func (r *Redis) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.RPush(ctx, key, values...).Err()
}

// LPop 从列表左侧弹出
func (r *Redis) LPop(ctx context.Context, key string) (string, error) {
	return r.client.LPop(ctx, key).Result()
}

// RPop 从列表右侧弹出
func (r *Redis) RPop(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

// LRange 获取列表范围内的元素
func (r *Redis) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.LRange(ctx, key, start, stop).Result()
}

// SAdd 添加集合成员
func (r *Redis) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

// SMembers 获取集合所有成员
func (r *Redis) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

// SIsMember 检查是否是集合成员
func (r *Redis) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.client.SIsMember(ctx, key, member).Result()
}

// SRem 删除集合成员
func (r *Redis) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SRem(ctx, key, members...).Err()
}

// ZAdd 添加有序集合成员
func (r *Redis) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return r.client.ZAdd(ctx, key, members...).Err()
}

// ZRange 获取有序集合范围内的成员
func (r *Redis) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(ctx, key, start, stop).Result()
}

// ZRem 删除有序集合成员
func (r *Redis) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.ZRem(ctx, key, members...).Err()
}

// RedisHealthChecker Redis 健康检查器
//
// 初级工程师学习要点：
// - 实现 health.HealthChecker 接口
// - 通过 Ping 检查 Redis 连接是否正常
type RedisHealthChecker struct {
	name   string
	client redis.UniversalClient
}

// Name 返回检查器名称
func (c *RedisHealthChecker) Name() string {
	return c.name
}

// Ping 执行轻量级健康检查（用于 Liveness）
//
// 初级工程师学习要点：
// - Ping 只检查 Redis 连接是否存活
// - 用于 Kubernetes Liveness Probe
func (c *RedisHealthChecker) Ping(ctx context.Context) error {
	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	return nil
}

// Check 执行完整健康检查（用于 Readiness）
//
// 初级工程师学习要点：
// - Check 执行完整的 Redis 检查（Ping + SET/GET 测试）
// - 用于 Kubernetes Readiness Probe
func (c *RedisHealthChecker) Check(ctx context.Context) error {
	// Ping 检查
	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	// 执行 SET/GET 测试验证 Redis 功能
	testKey := "__health_check__"
	if err := c.client.Set(ctx, testKey, "ok", 10*time.Second).Err(); err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}

	if _, err := c.client.Get(ctx, testKey).Result(); err != nil {
		return fmt.Errorf("redis get failed: %w", err)
	}

	return nil
}

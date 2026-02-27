// Package database 提供数据库管理功能
package database

import (
	"fmt"
	"sync"

	"github.com/jingpc/awesome-be/internal/config"
	"github.com/jingpc/awesome-be/internal/health"
	"github.com/jingpc/awesome-be/internal/logger"
)

// Manager 数据库管理器
//
// 初级工程师学习要点：
// - Manager 管理多个数据库实例
// - 使用 map 存储，通过名称快速查找
// - 使用 sync.RWMutex 保证并发安全
type Manager struct {
	databases map[string]*Database
	mu        sync.RWMutex
}

// NewManager 创建数据库管理器
//
// 初级工程师学习要点：
// - 根据配置初始化所有数据库实例
// - 每个数据库实例都会自动注册健康检查
// - 接收 logger 参数，传递给数据库实例
func NewManager(configs []config.DatabaseConfig, log *logger.Logger, healthMgr *health.Manager) (*Manager, error) {
	mgr := &Manager{
		databases: make(map[string]*Database),
	}

	// 初始化所有数据库实例
	for _, cfg := range configs {
		db, err := New(cfg, log, healthMgr)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database %s: %w", cfg.Name, err)
		}

		mgr.databases[cfg.Name] = db
	}

	return mgr, nil
}

// Get 获取指定名称的数据库实例
//
// 初级工程师学习要点：
// - 通过名称获取数据库实例
// - 如果不存在，返回 nil
func (m *Manager) Get(name string) *Database {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.databases[name]
}

// Close 关闭所有数据库连接
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, db := range m.databases {
		if err := db.Close(); err != nil {
			// 记录错误但继续关闭其他数据库
			// 注意：这里使用 fmt.Printf 是因为此时 logger 可能已经关闭
			fmt.Printf("Warning: failed to close database %s: %v\n", db.Name(), err)
		}
	}

	return nil
}

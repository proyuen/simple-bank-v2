package repository

import "gorm.io/gorm"

// TransactionManager 定义事务管理接口
type TransactionManager interface {
	Transaction(fc func(tx *gorm.DB) error) error
}

// GormTxManager 使用 GORM 实现事务管理
type GormTxManager struct {
	db *gorm.DB
}

// NewTxManager 创建事务管理器
func NewTxManager(db *gorm.DB) *GormTxManager {
	return &GormTxManager{db: db}
}

// Transaction 执行数据库事务
// 如果 fc 返回错误，事务会自动回滚；否则自动提交
func (t *GormTxManager) Transaction(fc func(tx *gorm.DB) error) error {
	return t.db.Transaction(fc)
}

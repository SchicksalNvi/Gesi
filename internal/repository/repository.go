package repository

import (
	"context"

	"gorm.io/gorm"
)

// BaseRepository 基础仓库接口
// 提供上下文和事务支持
type BaseRepository interface {
	// WithContext 返回带上下文的仓库实例
	WithContext(ctx context.Context) BaseRepository

	// WithTransaction 返回带事务的仓库实例
	WithTransaction(tx *gorm.DB) BaseRepository

	// GetDB 获取数据库实例
	GetDB() *gorm.DB
}

// baseRepository 基础仓库实现
type baseRepository struct {
	db  *gorm.DB
	ctx context.Context
}

// NewBaseRepository 创建基础仓库
func NewBaseRepository(db *gorm.DB) BaseRepository {
	return &baseRepository{
		db:  db,
		ctx: context.Background(),
	}
}

// WithContext 返回带上下文的仓库实例
func (r *baseRepository) WithContext(ctx context.Context) BaseRepository {
	return &baseRepository{
		db:  r.db,
		ctx: ctx,
	}
}

// WithTransaction 返回带事务的仓库实例
func (r *baseRepository) WithTransaction(tx *gorm.DB) BaseRepository {
	return &baseRepository{
		db:  tx,
		ctx: r.ctx,
	}
}

// GetDB 获取数据库实例
func (r *baseRepository) GetDB() *gorm.DB {
	if r.ctx != nil {
		return r.db.WithContext(r.ctx)
	}
	return r.db
}

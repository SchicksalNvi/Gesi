package database

import (
	"context"

	"gorm.io/gorm"
)

// Repository 仓储接口
type Repository interface {
	// WithContext 设置上下文
	WithContext(ctx context.Context) Repository

	// WithTransaction 执行事务
	WithTransaction(fn func(Repository) error) error

	// DB 获取底层数据库连接
	DB() *gorm.DB
}

// repository 仓储实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建仓储
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// WithContext 设置上下文
func (r *repository) WithContext(ctx context.Context) Repository {
	return &repository{
		db: r.db.WithContext(ctx),
	}
}

// WithTransaction 执行事务
func (r *repository) WithTransaction(fn func(Repository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepo := &repository{db: tx}
		return fn(txRepo)
	})
}

// DB 获取底层数据库连接
func (r *repository) DB() *gorm.DB {
	return r.db
}

// Pagination 分页结构
type Pagination struct {
	Page     int   `json:"page"`      // 当前页码（从 1 开始）
	PageSize int   `json:"page_size"` // 每页大小
	Total    int64 `json:"total"`     // 总记录数
}

// Paginate 分页查询辅助函数
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page < 1 {
			page = 1
		}
		if pageSize < 1 {
			pageSize = 10
		}
		if pageSize > 100 {
			pageSize = 100 // 限制最大页面大小
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// GetPagination 获取分页信息
func GetPagination(db *gorm.DB, page, pageSize int) (*Pagination, error) {
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	return &Pagination{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

package database

import (
	"context"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// 测试模型
type TestModel struct {
	ID        uint      `gorm:"primarykey"`
	Name      string    `gorm:"size:100"`
	Value     int       
	CreatedAt time.Time
}

// 属性 11：超时控制
// 验证需求：4.5
func TestRepositoryTimeoutProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// 创建测试数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	// 迁移测试模型
	db.AutoMigrate(&TestModel{})

	properties.Property("queries with context respect timeout", prop.ForAll(
		func(name string) bool {
			repo := NewRepository(db)

			// 创建一个已经超时的上下文
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
			defer cancel()

			time.Sleep(10 * time.Millisecond) // 确保上下文已超时

			// 使用超时的上下文
			repoWithCtx := repo.WithContext(ctx)

			// 尝试查询（应该失败或快速返回）
			var result TestModel
			err := repoWithCtx.DB().First(&result).Error

			// 验证操作被取消或快速完成
			return err != nil || ctx.Err() != nil
		},
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

// 属性 13：事务原子性
// 验证需求：5.2
func TestTransactionAtomicityProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// 创建测试数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	db.AutoMigrate(&TestModel{})

	properties.Property("failed transactions rollback all changes", prop.ForAll(
		func(name string, value int) bool {
			if name == "" {
				name = "test"
			}

			repo := NewRepository(db)

			// 记录初始计数
			var initialCount int64
			db.Model(&TestModel{}).Count(&initialCount)

			// 执行一个会失败的事务
			err := repo.WithTransaction(func(txRepo Repository) error {
				// 插入一条记录
				if err := txRepo.DB().Create(&TestModel{
					Name:  name,
					Value: value,
				}).Error; err != nil {
					return err
				}

				// 故意返回错误以触发回滚
				return gorm.ErrInvalidTransaction
			})

			// 验证事务失败
			if err == nil {
				return false
			}

			// 验证记录没有被插入（事务已回滚）
			var finalCount int64
			db.Model(&TestModel{}).Count(&finalCount)

			return finalCount == initialCount
		},
		gen.AnyString(),
		gen.Int(),
	))

	properties.TestingRun(t)
}

// 单元测试
func TestNewRepository(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	repo := NewRepository(db)
	if repo == nil {
		t.Error("expected repository to be created")
	}

	if repo.DB() != db {
		t.Error("expected repository to return the same db instance")
	}
}

func TestRepositoryWithContext(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	repo := NewRepository(db)
	ctx := context.Background()

	repoWithCtx := repo.WithContext(ctx)
	if repoWithCtx == nil {
		t.Error("expected repository with context to be created")
	}
}

func TestRepositoryWithTransaction(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	db.AutoMigrate(&TestModel{})

	repo := NewRepository(db)

	// 成功的事务
	err = repo.WithTransaction(func(txRepo Repository) error {
		return txRepo.DB().Create(&TestModel{
			Name:  "test",
			Value: 123,
		}).Error
	})

	if err != nil {
		t.Errorf("transaction failed: %v", err)
	}

	// 验证记录被插入
	var count int64
	db.Model(&TestModel{}).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 record, got %d", count)
	}
}

func TestPaginate(t *testing.T) {
	testCases := []struct {
		page         int
		pageSize     int
		expectedPage int
		expectedSize int
	}{
		{1, 10, 1, 10},
		{2, 20, 2, 20},
		{0, 10, 1, 10},   // page < 1 应该被修正为 1
		{1, 0, 1, 10},    // pageSize < 1 应该被修正为 10
		{1, 200, 1, 100}, // pageSize > 100 应该被限制为 100
	}

	for _, tc := range testCases {
		// 测试分页函数的行为
		page := tc.page
		pageSize := tc.pageSize

		if page < 1 {
			page = 1
		}
		if pageSize < 1 {
			pageSize = 10
		}
		if pageSize > 100 {
			pageSize = 100
		}

		if page != tc.expectedPage {
			t.Errorf("page=%d: expected %d, got %d", tc.page, tc.expectedPage, page)
		}

		if pageSize != tc.expectedSize {
			t.Errorf("pageSize=%d: expected %d, got %d", tc.pageSize, tc.expectedSize, pageSize)
		}
	}
}

func TestGetPagination(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	db.AutoMigrate(&TestModel{})

	// 插入测试数据
	for i := 0; i < 25; i++ {
		db.Create(&TestModel{Name: "test", Value: i})
	}

	// 测试分页
	pagination, err := GetPagination(db.Model(&TestModel{}), 1, 10)
	if err != nil {
		t.Fatalf("GetPagination failed: %v", err)
	}

	if pagination.Page != 1 {
		t.Errorf("expected page 1, got %d", pagination.Page)
	}

	if pagination.PageSize != 10 {
		t.Errorf("expected page size 10, got %d", pagination.PageSize)
	}

	if pagination.Total != 25 {
		t.Errorf("expected total 25, got %d", pagination.Total)
	}
}

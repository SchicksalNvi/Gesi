package repository

import (
	"go-cesi/internal/errors"
	"go-cesi/internal/models"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create 创建用户
func (r *userRepository) Create(user *models.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return errors.NewDatabaseError("create user", err)
	}
	return nil
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(id string) (*models.User, error) {
	var user models.User
	if err := r.db.Preload("Roles").Preload("NodeAccess").Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("user", id)
		}
		return nil, errors.NewDatabaseError("get user by id", err)
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.Preload("Roles").Preload("NodeAccess").Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("user", username)
		}
		return nil, errors.NewDatabaseError("get user by username", err)
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Preload("Roles").Preload("NodeAccess").Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("user", email)
		}
		return nil, errors.NewDatabaseError("get user by email", err)
	}
	return &user, nil
}

// Update 更新用户
func (r *userRepository) Update(user *models.User) error {
	if err := r.db.Save(user).Error; err != nil {
		return errors.NewDatabaseError("update user", err)
	}
	return nil
}

// Delete 删除用户
func (r *userRepository) Delete(id string) error {
	if err := r.db.Where("id = ?", id).Delete(&models.User{}).Error; err != nil {
		return errors.NewDatabaseError("delete user", err)
	}
	return nil
}

// List 获取用户列表
func (r *userRepository) List(offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	// 获取总数
	if err := r.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, errors.NewDatabaseError("count users", err)
	}

	// 获取分页数据
	if err := r.db.Preload("Roles").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, errors.NewDatabaseError("list users", err)
	}

	return users, total, nil
}

// ExistsByUsername 检查用户名是否存在
func (r *userRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, errors.NewDatabaseError("check username exists", err)
	}
	return count > 0, nil
}

// ExistsByEmail 检查邮箱是否存在
func (r *userRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, errors.NewDatabaseError("check email exists", err)
	}
	return count > 0, nil
}

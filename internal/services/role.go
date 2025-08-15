package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"go-cesi/internal/models"
)

type RoleService struct {
	db *gorm.DB
}

func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{db: db}
}

// CreateRole 创建角色
func (s *RoleService) CreateRole(role *models.Role) error {
	if role.ID == "" {
		role.ID = uuid.New().String()
	}

	// 检查角色名是否已存在
	var existingRole models.Role
	if err := s.db.Where("name = ?", role.Name).First(&existingRole).Error; err == nil {
		return errors.New("角色名已存在")
	}

	return s.db.Create(role).Error
}

// GetRoleByID 根据ID获取角色
func (s *RoleService) GetRoleByID(id string) (*models.Role, error) {
	var role models.Role
	err := s.db.Preload("Permissions").Preload("Users").First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRoleByName 根据名称获取角色
func (s *RoleService) GetRoleByName(name string) (*models.Role, error) {
	var role models.Role
	err := s.db.Preload("Permissions").First(&role, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetAllRoles 获取所有角色
func (s *RoleService) GetAllRoles() ([]models.Role, error) {
	var roles []models.Role
	err := s.db.Preload("Permissions").Find(&roles).Error
	return roles, err
}

// UpdateRole 更新角色
func (s *RoleService) UpdateRole(role *models.Role) error {
	// 系统角色不允许修改名称
	if role.IsSystem {
		return s.db.Model(role).Select("display_name", "description", "updated_at").Updates(role).Error
	}
	return s.db.Save(role).Error
}

// DeleteRole 删除角色
func (s *RoleService) DeleteRole(id string) error {
	var role models.Role
	if err := s.db.First(&role, "id = ?", id).Error; err != nil {
		return err
	}

	// 系统角色不允许删除
	if role.IsSystem {
		return errors.New("系统角色不允许删除")
	}

	// 检查是否有用户使用该角色
	var userCount int64
	s.db.Model(&models.UserRole{}).Where("role_id = ?", id).Count(&userCount)
	if userCount > 0 {
		return errors.New("该角色正在被用户使用，无法删除")
	}

	return s.db.Delete(&role).Error
}

// AssignPermissionsToRole 为角色分配权限
func (s *RoleService) AssignPermissionsToRole(roleID string, permissionIDs []string) error {
	// 先清除现有权限
	if err := s.db.Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error; err != nil {
		return err
	}

	// 添加新权限
	for _, permissionID := range permissionIDs {
		rolePermission := models.RolePermission{
			RoleID:       roleID,
			PermissionID: permissionID,
			CreatedAt:    time.Now(),
		}
		if err := s.db.Create(&rolePermission).Error; err != nil {
			return err
		}
	}

	return nil
}

// AssignRoleToUser 为用户分配角色
func (s *RoleService) AssignRoleToUser(userID, roleID, grantedBy string) error {
	// 检查是否已存在
	var existingUserRole models.UserRole
	if err := s.db.Where("user_id = ? AND role_id = ?", userID, roleID).First(&existingUserRole).Error; err == nil {
		return errors.New("用户已拥有该角色")
	}

	userRole := models.UserRole{
		UserID:    userID,
		RoleID:    roleID,
		GrantedBy: grantedBy,
		CreatedAt: time.Now(),
	}

	return s.db.Create(&userRole).Error
}

// RemoveRoleFromUser 移除用户角色
func (s *RoleService) RemoveRoleFromUser(userID, roleID string) error {
	return s.db.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&models.UserRole{}).Error
}

// GetUserRoles 获取用户的所有角色
func (s *RoleService) GetUserRoles(userID string) ([]models.Role, error) {
	var roles []models.Role
	err := s.db.Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Preload("Permissions").
		Find(&roles).Error
	return roles, err
}

// InitializeSystemRoles 初始化系统角色
func (s *RoleService) InitializeSystemRoles() error {
	// 定义系统角色
	systemRoles := []models.Role{
		{
			ID:          uuid.New().String(),
			Name:        models.RoleSuperAdmin,
			DisplayName: "超级管理员",
			Description: "拥有系统所有权限的超级管理员",
			IsSystem:    true,
		},
		{
			ID:          uuid.New().String(),
			Name:        models.RoleEnvironmentAdmin,
			DisplayName: "环境管理员",
			Description: "特定环境的管理员，可管理环境内所有资源",
			IsSystem:    true,
		},
		{
			ID:          uuid.New().String(),
			Name:        models.RoleNodeOperator,
			DisplayName: "节点操作员",
			Description: "可操作特定节点的进程管理",
			IsSystem:    true,
		},
		{
			ID:          uuid.New().String(),
			Name:        models.RoleReadOnlyUser,
			DisplayName: "只读用户",
			Description: "只能查看系统状态，无操作权限",
			IsSystem:    true,
		},
	}

	// 创建系统角色
	for _, role := range systemRoles {
		var existingRole models.Role
		if err := s.db.Where("name = ?", role.Name).First(&existingRole).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := s.db.Create(&role).Error; err != nil {
					return fmt.Errorf("创建系统角色 %s 失败: %v", role.Name, err)
				}
			}
		}
	}

	return nil
}

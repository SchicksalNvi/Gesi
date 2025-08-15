package services

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"go-cesi/internal/models"
)

type PermissionService struct {
	db *gorm.DB
}

func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{db: db}
}

// CreatePermission 创建权限
func (s *PermissionService) CreatePermission(permission *models.Permission) error {
	if permission.ID == "" {
		permission.ID = uuid.New().String()
	}

	// 检查权限名是否已存在
	var existingPermission models.Permission
	if err := s.db.Where("name = ?", permission.Name).First(&existingPermission).Error; err == nil {
		return errors.New("权限名已存在")
	}

	return s.db.Create(permission).Error
}

// GetPermissionByID 根据ID获取权限
func (s *PermissionService) GetPermissionByID(id string) (*models.Permission, error) {
	var permission models.Permission
	err := s.db.First(&permission, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// GetPermissionByName 根据名称获取权限
func (s *PermissionService) GetPermissionByName(name string) (*models.Permission, error) {
	var permission models.Permission
	err := s.db.First(&permission, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// GetAllPermissions 获取所有权限
func (s *PermissionService) GetAllPermissions() ([]models.Permission, error) {
	var permissions []models.Permission
	err := s.db.Find(&permissions).Error
	return permissions, err
}

// GetPermissionsByResource 根据资源类型获取权限
func (s *PermissionService) GetPermissionsByResource(resource string) ([]models.Permission, error) {
	var permissions []models.Permission
	err := s.db.Where("resource = ?", resource).Find(&permissions).Error
	return permissions, err
}

// UpdatePermission 更新权限
func (s *PermissionService) UpdatePermission(permission *models.Permission) error {
	// 系统权限不允许修改名称和资源类型
	if permission.IsSystem {
		return s.db.Model(permission).Select("display_name", "description", "updated_at").Updates(permission).Error
	}
	return s.db.Save(permission).Error
}

// DeletePermission 删除权限
func (s *PermissionService) DeletePermission(id string) error {
	var permission models.Permission
	if err := s.db.First(&permission, "id = ?", id).Error; err != nil {
		return err
	}

	// 系统权限不允许删除
	if permission.IsSystem {
		return errors.New("系统权限不允许删除")
	}

	// 检查是否有角色使用该权限
	var roleCount int64
	s.db.Model(&models.RolePermission{}).Where("permission_id = ?", id).Count(&roleCount)
	if roleCount > 0 {
		return errors.New("该权限正在被角色使用，无法删除")
	}

	return s.db.Delete(&permission).Error
}

// InitializeSystemPermissions 初始化系统权限
func (s *PermissionService) InitializeSystemPermissions() error {
	// 定义系统权限
	systemPermissions := []models.Permission{
		// 系统权限
		{
			ID:          uuid.New().String(),
			Name:        models.PermissionSystemManage,
			DisplayName: "系统管理",
			Description: "管理系统配置和全局设置",
			Resource:    "system",
			Action:      "manage",
			IsSystem:    true,
		},
		{
			ID:          uuid.New().String(),
			Name:        models.PermissionSystemConfig,
			DisplayName: "系统配置",
			Description: "修改系统配置参数",
			Resource:    "system",
			Action:      "config",
			IsSystem:    true,
		},
		// 用户权限
		{
			ID:          uuid.New().String(),
			Name:        models.PermissionUserRead,
			DisplayName: "查看用户",
			Description: "查看用户信息和列表",
			Resource:    "user",
			Action:      "read",
			IsSystem:    true,
		},
		{
			ID:          uuid.New().String(),
			Name:        models.PermissionUserWrite,
			DisplayName: "管理用户",
			Description: "创建、修改用户信息",
			Resource:    "user",
			Action:      "write",
			IsSystem:    true,
		},
		{
			ID:          uuid.New().String(),
			Name:        models.PermissionUserDelete,
			DisplayName: "删除用户",
			Description: "删除用户账户",
			Resource:    "user",
			Action:      "delete",
			IsSystem:    true,
		},
		// 节点权限
		{
			ID:          uuid.New().String(),
			Name:        models.PermissionNodeRead,
			DisplayName: "查看节点",
			Description: "查看节点信息和状态",
			Resource:    "node",
			Action:      "read",
			IsSystem:    true,
		},
		{
			ID:          uuid.New().String(),
			Name:        models.PermissionNodeWrite,
			DisplayName: "管理节点",
			Description: "创建、修改节点配置",
			Resource:    "node",
			Action:      "write",
			IsSystem:    true,
		},
		{
			ID:          uuid.New().String(),
			Name:        models.PermissionNodeDelete,
			DisplayName: "删除节点",
			Description: "删除节点配置",
			Resource:    "node",
			Action:      "delete",
			IsSystem:    true,
		},
		// 进程权限
		{
			ID:          uuid.New().String(),
			Name:        models.PermissionProcessRead,
			DisplayName: "查看进程",
			Description: "查看进程状态和信息",
			Resource:    "process",
			Action:      "read",
			IsSystem:    true,
		},
		{
			ID:          uuid.New().String(),
			Name:        models.PermissionProcessWrite,
			DisplayName: "管理进程",
			Description: "修改进程配置",
			Resource:    "process",
			Action:      "write",
			IsSystem:    true,
		},
		{
			ID:          uuid.New().String(),
			Name:        models.PermissionProcessExecute,
			DisplayName: "控制进程",
			Description: "启动、停止、重启进程",
			Resource:    "process",
			Action:      "execute",
			IsSystem:    true,
		},
		{
			ID:          uuid.New().String(),
			Name:        models.PermissionProcessDelete,
			DisplayName: "删除进程",
			Description: "删除进程配置",
			Resource:    "process",
			Action:      "delete",
			IsSystem:    true,
		},
		// 日志权限
		{
			ID:          uuid.New().String(),
			Name:        models.PermissionLogRead,
			DisplayName: "查看日志",
			Description: "查看系统和进程日志",
			Resource:    "log",
			Action:      "read",
			IsSystem:    true,
		},
		{
			ID:          uuid.New().String(),
			Name:        models.PermissionLogWrite,
			DisplayName: "管理日志",
			Description: "配置日志设置",
			Resource:    "log",
			Action:      "write",
			IsSystem:    true,
		},
		{
			ID:          uuid.New().String(),
			Name:        models.PermissionLogDelete,
			DisplayName: "删除日志",
			Description: "清理和删除日志文件",
			Resource:    "log",
			Action:      "delete",
			IsSystem:    true,
		},
		// 配置权限
		{
			ID:          uuid.New().String(),
			Name:        models.PermissionConfigRead,
			DisplayName: "查看配置",
			Description: "查看系统配置",
			Resource:    "config",
			Action:      "read",
			IsSystem:    true,
		},
		{
			ID:          uuid.New().String(),
			Name:        models.PermissionConfigWrite,
			DisplayName: "管理配置",
			Description: "修改系统配置",
			Resource:    "config",
			Action:      "write",
			IsSystem:    true,
		},
		{
			ID:          uuid.New().String(),
			Name:        models.PermissionConfigDelete,
			DisplayName: "删除配置",
			Description: "删除配置项",
			Resource:    "config",
			Action:      "delete",
			IsSystem:    true,
		},
	}

	// 创建系统权限
	for _, permission := range systemPermissions {
		var existingPermission models.Permission
		if err := s.db.Where("name = ?", permission.Name).First(&existingPermission).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := s.db.Create(&permission).Error; err != nil {
					return fmt.Errorf("创建系统权限 %s 失败: %v", permission.Name, err)
				}
			}
		}
	}

	return nil
}

// AssignDefaultPermissionsToRole 为角色分配默认权限
func (s *PermissionService) AssignDefaultPermissionsToRole(roleName string) error {
	roleService := NewRoleService(s.db)
	role, err := roleService.GetRoleByName(roleName)
	if err != nil {
		return err
	}

	var permissionNames []string

	// 根据角色分配不同的默认权限
	switch roleName {
	case models.RoleSuperAdmin:
		// 超级管理员拥有所有权限
		var permissions []models.Permission
		if err := s.db.Find(&permissions).Error; err != nil {
			return err
		}
		permissionIDs := make([]string, len(permissions))
		for i, p := range permissions {
			permissionIDs[i] = p.ID
		}
		return roleService.AssignPermissionsToRole(role.ID, permissionIDs)

	case models.RoleEnvironmentAdmin:
		permissionNames = []string{
			models.PermissionUserRead, models.PermissionUserWrite,
			models.PermissionNodeRead, models.PermissionNodeWrite,
			models.PermissionProcessRead, models.PermissionProcessWrite, models.PermissionProcessExecute,
			models.PermissionLogRead, models.PermissionLogWrite,
			models.PermissionConfigRead, models.PermissionConfigWrite,
		}

	case models.RoleNodeOperator:
		permissionNames = []string{
			models.PermissionNodeRead,
			models.PermissionProcessRead, models.PermissionProcessExecute,
			models.PermissionLogRead,
			models.PermissionConfigRead,
		}

	case models.RoleReadOnlyUser:
		permissionNames = []string{
			models.PermissionNodeRead,
			models.PermissionProcessRead,
			models.PermissionLogRead,
			models.PermissionConfigRead,
		}
	}

	// 获取权限ID
	var permissions []models.Permission
	if err := s.db.Where("name IN ?", permissionNames).Find(&permissions).Error; err != nil {
		return err
	}

	permissionIDs := make([]string, len(permissions))
	for i, p := range permissions {
		permissionIDs[i] = p.ID
	}

	return roleService.AssignPermissionsToRole(role.ID, permissionIDs)
}

package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"superview/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.RolePermission{},
		&models.NodeAccess{},
	)
	require.NoError(t, err)

	return db
}

// createTestUser 创建测试用户
func createTestUser(t *testing.T, db *gorm.DB, userID string) {
	user := &models.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
	}
	err := db.Create(user).Error
	require.NoError(t, err)
}

// createTestRole 创建测试角色
func createTestRole(t *testing.T, db *gorm.DB, roleID, roleName string) {
	role := &models.Role{
		ID:          roleID,
		Name:        roleName,
		DisplayName: roleName,
	}
	err := db.Create(role).Error
	require.NoError(t, err)
}

// createTestPermission 创建测试权限
func createTestPermission(t *testing.T, db *gorm.DB, permID, permName string) {
	perm := &models.Permission{
		ID:          permID,
		Name:        permName,
		DisplayName: permName,
	}
	err := db.Create(perm).Error
	require.NoError(t, err)
}

// assignRoleToUser 分配角色给用户
func assignRoleToUser(t *testing.T, db *gorm.DB, userID, roleID string) {
	userRole := &models.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	err := db.Create(userRole).Error
	require.NoError(t, err)
}

// assignPermissionToRole 分配权限给角色
func assignPermissionToRole(t *testing.T, db *gorm.DB, roleID, permID string) {
	rolePerm := &models.RolePermission{
		RoleID:       roleID,
		PermissionID: permID,
	}
	err := db.Create(rolePerm).Error
	require.NoError(t, err)
}

// 属性 18：权限验证
// 验证需求：6.5
func TestPermissionVerificationProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("users without permission are denied", prop.ForAll(
		func(userID string, permission string) bool {
			if userID == "" {
				userID = "user1"
			}
			if permission == "" {
				permission = "test:read"
			}

			db := setupTestDB(t)
			createTestUser(t, db, userID)

			pc := NewPermissionChecker(db)
			hasPermission, err := pc.CheckPermission(userID, permission)

			return err == nil && !hasPermission
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.Property("users with permission are allowed", prop.ForAll(
		func(userID string, permission string) bool {
			if userID == "" {
				userID = "user1"
			}
			if permission == "" {
				permission = "test:read"
			}

			db := setupTestDB(t)
			createTestUser(t, db, userID)
			createTestRole(t, db, "role1", "test_role")
			createTestPermission(t, db, "perm1", permission)
			assignRoleToUser(t, db, userID, "role1")
			assignPermissionToRole(t, db, "role1", "perm1")

			pc := NewPermissionChecker(db)
			hasPermission, err := pc.CheckPermission(userID, permission)

			return err == nil && hasPermission
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.Property("super admin has all permissions", prop.ForAll(
		func(permission string) bool {
			if permission == "" {
				permission = "any:permission"
			}

			db := setupTestDB(t)
			userID := "admin1"
			createTestUser(t, db, userID)
			createTestRole(t, db, "superadmin", models.RoleSuperAdmin)
			assignRoleToUser(t, db, userID, "superadmin")

			pc := NewPermissionChecker(db)
			hasPermission, err := pc.CheckPermission(userID, permission)

			return err == nil && hasPermission
		},
		gen.AlphaString(),
	))

	properties.Property("role assignment is transitive", prop.ForAll(
		func(userID string) bool {
			if userID == "" {
				userID = "user1"
			}

			db := setupTestDB(t)
			createTestUser(t, db, userID)
			createTestRole(t, db, "role1", "test_role")
			createTestPermission(t, db, "perm1", "test:read")
			assignRoleToUser(t, db, userID, "role1")
			assignPermissionToRole(t, db, "role1", "perm1")

			pc := NewPermissionChecker(db)

			// 用户有角色
			hasRole, err1 := pc.CheckRole(userID, "test_role")
			// 角色有权限，所以用户有权限
			hasPermission, err2 := pc.CheckPermission(userID, "test:read")

			return err1 == nil && err2 == nil && hasRole && hasPermission
		},
		gen.AlphaString(),
	))

	properties.Property("permission check is consistent", prop.ForAll(
		func(userID string, permission string) bool {
			if userID == "" {
				userID = "user1"
			}
			if permission == "" {
				permission = "test:read"
			}

			db := setupTestDB(t)
			createTestUser(t, db, userID)
			createTestRole(t, db, "role1", "test_role")
			createTestPermission(t, db, "perm1", permission)
			assignRoleToUser(t, db, userID, "role1")
			assignPermissionToRole(t, db, "role1", "perm1")

			pc := NewPermissionChecker(db)

			// 多次检查应该返回相同结果
			result1, err1 := pc.CheckPermission(userID, permission)
			result2, err2 := pc.CheckPermission(userID, permission)

			return err1 == nil && err2 == nil && result1 == result2 && result1 == true
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

// TestPermissionMiddleware 测试权限中间件
func TestPermissionMiddleware(t *testing.T) {
	db := setupTestDB(t)
	pc := NewPermissionChecker(db)

	// 创建测试数据
	userID := "user1"
	createTestUser(t, db, userID)
	createTestRole(t, db, "role1", "test_role")
	createTestPermission(t, db, "perm1", "test:read")
	assignRoleToUser(t, db, userID, "role1")
	assignPermissionToRole(t, db, "role1", "perm1")

	// 测试有权限的情况
	t.Run("with permission", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("userID", userID)
			c.Next()
		})
		router.GET("/test", pc.RequirePermission("test:read"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 测试无权限的情况
	t.Run("without permission", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("userID", userID)
			c.Next()
		})
		router.GET("/test", pc.RequirePermission("test:write"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	// 测试未认证的情况
	t.Run("not authenticated", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", pc.RequirePermission("test:read"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestRoleMiddleware 测试角色中间件
func TestRoleMiddleware(t *testing.T) {
	db := setupTestDB(t)
	pc := NewPermissionChecker(db)

	userID := "user1"
	createTestUser(t, db, userID)
	createTestRole(t, db, "role1", "admin")
	assignRoleToUser(t, db, userID, "role1")

	// 测试有角色的情况
	t.Run("with role", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("userID", userID)
			c.Next()
		})
		router.GET("/test", pc.RequireRole("admin"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 测试无角色的情况
	t.Run("without role", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("userID", userID)
			c.Next()
		})
		router.GET("/test", pc.RequireRole("superadmin"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// TestAnyPermissionMiddleware 测试任意权限中间件
func TestAnyPermissionMiddleware(t *testing.T) {
	db := setupTestDB(t)
	pc := NewPermissionChecker(db)

	userID := "user1"
	createTestUser(t, db, userID)
	createTestRole(t, db, "role1", "test_role")
	createTestPermission(t, db, "perm1", "test:read")
	assignRoleToUser(t, db, userID, "role1")
	assignPermissionToRole(t, db, "role1", "perm1")

	// 测试有任意一个权限的情况
	t.Run("with any permission", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("userID", userID)
			c.Next()
		})
		router.GET("/test", pc.RequireAnyPermission("test:read", "test:write"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 测试没有任何权限的情况
	t.Run("without any permission", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("userID", userID)
			c.Next()
		})
		router.GET("/test", pc.RequireAnyPermission("test:write", "test:delete"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// TestAllPermissionsMiddleware 测试所有权限中间件
func TestAllPermissionsMiddleware(t *testing.T) {
	db := setupTestDB(t)
	pc := NewPermissionChecker(db)

	userID := "user1"
	createTestUser(t, db, userID)
	createTestRole(t, db, "role1", "test_role")
	createTestPermission(t, db, "perm1", "test:read")
	createTestPermission(t, db, "perm2", "test:write")
	assignRoleToUser(t, db, userID, "role1")
	assignPermissionToRole(t, db, "role1", "perm1")
	assignPermissionToRole(t, db, "role1", "perm2")

	// 测试有所有权限的情况
	t.Run("with all permissions", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("userID", userID)
			c.Next()
		})
		router.GET("/test", pc.RequireAllPermissions("test:read", "test:write"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 测试缺少某个权限的情况
	t.Run("missing one permission", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("userID", userID)
			c.Next()
		})
		router.GET("/test", pc.RequireAllPermissions("test:read", "test:delete"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// TestSuperAdminPermissions 测试超级管理员权限
func TestSuperAdminPermissions(t *testing.T) {
	db := setupTestDB(t)
	pc := NewPermissionChecker(db)

	userID := "admin1"
	createTestUser(t, db, userID)
	createTestRole(t, db, "superadmin", models.RoleSuperAdmin)
	assignRoleToUser(t, db, userID, "superadmin")

	// 超级管理员应该有任何权限
	permissions := []string{
		"test:read",
		"test:write",
		"test:delete",
		"system:manage",
		"user:delete",
	}

	for _, perm := range permissions {
		hasPermission, err := pc.CheckPermission(userID, perm)
		assert.NoError(t, err)
		assert.True(t, hasPermission, "super admin should have permission: %s", perm)
	}
}

// TestGetUserPermissions 测试获取用户权限
func TestGetUserPermissions(t *testing.T) {
	db := setupTestDB(t)
	pc := NewPermissionChecker(db)

	userID := "user1"
	createTestUser(t, db, userID)
	createTestRole(t, db, "role1", "test_role")
	createTestPermission(t, db, "perm1", "test:read")
	createTestPermission(t, db, "perm2", "test:write")
	assignRoleToUser(t, db, userID, "role1")
	assignPermissionToRole(t, db, "role1", "perm1")
	assignPermissionToRole(t, db, "role1", "perm2")

	permissions, err := pc.GetUserPermissions(userID)
	assert.NoError(t, err)
	assert.Len(t, permissions, 2)
	assert.Contains(t, permissions, "test:read")
	assert.Contains(t, permissions, "test:write")
}

// TestGetUserRoles 测试获取用户角色
func TestGetUserRoles(t *testing.T) {
	db := setupTestDB(t)
	pc := NewPermissionChecker(db)

	userID := "user1"
	createTestUser(t, db, userID)
	createTestRole(t, db, "role1", "admin")
	createTestRole(t, db, "role2", "operator")
	assignRoleToUser(t, db, userID, "role1")
	assignRoleToUser(t, db, userID, "role2")

	roles, err := pc.GetUserRoles(userID)
	assert.NoError(t, err)
	assert.Len(t, roles, 2)
	assert.Contains(t, roles, "admin")
	assert.Contains(t, roles, "operator")
}

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go-cesi/internal/models"
	"go-cesi/internal/validation"
	"gorm.io/gorm"
)

type UserAPI struct {
	db *gorm.DB
}

func NewUserAPI(db *gorm.DB) *UserAPI {
	return &UserAPI{db: db}
}

// 检查当前用户是否为管理员
func (u *UserAPI) checkAdmin(c *gin.Context) bool {
	var currentUser models.User
	userID := c.GetString("user_id")
	if err := u.db.First(&currentUser, userID).Error; err != nil {
		return false
	}
	return currentUser.IsAdmin
}

func (u *UserAPI) ListUsers(c *gin.Context) {
	// 只有管理员可以列出所有用户
	if !u.checkAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "需要管理员权限",
		})
		return
	}

	var users []models.User
	if err := u.db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "获取用户列表失败",
		})
		return
	}

	// 不返回密码字段
	userList := make([]gin.H, len(users))
	for i, user := range users {
		userList[i] = gin.H{
			"username": user.Username,
			"is_admin": user.IsAdmin,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   userList,
	})
}

func (u *UserAPI) CreateUser(c *gin.Context) {
	// 只有管理员可以创建用户
	if !u.checkAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "需要管理员权限",
		})
		return
	}

	type createUserRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		IsAdmin  bool   `json:"is_admin"`
	}

	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "请求格式无效",
		})
		return
	}

	// 输入验证
	validator := validation.NewValidator()
	validator.ValidateRequired("username", req.Username)
	validator.ValidateLength("username", req.Username, 3, 50)
	validator.ValidateAlphanumeric("username", req.Username)
	validator.ValidateNoSQLInjection("username", req.Username)
	
	validator.ValidateRequired("password", req.Password)
	validator.ValidateLength("password", req.Password, 6, 100)
	validator.ValidateNoSQLInjection("password", req.Password)
	
	if validator.HasErrors() {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "输入验证失败",
			"errors":  validator.Errors(),
		})
		return
	}
	
	// 清理输入
	req.Username = validation.SanitizeInput(req.Username)
	req.Password = validation.SanitizeInput(req.Password)

	// 检查用户名是否已存在
	var existingUser models.User
	if err := u.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": "用户名已存在",
		})
		return
	}

	user := models.User{
		Username: req.Username,
		IsAdmin:  req.IsAdmin,
	}

	if err := user.SetPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "设置密码失败",
		})
		return
	}

	if err := u.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "创建用户失败",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data": gin.H{
			"username": user.Username,
			"is_admin": user.IsAdmin,
		},
	})
}

func (u *UserAPI) ChangePassword(c *gin.Context) {
	username := c.Param("username")

	type changePasswordRequest struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "请求格式无效",
		})
		return
	}

	// 获取当前用户
	var currentUser models.User
	currentUserID := c.GetString("user_id")
	if err := u.db.First(&currentUser, currentUserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "当前用户不存在",
		})
		return
	}

	// 获取目标用户
	var targetUser models.User
	if err := u.db.Where("username = ?", username).First(&targetUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "目标用户不存在",
		})
		return
	}

	// 只有管理员或用户本人可以修改密码
	if !currentUser.IsAdmin && currentUser.ID != targetUser.ID {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "无权修改其他用户的密码",
		})
		return
	}

	// 验证旧密码
	if !targetUser.VerifyPassword(req.OldPassword) {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "当前密码错误",
		})
		return
	}

	// 设置新密码
	if err := targetUser.SetPassword(req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "设置新密码失败",
		})
		return
	}

	// 保存更新
	if err := u.db.Save(&targetUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "更新密码失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "密码修改成功",
	})
}

func (u *UserAPI) DeleteUser(c *gin.Context) {
	// 只有管理员可以删除用户
	if !u.checkAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "需要管理员权限",
		})
		return
	}

	username := c.Param("username")

	// 检查用户是否存在
	var user models.User
	if err := u.db.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "用户不存在",
		})
		return
	}

	// 不允许删除自己
	currentUserID := c.GetString("user_id")
	if user.ID == currentUserID {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "不能删除自己",
		})
		return
	}

	if err := u.db.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "删除用户失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "用户删除成功",
	})
}

// GetProfile 获取用户个人资料
func (u *UserAPI) GetProfile(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	var user models.User
	if err := u.db.First(&user, currentUserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"full_name":  user.FullName,
			"is_admin":   user.IsAdmin,
			"is_active":  user.IsActive,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	})
}

// UpdateProfile 更新用户个人资料
func (u *UserAPI) UpdateProfile(c *gin.Context) {
	type updateProfileRequest struct {
		Email    string `json:"email"`
		FullName string `json:"full_name"`
	}

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "请求格式无效",
		})
		return
	}

	currentUserID := c.GetString("user_id")
	var user models.User
	if err := u.db.First(&user, currentUserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "用户不存在",
		})
		return
	}

	// 更新用户信息（这里假设User模型有Email和FullName字段）
	// 如果模型中没有这些字段，可以根据实际需求调整
	updateData := map[string]interface{}{}
	if req.Email != "" {
		updateData["email"] = req.Email
	}
	if req.FullName != "" {
		updateData["full_name"] = req.FullName
	}

	if len(updateData) > 0 {
		if err := u.db.Model(&user).Updates(updateData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "更新个人资料失败",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "个人资料更新成功",
		"data": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"full_name":  user.FullName,
			"is_admin":   user.IsAdmin,
			"is_active":  user.IsActive,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	})
}
package validation

import (
	"go-cesi/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
	IsAdmin  *bool  `json:"is_admin,omitempty"`
}

// Validate 验证创建用户请求
func (r *CreateUserRequest) Validate() ValidationErrors {
	var errors ValidationErrors
	
	// 清理输入
	r.Username = SanitizeString(r.Username)
	r.Email = SanitizeString(r.Email)
	
	// 验证用户名
	if err := ValidateUsername(r.Username); err != nil {
		if validationErr, ok := err.(ValidationError); ok {
			errors = append(errors, validationErr)
		}
	}
	
	// 验证密码
	if err := ValidatePassword(r.Password); err != nil {
		if validationErr, ok := err.(ValidationError); ok {
			errors = append(errors, validationErr)
		}
	}
	
	// 验证邮箱
	if err := ValidateEmail(r.Email); err != nil {
		if validationErr, ok := err.(ValidationError); ok {
			errors = append(errors, validationErr)
		}
	}
	
	return errors
}

// ToModel 转换为模型
func (r *CreateUserRequest) ToModel() (*models.User, error) {
	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(r.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	
	return &models.User{
		Username: r.Username,
		Password: string(hashedPassword),
		Email:    r.Email,
		IsAdmin:  r.IsAdmin != nil && *r.IsAdmin,
	}, nil
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
	IsAdmin  *bool   `json:"is_admin,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// Validate 验证更新用户请求
func (r *UpdateUserRequest) Validate() ValidationErrors {
	var errors ValidationErrors
	
	// 验证用户名
	if r.Username != nil {
		*r.Username = SanitizeString(*r.Username)
		if err := ValidateUsername(*r.Username); err != nil {
			if validationErr, ok := err.(ValidationError); ok {
				errors = append(errors, validationErr)
			}
		}
	}
	
	// 验证邮箱
	if r.Email != nil {
		*r.Email = SanitizeString(*r.Email)
		if err := ValidateEmail(*r.Email); err != nil {
			if validationErr, ok := err.(ValidationError); ok {
				errors = append(errors, validationErr)
			}
		}
	}
	
	return errors
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Validate 验证登录请求
func (r *LoginRequest) Validate() ValidationErrors {
	var errors ValidationErrors
	validator := NewValidator()
	
	// 清理输入
	r.Username = SanitizeString(r.Username)
	
	// 验证用户名
	validator.ValidateRequired("username", r.Username)
	
	// 验证密码
	validator.ValidateRequired("password", r.Password)
	
	// 转换验证错误
	if validator.HasErrors() {
		errors = append(errors, validator.Errors()...)
	}
	
	return errors
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// Validate 验证修改密码请求
func (r *ChangePasswordRequest) Validate() ValidationErrors {
	var errors ValidationErrors
	validator := NewValidator()
	
	// 验证旧密码
	validator.ValidateRequired("old_password", r.OldPassword)
	
	// 验证新密码
	validator.ValidatePassword("new_password", r.NewPassword)
	
	// 转换验证错误
	if validator.HasErrors() {
		errors = append(errors, validator.Errors()...)
	}
	
	return errors
}

// CreateNodeRequest 创建节点请求
type CreateNodeRequest struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	Environment string `json:"environment,omitempty"`
	Description string `json:"description,omitempty"`
}

// Validate 验证创建节点请求
func (r *CreateNodeRequest) Validate() ValidationErrors {
	var errors ValidationErrors
	
	// 清理输入
	r.Name = SanitizeString(r.Name)
	r.Host = SanitizeString(r.Host)
	r.Username = SanitizeString(r.Username)
	r.Environment = SanitizeString(r.Environment)
	r.Description = SanitizeString(r.Description)
	
	// 验证节点名称
	validator := NewValidator()
	validator.ValidateNodeName("name", r.Name)
	
	// 转换验证错误
	if validator.HasErrors() {
		errors = append(errors, validator.Errors()...)
	}
	
	// 验证主机地址
	validator.ValidateHost("host", r.Host)
	
	// 验证端口
	validator.ValidatePort("port", r.Port)
	
	// 验证描述长度
	validator.ValidateMaxLength("description", r.Description, 500)
	
	return errors
}

// ToModel 转换为模型
func (r *CreateNodeRequest) ToModel() *models.Node {
	return &models.Node{
		Name:        r.Name,
		Host:        r.Host,
		Port:        r.Port,
		Username:    r.Username,
		Password:    r.Password,
		Environment: r.Environment,
		Description: r.Description,
		Status:      "inactive", // 默认状态
	}
}

// UpdateNodeRequest 更新节点请求
type UpdateNodeRequest struct {
	Name        *string `json:"name,omitempty"`
	Host        *string `json:"host,omitempty"`
	Port        *int    `json:"port,omitempty"`
	Username    *string `json:"username,omitempty"`
	Password    *string `json:"password,omitempty"`
	Environment *string `json:"environment,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
}

// Validate 验证更新节点请求
func (r *UpdateNodeRequest) Validate() ValidationErrors {
	var errors ValidationErrors
	
	validator := NewValidator()
	
	// 验证节点名称
	if r.Name != nil {
		*r.Name = SanitizeString(*r.Name)
		validator.ValidateNodeName("name", *r.Name)
	}
	
	// 验证主机地址
	if r.Host != nil {
		*r.Host = SanitizeString(*r.Host)
		validator.ValidateHost("host", *r.Host)
	}
	
	// 验证端口
	if r.Port != nil {
		validator.ValidatePort("port", *r.Port)
	}
	
	// 验证描述长度
	if r.Description != nil {
		*r.Description = SanitizeString(*r.Description)
		validator.ValidateMaxLength("description", *r.Description, 500)
	}
	
	// 转换验证错误
	if validator.HasErrors() {
		errors = append(errors, validator.Errors()...)
	}
	
	// 验证状态
	if r.Status != nil {
		*r.Status = SanitizeString(*r.Status)
		validStatuses := []string{"active", "inactive", "error"}
		valid := false
		for _, status := range validStatuses {
			if *r.Status == status {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, ValidationError{
				Field:   "status",
				Message: "must be one of: active, inactive, error",
			})
		}
	}
	
	return errors
}

// PaginationRequest 分页请求
type PaginationRequest struct {
	Page     int `json:"page,omitempty" form:"page"`
	PageSize int `json:"page_size,omitempty" form:"page_size"`
}

// Validate 验证分页请求
func (r *PaginationRequest) Validate() ValidationErrors {
	var errors ValidationErrors
	
	if r.Page < 1 {
		r.Page = 1
	}
	
	if r.PageSize < 1 {
		r.PageSize = 20
	} else if r.PageSize > 100 {
		errors = append(errors, ValidationError{
			Field:   "page_size",
			Message: "must be at most 100",
		})
	}
	
	return errors
}
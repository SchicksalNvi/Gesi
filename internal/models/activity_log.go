package models

import (
	"time"

	"gorm.io/gorm"
)

type ActivityLog struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 日志基本信息
	Level    string `json:"level" gorm:"size:20;not null"`     // INFO, WARNING, ERROR
	Message  string `json:"message" gorm:"type:text;not null"` // 日志消息
	Action   string `json:"action" gorm:"size:50;index"`       // 操作类型：start, stop, restart, login, logout等
	Resource string `json:"resource" gorm:"size:100"`          // 资源类型：process, node, user等
	Target   string `json:"target" gorm:"size:200"`            // 目标对象：进程名、节点名、用户名等

	// 用户信息
	UserID   string `json:"user_id" gorm:"size:50;index"`   // 操作用户ID
	Username string `json:"username" gorm:"size:100;index"` // 操作用户名

	// 请求信息
	IPAddress string `json:"ip_address" gorm:"size:45"`   // 客户端IP地址
	UserAgent string `json:"user_agent" gorm:"type:text"` // 用户代理

	// 额外信息
	Details  string `json:"details" gorm:"type:text"`                // 详细信息（JSON格式）
	Status   string `json:"status" gorm:"size:20;default:'success'"` // success, error, warning
	Duration int64  `json:"duration"`                                // 操作耗时（毫秒）
}

// TableName 指定表名
func (ActivityLog) TableName() string {
	return "activity_logs"
}

// LogLevel 定义日志级别常量
const (
	LogLevelInfo    = "INFO"
	LogLevelWarning = "WARNING"
	LogLevelError   = "ERROR"
)

// LogAction 定义操作类型常量
const (
	ActionLogin            = "login"
	ActionLogout           = "logout"
	ActionStartProcess     = "start_process"
	ActionStopProcess      = "stop_process"
	ActionRestartProcess   = "restart_process"
	ActionStartGroup       = "start_group"
	ActionStopGroup        = "stop_group"
	ActionRestartGroup     = "restart_group"
	ActionCreateUser       = "create_user"
	ActionDeleteUser       = "delete_user"
	ActionChangePassword   = "change_password"
	ActionViewLogs         = "view_logs"
	ActionViewNodes        = "view_nodes"
	ActionViewGroups       = "view_groups"
	ActionViewEnvironments = "view_environments"
)

// LogResource 定义资源类型常量
const (
	ResourceProcess     = "process"
	ResourceGroup       = "group"
	ResourceNode        = "node"
	ResourceUser        = "user"
	ResourceAuth        = "auth"
	ResourceSystem      = "system"
	ResourceEnvironment = "environment"
)

// LogStatus 定义状态常量
const (
	StatusSuccess = "success"
	StatusError   = "error"
	StatusWarning = "warning"
)

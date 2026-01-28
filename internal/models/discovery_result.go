package models

import (
	"time"

	"gorm.io/gorm"
)

// Discovery result status constants
const (
	ResultStatusSuccess           = "success"
	ResultStatusTimeout           = "timeout"
	ResultStatusConnectionRefused = "connection_refused"
	ResultStatusAuthFailed        = "auth_failed"
	ResultStatusError             = "error"
)

// DiscoveryResult represents the outcome of probing a single IP address.
// Each result is linked to a parent DiscoveryTask.
type DiscoveryResult struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"not null" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index:idx_discovery_result_deleted_at" json:"-"`

	TaskID uint   `gorm:"index:idx_discovery_result_task_id;not null" json:"task_id"`
	IP     string `gorm:"size:50;not null" json:"ip"`
	Port   int    `gorm:"not null;check:port > 0 AND port <= 65535" json:"port"`

	Status string `gorm:"size:20;not null" json:"status"`
	// success, timeout, connection_refused, auth_failed, error

	NodeID   *uint  `json:"node_id,omitempty"`                       // If registered
	NodeName string `gorm:"size:100" json:"node_name,omitempty"`     // Generated name
	Version  string `gorm:"size:50" json:"version,omitempty"`        // Supervisor version
	ErrorMsg string `gorm:"size:500" json:"error_msg,omitempty"`     // Error details

	Duration int64 `json:"duration_ms"` // Probe duration in milliseconds
}

// IsSuccess returns true if the probe succeeded.
func (r *DiscoveryResult) IsSuccess() bool {
	return r.Status == ResultStatusSuccess
}

// IsTimeout returns true if the probe timed out.
func (r *DiscoveryResult) IsTimeout() bool {
	return r.Status == ResultStatusTimeout
}

// IsConnectionRefused returns true if the connection was refused.
func (r *DiscoveryResult) IsConnectionRefused() bool {
	return r.Status == ResultStatusConnectionRefused
}

// IsAuthFailed returns true if authentication failed.
func (r *DiscoveryResult) IsAuthFailed() bool {
	return r.Status == ResultStatusAuthFailed
}

// IsError returns true if an error occurred during probing.
func (r *DiscoveryResult) IsError() bool {
	return r.Status == ResultStatusError
}

// IsFailed returns true if the probe failed for any reason.
func (r *DiscoveryResult) IsFailed() bool {
	return !r.IsSuccess()
}

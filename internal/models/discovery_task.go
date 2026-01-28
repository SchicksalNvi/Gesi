package models

import (
	"time"

	"gorm.io/gorm"
)

// Discovery task status constants
const (
	DiscoveryStatusPending   = "pending"
	DiscoveryStatusRunning   = "running"
	DiscoveryStatusCompleted = "completed"
	DiscoveryStatusCancelled = "cancelled"
	DiscoveryStatusFailed    = "failed"
)

// DiscoveryTask represents a network discovery scan operation.
// It tracks the progress and results of scanning a CIDR range for Supervisor nodes.
type DiscoveryTask struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"not null;index:idx_discovery_task_created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index:idx_discovery_task_deleted_at" json:"-"`

	CIDR     string `gorm:"size:50;not null" json:"cidr"`
	Port     int    `gorm:"not null;check:port > 0 AND port <= 65535" json:"port"`
	Username string `gorm:"size:50" json:"username"`
	// Password NOT stored - security requirement

	Status string `gorm:"size:20;not null;default:'pending';index:idx_discovery_task_status" json:"status"`

	TotalIPs   int `gorm:"not null;default:0" json:"total_ips"`
	ScannedIPs int `gorm:"not null;default:0" json:"scanned_ips"`
	FoundNodes int `gorm:"not null;default:0" json:"found_nodes"`
	FailedIPs  int `gorm:"not null;default:0" json:"failed_ips"`

	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	ErrorMsg    string     `gorm:"size:500" json:"error_msg,omitempty"`

	CreatedBy string `gorm:"size:100;not null" json:"created_by"`
}

// IsPending returns true if the task is waiting to start.
func (t *DiscoveryTask) IsPending() bool {
	return t.Status == DiscoveryStatusPending
}

// IsRunning returns true if the task is currently scanning.
func (t *DiscoveryTask) IsRunning() bool {
	return t.Status == DiscoveryStatusRunning
}

// IsCompleted returns true if the task finished successfully.
func (t *DiscoveryTask) IsCompleted() bool {
	return t.Status == DiscoveryStatusCompleted
}

// IsCancelled returns true if the task was cancelled by user.
func (t *DiscoveryTask) IsCancelled() bool {
	return t.Status == DiscoveryStatusCancelled
}

// IsFailed returns true if the task failed with an error.
func (t *DiscoveryTask) IsFailed() bool {
	return t.Status == DiscoveryStatusFailed
}

// IsTerminal returns true if the task is in a final state (completed, cancelled, or failed).
func (t *DiscoveryTask) IsTerminal() bool {
	return t.IsCompleted() || t.IsCancelled() || t.IsFailed()
}

// Progress returns the scan progress as a percentage (0-100).
func (t *DiscoveryTask) Progress() float64 {
	if t.TotalIPs == 0 {
		return 0
	}
	return float64(t.ScannedIPs) / float64(t.TotalIPs) * 100
}

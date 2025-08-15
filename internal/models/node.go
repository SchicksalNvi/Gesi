package models

import (
	"fmt"
	"time"
	"gorm.io/gorm"
)

type Node struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"not null;index:idx_node_created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index:idx_node_deleted_at" json:"-"`
	Name        string `gorm:"size:100;not null;uniqueIndex:idx_node_name" json:"name" validate:"required,min=1,max=100"`
	Host        string `gorm:"size:100;not null;index:idx_host" json:"host" validate:"required,hostname_rfc1123|ip"`
	Port        int    `gorm:"not null;check:port > 0 AND port <= 65535" json:"port" validate:"required,min=1,max=65535"`
	Username    string `gorm:"size:50" json:"username" validate:"omitempty,max=50"`
	Password    string `gorm:"size:100" json:"-" validate:"omitempty,max=100"`
	Status      string `gorm:"size:20;default:'unknown';index:idx_status" json:"status" validate:"omitempty,oneof=unknown active inactive connected disconnected"`
	Environment string `gorm:"size:50;index:idx_environment" json:"environment" validate:"omitempty,max=50"`
	Description string `gorm:"size:500" json:"description" validate:"omitempty,max=500"`
}

func (n *Node) GetConnectionString() string {
	return fmt.Sprintf("%s:%d", n.Host, n.Port)
}

func (n *Node) IsActive() bool {
	return n.Status == "active"
}

func (n *Node) IsConnected() bool {
	return n.Status == "connected"
}

package repository

import (
	"go-cesi/internal/errors"
	"go-cesi/internal/models"
	"gorm.io/gorm"
)

type nodeRepository struct {
	db *gorm.DB
}

// NewNodeRepository 创建节点仓库实例
func NewNodeRepository(db *gorm.DB) NodeRepository {
	return &nodeRepository{db: db}
}

// Create 创建节点
func (r *nodeRepository) Create(node *models.Node) error {
	if err := r.db.Create(node).Error; err != nil {
		return errors.NewDatabaseError("create node", err)
	}
	return nil
}

// GetByID 根据ID获取节点
func (r *nodeRepository) GetByID(id uint) (*models.Node, error) {
	var node models.Node
	if err := r.db.Where("id = ?", id).First(&node).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("node", string(rune(id)))
		}
		return nil, errors.NewDatabaseError("get node by id", err)
	}
	return &node, nil
}

// GetByName 根据名称获取节点
func (r *nodeRepository) GetByName(name string) (*models.Node, error) {
	var node models.Node
	if err := r.db.Where("name = ?", name).First(&node).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("node", name)
		}
		return nil, errors.NewDatabaseError("get node by name", err)
	}
	return &node, nil
}

// Update 更新节点
func (r *nodeRepository) Update(node *models.Node) error {
	if err := r.db.Save(node).Error; err != nil {
		return errors.NewDatabaseError("update node", err)
	}
	return nil
}

// Delete 删除节点
func (r *nodeRepository) Delete(id uint) error {
	if err := r.db.Where("id = ?", id).Delete(&models.Node{}).Error; err != nil {
		return errors.NewDatabaseError("delete node", err)
	}
	return nil
}

// List 获取节点列表
func (r *nodeRepository) List(offset, limit int) ([]*models.Node, int64, error) {
	var nodes []*models.Node
	var total int64
	
	// 获取总数
	if err := r.db.Model(&models.Node{}).Count(&total).Error; err != nil {
		return nil, 0, errors.NewDatabaseError("count nodes", err)
	}
	
	// 获取分页数据
	if err := r.db.Offset(offset).Limit(limit).Find(&nodes).Error; err != nil {
		return nil, 0, errors.NewDatabaseError("list nodes", err)
	}
	
	return nodes, total, nil
}

// GetByStatus 根据状态获取节点列表
func (r *nodeRepository) GetByStatus(status string) ([]*models.Node, error) {
	var nodes []*models.Node
	if err := r.db.Where("status = ?", status).Find(&nodes).Error; err != nil {
		return nil, errors.NewDatabaseError("get nodes by status", err)
	}
	return nodes, nil
}

// ExistsByName 检查节点名称是否存在
func (r *nodeRepository) ExistsByName(name string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Node{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return false, errors.NewDatabaseError("check node name exists", err)
	}
	return count > 0, nil
}
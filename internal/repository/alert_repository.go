package repository

import (
	"go-cesi/internal/errors"
	"go-cesi/internal/models"
	"gorm.io/gorm"
)

type alertRepository struct {
	db *gorm.DB
}

// NewAlertRepository 创建告警仓库实例
func NewAlertRepository(db *gorm.DB) AlertRepository {
	return &alertRepository{db: db}
}

// CreateRule 创建告警规则
func (r *alertRepository) CreateRule(rule *models.AlertRule) error {
	if err := r.db.Create(rule).Error; err != nil {
		return errors.NewDatabaseError("create alert rule", err)
	}
	return nil
}

// GetRuleByID 根据ID获取告警规则
func (r *alertRepository) GetRuleByID(id uint) (*models.AlertRule, error) {
	var rule models.AlertRule
	if err := r.db.Preload("NotificationChannels").Where("id = ?", id).First(&rule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("alert rule", string(rune(id)))
		}
		return nil, errors.NewDatabaseError("get alert rule by id", err)
	}
	return &rule, nil
}

// UpdateRule 更新告警规则
func (r *alertRepository) UpdateRule(rule *models.AlertRule) error {
	if err := r.db.Save(rule).Error; err != nil {
		return errors.NewDatabaseError("update alert rule", err)
	}
	return nil
}

// DeleteRule 删除告警规则
func (r *alertRepository) DeleteRule(id uint) error {
	if err := r.db.Where("id = ?", id).Delete(&models.AlertRule{}).Error; err != nil {
		return errors.NewDatabaseError("delete alert rule", err)
	}
	return nil
}

// ListRules 获取告警规则列表
func (r *alertRepository) ListRules(offset, limit int) ([]*models.AlertRule, int64, error) {
	var rules []*models.AlertRule
	var total int64

	// 获取总数
	if err := r.db.Model(&models.AlertRule{}).Count(&total).Error; err != nil {
		return nil, 0, errors.NewDatabaseError("count alert rules", err)
	}

	// 获取分页数据
	if err := r.db.Preload("NotificationChannels").Offset(offset).Limit(limit).Find(&rules).Error; err != nil {
		return nil, 0, errors.NewDatabaseError("list alert rules", err)
	}

	return rules, total, nil
}

// GetActiveRules 获取活跃的告警规则
func (r *alertRepository) GetActiveRules() ([]*models.AlertRule, error) {
	var rules []*models.AlertRule
	if err := r.db.Preload("NotificationChannels").Where("is_enabled = ?", true).Find(&rules).Error; err != nil {
		return nil, errors.NewDatabaseError("get active alert rules", err)
	}
	return rules, nil
}

// CreateAlert 创建告警
func (r *alertRepository) CreateAlert(alert *models.Alert) error {
	if err := r.db.Create(alert).Error; err != nil {
		return errors.NewDatabaseError("create alert", err)
	}
	return nil
}

// GetAlertByID 根据ID获取告警
func (r *alertRepository) GetAlertByID(id uint) (*models.Alert, error) {
	var alert models.Alert
	if err := r.db.Preload("Rule").Where("id = ?", id).First(&alert).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("alert", string(rune(id)))
		}
		return nil, errors.NewDatabaseError("get alert by id", err)
	}
	return &alert, nil
}

// UpdateAlert 更新告警
func (r *alertRepository) UpdateAlert(alert *models.Alert) error {
	if err := r.db.Save(alert).Error; err != nil {
		return errors.NewDatabaseError("update alert", err)
	}
	return nil
}

// DeleteAlert 删除告警
func (r *alertRepository) DeleteAlert(id uint) error {
	if err := r.db.Where("id = ?", id).Delete(&models.Alert{}).Error; err != nil {
		return errors.NewDatabaseError("delete alert", err)
	}
	return nil
}

// ListAlerts 获取告警列表
func (r *alertRepository) ListAlerts(offset, limit int) ([]*models.Alert, int64, error) {
	var alerts []*models.Alert
	var total int64

	// 获取总数
	if err := r.db.Model(&models.Alert{}).Count(&total).Error; err != nil {
		return nil, 0, errors.NewDatabaseError("count alerts", err)
	}

	// 获取分页数据
	if err := r.db.Preload("Rule").Offset(offset).Limit(limit).Find(&alerts).Error; err != nil {
		return nil, 0, errors.NewDatabaseError("list alerts", err)
	}

	return alerts, total, nil
}

// GetActiveAlerts 获取活跃的告警
func (r *alertRepository) GetActiveAlerts() ([]*models.Alert, error) {
	var alerts []*models.Alert
	if err := r.db.Preload("Rule").Where("status = ?", models.AlertStatusFiring).Find(&alerts).Error; err != nil {
		return nil, errors.NewDatabaseError("get active alerts", err)
	}
	return alerts, nil
}

// GetAlertsByRuleID 根据规则ID获取告警列表
func (r *alertRepository) GetAlertsByRuleID(ruleID uint) ([]*models.Alert, error) {
	var alerts []*models.Alert
	if err := r.db.Preload("Rule").Where("rule_id = ?", ruleID).Find(&alerts).Error; err != nil {
		return nil, errors.NewDatabaseError("get alerts by rule id", err)
	}
	return alerts, nil
}

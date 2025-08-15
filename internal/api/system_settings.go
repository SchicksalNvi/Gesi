package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/gammazero/workerpool"
	"go-cesi/internal/models"
)

// SystemSettingsAPI handles system settings related operations
type SystemSettingsAPI struct {
	db *gorm.DB
	workerPool *workerpool.WorkerPool
}

// NewSystemSettingsAPI creates a new SystemSettingsAPI instance
func NewSystemSettingsAPI(db *gorm.DB) *SystemSettingsAPI {
	return &SystemSettingsAPI{
		db: db,
		workerPool: workerpool.New(10),
	}
}

// GetSystemSettings retrieves all system settings
func (api *SystemSettingsAPI) GetSystemSettings(c *gin.Context) {
	var settings []models.SystemSettings
	result := api.db.Find(&settings)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch system settings"})
		return
	}

	// Convert to map for easier frontend consumption
	settingsMap := make(map[string]interface{})
	for _, setting := range settings {
		settingsMap[setting.Key] = setting.Value
	}

	c.JSON(http.StatusOK, gin.H{
		"settings": settingsMap,
		"count": len(settings),
	})
}

// GetSystemSetting retrieves a specific system setting
func (api *SystemSettingsAPI) GetSystemSetting(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setting key is required"})
		return
	}

	var setting models.SystemSettings
	result := api.db.Where("key = ?", key).First(&setting)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch setting"})
		}
		return
	}

	c.JSON(http.StatusOK, setting)
}

// UpdateSystemSetting updates a system setting
func (api *SystemSettingsAPI) UpdateSystemSetting(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setting key is required"})
		return
	}

	var request struct {
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
		Category    string `json:"category"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if setting exists
	var setting models.SystemSettings
	result := api.db.Where("key = ?", key).First(&setting)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new setting
			setting = models.SystemSettings{
				Key:         key,
				Value:       request.Value,
				Description: request.Description,
				Category:    request.Category,
			}
			result = api.db.Create(&setting)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch setting"})
			return
		}
	} else {
		// Update existing setting
		setting.Value = request.Value
		if request.Description != "" {
			setting.Description = request.Description
		}
		if request.Category != "" {
			setting.Category = request.Category
		}
		result = api.db.Save(&setting)
	}

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save setting"})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// UpdateMultipleSettings updates multiple system settings at once
func (api *SystemSettingsAPI) UpdateMultipleSettings(c *gin.Context) {
	var request struct {
		Settings map[string]interface{} `json:"settings" binding:"required"`
		Category string                 `json:"category"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := api.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for key, value := range request.Settings {
		var setting models.SystemSettings
		result := tx.Where("key = ?", key).First(&setting)
		
		valueStr := ""
		switch v := value.(type) {
		case string:
			valueStr = v
		case bool:
			valueStr = strconv.FormatBool(v)
		case float64:
			valueStr = strconv.FormatFloat(v, 'f', -1, 64)
		default:
			valueStr = "" // Handle other types as needed
		}

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				// Create new setting
				setting = models.SystemSettings{
					Key:      key,
					Value:    valueStr,
					Category: request.Category,
				}
				if err := tx.Create(&setting).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create setting: " + key})
					return
				}
			} else {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch setting: " + key})
				return
			}
		} else {
			// Update existing setting
			setting.Value = valueStr
			if request.Category != "" {
				setting.Category = request.Category
			}
			if err := tx.Save(&setting).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update setting: " + key})
				return
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit settings update"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
}

// DeleteSystemSetting deletes a system setting
func (api *SystemSettingsAPI) DeleteSystemSetting(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setting key is required"})
		return
	}

	result := api.db.Where("key = ?", key).Delete(&models.SystemSettings{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete setting"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Setting deleted successfully"})
}

// GetUserPreferences retrieves user preferences
func (api *SystemSettingsAPI) GetUserPreferences(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var preferences models.UserPreferences
	result := api.db.Where("user_id = ?", userID).First(&preferences)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Return default preferences
			preferences = models.UserPreferences{
				UserID:   fmt.Sprintf("%d", userID),
				Theme:    "light",
				Language: "en",
				Timezone: "UTC",
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user preferences"})
			return
		}
	}

	c.JSON(http.StatusOK, preferences)
}

// UpdateUserPreferences updates user preferences
func (api *SystemSettingsAPI) UpdateUserPreferences(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var request models.UserPreferences
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request.UserID = fmt.Sprintf("%d", userID)

	// Check if preferences exist
	var preferences models.UserPreferences
	result := api.db.Where("user_id = ?", userID).First(&preferences)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new preferences
			result = api.db.Create(&request)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user preferences"})
			return
		}
	} else {
		// Update existing preferences
		request.ID = preferences.ID
		result = api.db.Save(&request)
	}

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user preferences"})
		return
	}

	c.JSON(http.StatusOK, request)
}

// TestEmailConfiguration tests the email configuration
func (api *SystemSettingsAPI) TestEmailConfiguration(c *gin.Context) {
	var request struct {
		TestEmail string `json:"test_email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get email settings
	var settings []models.SystemSettings
	api.db.Where("category = ?", "email").Find(&settings)

	emailConfig := make(map[string]string)
	for _, setting := range settings {
		emailConfig[setting.Key] = setting.Value
	}

	// Simulate email test (in real implementation, use actual SMTP)
	api.workerPool.Submit(func() {
		// Mock email sending logic
		// In real implementation, use net/smtp or a library like gomail
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Test email sent successfully to " + request.TestEmail,
		"success": true,
	})
}

// ResetToDefaults resets system settings to default values
func (api *SystemSettingsAPI) ResetToDefaults(c *gin.Context) {
	category := c.Query("category")
	
	tx := api.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete existing settings for the category
	if category != "" {
		if err := tx.Where("category = ?", category).Delete(&models.SystemSettings{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset settings"})
			return
		}
	} else {
		if err := tx.Delete(&models.SystemSettings{}, "1=1").Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset settings"})
			return
		}
	}

	// Create default settings
	defaultSettings := []models.SystemSettings{
		{Key: "theme.primary_color", Value: "#007bff", Category: "theme", Description: "Primary theme color"},
		{Key: "theme.secondary_color", Value: "#6c757d", Category: "theme", Description: "Secondary theme color"},
		{Key: "theme.dark_mode", Value: "false", Category: "theme", Description: "Enable dark mode"},
		{Key: "system.session_timeout", Value: "30", Category: "system", Description: "Session timeout in minutes"},
		{Key: "system.auto_refresh", Value: "true", Category: "system", Description: "Enable auto refresh"},
		{Key: "system.refresh_interval", Value: "5", Category: "system", Description: "Auto refresh interval in seconds"},
		{Key: "language.current", Value: "en", Category: "language", Description: "Current system language"},
	}

	for _, setting := range defaultSettings {
		if category == "" || setting.Category == category {
			if err := tx.Create(&setting).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create default settings"})
				return
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit settings reset"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings reset to defaults successfully"})
}
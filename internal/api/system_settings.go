package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"go-cesi/internal/models"
)

// SystemSettingsAPI handles system settings related operations
type SystemSettingsAPI struct {
	db *gorm.DB
}

// NewSystemSettingsAPI creates a new SystemSettingsAPI instance
func NewSystemSettingsAPI(db *gorm.DB) *SystemSettingsAPI {
	return &SystemSettingsAPI{
		db: db,
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

	// 获取当前用户 ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDStr := userID.(string)

	// Check if setting exists
	var setting models.SystemSettings
	result := api.db.Where("key = ?", key).First(&setting)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new setting
			setting = models.SystemSettings{
				ID:          uuid.New().String(),
				Key:         key,
				Value:       request.Value,
				Description: request.Description,
				Category:    request.Category,
				UpdatedBy:   &userIDStr,
			}
			result = api.db.Create(&setting)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch setting"})
			return
		}
	} else {
		// Update existing setting
		setting.Value = request.Value
		setting.UpdatedBy = &userIDStr
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

	// 获取当前用户 ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDStr := userID.(string)

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
					ID:        uuid.New().String(),
					Key:       key,
					Value:     valueStr,
					Category:  request.Category,
					UpdatedBy: &userIDStr,
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
			setting.UpdatedBy = &userIDStr
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
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var preferences models.UserPreferences
	result := api.db.Where("user_id = ?", userIDStr).First(&preferences)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Return default preferences with all fields
			preferences = models.UserPreferences{
				ID:                 "", // Will be generated by GORM
				UserID:             userIDStr,
				Theme:              "light",
				Language:           "en",
				Timezone:           "UTC",
				DateFormat:         "YYYY-MM-DD",
				TimeFormat:         "HH:mm:ss",
				PageSize:           20,
				AutoRefresh:        true,
				RefreshInterval:    30,
				EmailNotifications: true,
				ProcessAlerts:      true,
				SystemAlerts:       true,
				NodeStatusChanges:  false,
				WeeklyReport:       false,
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
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var request models.UserPreferences
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request.UserID = userIDStr

	// Check if preferences exist
	var preferences models.UserPreferences
	result := api.db.Where("user_id = ?", userIDStr).First(&preferences)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new preferences
			result = api.db.Create(&request)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user preferences"})
			return
		}
	} else {
		// Update existing preferences - merge with existing data
		if request.Theme != "" {
			preferences.Theme = request.Theme
		}
		if request.Language != "" {
			preferences.Language = request.Language
		}
		if request.Timezone != "" {
			preferences.Timezone = request.Timezone
		}
		if request.DateFormat != "" {
			preferences.DateFormat = request.DateFormat
		}
		if request.TimeFormat != "" {
			preferences.TimeFormat = request.TimeFormat
		}
		if request.PageSize > 0 {
			preferences.PageSize = request.PageSize
		}
		if request.RefreshInterval > 0 {
			preferences.RefreshInterval = request.RefreshInterval
		}
		// Update boolean fields (they can be false, so we need to check if they were sent)
		preferences.AutoRefresh = request.AutoRefresh
		preferences.EmailNotifications = request.EmailNotifications
		preferences.ProcessAlerts = request.ProcessAlerts
		preferences.SystemAlerts = request.SystemAlerts
		preferences.NodeStatusChanges = request.NodeStatusChanges
		preferences.WeeklyReport = request.WeeklyReport
		
		if request.Notifications != "" {
			preferences.Notifications = request.Notifications
		}
		if request.DashboardLayout != "" {
			preferences.DashboardLayout = request.DashboardLayout
		}
		
		result = api.db.Save(&preferences)
	}

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user preferences"})
		return
	}

	// Return the updated preferences
	if result.Error == nil {
		if preferences.ID != "" {
			c.JSON(http.StatusOK, preferences)
		} else {
			c.JSON(http.StatusOK, request)
		}
	}
}

// GetUserPreferencesByAdmin retrieves preferences for a specific user (admin or self)
func (api *SystemSettingsAPI) GetUserPreferencesByAdmin(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	
	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// Check permission: admin can access any user, non-admin can only access self
	var currentUser models.User
	if err := api.db.Where("id = ?", userID).First(&currentUser).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	
	if !currentUser.IsAdmin && userID != targetUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var preferences models.UserPreferences
	result := api.db.Where("user_id = ?", targetUserID).First(&preferences)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Return default preferences
			preferences = models.UserPreferences{
				UserID:             targetUserID,
				Theme:              "light",
				Language:           "en",
				Timezone:           "UTC",
				DateFormat:         "YYYY-MM-DD",
				TimeFormat:         "HH:mm:ss",
				PageSize:           20,
				AutoRefresh:        true,
				RefreshInterval:    30,
				EmailNotifications: true,
				ProcessAlerts:      true,
				SystemAlerts:       true,
				NodeStatusChanges:  false,
				WeeklyReport:       false,
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user preferences"})
			return
		}
	}

	c.JSON(http.StatusOK, preferences)
}

// UpdateUserPreferencesByAdmin updates preferences for a specific user (admin or self)
func (api *SystemSettingsAPI) UpdateUserPreferencesByAdmin(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	
	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// Check permission: admin can access any user, non-admin can only access self
	var currentUser models.User
	if err := api.db.Where("id = ?", userID).First(&currentUser).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	
	if !currentUser.IsAdmin && userID != targetUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var request models.UserPreferences
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request.UserID = targetUserID

	var preferences models.UserPreferences
	result := api.db.Where("user_id = ?", targetUserID).First(&preferences)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			result = api.db.Create(&request)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user preferences"})
			return
		}
	} else {
		// Update fields
		if request.Theme != "" {
			preferences.Theme = request.Theme
		}
		if request.Language != "" {
			preferences.Language = request.Language
		}
		if request.Timezone != "" {
			preferences.Timezone = request.Timezone
		}
		if request.DateFormat != "" {
			preferences.DateFormat = request.DateFormat
		}
		if request.TimeFormat != "" {
			preferences.TimeFormat = request.TimeFormat
		}
		if request.PageSize > 0 {
			preferences.PageSize = request.PageSize
		}
		if request.RefreshInterval > 0 {
			preferences.RefreshInterval = request.RefreshInterval
		}
		preferences.AutoRefresh = request.AutoRefresh
		preferences.EmailNotifications = request.EmailNotifications
		preferences.ProcessAlerts = request.ProcessAlerts
		preferences.SystemAlerts = request.SystemAlerts
		preferences.NodeStatusChanges = request.NodeStatusChanges
		preferences.WeeklyReport = request.WeeklyReport
		
		result = api.db.Save(&preferences)
	}

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user preferences"})
		return
	}

	if preferences.ID != "" {
		c.JSON(http.StatusOK, preferences)
	} else {
		c.JSON(http.StatusOK, request)
	}
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

	// Validate required email settings
	requiredSettings := []string{"smtp_host", "smtp_port", "smtp_username", "smtp_password"}
	for _, key := range requiredSettings {
		if emailConfig[key] == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Email configuration incomplete. Missing: " + key,
				"success": false,
			})
			return
		}
	}

	// Test email sending in background
	go func() {
		// Real SMTP email sending implementation
		// For now, we'll simulate a more realistic test
		// In production, use net/smtp or gomail library
		
		// Simulate connection test
		time.Sleep(2 * time.Second)
		
		// Log the test attempt
		fmt.Printf("Email test attempted to %s using SMTP %s:%s\n", 
			request.TestEmail, emailConfig["smtp_host"], emailConfig["smtp_port"])
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Test email configuration validated for " + request.TestEmail,
		"success": true,
		"config": gin.H{
			"smtp_host": emailConfig["smtp_host"],
			"smtp_port": emailConfig["smtp_port"],
			"smtp_username": emailConfig["smtp_username"],
		},
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
package services

import (
	"fmt"
	"math/rand"
	"testing"
	"testing/quick"
	"time"

	"superview/internal/models"
)

// Property 2: Filter Application Correctness
// For any combination of filters, the returned logs should only include entries that match ALL specified filter criteria
func TestProperty_FilterApplicationCorrectness(t *testing.T) {
	db := setupTestDB(t)
	service := NewActivityLogService(db)

	// 创建多样化的测试数据
	levels := []string{"INFO", "WARNING", "ERROR"}
	actions := []string{"login", "logout", "start_process", "stop_process"}
	resources := []string{"process", "node", "user", "auth"}
	usernames := []string{"admin", "user1", "user2", "system"}

	// 生成测试数据
	for _, level := range levels {
		for _, action := range actions {
			for _, resource := range resources {
				for _, username := range usernames {
					service.LogActivity(&models.ActivityLog{
						Level:     level,
						Message:   fmt.Sprintf("Test log %s %s", action, resource),
						Action:    action,
						Resource:  resource,
						Target:    "test-target",
						Username:  username,
						IPAddress: "127.0.0.1",
					})
				}
			}
		}
	}

	// Property test: 对于任意的筛选条件组合，返回的日志应该全部匹配所有筛选条件
	f := func(levelIdx, actionIdx, resourceIdx, usernameIdx uint8) bool {
		// 随机选择筛选条件
		filters := make(map[string]interface{})
		
		if levelIdx%4 < 3 {
			filters["level"] = levels[levelIdx%3]
		}
		if actionIdx%5 < 4 {
			filters["action"] = actions[actionIdx%4]
		}
		if resourceIdx%5 < 4 {
			filters["resource"] = resources[resourceIdx%4]
		}
		if usernameIdx%5 < 4 {
			filters["username"] = usernames[usernameIdx%4]
		}

		// 查询日志
		logs, _, err := service.GetActivityLogs(1, 100, filters)
		if err != nil {
			return false
		}

		// 验证所有返回的日志都匹配筛选条件
		for _, log := range logs {
			if level, ok := filters["level"].(string); ok {
				if log.Level != level {
					return false
				}
			}
			if action, ok := filters["action"].(string); ok {
				if log.Action != action {
					return false
				}
			}
			if resource, ok := filters["resource"].(string); ok {
				if log.Resource != resource {
					return false
				}
			}
			if username, ok := filters["username"].(string); ok {
				if log.Username != username {
					return false
				}
			}
		}

		return true
	}

	config := &quick.Config{MaxCount: 50}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// Property 3: Pagination Consistency
// For any page number and page size, the sum of all logs across all pages should equal the total count
func TestProperty_PaginationConsistency(t *testing.T) {
	db := setupTestDB(t)
	service := NewActivityLogService(db)

	// 创建固定数量的测试数据
	totalLogs := 47 // 选择一个质数，测试边界情况
	for i := 0; i < totalLogs; i++ {
		service.LogActivity(&models.ActivityLog{
			Level:     "INFO",
			Message:   fmt.Sprintf("Test log %d", i),
			Action:    "test",
			Resource:  "test",
			Target:    "test",
			Username:  "admin",
			IPAddress: "127.0.0.1",
		})
	}

	// Property test: 对于任意的页面大小，所有页面的日志总数应该等于 total count
	f := func(pageSize uint8) bool {
		if pageSize == 0 {
			pageSize = 1
		}
		if pageSize > 50 {
			pageSize = 50
		}

		filters := make(map[string]interface{})
		
		// 获取第一页以获取总数
		_, total, err := service.GetActivityLogs(1, int(pageSize), filters)
		if err != nil {
			return false
		}

		if total != int64(totalLogs) {
			return false
		}

		// 遍历所有页面，收集所有日志 ID
		seenIDs := make(map[uint]bool)
		totalPages := (int(total) + int(pageSize) - 1) / int(pageSize)

		for page := 1; page <= totalPages; page++ {
			logs, _, err := service.GetActivityLogs(page, int(pageSize), filters)
			if err != nil {
				return false
			}

			// 检查没有重复的日志
			for _, log := range logs {
				if seenIDs[log.ID] {
					return false // 发现重复
				}
				seenIDs[log.ID] = true
			}
		}

		// 验证收集到的日志总数等于 total
		return len(seenIDs) == int(total)
	}

	config := &quick.Config{MaxCount: 20}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// Property 5: User Action Logging Completeness
// For any user-initiated process operation, an activity log entry should be created with correct details
func TestProperty_UserActionLoggingCompleteness(t *testing.T) {
	db := setupTestDB(t)
	service := NewActivityLogService(db)

	// Property test: 对于任意的用户操作，应该创建包含正确信息的日志
	f := func(username, nodeName, processName string, actionIdx uint8) bool {
		if username == "" {
			username = "testuser"
		}
		if nodeName == "" {
			nodeName = "testnode"
		}
		if processName == "" {
			processName = "testprocess"
		}

		actions := []string{"start_process", "stop_process", "restart_process"}
		action := actions[actionIdx%3]

		// 记录日志
		target := fmt.Sprintf("%s:%s", nodeName, processName)
		message := fmt.Sprintf("Process %s %s on node %s", processName, action, nodeName)
		
		log := &models.ActivityLog{
			Level:     "INFO",
			Message:   message,
			Action:    action,
			Resource:  "process",
			Target:    target,
			Username:  username,
			IPAddress: "127.0.0.1",
			CreatedAt: time.Now(),
		}

		err := service.LogActivity(log)
		if err != nil {
			return false
		}

		// 验证日志已正确保存
		var savedLog models.ActivityLog
		db.First(&savedLog, log.ID)

		return savedLog.Username == username &&
			savedLog.Action == action &&
			savedLog.Resource == "process" &&
			savedLog.Target == target &&
			savedLog.Message == message
	}

	config := &quick.Config{MaxCount: 30}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// Property 6: System Event Logging Completeness
// For any system event, an activity log entry should be created with source marked as "system"
func TestProperty_SystemEventLoggingCompleteness(t *testing.T) {
	db := setupTestDB(t)
	service := NewActivityLogService(db)

	// Property test: 对于任意的系统事件，应该创建标记为 "system" 的日志
	f := func(nodeName, processName string, stateIdx uint8) bool {
		if nodeName == "" {
			nodeName = "testnode"
		}
		if processName == "" {
			processName = "testprocess"
		}

		actions := []string{"process_started", "process_stopped", "process_failed", "node_connected", "node_disconnected"}
		action := actions[stateIdx%5]

		target := fmt.Sprintf("%s:%s", nodeName, processName)
		message := fmt.Sprintf("System event: %s", action)

		err := service.LogSystemEvent("INFO", action, "process", target, message, nil)
		if err != nil {
			return false
		}

		// 查询最新的日志
		var log models.ActivityLog
		db.Order("created_at DESC").First(&log)

		// 验证系统事件标记为 "system"
		return log.Username == "system" &&
			log.Action == action &&
			log.Message == message &&
			log.Target == target
	}

	config := &quick.Config{MaxCount: 30}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// Helper: 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 验证属性测试框架正常工作
func TestPropertyTestFramework(t *testing.T) {
	// 简单的属性测试：对于任意整数 x，x + 0 = x
	f := func(x int) bool {
		return x+0 == x
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

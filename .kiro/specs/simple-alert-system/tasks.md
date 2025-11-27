# Implementation Plan

- [x] 1. Implement Alert Monitor service
  - [x] 1.1 Create AlertMonitor struct and initialization
    - Create `internal/services/alert_monitor.go`
    - Define AlertMonitor struct with dependencies
    - Implement NewAlertMonitor constructor
    - _Requirements: 1.1, 2.1_

  - [x] 1.2 Implement node status change handler
    - Create handleNodeStatusChange method
    - Check for existing active node offline alerts
    - Create alert when node goes offline
    - Resolve alert when node comes online
    - _Requirements: 1.1, 1.2, 1.5_

  - [x] 1.3 Implement process status change handler
    - Create handleProcessStatusChange method
    - Check for existing active process stopped alerts
    - Create alert when process stops (state == 0)
    - Resolve alert when process starts (state == 20)
    - _Requirements: 2.1, 2.2, 2.5_

  - [x] 1.4 Implement Start and Stop methods
    - Create background goroutine for monitoring
    - Subscribe to Supervisor service events
    - Handle graceful shutdown
    - _Requirements: 1.1, 2.1_

- [x] 2. Extend Alert Service with helper methods
  - [x] 2.1 Implement CreateNodeOfflineAlert
    - Check for existing active alert using metadata
    - Create alert with critical severity
    - Include node name in message and metadata
    - _Requirements: 1.1, 1.3, 1.5_

  - [x] 2.2 Implement ResolveNodeOfflineAlert
    - Find active alert by node name in metadata
    - Update status to resolved
    - Set end time and resolved_by
    - _Requirements: 1.2_

  - [x] 2.3 Implement CreateProcessStoppedAlert
    - Check for existing active alert using metadata
    - Create alert with high severity
    - Include node name and process name in message and metadata
    - _Requirements: 2.1, 2.3, 2.5_

  - [x] 2.4 Implement ResolveProcessStoppedAlert
    - Find active alert by node name and process name in metadata
    - Update status to resolved
    - Set end time and resolved_by
    - _Requirements: 2.2_

  - [x] 2.5 Implement GetActiveAlerts helper
    - Query alerts with status "active" or "acknowledged"
    - Order by created_at DESC
    - Include necessary preloads
    - _Requirements: 3.1, 3.3_

- [x] 3. Update Alert API to use real data
  - [x] 3.1 Update GetAlerts endpoint
    - Remove mock data
    - Call AlertService.GetAlerts with filters
    - Support severity and status filtering
    - Return real alerts from database
    - _Requirements: 3.1, 3.2, 3.4, 3.5_

  - [x] 3.2 Implement AcknowledgeAlert endpoint
    - Extract alert ID from URL
    - Extract user ID from context
    - Update alert status to "acknowledged"
    - Record acked_by and acked_at
    - _Requirements: 4.1, 4.3_

  - [x] 3.3 Implement ResolveAlert endpoint
    - Extract alert ID from URL
    - Extract user ID from context
    - Update alert status to "resolved"
    - Record resolved_by, resolved_at, and end_time
    - _Requirements: 4.2, 4.4, 4.5_

- [x] 4. Integrate Alert Monitor with main application
  - [x] 4.1 Initialize Alert Monitor in main.go
    - Create AlertService instance
    - Create AlertMonitor with dependencies
    - Start AlertMonitor after Supervisor service
    - _Requirements: 1.1, 2.1_

  - [x] 4.2 Add graceful shutdown for Alert Monitor
    - Call AlertMonitor.Stop() during shutdown
    - Wait for monitor goroutine to finish
    - _Requirements: 1.1, 2.1_

- [x] 5. Update frontend Alert page
  - [x] 5.1 Replace mock data with API calls
    - Update `web/react-app/src/pages/Alerts/index.tsx`
    - Call `/api/alerts` endpoint
    - Parse and display real alert data
    - _Requirements: 3.1, 3.2_

  - [x] 5.2 Implement acknowledge action
    - Add onClick handler for Acknowledge button
    - Call `/api/alerts/:id/acknowledge` endpoint
    - Refresh alert list after success
    - _Requirements: 4.1_

  - [x] 5.3 Implement resolve action
    - Add onClick handler for Resolve button
    - Call `/api/alerts/:id/resolve` endpoint
    - Refresh alert list after success
    - _Requirements: 4.2_

  - [x] 5.4 Update Alert Rules page
    - Remove or hide Alert Rules page (not needed for simple monitoring)
    - Or add note that rules are automatic
    - _Requirements: N/A_

- [x] 6. Testing and verification
  - [x] 6.1 Test node offline alert
    - Stop a configured node
    - Verify alert appears in Alert page
    - Restart node
    - Verify alert is resolved
    - _Requirements: 1.1, 1.2_

  - [x] 6.2 Test process stopped alert
    - Stop a process via Supervisor
    - Verify alert appears in Alert page
    - Start process
    - Verify alert is resolved
    - _Requirements: 2.1, 2.2_

  - [x] 6.3 Test acknowledge and resolve actions
    - Create test alerts
    - Acknowledge an alert via UI
    - Resolve an alert via UI
    - Verify status updates in database
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [x] 7. Final checkpoint
  - Ensure all tests pass, ask the user if questions arise.

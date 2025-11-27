# Design Document

## Overview

This design implements a simple, event-driven alert system that monitors node connectivity and process status. The system automatically creates alerts when nodes go offline or processes stop, and automatically resolves them when the issues are fixed. The implementation leverages existing Supervisor service events and adds minimal new code.

## Architecture

The alert system consists of three main components:

1. **Alert Monitor**: Listens to node and process status changes from the Supervisor service
2. **Alert Manager**: Creates, updates, and resolves alerts in the database
3. **Alert API**: Provides endpoints for the frontend to fetch and manage alerts

### Data Flow

```
Supervisor Service → Status Change Event → Alert Monitor → Alert Manager → Database
                                                                              ↓
Frontend ← Alert API ← Database
```

## Components and Interfaces

### Alert Monitor

**Location:** `internal/services/alert_monitor.go`

**Responsibilities:**
- Subscribe to node connection/disconnection events
- Subscribe to process state change events
- Trigger alert creation/resolution based on events

**Interface:**
```go
type AlertMonitor struct {
    alertService *AlertService
    supervisorService *supervisor.SupervisorService
}

func NewAlertMonitor(alertService *AlertService, supervisorService *supervisor.SupervisorService) *AlertMonitor
func (m *AlertMonitor) Start()
func (m *AlertMonitor) Stop()
func (m *AlertMonitor) handleNodeStatusChange(nodeName string, isConnected bool)
func (m *AlertMonitor) handleProcessStatusChange(nodeName, processName string, state int)
```

### Alert Manager

**Location:** Extend `internal/services/alert.go`

**New Methods:**
```go
func (s *AlertService) CreateNodeOfflineAlert(nodeName string) error
func (s *AlertService) ResolveNodeOfflineAlert(nodeName string) error
func (s *AlertService) CreateProcessStoppedAlert(nodeName, processName string) error
func (s *AlertService) ResolveProcessStoppedAlert(nodeName, processName string) error
func (s *AlertService) GetActiveAlerts() ([]Alert, error)
```

### Alert API

**Location:** `internal/api/alerts.go`

**Endpoints:**
- `GET /api/alerts` - Get all alerts with filtering
- `GET /api/alerts/:id` - Get alert by ID
- `POST /api/alerts/:id/acknowledge` - Acknowledge an alert
- `POST /api/alerts/:id/resolve` - Resolve an alert

## Data Models

Using existing models from `internal/models/alert.go`. No new models needed.

### Alert Types

We'll use a simplified approach without AlertRule:

**Node Offline Alert:**
```go
Alert{
    Message: "Node 'node-name' is offline",
    Severity: "critical",
    Status: "active",
    NodeID: nil,  // We don't have node IDs yet
    ProcessName: nil,
    Metadata: `{"type": "node_offline", "node_name": "node-name"}`,
}
```

**Process Stopped Alert:**
```go
Alert{
    Message: "Process 'process-name' on node 'node-name' has stopped",
    Severity: "high",
    Status: "active",
    NodeID: nil,
    ProcessName: &processName,
    Metadata: `{"type": "process_stopped", "node_name": "node-name", "process_name": "process-name"}`,
}
```

## Implementation Strategy

### Phase 1: Backend Alert Creation

1. Create AlertMonitor service
2. Hook into Supervisor service events
3. Implement alert creation/resolution logic
4. Add Alert API endpoints

### Phase 2: Frontend Integration

1. Replace mock data with real API calls
2. Implement acknowledge/resolve actions
3. Add real-time updates via WebSocket (optional)

## Error Handling

### Database Errors

**Scenario:** Failed to create or update alert in database

**Response:**
- Log error with context (node name, process name)
- Continue monitoring (don't crash the monitor)
- Retry on next status change

**Logging:**
- Log level: Error
- Message: "Failed to create/update alert"
- Fields: node_name, process_name, error

### Duplicate Alert Prevention

**Scenario:** Attempting to create an alert that already exists

**Response:**
- Check for existing active alert before creating
- If exists, update timestamp instead of creating new
- Log as debug, not error

## Testing Strategy

### Unit Testing

1. **Test: Alert creation for node offline**
   - Given: Node disconnects
   - Expected: Alert created with correct severity and message

2. **Test: Alert resolution for node online**
   - Given: Node reconnects
   - Expected: Existing alert resolved with end time

3. **Test: Alert creation for process stopped**
   - Given: Process state changes to stopped
   - Expected: Alert created with node and process info

4. **Test: Duplicate alert prevention**
   - Given: Node offline alert already exists
   - Expected: No new alert created

5. **Test: Alert acknowledge**
   - Given: User acknowledges alert
   - Expected: Status updated, user ID and timestamp recorded

### Integration Testing

1. **Test: End-to-end node offline alert**
   - Disconnect a node
   - Verify alert appears in API response
   - Reconnect node
   - Verify alert is resolved

2. **Test: End-to-end process stopped alert**
   - Stop a process
   - Verify alert appears in API response
   - Start process
   - Verify alert is resolved

## Implementation Notes

### Minimal Code Changes

This design minimizes code changes by:
- Reusing existing Alert models and service
- Leveraging existing Supervisor service events
- Not requiring AlertRule for simple monitoring
- Using metadata JSON field for flexible alert types

### Backward Compatibility

- Existing Alert API endpoints remain unchanged
- Frontend can still use existing Alert models
- No database migration needed (models already exist)

### Performance Considerations

- Alert monitor runs in background goroutine
- Database writes are async (don't block monitoring)
- Alert queries use indexes on status and created_at
- No polling - event-driven architecture

### Future Enhancements

- Add AlertRule support for custom thresholds
- Add notification channels (email, Slack)
- Add alert history and statistics
- Add alert grouping and deduplication

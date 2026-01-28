# Design Document: Node Discovery

## Overview

Node Discovery automates the process of finding and registering Supervisor nodes across a network range. Instead of manually editing config files for each node, administrators input a CIDR range, port, and credentials. The system scans the range concurrently, probes each IP via XML-RPC, and registers successful connections as new nodes.

The design prioritizes:
- **Simplicity**: Reuse existing patterns (worker pool, WebSocket hub, repository layer)
- **Pragmatism**: No over-engineering; scan → probe → register
- **Observability**: Real-time progress via WebSocket, full audit trail

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   React Page    │────▶│   Discovery API  │────▶│ Discovery       │
│  NodeDiscovery  │     │   /api/discovery │     │ Service         │
└─────────────────┘     └──────────────────┘     └─────────────────┘
        │                        │                       │
        │ WebSocket              │                       │
        ▼                        ▼                       ▼
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   WebSocket     │◀────│   Progress       │◀────│  Worker Pool    │
│   Hub           │     │   Events         │     │  (Scanner)      │
└─────────────────┘     └──────────────────┘     └─────────────────┘
                                                        │
                                                        ▼
                                                ┌─────────────────┐
                                                │  XML-RPC Client │
                                                │  (Probe)        │
                                                └─────────────────┘
```

## Components and Interfaces

### 1. Discovery API Handler (`internal/api/discovery.go`)

HTTP endpoints for discovery operations:

```go
type DiscoveryAPI struct {
    service            *services.DiscoveryService
    activityLogService *services.ActivityLogService
    hub                WebSocketHub
}

// Endpoints:
// POST   /api/discovery/tasks      - Start a new discovery task
// GET    /api/discovery/tasks      - List discovery tasks (paginated)
// GET    /api/discovery/tasks/:id  - Get task details with results
// POST   /api/discovery/tasks/:id/cancel - Cancel running task
// DELETE /api/discovery/tasks/:id  - Delete task and results
```

### 2. Discovery Service (`internal/services/discovery.go`)

Core business logic:

```go
type DiscoveryService struct {
    db             *gorm.DB
    nodeRepo       repository.NodeRepository
    hub            WebSocketHub
    activeScans    map[uint]*ScanContext  // task_id -> context
    mu             sync.RWMutex
}

type ScanContext struct {
    TaskID     uint
    Cancel     context.CancelFunc
    Pool       *utils.WorkerPool
}

// Key methods:
func (s *DiscoveryService) StartDiscovery(req *DiscoveryRequest) (*DiscoveryTask, error)
func (s *DiscoveryService) CancelDiscovery(taskID uint) error
func (s *DiscoveryService) GetTask(taskID uint) (*DiscoveryTask, error)
func (s *DiscoveryService) ListTasks(offset, limit int) ([]*DiscoveryTask, int64, error)
```

### 3. Scanner (`internal/services/scanner.go`)

Network scanning logic using existing worker pool:

```go
type ProbeTask struct {
    TaskID   uint
    IP       string
    Port     int
    Username string
    Password string
    Timeout  time.Duration
}

func (t *ProbeTask) Execute(ctx context.Context) error
func (t *ProbeTask) ID() string
```

### 4. CIDR Parser (`internal/utils/cidr.go`)

Simple CIDR validation and IP enumeration:

```go
func ParseCIDR(cidr string) (*CIDRRange, error)
func (r *CIDRRange) IPs() []string
func (r *CIDRRange) Count() int
```

## Data Models

### DiscoveryTask (`internal/models/discovery_task.go`)

```go
type DiscoveryTask struct {
    ID          uint           `gorm:"primaryKey" json:"id"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
    
    CIDR        string `gorm:"size:50;not null" json:"cidr"`
    Port        int    `gorm:"not null" json:"port"`
    Username    string `gorm:"size:50" json:"username"`
    // Password NOT stored - security
    
    Status      string `gorm:"size:20;default:'pending'" json:"status"`
    // pending, running, completed, cancelled, failed
    
    TotalIPs    int    `gorm:"default:0" json:"total_ips"`
    ScannedIPs  int    `gorm:"default:0" json:"scanned_ips"`
    FoundNodes  int    `gorm:"default:0" json:"found_nodes"`
    FailedIPs   int    `gorm:"default:0" json:"failed_ips"`
    
    StartedAt   *time.Time `json:"started_at"`
    CompletedAt *time.Time `json:"completed_at"`
    ErrorMsg    string     `gorm:"size:500" json:"error_msg,omitempty"`
    
    CreatedBy   string `gorm:"size:100" json:"created_by"`
}
```

### DiscoveryResult (`internal/models/discovery_result.go`)

```go
type DiscoveryResult struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    
    TaskID    uint   `gorm:"index;not null" json:"task_id"`
    IP        string `gorm:"size:50;not null" json:"ip"`
    Port      int    `gorm:"not null" json:"port"`
    
    Status    string `gorm:"size:20;not null" json:"status"`
    // success, timeout, connection_refused, auth_failed, error
    
    NodeID    *uint  `json:"node_id,omitempty"`      // If registered
    NodeName  string `gorm:"size:100" json:"node_name,omitempty"`
    Version   string `gorm:"size:50" json:"version,omitempty"`
    ErrorMsg  string `gorm:"size:500" json:"error_msg,omitempty"`
    
    Duration  int64  `json:"duration_ms"` // Probe duration in ms
}
```

## API Contracts

### Start Discovery

```
POST /api/discovery/tasks
Content-Type: application/json

Request:
{
    "cidr": "192.168.1.0/24",
    "port": 9001,
    "username": "admin",
    "password": "secret",
    "timeout_seconds": 3,    // optional, default 3
    "max_workers": 50        // optional, default 50
}

Response (201):
{
    "status": "success",
    "task": {
        "id": 1,
        "cidr": "192.168.1.0/24",
        "port": 9001,
        "status": "pending",
        "total_ips": 254,
        "created_at": "2024-01-15T10:00:00Z"
    }
}

Errors:
- 400: Invalid CIDR format
- 400: CIDR range too large (>65536)
- 400: Invalid port (must be 1-65535)
- 409: Another scan already running for this range
```

### List Tasks

```
GET /api/discovery/tasks?page=1&limit=20&status=completed

Response (200):
{
    "status": "success",
    "tasks": [...],
    "total": 42,
    "page": 1,
    "limit": 20
}
```

### Get Task Details

```
GET /api/discovery/tasks/:id

Response (200):
{
    "status": "success",
    "task": {
        "id": 1,
        "cidr": "192.168.1.0/24",
        "status": "completed",
        "total_ips": 254,
        "scanned_ips": 254,
        "found_nodes": 5,
        "failed_ips": 249,
        ...
    },
    "results": [
        {
            "ip": "192.168.1.10",
            "status": "success",
            "node_name": "node-192-168-1-10",
            "version": "4.2.0"
        },
        ...
    ]
}
```

### Cancel Task

```
POST /api/discovery/tasks/:id/cancel

Response (200):
{
    "status": "success",
    "message": "Task cancelled"
}
```

## WebSocket Events

Discovery progress broadcasts to all connected clients:

```go
// Progress update (every N scanned IPs or on discovery)
{
    "type": "discovery_progress",
    "data": {
        "task_id": 1,
        "scanned_ips": 50,
        "total_ips": 254,
        "found_nodes": 2,
        "failed_ips": 48,
        "percent": 19.7
    }
}

// Node discovered
{
    "type": "node_discovered",
    "data": {
        "task_id": 1,
        "ip": "192.168.1.10",
        "port": 9001,
        "node_name": "node-192-168-1-10",
        "version": "4.2.0"
    }
}

// Task completed
{
    "type": "discovery_completed",
    "data": {
        "task_id": 1,
        "status": "completed",
        "total_ips": 254,
        "found_nodes": 5,
        "duration_seconds": 45
    }
}
```

## Scanning Algorithm

```
1. Parse and validate CIDR
2. Create DiscoveryTask record (status=pending)
3. Calculate IP list from CIDR
4. Create worker pool (configurable workers, default 50)
5. Set task status=running, record started_at
6. For each IP in range:
   a. Submit ProbeTask to worker pool
   b. ProbeTask.Execute():
      - Create XML-RPC client with timeout
      - Call supervisor.getState()
      - On success: extract version, return success
      - On failure: categorize error (timeout/refused/auth/other)
7. Collect results from worker pool
8. For each successful probe:
   a. Check if node exists (by host:port)
   b. If not exists: create Node record
   c. Create DiscoveryResult record
9. Broadcast progress every 10 IPs or on discovery
10. On completion: set status=completed, record completed_at
11. Cleanup: stop worker pool
```

## Error Handling

| Error Type | HTTP Code | Handling |
|------------|-----------|----------|
| Invalid CIDR | 400 | Return validation error |
| Range too large | 400 | Return limit error (max 65536) |
| Task not found | 404 | Return not found |
| Task already cancelled | 409 | Return conflict |
| Probe timeout | - | Record as "timeout" result |
| Connection refused | - | Record as "connection_refused" |
| Auth failed | - | Record as "auth_failed" |
| XML-RPC error | - | Record as "error" with message |
| DB error | 500 | Log and return internal error |

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: CIDR Validation Correctness

*For any* string input, the CIDR parser SHALL either:
- Accept valid RFC 4632 IPv4 CIDR notation and return a parsed range, OR
- Reject invalid input with a non-empty error message

Additionally, *for any* valid CIDR with prefix length < 16, the parser SHALL reject it as too large.

**Validates: Requirements 1.1, 1.2, 1.4**

### Property 2: CIDR IP Count Calculation

*For any* valid CIDR with prefix length N (where 16 ≤ N ≤ 32), the calculated IP count SHALL equal 2^(32-N).

**Validates: Requirements 1.3**

### Property 3: Task Creation Invariants

*For any* valid discovery request, creating a task SHALL produce:
- A unique task ID (no two tasks share the same ID)
- Initial status of "pending"
- Created timestamp ≤ current time

**Validates: Requirements 2.1, 2.2**

### Property 4: Progress Tracking Invariant

*For any* discovery task at any point during execution:
- scanned_ips + found_nodes + failed_ips components are consistent
- scanned_ips ≤ total_ips
- When status is "completed": scanned_ips == total_ips

**Validates: Requirements 2.3, 2.5**

### Property 5: Node Registration Correctness

*For any* successful probe result:
- A Node record SHALL be created with name matching pattern "node-{ip-with-dashes}"
- The Node SHALL have status "discovered"
- If a node with same host:port exists, no duplicate SHALL be created

**Validates: Requirements 5.1, 5.2, 5.3, 5.5**

### Property 6: Result-Task Relationship Integrity

*For any* DiscoveryResult record:
- It SHALL have a valid TaskID referencing an existing DiscoveryTask
- The result's IP SHALL be within the task's CIDR range

**Validates: Requirements 7.1, 7.2**

### Property 7: Credential Security

*For any* DiscoveryTask record in the database:
- The password field SHALL be empty or omitted
- Activity logs SHALL NOT contain the password in plain text

**Validates: Requirements 9.1, 9.2**

### Property 8: API Authentication Enforcement

*For any* discovery API endpoint:
- Requests without valid JWT SHALL receive 401 Unauthorized
- Requests from non-admin users SHALL receive 403 Forbidden

**Validates: Requirements 9.3, 9.4**

## Testing Strategy

Testing follows a dual approach: unit tests for specific examples and edge cases, property tests for universal correctness.

### Unit Tests
- CIDR parsing edge cases (invalid formats, boundary values like /0, /32)
- API request validation (missing fields, invalid port numbers)
- Probe error categorization (timeout, connection_refused, auth_failed)
- WebSocket event formatting
- Task cancellation flow

### Property Tests
Property-based tests use `gopter` library for Go. Each test runs minimum 100 iterations.

| Property | Test Focus |
|----------|------------|
| Property 1 | Generate random strings, verify parse/reject behavior |
| Property 2 | Generate valid CIDRs, verify IP count formula |
| Property 3 | Create multiple tasks, verify uniqueness and initial state |
| Property 4 | Simulate progress updates, verify invariants hold |
| Property 5 | Generate probe results, verify node creation rules |
| Property 6 | Create results, verify task linkage |
| Property 7 | Create tasks with passwords, verify not persisted |
| Property 8 | Send requests with various auth states, verify rejection |

### Test Configuration
- Property tests: minimum 100 iterations per property
- Each property test tagged with: `Feature: node-discovery, Property N: {description}`
- Integration tests for WebSocket events and XML-RPC probing (separate from property tests)


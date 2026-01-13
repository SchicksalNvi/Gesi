# Design Document

## Overview

This design addresses critical concurrent safety issues in Go-CESI through systematic refactoring of shared data structures, elimination of race conditions, and implementation of proper synchronization patterns. The approach prioritizes correctness over performance, following the principle that a slow correct program is infinitely better than a fast incorrect one.

## Architecture

### Core Design Principles

1. **Never modify maps during iteration** - Use collect-then-modify pattern
2. **Minimize lock scope** - Hold locks for the shortest time possible  
3. **Use channels for coordination** - Prefer channels over shared memory where appropriate
4. **Fail fast with clear errors** - Don't hide problems, expose them immediately
5. **One responsibility per component** - Each component should have a single, clear purpose

### Synchronization Strategy

- **WebSocket Hub**: Use separate cleanup goroutine with channel-based coordination
- **Supervisor Service**: Protect all map operations with RWMutex, use atomic operations for counters
- **Configuration**: Use atomic.Value for lock-free reads with synchronized updates
- **Database**: Implement connection pool monitoring and circuit breaker pattern

## Components and Interfaces

### 1. WebSocket Hub Redesign

```go
type Hub struct {
    clients       map[*Client]bool
    clientsMu     sync.RWMutex
    
    broadcast     chan []byte
    register      chan *Client
    unregister    chan *Client
    cleanup       chan *Client  // New: separate cleanup channel
    
    service       *supervisor.SupervisorService
    config        *WebSocketConfig
    
    connectionCount int64  // atomic
    
    ctx           context.Context
    cancel        context.CancelFunc
    wg            sync.WaitGroup
}
```

**Key Changes:**
- Separate `cleanup` channel to avoid iteration-modification conflicts
- Atomic counter for connection tracking
- Proper WaitGroup for graceful shutdown
- Clear separation of concerns between registration and cleanup

### 2. Supervisor Service Thread Safety

```go
type SupervisorService struct {
    nodes              map[string]*Node
    nodesMu            sync.RWMutex
    
    processStates      map[string]map[string]int
    nodeStates         map[string]bool
    statesMu           sync.RWMutex
    
    stopChan           chan struct{}
    wg                 sync.WaitGroup
    shutdown           int32  // atomic
    
    activityLogger     ActivityLogger
    
    // New: connection pool management
    connectionSemaphore chan struct{}  // Limit concurrent connections
}
```

**Key Changes:**
- Separate mutexes for different data structures
- Atomic shutdown flag
- Connection semaphore to limit concurrent operations
- Proper WaitGroup management

### 3. Unified Authentication Middleware

```go
type AuthMiddleware struct {
    db          *gorm.DB
    tokenParser TokenParser
    config      AuthConfig
}

type AuthConfig struct {
    SkipPaths      []string
    TokenSources   []TokenSource  // Header, Cookie, Query
    ErrorHandler   ErrorHandler
}
```

**Key Changes:**
- Single implementation replacing duplicates
- Configurable token sources
- Consistent error handling
- Clear separation of concerns

### 4. Safe Configuration Management

```go
type ConfigManager struct {
    current     atomic.Value  // stores *Config
    updateMu    sync.Mutex    // Serializes updates
    validators  []ConfigValidator
    logger      Logger
}
```

**Key Changes:**
- Lock-free reads using atomic.Value
- Serialized updates with validation
- Rollback capability on validation failure

## Data Models

### Client Connection State

```go
type Client struct {
    hub        *Hub
    conn       *websocket.Conn
    send       chan []byte
    
    // Thread-safe fields
    userID     string
    subscribed sync.Map  // map[string]bool - thread-safe
    
    // Protected by mutex
    mu           sync.RWMutex
    lastPong     time.Time
    violationCount int
    closed       bool
}
```

### Node Connection State

```go
type Node struct {
    // Immutable after creation
    Name         string
    Environment  string
    Host         string
    Port         int
    Username     string
    Password     string
    
    // Protected by mutex
    mu           sync.RWMutex
    IsConnected  bool
    LastPing     time.Time
    Processes    []Process
    
    client       *xmlrpc.SupervisorClient
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

<function_calls>
<invoke name="prework">
<parameter name="featureName">concurrent-safety-fixes

### Property Reflection

After reviewing all properties identified in the prework analysis, I've identified several areas where properties can be consolidated to eliminate redundancy:

**Consolidation Areas:**
- Properties 1.1, 1.3, 2.1, 2.2, 2.3 all test thread safety - can be combined into comprehensive race detection properties
- Properties 3.2, 3.3, 3.4 all test authentication consistency - can be combined into single auth behavior property  
- Properties 4.1, 4.2, 4.4 all test configuration update safety - can be combined into atomic config update property
- Properties 8.1, 8.3, 8.5 all test goroutine lifecycle - can be combined into comprehensive resource management property
- Properties 9.1, 9.2, 9.3, 9.4, 9.5 all test input validation - can be combined into comprehensive input security property

**Final Property Set (after consolidation):**

Property 1: **WebSocket Hub Race-Free Operations**
*For any* sequence of concurrent client connections, disconnections, and message broadcasts, the WebSocket hub should handle all operations without data races or panics
**Validates: Requirements 1.1, 1.2, 1.3**

Property 2: **Message Delivery Ordering**
*For any* client connection, messages broadcast to that client should arrive in the same order they were sent
**Validates: Requirements 1.4**

Property 3: **Separate Cleanup Architecture**
*For any* client cleanup operation, the WebSocket hub should use a separate cleanup goroutine to avoid iteration-modification conflicts
**Validates: Requirements 1.5**

Property 4: **Supervisor Service Thread Safety**
*For any* concurrent operations on the supervisor service, all map operations should be protected by appropriate synchronization without data races
**Validates: Requirements 2.1, 2.2, 2.3**

Property 5: **Resource-Limited Concurrent Operations**
*For any* auto-refresh cycle, the system should limit concurrent connection attempts to prevent resource exhaustion
**Validates: Requirements 2.4**

Property 6: **Graceful Service Shutdown**
*For any* service shutdown request, all spawned goroutines should complete within the shutdown timeout
**Validates: Requirements 2.5**

Property 7: **Single Authentication Implementation**
*For any* codebase scan, exactly one AuthMiddleware implementation should exist
**Validates: Requirements 3.1, 3.5**

Property 8: **Consistent Authentication Behavior**
*For any* authentication request across all endpoints, the same token extraction, validation, and error response logic should be applied
**Validates: Requirements 3.2, 3.3, 3.4**

Property 9: **Atomic Configuration Updates**
*For any* configuration update operation, readers should get either complete old or complete new configuration, never partial state, and updates should be serialized
**Validates: Requirements 4.1, 4.2, 4.4**

Property 10: **Configuration Validation and Rollback**
*For any* invalid configuration provided, the system should reject it and maintain the previous valid configuration
**Validates: Requirements 4.3, 4.5**

Property 11: **Database Resilience**
*For any* database failure scenario, the system should implement retry logic with exponential backoff and handle resource exhaustion gracefully
**Validates: Requirements 5.1, 5.2, 5.4**

Property 12: **Transaction Resource Management**
*For any* database transaction, timeout should trigger proper resource cleanup and appropriate error responses
**Validates: Requirements 5.3**

Property 13: **Hierarchical Operation Timeouts**
*For any* batch operation, individual operations should have their own timeouts in addition to batch-level timeouts
**Validates: Requirements 6.1, 6.2**

Property 14: **Circuit Breaker Pattern**
*For any* node that fails repeatedly, the system should implement circuit breaker pattern to prevent cascading failures
**Validates: Requirements 6.3**

Property 15: **Timeout Resource Cleanup**
*For any* operation cancelled due to timeout, all associated resources should be properly cleaned up
**Validates: Requirements 6.4**

Property 16: **Consistent Error Response Format**
*For any* API error across all endpoints, the response should use consistent JSON format and appropriate HTTP status codes
**Validates: Requirements 7.1, 7.2**

Property 17: **Detailed Validation Errors**
*For any* validation failure, the system should provide detailed field-level error information without exposing sensitive data
**Validates: Requirements 7.3, 7.5**

Property 18: **Contextual Error Logging**
*For any* system error, log entries should contain sufficient context for debugging including correlation IDs
**Validates: Requirements 7.4, 10.2**

Property 19: **Comprehensive Resource Management**
*For any* background task or long-running operation, the system should ensure proper cancellation via context and resource cleanup
**Validates: Requirements 8.1, 8.3, 8.5**

Property 20: **WebSocket Resource Cleanup**
*For any* WebSocket connection closure, all associated resources should be cleaned up properly
**Validates: Requirements 8.2**

Property 21: **Connection Limiting**
*For any* auto-refresh operation, the number of concurrent connection attempts should be limited by semaphore
**Validates: Requirements 8.4**

Property 22: **Comprehensive Input Security**
*For any* user input across all endpoints, the system should validate parameters, prevent injection attacks, and sanitize data before processing
**Validates: Requirements 9.1, 9.2, 9.3, 9.4, 9.5**

Property 23: **System Observability**
*For any* system operation, appropriate metrics should be exposed for goroutine counts, resource usage, and operation latencies
**Validates: Requirements 10.1, 10.3**

Property 24: **Health Check Detail**
*For any* system health degradation, health check endpoints should expose detailed status information
**Validates: Requirements 10.4**

Property 25: **Structured Logging**
*For any* log entry, the system should use structured logging with appropriate log levels
**Validates: Requirements 10.5**

## Error Handling

### Error Categories and Responses

1. **Concurrent Access Errors**
   - Detection: Go race detector in tests
   - Response: Immediate panic with clear stack trace
   - Recovery: Not applicable - these are programming errors

2. **Resource Exhaustion Errors**
   - Detection: Connection pool monitoring, goroutine counting
   - Response: Graceful degradation with appropriate HTTP status
   - Recovery: Circuit breaker pattern, exponential backoff

3. **Configuration Errors**
   - Detection: Validation during config load
   - Response: Log error, maintain previous valid config
   - Recovery: Automatic retry on next SIGHUP

4. **Network/Database Errors**
   - Detection: Connection timeouts, query failures
   - Response: Retry with exponential backoff
   - Recovery: Circuit breaker, fallback to cached data

### Error Response Format

```go
type ErrorResponse struct {
    Status    string                 `json:"status"`
    Message   string                 `json:"message"`
    Code      string                 `json:"code,omitempty"`
    Details   map[string]interface{} `json:"details,omitempty"`
    RequestID string                 `json:"request_id,omitempty"`
}
```

## Testing Strategy

### Dual Testing Approach

The system requires both unit tests and property-based tests to ensure comprehensive coverage:

**Unit Tests:**
- Specific examples of correct behavior
- Edge cases and error conditions  
- Integration points between components
- Mock-based testing for external dependencies

**Property-Based Tests:**
- Universal properties that hold for all inputs
- Comprehensive input coverage through randomization
- Race condition detection using Go's race detector
- Resource leak detection using runtime metrics

### Property-Based Testing Configuration

- **Testing Framework**: Use `testing/quick` for basic property tests, `github.com/leanovate/gopter` for advanced scenarios
- **Test Iterations**: Minimum 1000 iterations per property test due to concurrent nature
- **Race Detection**: All property tests must run with `-race` flag
- **Resource Monitoring**: Tests must monitor goroutine counts and memory usage

### Test Tagging Format

Each property-based test must include a comment referencing its design document property:

```go
// Feature: concurrent-safety-fixes, Property 1: WebSocket Hub Race-Free Operations
func TestWebSocketHubConcurrentSafety(t *testing.T) {
    // Property test implementation
}
```

### Critical Test Scenarios

1. **Concurrent Client Operations**: Simulate hundreds of clients connecting/disconnecting simultaneously
2. **Configuration Hot-Reload**: Send rapid SIGHUP signals while system is under load
3. **Database Failure Recovery**: Simulate database failures during high-concurrency operations
4. **Resource Exhaustion**: Test behavior when connection pools, goroutines, or memory are exhausted
5. **Graceful Shutdown**: Verify clean shutdown under various load conditions

### Performance Benchmarks

- **WebSocket Throughput**: Measure messages/second with varying client counts
- **Database Connection Pool**: Measure query latency under pool exhaustion
- **Memory Usage**: Monitor for goroutine and memory leaks during long-running tests
- **CPU Usage**: Ensure synchronization overhead doesn't exceed 10% of total CPU

The testing strategy emphasizes correctness over performance, with the understanding that a correct slow system can be optimized, but an incorrect fast system is worthless.
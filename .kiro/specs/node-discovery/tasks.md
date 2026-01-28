index.tsx:58  Warning: Instance created by `useForm` is not connected to any Form element. Forget to pass `form` prop?
warning @ warning.js:30
call @ warning.js:51
warningOnce @ warning.js:58
(匿名) @ useForm.js:160
setTimeout
(匿名) @ useForm.js:157
(匿名) @ useForm.js:645
(匿名) @ index.tsx:58
commitHookEffectListMount @ react-dom.development.js:23189
commitPassiveMountOnFiber @ react-dom.development.js:24965
commitPassiveMountEffects_complete @ react-dom.development.js:24930
commitPassiveMountEffects_begin @ react-dom.development.js:24917
commitPassiveMountEffects @ react-dom.development.js:24905
flushPassiveEffectsImpl @ react-dom.development.js:27078
flushPassiveEffects @ react-dom.development.js:27023
commitRootImpl @ react-dom.development.js:26974
commitRoot @ react-dom.development.js:26721
performSyncWorkOnRoot @ react-dom.development.js:26156
flushSyncCallbacks @ react-dom.development.js:12042
(匿名) @ react-dom.development.js:25690
XMLHttpRequest.send
dispatchXhrRequest @ xhr.js:198
xhr @ xhr.js:15
dispatchRequest @ dispatchRequest.js:51
Promise.then
_request @ Axios.js:163
request @ Axios.js:40
Axios.<computed> @ Axios.js:211
wrap @ bind.js:12
get @ client.ts:48
getUserPreferences @ settings.ts:52
loadUserPreferences @ index.tsx:72
(匿名) @ index.tsx:96
commitHookEffectListMount @ react-dom.development.js:23189
commitPassiveMountOnFiber @ react-dom.development.js:24965
commitPassiveMountEffects_complete @ react-dom.development.js:24930
commitPassiveMountEffects_begin @ react-dom.development.js:24917
commitPassiveMountEffects @ react-dom.development.js:24905
flushPassiveEffectsImpl @ react-dom.development.js:27078
flushPassiveEffects @ react-dom.development.js:27023
commitRootImpl @ react-dom.development.js:26974
commitRoot @ react-dom.development.js:26721
performSyncWorkOnRoot @ react-dom.development.js:26156
flushSyncCallbacks @ react-dom.development.js:12042
(匿名) @ react-dom.development.js:25690


## Overview

Implementation follows a bottom-up approach: data models → utilities → service → API → frontend. Each task builds on previous work, with property tests validating correctness at each layer.

## Tasks

- [x] 1. Create data models and database migration
  - [x] 1.1 Create DiscoveryTask model in `internal/models/discovery_task.go`
    - Define struct with GORM tags
    - Add status constants (pending, running, completed, cancelled, failed)
    - _Requirements: 2.1, 2.2_
  
  - [x] 1.2 Create DiscoveryResult model in `internal/models/discovery_result.go`
    - Define struct with foreign key to DiscoveryTask
    - Add status constants (success, timeout, connection_refused, auth_failed, error)
    - _Requirements: 7.1, 7.2_
  
  - [x] 1.3 Add auto-migration in `internal/database/database.go`
    - Register DiscoveryTask and DiscoveryResult models
    - _Requirements: 7.1_

- [x] 2. Implement CIDR utilities
  - [x] 2.1 Create CIDR parser in `internal/utils/cidr.go`
    - ParseCIDR function with validation
    - IPs() method to enumerate addresses
    - Count() method for total count
    - Reject ranges larger than /16
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_
  
  - [x] 2.2 Write property tests for CIDR parser
    - **Property 1: CIDR Validation Correctness**
    - **Property 2: CIDR IP Count Calculation**
    - **Validates: Requirements 1.1, 1.2, 1.3, 1.4**

- [x] 3. Checkpoint - Verify models and CIDR utilities
  - Ensure all tests pass, ask the user if questions arise.

- [x] 4. Implement Discovery repository
  - [x] 4.1 Add DiscoveryRepository interface in `internal/repository/interfaces.go`
    - CreateTask, GetTask, UpdateTask, DeleteTask
    - ListTasks with pagination and status filter
    - CreateResult, GetResultsByTaskID
    - _Requirements: 7.1, 7.2, 7.3_
  
  - [x] 4.2 Implement DiscoveryRepository in `internal/repository/discovery_repository.go`
    - Implement all interface methods
    - _Requirements: 7.1, 7.2, 7.3_

- [x] 5. Implement Discovery service
  - [x] 5.1 Create DiscoveryService in `internal/services/discovery.go`
    - StartDiscovery: validate input, create task, start scan
    - CancelDiscovery: stop worker pool, update status
    - GetTask, ListTasks: delegate to repository
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_
  
  - [x] 5.2 Implement scanner logic in `internal/services/scanner.go`
    - ProbeTask struct implementing utils.Task interface
    - Execute method: XML-RPC connection and supervisor.getState call
    - Error categorization (timeout, refused, auth_failed)
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 4.1, 4.2, 4.3, 4.4, 4.5, 4.6_
  
  - [x] 5.3 Implement node registration logic
    - Check for existing node by host:port
    - Generate node name from IP
    - Create node with status "discovered"
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_
  
  - [x] 5.4 Implement WebSocket progress broadcasting
    - Emit discovery_progress every 10 IPs
    - Emit node_discovered on successful probe
    - Emit discovery_completed on finish
    - _Requirements: 6.1, 6.2, 6.3_
  
  - [x] 5.5 Write property tests for Discovery service
    - **Property 3: Task Creation Invariants**
    - **Property 4: Progress Tracking Invariant**
    - **Property 5: Node Registration Correctness**
    - **Property 6: Result-Task Relationship Integrity**
    - **Validates: Requirements 2.1, 2.2, 2.3, 2.5, 5.1, 5.2, 5.3, 5.5, 7.1, 7.2**

- [x] 6. Checkpoint - Verify service layer
  - Ensure all tests pass, ask the user if questions arise.

- [x] 7. Implement Discovery API
  - [x] 7.1 Create DiscoveryAPI handler in `internal/api/discovery.go`
    - POST /api/discovery/tasks - StartDiscovery
    - GET /api/discovery/tasks - ListTasks
    - GET /api/discovery/tasks/:id - GetTask
    - POST /api/discovery/tasks/:id/cancel - CancelTask
    - DELETE /api/discovery/tasks/:id - DeleteTask
    - _Requirements: 2.1, 2.4, 6.4, 7.3_
  
  - [x] 7.2 Register routes in `internal/api/api.go`
    - Add discovery routes under /api/discovery
    - Apply auth middleware
    - _Requirements: 9.3, 9.4_
  
  - [x] 7.3 Add activity logging for discovery actions
    - Log task start, completion, cancellation, failure
    - Log node registrations
    - _Requirements: 8.1, 8.2, 8.3, 8.4_
  
  - [x] 7.4 Write property tests for API security
    - **Property 7: Credential Security**
    - **Property 8: API Authentication Enforcement**
    - **Validates: Requirements 9.1, 9.2, 9.3, 9.4**

- [x] 8. Checkpoint - Verify API layer
  - Ensure all tests pass, ask the user if questions arise.

- [x] 9. Implement React frontend
  - [x] 9.1 Create discovery API client in `web/react-app/src/api/discovery.ts`
    - startDiscovery, getTasks, getTask, cancelTask, deleteTask
    - _Requirements: 2.1, 2.4, 6.4_
  
  - [x] 9.2 Create NodeDiscovery page in `web/react-app/src/pages/Discovery/index.tsx`
    - Form for CIDR, port, username, password input
    - CIDR validation feedback
    - Start discovery button
    - _Requirements: 1.1, 1.2, 1.3_
  
  - [x] 9.3 Implement discovery progress display
    - Progress bar showing scanned/total
    - Real-time updates via WebSocket
    - List of discovered nodes
    - _Requirements: 6.1, 6.2, 6.3_
  
  - [x] 9.4 Implement discovery history view
    - Table of past discovery tasks
    - Pagination and status filtering
    - View details and results
    - _Requirements: 7.3_
  
  - [x] 9.5 Add route and navigation
    - Add /discovery route in App.tsx
    - Add navigation link in sidebar
    - _Requirements: N/A (UI integration)_

- [x] 10. Final checkpoint - End-to-end verification
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- All property tests are required (comprehensive testing)
- Property tests use `gopter` library with minimum 100 iterations
- WebSocket events follow existing hub.Broadcast pattern
- Node naming: `node-{ip-with-dashes}` (e.g., node-192-168-1-100)
- Credentials are NOT stored in DiscoveryTask records for security

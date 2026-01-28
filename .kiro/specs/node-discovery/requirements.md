# Requirements Document

## Introduction

Node Discovery is a feature that enables automated scanning of network ranges to discover and register Supervisor nodes. Instead of manually configuring each node in config files, administrators can input a CIDR network range, target port, and credentials to automatically probe and add reachable Supervisor instances to the database.

## Glossary

- **Discovery_Service**: The backend service responsible for scanning network ranges and probing Supervisor nodes
- **Scanner**: The component that iterates through IP addresses in a CIDR range and attempts connections
- **Probe**: A single connection attempt to verify if a Supervisor instance is running at a given IP:port
- **Discovery_Task**: A background job representing a complete network scan operation
- **Discovery_Result**: The outcome of probing a single IP address (success/failure with details)
- **CIDR**: Classless Inter-Domain Routing notation for specifying network ranges (e.g., 192.168.1.0/24)

## Requirements

### Requirement 1: CIDR Input and Validation

**User Story:** As an administrator, I want to input a network range in CIDR notation, so that I can specify which IP addresses to scan for Supervisor nodes.

#### Acceptance Criteria

1. WHEN a user submits a CIDR string, THE Discovery_Service SHALL validate it against RFC 4632 CIDR format
2. WHEN an invalid CIDR string is provided, THE Discovery_Service SHALL return a descriptive error message indicating the validation failure
3. WHEN a valid CIDR is provided, THE Discovery_Service SHALL calculate and return the total number of IP addresses to scan
4. IF the CIDR range exceeds 65536 addresses (/16 or larger), THEN THE Discovery_Service SHALL reject the request with a warning about scan size
5. THE Discovery_Service SHALL support both IPv4 CIDR notation (e.g., 192.168.1.0/24)

### Requirement 2: Discovery Task Management

**User Story:** As an administrator, I want to start, monitor, and cancel discovery tasks, so that I can control the scanning process.

#### Acceptance Criteria

1. WHEN a user initiates a discovery scan, THE Discovery_Service SHALL create a Discovery_Task with a unique identifier
2. WHEN a Discovery_Task is created, THE Discovery_Service SHALL persist it to the database with status "pending"
3. WHILE a Discovery_Task is running, THE Discovery_Service SHALL update progress (scanned count, found count, failed count)
4. WHEN a user requests task cancellation, THE Discovery_Service SHALL stop the scan gracefully and mark status as "cancelled"
5. WHEN a Discovery_Task completes, THE Discovery_Service SHALL update status to "completed" with final statistics
6. IF an error occurs during scanning, THEN THE Discovery_Service SHALL mark the task as "failed" with error details

### Requirement 3: Concurrent Network Scanning

**User Story:** As an administrator, I want the system to scan multiple IPs concurrently, so that discovery completes in reasonable time.

#### Acceptance Criteria

1. WHEN scanning a network range, THE Scanner SHALL use a worker pool for concurrent probing
2. THE Scanner SHALL limit concurrent connections to a configurable maximum (default: 50 workers)
3. WHEN probing an IP address, THE Scanner SHALL apply a connection timeout (default: 3 seconds)
4. WHEN a probe times out, THE Scanner SHALL mark that IP as unreachable and continue scanning
5. THE Scanner SHALL process IPs in sequential order within the CIDR range

### Requirement 4: Supervisor Node Probing

**User Story:** As an administrator, I want the system to verify Supervisor connectivity using XML-RPC, so that only valid Supervisor instances are discovered.

#### Acceptance Criteria

1. WHEN probing an IP:port combination, THE Scanner SHALL attempt an XML-RPC connection using provided credentials
2. WHEN connection succeeds, THE Scanner SHALL call the `supervisor.getState` method to verify Supervisor is running
3. WHEN `supervisor.getState` returns successfully, THE Scanner SHALL extract Supervisor version and state information
4. IF authentication fails, THEN THE Scanner SHALL record the IP as "auth_failed" in Discovery_Result
5. IF connection is refused, THEN THE Scanner SHALL record the IP as "connection_refused" in Discovery_Result
6. IF timeout occurs, THEN THE Scanner SHALL record the IP as "timeout" in Discovery_Result

### Requirement 5: Node Registration

**User Story:** As an administrator, I want successfully discovered nodes to be automatically added to the database, so that I can manage them immediately.

#### Acceptance Criteria

1. WHEN a Supervisor node is successfully probed, THE Discovery_Service SHALL create a Node record in the database
2. WHEN creating a Node record, THE Discovery_Service SHALL generate a unique name based on IP address (e.g., "node-192-168-1-100")
3. WHEN a node with the same host:port already exists, THE Discovery_Service SHALL skip registration and mark as "duplicate"
4. THE Discovery_Service SHALL store the provided credentials (username/password) with the new Node record
5. WHEN a node is registered, THE Discovery_Service SHALL set initial status to "discovered"

### Requirement 6: Progress Feedback

**User Story:** As an administrator, I want real-time progress updates during scanning, so that I can monitor the discovery process.

#### Acceptance Criteria

1. WHILE a Discovery_Task is running, THE Discovery_Service SHALL emit progress events via WebSocket
2. WHEN a node is discovered, THE Discovery_Service SHALL emit a "node_discovered" event with node details
3. WHEN scanning completes, THE Discovery_Service SHALL emit a "discovery_completed" event with summary statistics
4. THE Discovery_Service SHALL provide a polling endpoint for clients that cannot use WebSocket

### Requirement 7: Discovery History

**User Story:** As an administrator, I want to view past discovery tasks and their results, so that I can audit and review scanning history.

#### Acceptance Criteria

1. THE Discovery_Service SHALL persist all Discovery_Task records with timestamps
2. THE Discovery_Service SHALL persist Discovery_Result records linked to their parent Discovery_Task
3. WHEN querying discovery history, THE Discovery_Service SHALL support pagination and filtering by status
4. THE Discovery_Service SHALL retain discovery history for a configurable period (default: 30 days)

### Requirement 8: Activity Logging

**User Story:** As an administrator, I want all discovery actions logged, so that I have an audit trail for security compliance.

#### Acceptance Criteria

1. WHEN a discovery task is started, THE Discovery_Service SHALL log the action with user, CIDR range, and timestamp
2. WHEN nodes are discovered and registered, THE Discovery_Service SHALL log each registration
3. WHEN a discovery task is cancelled, THE Discovery_Service SHALL log the cancellation with reason
4. WHEN a discovery task fails, THE Discovery_Service SHALL log the failure with error details

### Requirement 9: Security Considerations

**User Story:** As an administrator, I want credentials handled securely, so that sensitive information is protected.

#### Acceptance Criteria

1. THE Discovery_Service SHALL NOT log credentials in plain text
2. WHEN storing credentials in Discovery_Task records, THE Discovery_Service SHALL encrypt or omit them
3. THE Discovery_Service SHALL require authentication for all discovery API endpoints
4. THE Discovery_Service SHALL validate that the requesting user has admin privileges

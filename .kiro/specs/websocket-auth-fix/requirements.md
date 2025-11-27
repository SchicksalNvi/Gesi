# Requirements Document

## Introduction

The WebSocket connection on the nodes page is failing with the error "WebSocket is closed before the connection is established." This occurs because the authentication middleware expects JWT tokens in the Authorization header, but WebSocket connections from browsers cannot set custom headers during the initial handshake. The token is being passed as a query parameter (`?token=...`), but the middleware is not configured to accept tokens from query parameters, causing the connection to be rejected before the WebSocket upgrade completes.

## Glossary

- **WebSocket**: A communication protocol providing full-duplex communication channels over a single TCP connection
- **JWT (JSON Web Token)**: A compact, URL-safe means of representing claims to be transferred between two parties
- **Authentication Middleware**: Server-side code that validates user credentials before allowing access to protected resources
- **Query Parameter**: A key-value pair appended to a URL after a question mark (e.g., `?token=abc123`)
- **Authorization Header**: An HTTP header used to send credentials, typically in the format `Authorization: Bearer <token>`
- **WebSocket Upgrade**: The HTTP handshake process that transitions a connection from HTTP to WebSocket protocol
- **Hub**: The WebSocket connection manager that handles client registration, message broadcasting, and connection lifecycle

## Requirements

### Requirement 1

**User Story:** As a user viewing the nodes page, I want the WebSocket connection to establish successfully, so that I can receive real-time updates about node and process status.

#### Acceptance Criteria

1. WHEN a user navigates to the nodes page with a valid JWT token THEN the system SHALL establish a WebSocket connection successfully
2. WHEN the WebSocket handshake occurs THEN the system SHALL accept JWT tokens from the query parameter `token`
3. WHEN the WebSocket handshake occurs with a token in the Authorization header THEN the system SHALL continue to accept tokens from the Authorization header for backward compatibility
4. WHEN an invalid or missing token is provided THEN the system SHALL reject the WebSocket connection with an appropriate error message
5. WHEN an expired token is provided THEN the system SHALL reject the WebSocket connection and return an authentication error

### Requirement 2

**User Story:** As a system administrator, I want WebSocket authentication to be secure, so that unauthorized users cannot access real-time system data.

#### Acceptance Criteria

1. WHEN a WebSocket connection is attempted without a token THEN the system SHALL reject the connection
2. WHEN a WebSocket connection is attempted with a malformed token THEN the system SHALL reject the connection and log the attempt
3. WHEN a WebSocket connection is established THEN the system SHALL extract and validate the user ID from the token claims
4. WHEN a token validation fails THEN the system SHALL not upgrade the connection to WebSocket protocol
5. WHEN a valid token is provided THEN the system SHALL set the user context for the WebSocket client

### Requirement 3

**User Story:** As a developer, I want the authentication logic to be maintainable and testable, so that future changes can be made safely.

#### Acceptance Criteria

1. WHEN implementing token extraction THEN the system SHALL check the query parameter first, then fall back to the Authorization header
2. WHEN extracting tokens from the Authorization header THEN the system SHALL properly parse the "Bearer " prefix
3. WHEN the WebSocket middleware is invoked THEN the system SHALL use the existing JWT parsing functions from the auth package
4. WHEN authentication fails THEN the system SHALL return appropriate HTTP status codes before the WebSocket upgrade
5. WHEN the WebSocket connection is established THEN the system SHALL pass the authenticated user ID to the WebSocket client handler

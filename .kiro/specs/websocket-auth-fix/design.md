# Design Document

## Overview

This design addresses the WebSocket authentication failure by creating a specialized authentication middleware for WebSocket connections that accepts JWT tokens from both query parameters and Authorization headers. The solution maintains backward compatibility while enabling browser-based WebSocket connections that cannot set custom headers during the initial handshake.

## Architecture

The solution involves modifying the WebSocket route handler in `cmd/main.go` to use a custom authentication function instead of the standard HTTP middleware. This custom function will:

1. Extract the JWT token from either the query parameter or Authorization header
2. Validate the token using existing JWT parsing logic
3. Set the user context in the Gin context
4. Allow the WebSocket upgrade to proceed if authentication succeeds
5. Return an HTTP error response if authentication fails

The existing WebSocket Hub and Client implementations remain unchanged, as they already properly handle authenticated connections once the user context is set.

## Components and Interfaces

### WebSocket Authentication Handler

**Location:** `cmd/main.go` (inline function in WebSocket route setup)

**Responsibilities:**
- Extract JWT token from request (query parameter or header)
- Validate token using `auth.ParseToken()`
- Set user context in Gin context
- Handle authentication errors before WebSocket upgrade

**Interface:**
```go
// Inline handler function signature
func(c *gin.Context) {
    // Extract token
    // Validate token
    // Set user context
    // Call hub.HandleWebSocket(c)
}
```

### Token Extraction Logic

**Function:** `extractToken(c *gin.Context) string`

**Logic:**
1. Check for `token` query parameter
2. If not found, check `Authorization` header
3. If Authorization header exists, strip "Bearer " prefix
4. Return token string or empty string if not found

### Existing Components (No Changes Required)

- **auth.ParseToken()**: Already validates JWT tokens and returns claims
- **websocket.Hub.HandleWebSocket()**: Already expects user_id in Gin context
- **websocket.Client**: Already uses user_id from context for logging and tracking

## Data Models

No new data models are required. The existing JWT Claims structure is sufficient:

```go
type Claims struct {
    UserID string `json:"user_id"`
    jwt.RegisteredClaims
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Token extraction precedence

*For any* HTTP request to the WebSocket endpoint, if both a query parameter token and an Authorization header token are present, the query parameter token should be used for authentication.
**Validates: Requirements 3.1**

### Property 2: Valid token acceptance

*For any* valid JWT token (not expired, properly signed), when provided via either query parameter or Authorization header, the WebSocket connection should be established successfully.
**Validates: Requirements 1.1, 1.2, 1.3**

### Property 3: Invalid token rejection

*For any* invalid JWT token (expired, malformed, or improperly signed), the WebSocket connection attempt should be rejected with an HTTP 401 status code before the WebSocket upgrade occurs.
**Validates: Requirements 1.4, 1.5, 2.2, 2.4**

### Property 4: Missing token rejection

*For any* WebSocket connection attempt without a token in either the query parameter or Authorization header, the connection should be rejected with an HTTP 401 status code.
**Validates: Requirements 2.1**

### Property 5: User context propagation

*For any* successfully authenticated WebSocket connection, the user ID from the JWT claims should be available in the Gin context and passed to the WebSocket client handler.
**Validates: Requirements 2.3, 2.5, 3.5**

### Property 6: Bearer prefix handling

*For any* Authorization header containing a token with the "Bearer " prefix, the system should correctly extract the token by removing the prefix before validation.
**Validates: Requirements 3.2**

## Error Handling

### Authentication Errors

**Scenario:** Token is missing, invalid, expired, or malformed

**Response:**
- HTTP Status: 401 Unauthorized
- JSON Body: `{"error": "Authentication failed: <specific reason>"}`
- Action: Connection is not upgraded to WebSocket

**Logging:**
- Log level: Warn
- Message: "WebSocket authentication failed"
- Fields: remote_addr, error_reason

### Token Parsing Errors

**Scenario:** JWT parsing fails due to signature mismatch or invalid format

**Response:**
- HTTP Status: 401 Unauthorized
- JSON Body: `{"error": "Invalid token"}`
- Action: Connection is not upgraded to WebSocket

**Logging:**
- Log level: Warn
- Message: "Failed to parse WebSocket token"
- Fields: remote_addr, error

### Successful Authentication

**Logging:**
- Log level: Debug
- Message: "WebSocket authentication successful"
- Fields: user_id, remote_addr

## Testing Strategy

### Unit Testing

Unit tests will verify specific authentication scenarios:

1. **Test: Token extraction from query parameter**
   - Given: Request with `?token=valid_jwt`
   - Expected: Token is extracted correctly

2. **Test: Token extraction from Authorization header**
   - Given: Request with `Authorization: Bearer valid_jwt`
   - Expected: Token is extracted correctly after removing "Bearer " prefix

3. **Test: Query parameter takes precedence**
   - Given: Request with both query parameter and header tokens
   - Expected: Query parameter token is used

4. **Test: Missing token rejection**
   - Given: Request with no token
   - Expected: 401 response, no WebSocket upgrade

5. **Test: Expired token rejection**
   - Given: Request with expired JWT
   - Expected: 401 response with appropriate error message

### Property-Based Testing

Property-based tests will use the `testing/quick` package from Go's standard library to verify correctness properties across many randomly generated inputs.

**Configuration:**
- Minimum 100 iterations per property test
- Use `testing/quick` for random input generation
- Each property test will be tagged with its corresponding design property number

**Test Properties:**

1. **Property 1: Token extraction precedence** (Property 1)
   - Generate random valid JWT tokens
   - Create requests with both query and header tokens
   - Verify query parameter is always used

2. **Property 2: Valid token acceptance** (Property 2)
   - Generate random valid JWT tokens with various user IDs
   - Test both query parameter and header delivery
   - Verify all result in successful authentication

3. **Property 3: Invalid token rejection** (Property 3)
   - Generate random invalid tokens (malformed, wrong signature, expired)
   - Verify all are rejected with 401 status

4. **Property 4: Missing token rejection** (Property 4)
   - Generate requests without tokens
   - Verify all are rejected with 401 status

5. **Property 5: User context propagation** (Property 5)
   - Generate random valid tokens with various user IDs
   - Verify user ID from token matches user ID in context

6. **Property 6: Bearer prefix handling** (Property 6)
   - Generate random valid tokens
   - Test with and without "Bearer " prefix
   - Verify both are handled correctly

### Integration Testing

Integration tests will verify the complete WebSocket connection flow:

1. **Test: End-to-end WebSocket connection with valid token**
   - Authenticate user via login endpoint
   - Use returned JWT to establish WebSocket connection
   - Verify connection succeeds and receives initial data

2. **Test: WebSocket connection rejection with invalid token**
   - Attempt WebSocket connection with invalid token
   - Verify connection is rejected before upgrade

3. **Test: WebSocket connection with expired token**
   - Create token with 1-second expiration
   - Wait for expiration
   - Attempt connection
   - Verify rejection

## Implementation Notes

### Minimal Code Changes

The fix requires changes only to `cmd/main.go` in the WebSocket route setup section. No changes are needed to:
- `internal/websocket/hub.go`
- `internal/websocket/client.go`
- `internal/auth/jwt.go`
- `internal/auth/middleware.go`

### Backward Compatibility

The solution maintains backward compatibility by:
- Continuing to accept tokens from Authorization headers
- Using the same JWT validation logic
- Preserving the existing WebSocket Hub behavior
- Not changing any API contracts

### Security Considerations

- Query parameters are logged in many systems; however, this is already the current behavior in the frontend
- The token is still validated using the same secure JWT parsing logic
- Token expiration is still enforced
- Failed authentication attempts are logged for security monitoring

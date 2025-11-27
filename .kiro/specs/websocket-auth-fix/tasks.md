# Implementation Plan

- [x] 1. Implement WebSocket authentication handler in main.go
  - [x] 1.1 Create token extraction function
    - Extract token from query parameter `token`
    - Fall back to Authorization header if query parameter not present
    - Strip "Bearer " prefix from Authorization header
    - _Requirements: 1.2, 1.3, 3.1, 3.2_

  - [x] 1.2 Implement authentication logic in WebSocket route handler
    - Call token extraction function
    - Validate token using auth.ParseToken()
    - Set user_id in Gin context on success
    - Return 401 JSON error on authentication failure
    - Log authentication attempts (success and failure)
    - _Requirements: 1.1, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 3.4, 3.5_

  - [x] 1.3 Write property test for token extraction precedence
    - **Property 1: Token extraction precedence**
    - **Validates: Requirements 3.1**

  - [x] 1.4 Write property test for valid token acceptance
    - **Property 2: Valid token acceptance**
    - **Validates: Requirements 1.1, 1.2, 1.3**

  - [x] 1.5 Write property test for invalid token rejection
    - **Property 3: Invalid token rejection**
    - **Validates: Requirements 1.4, 1.5, 2.2, 2.4**

  - [x] 1.6 Write property test for missing token rejection
    - **Property 4: Missing token rejection**
    - **Validates: Requirements 2.1**

  - [x] 1.7 Write property test for user context propagation
    - **Property 5: User context propagation**
    - **Validates: Requirements 2.3, 2.5, 3.5**

  - [x] 1.8 Write property test for Bearer prefix handling
    - **Property 6: Bearer prefix handling**
    - **Validates: Requirements 3.2**

- [x] 2. Verify WebSocket connection in browser
  - Manually test WebSocket connection from nodes page
  - Verify real-time updates are received
  - Check browser console for connection errors
  - Confirm no authentication warnings appear
  - _Requirements: 1.1_

- [x] 3. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

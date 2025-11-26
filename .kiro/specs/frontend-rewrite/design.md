# Design Document

## Overview

The Go-CESI frontend will be rewritten using a modern, lightweight technology stack consisting of Alpine.js for reactive components, Tailwind CSS for styling, and native browser APIs for routing and WebSocket communication. This design eliminates the complexity of React while providing a clean, maintainable, and performant user interface.

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Browser                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │              HTML Pages (SPA)                         │  │
│  │  ┌─────────────┐  ┌──────────────┐  ┌─────────────┐  │  │
│  │  │  Alpine.js  │  │  Tailwind    │  │   Native    │  │  │
│  │  │  Components │  │  CSS Styles  │  │   Router    │  │  │
│  │  └─────────────┘  └──────────────┘  └─────────────┘  │  │
│  └───────────────────────────────────────────────────────┘  │
│                           │                                  │
│                           │ HTTP/WebSocket                   │
│                           ▼                                  │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Go Backend Server                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  REST API    │  │  WebSocket   │  │  Static File │      │
│  │  Endpoints   │  │  Hub         │  │  Server      │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

### Technology Stack

1. **Alpine.js (v3.x)**: Lightweight reactive framework for component behavior
2. **Tailwind CSS (v3.x)**: Utility-first CSS framework
3. **Native Fetch API**: For HTTP requests to backend
4. **Native WebSocket API**: For real-time updates
5. **History API**: For client-side routing
6. **LocalStorage**: For JWT token and user preferences

### Directory Structure

```
web/
├── static/
│   ├── css/
│   │   ├── tailwind.css          # Tailwind base
│   │   └── custom.css            # Custom styles
│   ├── js/
│   │   ├── alpine.min.js         # Alpine.js library
│   │   ├── app.js                # Main application logic
│   │   ├── api.js                # API client
│   │   ├── auth.js               # Authentication logic
│   │   ├── websocket.js          # WebSocket manager
│   │   ├── router.js             # Client-side router
│   │   └── components/
│   │       ├── navbar.js         # Navbar component
│   │       ├── sidebar.js        # Sidebar component
│   │       └── notifications.js  # Toast notifications
│   ├── img/
│   │   └── logo.png
│   └── favicon.ico
├── templates/
│   └── index.html                # Single HTML file (SPA)
└── tailwind.config.js            # Tailwind configuration
```

## Components and Interfaces

### 1. Application Shell (index.html)

The main HTML file that serves as the SPA container:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go-CESI</title>
    <link href="/static/css/tailwind.css" rel="stylesheet">
    <link href="/static/css/custom.css" rel="stylesheet">
</head>
<body class="bg-gray-50">
    <div id="app" x-data="app()" x-init="init()">
        <!-- Login Page -->
        <template x-if="!isAuthenticated">
            <div id="login-page"></div>
        </template>
        
        <!-- Main Application -->
        <template x-if="isAuthenticated">
            <div class="flex h-screen">
                <!-- Sidebar -->
                <aside id="sidebar"></aside>
                
                <!-- Main Content -->
                <div class="flex-1 flex flex-col">
                    <!-- Navbar -->
                    <nav id="navbar"></nav>
                    
                    <!-- Page Content -->
                    <main id="content" class="flex-1 overflow-auto p-6"></main>
                </div>
            </div>
        </template>
        
        <!-- Notifications -->
        <div id="notifications"></div>
    </div>
    
    <script src="/static/js/alpine.min.js" defer></script>
    <script src="/static/js/app.js" type="module"></script>
</body>
</html>
```

### 2. API Client (api.js)

Handles all HTTP communication with the backend:

```javascript
class ApiClient {
    constructor(baseURL) {
        this.baseURL = baseURL;
    }
    
    async request(endpoint, options = {}) {
        const token = localStorage.getItem('token');
        const headers = {
            'Content-Type': 'application/json',
            ...(token && { 'Authorization': `Bearer ${token}` }),
            ...options.headers
        };
        
        const response = await fetch(`${this.baseURL}${endpoint}`, {
            ...options,
            headers
        });
        
        if (response.status === 401) {
            // Token expired, redirect to login
            window.location.href = '/login';
            throw new Error('Unauthorized');
        }
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Request failed');
        }
        
        return response.json();
    }
    
    get(endpoint) {
        return this.request(endpoint, { method: 'GET' });
    }
    
    post(endpoint, data) {
        return this.request(endpoint, {
            method: 'POST',
            body: JSON.stringify(data)
        });
    }
    
    put(endpoint, data) {
        return this.request(endpoint, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
    }
    
    delete(endpoint) {
        return this.request(endpoint, { method: 'DELETE' });
    }
}

export const api = new ApiClient('/api');
```

### 3. WebSocket Manager (websocket.js)

Manages real-time communication:

```javascript
class WebSocketManager {
    constructor(url) {
        this.url = url;
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.listeners = new Map();
    }
    
    connect() {
        const token = localStorage.getItem('token');
        this.ws = new WebSocket(`${this.url}?token=${token}`);
        
        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.reconnectAttempts = 0;
        };
        
        this.ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            this.emit(data.type, data.payload);
        };
        
        this.ws.onclose = () => {
            console.log('WebSocket disconnected');
            this.reconnect();
        };
        
        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }
    
    reconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            setTimeout(() => this.connect(), 1000 * this.reconnectAttempts);
        }
    }
    
    on(event, callback) {
        if (!this.listeners.has(event)) {
            this.listeners.set(event, []);
        }
        this.listeners.get(event).push(callback);
    }
    
    emit(event, data) {
        if (this.listeners.has(event)) {
            this.listeners.get(event).forEach(callback => callback(data));
        }
    }
    
    disconnect() {
        if (this.ws) {
            this.ws.close();
        }
    }
}

export const ws = new WebSocketManager('ws://localhost:8081/ws');
```

### 4. Router (router.js)

Client-side routing using History API:

```javascript
class Router {
    constructor() {
        this.routes = new Map();
        this.currentRoute = null;
        
        window.addEventListener('popstate', () => this.handleRoute());
    }
    
    register(path, handler) {
        this.routes.set(path, handler);
    }
    
    navigate(path) {
        window.history.pushState({}, '', path);
        this.handleRoute();
    }
    
    handleRoute() {
        const path = window.location.pathname;
        const handler = this.routes.get(path) || this.routes.get('/404');
        
        if (handler) {
            this.currentRoute = path;
            handler();
        }
    }
    
    start() {
        this.handleRoute();
    }
}

export const router = new Router();
```

### 5. Authentication Manager (auth.js)

Handles user authentication:

```javascript
class AuthManager {
    constructor() {
        this.token = localStorage.getItem('token');
        this.user = JSON.parse(localStorage.getItem('user') || 'null');
    }
    
    async login(username, password) {
        const response = await api.post('/auth/login', { username, password });
        this.token = response.data.token;
        this.user = response.data.user;
        
        localStorage.setItem('token', this.token);
        localStorage.setItem('user', JSON.stringify(this.user));
        
        return response;
    }
    
    logout() {
        this.token = null;
        this.user = null;
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        window.location.href = '/login';
    }
    
    isAuthenticated() {
        return !!this.token;
    }
    
    getUser() {
        return this.user;
    }
}

export const auth = new AuthManager();
```

## Data Models

### User Model
```javascript
{
    id: string,
    username: string,
    email: string,
    fullName: string,
    isAdmin: boolean,
    isActive: boolean,
    lastLogin: string (ISO 8601),
    createdAt: string (ISO 8601)
}
```

### Node Model
```javascript
{
    name: string,
    hostname: string,
    port: number,
    username: string,
    environment: string,
    status: 'online' | 'offline',
    processCount: number,
    lastCheck: string (ISO 8601)
}
```

### Process Model
```javascript
{
    name: string,
    group: string,
    state: 'RUNNING' | 'STOPPED' | 'STARTING' | 'STOPPING' | 'FATAL',
    statename: string,
    description: string,
    pid: number,
    uptime: number,
    cpu: number,
    memory: number
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Authentication Token Persistence
*For any* authenticated user session, when the user refreshes the page, the authentication state should be preserved from localStorage
**Validates: Requirements 3.2, 3.3**

### Property 2: API Request Authorization
*For any* API request after authentication, the request should include the JWT token in the Authorization header
**Validates: Requirements 3.5**

### Property 3: WebSocket Reconnection
*For any* WebSocket disconnection, the system should attempt to reconnect with exponential backoff up to the maximum retry limit
**Validates: Requirements 7.4, 7.5**

### Property 4: Real-time UI Updates
*For any* WebSocket event received, the corresponding UI component should update within 100ms
**Validates: Requirements 7.2, 7.3**

### Property 5: Responsive Layout Consistency
*For any* viewport size change, all UI components should maintain proper spacing without overlap
**Validates: Requirements 2.2, 2.3**

### Property 6: Form Validation Feedback
*For any* form submission with invalid data, the system should highlight all problematic fields before sending the request
**Validates: Requirements 10.4**

### Property 7: Loading State Indication
*For any* asynchronous operation, a loading indicator should be displayed until the operation completes or fails
**Validates: Requirements 10.3**

### Property 8: Notification Auto-dismiss
*For any* success notification, the notification should automatically dismiss after exactly 3 seconds
**Validates: Requirements 10.5**

### Property 9: Route Navigation State
*For any* route navigation, the browser history should be updated and the back button should work correctly
**Validates: Requirements 11.2**

### Property 10: Keyboard Navigation
*For any* interactive element, pressing Tab should move focus to the next element in logical order
**Validates: Requirements 12.2**

## Error Handling

### API Errors
- Network errors: Display "Connection failed" message with retry option
- 401 Unauthorized: Redirect to login page
- 403 Forbidden: Display "Access denied" message
- 404 Not Found: Display "Resource not found" message
- 500 Server Error: Display "Server error" message with error details

### WebSocket Errors
- Connection failed: Attempt automatic reconnection
- Max retries exceeded: Display warning banner with manual reconnect button
- Message parsing error: Log error and continue operation

### Validation Errors
- Empty required fields: Highlight field with red border and show error message
- Invalid format: Show format hint below field
- Server validation errors: Display error messages from API response

## Testing Strategy

### Unit Testing
- Test API client methods with mocked fetch responses
- Test WebSocket manager connection and reconnection logic
- Test router path matching and navigation
- Test authentication manager token storage and retrieval
- Test form validation functions

### Integration Testing
- Test login flow from form submission to dashboard redirect
- Test process control actions (start/stop/restart)
- Test real-time updates via WebSocket
- Test navigation between pages
- Test logout and session cleanup

### Property-Based Testing
- Use **fast-check** library for JavaScript property-based testing
- Configure each test to run minimum 100 iterations
- Tag each test with format: **Feature: frontend-rewrite, Property {number}: {property_text}**
- Each correctness property must have ONE corresponding property-based test

### Manual Testing
- Test responsive design on different screen sizes
- Test keyboard navigation
- Test screen reader compatibility
- Test browser compatibility (Chrome, Firefox, Safari, Edge)
- Test performance with large datasets

### Performance Testing
- Measure initial page load time (target: < 2s)
- Measure time to interactive (target: < 1s)
- Measure JavaScript bundle size (target: < 200KB)
- Measure API response handling time (target: < 100ms)

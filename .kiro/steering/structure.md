# Project Structure

## Root Directory Layout

```
go-cesi/
├── cmd/                    # Application entry points
├── internal/               # Private application code
├── web/                    # Frontend assets
├── config/                 # Configuration files
├── data/                   # SQLite database files
├── logs/                   # Application logs
├── migrations/             # Database migrations
├── scripts/                # Utility scripts
├── tools/                  # CLI tools
├── config.toml             # Main configuration
├── .env                    # Environment variables
└── build-react.sh          # React build script
```

## Internal Package Organization

The `internal/` directory follows clean architecture principles:

### Core Layers

- **api/**: HTTP handlers and route definitions
  - One file per resource (e.g., `nodes.go`, `users.go`, `alerts.go`)
  - `api.go` contains route setup
  - `responses.go` for common response structures

- **models/**: Data models and domain entities
  - GORM models with database tags
  - Business logic methods on models

- **services/**: Business logic layer
  - Orchestrates operations between repositories
  - Contains domain-specific business rules

- **repository/**: Data access layer
  - Database operations abstraction
  - `interfaces.go` defines repository contracts

### Supporting Packages

- **auth/**: Authentication and authorization
  - JWT token generation/validation
  - Auth middleware
  
- **middleware/**: HTTP middleware
  - Performance monitoring
  - Request validation
  
- **config/**: Configuration management
  - Viper-based config loading
  - Config structs and defaults

- **database/**: Database initialization
  - GORM setup and migrations
  - Connection management

- **supervisor/**: Supervisor integration
  - `service.go`: Main supervisor service
  - `node.go`: Node management
  - `xmlrpc/`: XML-RPC client for Supervisor API

- **websocket/**: Real-time communication
  - `hub.go`: WebSocket connection hub
  - `client.go`: Client connection management

- **logger/**: Logging infrastructure
  - Zap logger configuration
  - Dynamic log level management

- **loggers/**: Application-specific loggers
  - Activity logging
  - File-based logging

- **validation/**: Input validation
  - DTO validation
  - Custom validators

- **utils/**: Utility functions
  - Log readers
  - Query optimizers
  - Resource managers

- **errors/**: Error handling
  - Custom error types
  - Error responses

## Frontend Structure

```
web/
├── react-app/              # React application
│   ├── src/
│   │   ├── components/     # Reusable React components
│   │   │   ├── Auth/       # Authentication components
│   │   │   ├── Layout/     # Layout components
│   │   │   └── ui/         # UI primitives (buttons, cards, etc.)
│   │   ├── contexts/       # React contexts (Auth, WebSocket)
│   │   ├── pages/          # Page components (one per route)
│   │   ├── services/       # API service layer (api.js)
│   │   ├── App.js          # Main app component
│   │   └── index.js        # Entry point
│   ├── public/             # Static assets
│   └── build/              # Production build output
├── static/                 # Legacy static assets
│   ├── css/
│   └── js/
└── templates/              # Legacy Go templates
```

## Naming Conventions

### Go Code

- **Files**: Snake_case or descriptive names (e.g., `activity_logs.go`, `process_enhanced.go`)
- **Packages**: Lowercase, single word when possible
- **Structs**: PascalCase (e.g., `User`, `ProcessGroup`)
- **Functions/Methods**: PascalCase for exported, camelCase for private
- **Constants**: PascalCase or SCREAMING_SNAKE_CASE for package-level

### React Code

- **Files**: PascalCase for components (e.g., `Dashboard.js`, `NodeDetail.js`)
- **Components**: PascalCase (e.g., `ProtectedRoute`, `ActivityLogs`)
- **Functions**: camelCase (e.g., `fetchNodes`, `handleSubmit`)
- **CSS**: kebab-case for classes

## API Route Structure

All API routes are prefixed with `/api` and organized by resource:

- `/api/auth/*` - Authentication endpoints
- `/api/nodes/*` - Node management
- `/api/users/*` - User management
- `/api/profile/*` - User profile
- `/api/environments/*` - Environment grouping
- `/api/groups/*` - Process groups
- `/api/activity-logs/*` - Activity logging
- `/api/roles/*` - Role management
- `/api/alerts/*` - Alert system
- `/api/process-enhanced/*` - Advanced process features
- `/api/configuration/*` - Configuration management
- `/api/logs/*` - Log analysis
- `/api/data-management/*` - Data export/import
- `/api/system-settings/*` - System settings
- `/api/developer/*` - Developer tools
- `/api/health/*` - Health checks

## Configuration Files

- `config.toml`: Main application config (nodes, server port, admin user)
- `.env`: Sensitive environment variables (JWT_SECRET, passwords)
- `web/react-app/.env`: React environment variables
- `data/cesi.db`: SQLite database file

## Key Conventions

1. **Error Handling**: Use structured logging with Zap, return errors up the stack
2. **Database Transactions**: Use GORM transactions for multi-step operations
3. **API Responses**: Consistent JSON structure via `responses.go`
4. **Authentication**: JWT middleware on protected routes
5. **WebSocket Events**: Typed event names (e.g., `node_update`, `process_status_change`)
6. **Activity Logging**: Log all user actions via activity log service

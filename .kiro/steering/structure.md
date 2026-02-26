# Project Structure

## Root Directory Layout

```
superview/
├── cmd/                    # Application entry points
├── internal/               # Private application code
├── web/                    # Frontend assets
├── config/                 # Configuration files
├── data/                   # SQLite database files
├── logs/                   # Application logs
├── tools/                  # CLI tools
├── superview.sh            # Management script
└── README.md               # Project documentation
```

## Internal Package Organization

The `internal/` directory follows clean architecture principles:

### Core Layers

- **api/**: HTTP handlers and route definitions
- **models/**: Data models and domain entities
- **services/**: Business logic layer
- **repository/**: Data access layer

### Supporting Packages

- **auth/**: Authentication and authorization
- **middleware/**: HTTP middleware
- **config/**: Configuration management
- **database/**: Database initialization
- **supervisor/**: Supervisor integration
- **websocket/**: Real-time communication
- **logger/**: Logging infrastructure
- **loggers/**: Application-specific loggers
- **validation/**: Input validation
- **utils/**: Utility functions
- **errors/**: Error handling
- **metrics/**: Prometheus metrics

## Frontend Structure

```
web/react-app/src/
├── components/     # Reusable React components
├── pages/          # Page components
├── layouts/        # Layout components
├── api/            # API client layer
├── store/          # State management (Zustand)
├── hooks/          # Custom hooks
├── types/          # TypeScript types
├── i18n/           # Internationalization
└── main.tsx        # Entry point
```

## Configuration Files

- `config/config.toml`: Main application config
- `config/nodelist.toml`: Supervisor node definitions
- `config/.env`: Sensitive environment variables
- `data/superview.db`: SQLite database file

## Key Conventions

1. Error Handling: Use structured logging with Zap, return errors up the stack
2. Database Transactions: Use GORM transactions for multi-step operations
3. API Responses: Consistent JSON structure via `responses.go`
4. Authentication: JWT middleware on protected routes
5. WebSocket Events: Typed event names (e.g., `node_update`, `process_status_change`)
6. Activity Logging: Log all user actions via activity log service

# Technology Stack

## Backend

- **Language**: Go 1.24.1
- **Web Framework**: Gin (v1.10.1)
- **Database**: SQLite with GORM ORM (v1.30.0)
- **Authentication**: JWT-based (golang-jwt/jwt v4.5.0)
- **Real-time Communication**: WebSocket (gorilla/websocket v1.5.3)
- **Configuration**: Viper (v1.20.1) with TOML format
- **Logging**: Zap (v1.27.0) - structured logging
- **Password Hashing**: bcrypt (golang.org/x/crypto)
- **Task Scheduling**: Cron (robfig/cron v3.0.0)
- **Worker Pools**: gammazero/workerpool (v1.1.3)
- **Rate Limiting**: golang.org/x/time

## Frontend

- **Framework**: React 18.2.0
- **UI Library**: Bootstrap 5.2.0 + React-Bootstrap 2.5.0
- **Routing**: React Router DOM v6.3.0
- **HTTP Client**: Axios 0.27.2
- **Charts**: Recharts 2.5.0
- **Icons**: Bootstrap Icons 1.9.1 + Lucide React 0.526.0
- **Date Handling**: Moment.js 2.29.4
- **Notifications**: Sonner 1.4.0
- **Build Tool**: Create React App (react-scripts 5.0.1)

## Common Commands

### Development

```bash
# Start backend (development)
go run cmd/main.go

# Start React dev server (optional for frontend development)
cd web/react-app && npm start

# Build React frontend
./build-react.sh
# or manually:
cd web/react-app && npm install && npm run build
```

### Production

```bash
# Build Go binary
go build -o go-cesi cmd/main.go

# Run production server
./go-cesi
```

### Database

```bash
# Create admin user via CLI
go run cmd/main.go create-admin --username admin --password pass123 --email admin@example.com
```

### Configuration

- Main config: `config.toml` (TOML format)
- Environment variables: `.env` file
- Required env vars: `JWT_SECRET`, `ADMIN_PASSWORD`, `NODE_PASSWORD`
- Config hot-reload: Send `SIGHUP` signal to process

### Testing & Utilities

```bash
# Run Go tests
go test ./...

# Download Go dependencies
go mod download

# Tidy Go modules
go mod tidy

# Verify admin password
go run tools/verify_admin_password.go

# Dump users
go run tools/dump_users.go
```

## Architecture Patterns

- **Clean Architecture**: Separation of concerns with internal packages
- **Repository Pattern**: Data access abstraction in `internal/repository`
- **Service Layer**: Business logic in `internal/services`
- **Middleware**: Cross-cutting concerns (auth, performance, validation)
- **WebSocket Hub**: Centralized real-time communication management
- **XML-RPC Client**: Communication with Supervisor instances

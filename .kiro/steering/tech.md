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
- **Language**: TypeScript 5.3
- **UI Library**: Ant Design 5.x
- **Routing**: React Router DOM v6
- **HTTP Client**: Axios
- **Charts**: ECharts
- **State Management**: Zustand
- **Build Tool**: Vite 5.x

## Common Commands

### Development

```bash
# Start backend (development)
go run cmd/main.go

# Start React dev server
cd web/react-app && npm run dev

# Build frontend
cd web/react-app && npm install && npm run build
```

### Production

```bash
# Build Go binary
go build -o superview cmd/main.go

# Run production server
./superview

# Or use management script
./superview.sh build
./superview.sh start
```

### Database

```bash
# Create admin user via CLI
go run cmd/main.go create-admin --username admin --password pass123 --email admin@example.com
```

### Configuration

- Main config: `config/config.toml` (TOML format)
- Node list: `config/nodelist.toml`
- Environment variables: `config/.env`
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
```

## Architecture Patterns

- **Clean Architecture**: Separation of concerns with internal packages
- **Repository Pattern**: Data access abstraction in `internal/repository`
- **Service Layer**: Business logic in `internal/services`
- **Middleware**: Cross-cutting concerns (auth, performance, validation)
- **WebSocket Hub**: Centralized real-time communication management
- **XML-RPC Client**: Communication with Supervisor instances

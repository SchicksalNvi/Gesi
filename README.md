# Go-CESI - Centralized Supervisor Interface

A modern web-based interface for managing multiple Supervisor instances with real-time monitoring capabilities.

## Features

### Core Functionality
- **Multi-Node Management**: Manage multiple Supervisor instances from a single interface
- **Real-time Monitoring**: WebSocket-based real-time updates for process status
- **Process Control**: Start, stop, restart individual processes or all processes on a node
- **User Management**: Role-based access control with admin and regular user roles
- **Activity Logging**: Comprehensive logging of all user actions and system events

### New High-Priority Features
1. **Batch Process Operations**: Start/stop/restart all processes on a node with single click
2. **Modern React Frontend**: Complete React-based user interface with real-time updates
3. **User Profile Management**: Users can view and update their personal profiles

## Architecture

- **Backend**: Go with Gin framework
- **Frontend**: React with Bootstrap for styling
- **Database**: SQLite with GORM ORM
- **Real-time Communication**: WebSocket for live updates
- **Authentication**: JWT-based authentication

## Prerequisites

- Go 1.19 or higher
- Node.js 16 or higher
- npm or yarn
- Supervisor instances to manage

## Installation

### 1. Clone the Repository
```bash
git clone <repository-url>
cd go-cesi
```

### 2. Install Go Dependencies
```bash
go mod download
```

### 3. Build React Frontend
```bash
# Make the build script executable
chmod +x build-react.sh

# Build the React app
./build-react.sh
```

### 4. Configuration

Create a `config.toml` file in the project root:

```toml
[server]
port = 8081

[admin]
username = "admin"
password = "admin123"
email = "admin@example.com"

[[nodes]]
name = "production-server"
environment = "production"
host = "192.168.1.100"
port = 9001
username = "supervisor"
password = "supervisor123"

[[nodes]]
name = "staging-server"
environment = "staging"
host = "192.168.1.101"
port = 9001
username = "supervisor"
password = "supervisor123"
```

## Running the Application

### Development Mode

1. **Start the Go backend**:
```bash
go run cmd/main.go
```

2. **For React development** (optional, if you want to modify frontend):
```bash
cd web/react-app
npm start
```

### Production Mode

1. **Build the React frontend** (if not already done):
```bash
./build-react.sh
```

2. **Build and run the Go application**:
```bash
go build -o go-cesi cmd/main.go
./go-cesi
```

## Usage

### Accessing the Application

1. Open your browser and navigate to `http://localhost:8081`
2. Login with the admin credentials configured in `config.toml`
3. The React frontend will load with the dashboard

### Key Features

#### Dashboard
- Overview of all nodes and their status
- Real-time process statistics
- Recent activity logs
- System health indicators

#### Node Management
- View all configured Supervisor nodes
- Monitor node connectivity and status
- **NEW**: Batch operations to start/stop/restart all processes on a node
- Real-time process status updates

#### User Management (Admin only)
- Create and manage user accounts
- Assign admin privileges
- View user activity

#### Profile Management
- **NEW**: Users can view and update their personal profiles
- Change password and email
- View account information

#### Activity Logs
- Comprehensive logging of all user actions
- System events and process changes
- Searchable and filterable logs

### API Endpoints

The application provides a comprehensive REST API:

- `POST /api/auth/login` - User authentication
- `GET /api/nodes` - List all nodes
- `GET /api/nodes/:name/processes` - Get processes for a node
- `POST /api/nodes/:name/start-all` - **NEW**: Start all processes on a node
- `POST /api/nodes/:name/stop-all` - **NEW**: Stop all processes on a node
- `POST /api/nodes/:name/restart-all` - **NEW**: Restart all processes on a node
- `GET /api/profile` - **NEW**: Get user profile
- `PUT /api/profile` - **NEW**: Update user profile
- `GET /api/users` - List users (admin only)
- `GET /api/activity-logs` - Get activity logs
- `GET /ws` - WebSocket endpoint for real-time updates

### WebSocket Events

Real-time updates are provided through WebSocket connections:

- `node_update` - Node status changes
- `process_update` - Process status changes
- `process_status_change` - Individual process state changes
- `system_stats` - System statistics updates
- `activity_log` - New activity log entries

## Development

### Project Structure

```
go-cesi/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── api/                    # API handlers
│   ├── auth/                   # Authentication service
│   ├── database/               # Database configuration
│   ├── loggers/                # Activity logging
│   ├── models/                 # Data models
│   ├── supervisor/             # Supervisor service
│   └── websocket/              # WebSocket implementation
├── web/
│   ├── react-app/              # React frontend application
│   │   ├── src/
│   │   │   ├── components/     # React components
│   │   │   ├── contexts/       # React contexts
│   │   │   ├── pages/          # Page components
│   │   │   └── services/       # API services
│   │   └── build/              # Built React app
│   ├── static/                 # Static assets
│   └── templates/              # Go templates (legacy)
├── config.toml                 # Configuration file
└── build-react.sh              # React build script
```

### Adding New Features

1. **Backend**: Add new API endpoints in `internal/api/`
2. **Frontend**: Add new React components in `web/react-app/src/`
3. **Database**: Add new models in `internal/models/`
4. **WebSocket**: Extend WebSocket events in `internal/websocket/`

## Troubleshooting

### Common Issues

1. **React build fails**: Ensure Node.js and npm are properly installed
2. **WebSocket connection fails**: Check firewall settings and authentication
3. **Supervisor connection fails**: Verify Supervisor XML-RPC is enabled and accessible
4. **Database errors**: Ensure write permissions in the application directory

### Logs

The application logs important events to the console. For production deployments, consider redirecting logs to files:

```bash
./go-cesi > app.log 2>&1
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

This project is licensed under the MIT License.
# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Architecture

lazy-ctrl is an integrated remote computer control system consisting of:

1. **Controller Agent (Go)** - Integrated control server with multi-protocol support
   - Location: `controller-agent/`
   - **HTTP Server**: REST API on port 7070
   - **gRPC Server**: gRPC API on port 7071
   - **MQTT Client**: IoT platform integration support
   - **Integrated Features**:
     - Command configuration management (CRUD operations)
     - User authentication (PIN-based security)
     - Security controls (rate limiting, whitelist)
     - Homepage layout management
     - Real-time command execution
     - Swagger API documentation

2. **Control Frontend (React + Vite)** - Complete web interface
   - Location: `lazy-ctrl-ui/`
   - Web interface for controlling local computer
   - Features: command templates, layout management, real-time execution
   - Responsive design with mobile support

**Note**: Config Service and Cloud Gateway have been integrated into the Controller Agent for simplified deployment.

## Development Commands

### Controller Agent (Go)
```bash
cd controller-agent
go run main.go                    # Start the agent server
go build -o lazy-ctrl-agent main.go  # Build executable
```

### Frontend Development (React + Vite)
```bash
cd lazy-ctrl-ui
npm run dev           # Start development server
npm run build         # Build for production
npm run lint          # Run linting
```

### Workspace Management
```bash
pnpm install          # Install dependencies for all packages
```

## Key Configuration Files

- `controller-agent/config/commands.json` - Command ID to shell command mappings
- `pnpm-workspace.yaml` - Defines the monorepo package structure
- `项目需求模块.md` - Complete project requirements and feature specifications

## API Endpoints

### Controller Agent (HTTP API - Port 7070)
- `GET /api/v1/commands` - Get all available commands
- `POST /api/v1/commands` - Create new command
- `PUT /api/v1/commands/{id}` - Update command configuration
- `DELETE /api/v1/commands/{id}` - Delete command
- `GET /api/v1/execute?id={command_id}` - Execute registered command by ID
- `POST /api/v1/reload` - Reload command configuration
- `GET /api/v1/health` - Health check
- `POST /api/v1/auth/verify` - PIN verification
- `GET /api/v1/docs` - Swagger API documentation

### Controller Agent (gRPC API - Port 7071)
- Full gRPC service mirror of HTTP API
- Protocol buffer definitions in `/proto` directory

### Controller Agent (MQTT Client)
- Configurable MQTT broker connection
- Support for third-party IoT platforms (Aliyun, Tencent Cloud)

### Frontend Architecture
- **Command Templates**: Built-in templates for common commands
- **Configuration Flow**: Template → Parameters → Save to Backend → Display in Command List
- **Layout Management**: Commands can be added to homepage for quick access
- **Real-time Execution**: Live feedback for command execution status

## Command Registration

Commands are registered via API or stored in `controller-agent/config/commands.json` with v3.0 format:
```json
{
  "command_id": {
    "name": "Command Name",
    "command": "shell_command_or_script_path",
    "description": "Command description",
    "showOnHomePage": true,
    "position": {"x": 0, "y": 0, "w": 2, "h": 1},
    "color": "#blue"
  }
}
```

## Module Dependencies

The controller agent uses a modular architecture with internal packages:
- `internal/handler` - HTTP/gRPC request handling and routing
- `internal/executor` - Cross-platform shell command execution
- `internal/config` - Configuration management and validation
- `internal/security` - Authentication, rate limiting, and security controls
- `internal/mqtt` - MQTT client integration
- `proto/` - Protocol buffer definitions for gRPC
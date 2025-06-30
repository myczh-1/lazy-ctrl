# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Architecture

lazy-ctrl is a multi-component system for remote computer control consisting of:

1. **Controller Agent (Go)** - Local computer agent that executes system commands
   - Location: `controller-agent/`
   - Serves HTTP API on port 7070
   - Loads command mappings from `config/commands.json`
   - Executes shell commands via executor package

2. **Config Service (Node.js + Midway)** - Configuration management service
   - Location: `config-service/` (to be implemented)
   - Manages command configurations via REST API
   - Provides bridge between frontend and controller agent

3. **Control Frontend (React + Vite)** - User interface
   - Location: `lazy-ctrl-ui/`
   - Web interface for controlling local computer
   - Features: command templates, layout management, real-time execution

4. **Cloud Gateway (Node.js)** - Optional cloud relay service
   - Location: `cloud-gateway/` (to be implemented)
   - Handles remote device control and user authentication

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

### Controller Agent
- `GET /api/v1/commands` - Get all available commands
- `GET /api/v1/execute?id={command_id}` - Execute registered command by ID
- `POST /api/v1/reload` - Reload command configuration
- `GET /api/v1/health` - Health check

### Frontend Architecture
- **Command Templates**: Built-in templates for common commands
- **Configuration Flow**: Template → Parameters → Save to Backend → Display in Command List
- **Layout Management**: Commands can be added to homepage for quick access
- **Real-time Execution**: Live feedback for command execution status

## Command Registration

Commands are registered in `controller-agent/config/commands.json` with format:
```json
{
  "command_id": "shell_command_or_script_path"
}
```

## Module Dependencies

The controller agent uses a simple HTTP server with internal packages:
- `internal/handler` - HTTP request handling and command lookup
- `internal/executor` - Shell command execution wrapper

Future Node.js services will use Midway framework for dependency injection and service architecture.
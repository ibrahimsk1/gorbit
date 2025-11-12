# Orbital Rush

A server-authoritative multiplayer space game built with Go (server) and TypeScript/PixiJS (client).

## Overview

Orbital Rush is a deterministic, server-authoritative game where players navigate a ship through space, collecting energy pallets while avoiding the sun. The game uses a fixed-tick simulation running at 30Hz on the server, with client-side prediction and server reconciliation.

## Architecture

- **Server**: Go-based authoritative simulation engine
- **Client**: TypeScript/PixiJS game client with WebSocket communication
- **Protocol**: JSON over WebSocket for real-time communication

## Prerequisites

- **Go**: 1.23 or later
- **Node.js**: 20.x or later
- **Docker**: 24.x or later 
- **Make**: For running common tasks

## Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd gorbit
   ```

2. **Set up the server**
   ```bash
   cd server
   go mod download
   make build
   ```

3. **Set up the client**
   ```bash
   cd client
   npm install
   npm run build
   ```

4. **Run locally**
   ```bash
   # Terminal 1: Server
   make run-server  # or: cd server && ./bin/server
   
   # Terminal 2: Client
   make run-client  # or: cd client && npm run dev
   ```

### Docker Development

1. **Build and start services**
   ```bash
   docker compose up --build
   ```

2. **Access the application**
   - Client: http://localhost:3000
   - Server: http://localhost:8080

## Development Workflow

### Running Tests

```bash
# Run all tests
make test

# Run Go tests only
cd server && make test

# Run TypeScript tests only
cd client && npm test
```

### Linting and Formatting

```bash
# Lint all code
make lint

# Format all code
make format

# Lint Go code only
cd server && make lint

# Lint TypeScript code only
cd client && npm run lint
```

### Building

```bash
# Build server
cd server && make build

# Build client
cd client && npm run build
```

## Project Structure

```
gorbit/
├── server/          # Go server application
│   ├── cmd/         # Application entry points
│   ├── internal/     # Internal packages
│   ├── go.mod       # Go module definition
│   └── Makefile     # Server build/test targets
├── client/          # TypeScript client application
│   ├── src/         # Source code
│   ├── package.json # Node.js dependencies
│   └── vite.config.ts # Vite configuration
├── docs/            # Documentation
└── docker-compose.yml # Docker services
```

## Environment Variables

### Server

The server uses the `PORT` environment variable (defaults to 8080 if not set).

### Client

The client uses Vite's default configuration. Environment variables can be configured via `.env` files if needed.

## Testing Strategy

The project follows a test-first approach with contract tests for infrastructure:

- **Unit Tests**: Pure function tests for simulation logic
- **Integration Tests**: Tests for adapters and transport layers
- **Contract Tests**: Infrastructure and workspace validation
- **E2E Tests**: Full system validation scenarios

Tests are labeled with:
- `scope:` - Test size (unit, integration, contract, e2e)
- `loop:` - G-loop being proven (g0-work, g1-physics, etc.)
- `layer:` - System layer (sim, server, client, infra)

## Contributing

1. Create a branch from `main`
2. Make your changes following the test-first approach
3. Ensure all tests pass: `make test`
4. Ensure linting passes: `make lint`
5. Submit a pull request

## G-Loop Progression

The project follows an inside-out development approach with G-loops:

- **G0**: Workspace bootstrap (current)
- **G1**: Math & Physics Core
- **G2**: Game Rules
- **G3**: Orchestration
- **G4**: Protocol & Contracts
- **G5**: Adapters (Server)
- **G6**: Client Rendering
- **G7**: Observability & Ops
- **G8**: Scenarios, Perf, Anti-Cheat



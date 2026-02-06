# ⚡ Workflow API

A Go-based API for managing and executing workflow automations. Provides endpoints to retrieve workflow definitions and execute workflows, with PostgreSQL for persistent storage.

## 🛠️ Tech Stack

- Go 1.25+
- PostgreSQL
- Docker (for development and deployment)

## 🚀 Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL
- Docker & Docker Compose (recommended for development)

### 1. Configure Database

Set the `DATABASE_URL` environment variable:

```
DATABASE_URL=postgres://user:password@host:port/dbname?sslmode=disable
```

Ensure PostgreSQL is running and accessible.

### 2. Run the API

- With Docker Compose (recommended):
  ```bash
  docker-compose up --build api
  ```
- Or run locally:
  ```bash
  go run main.go
  ```

## 📋 API Endpoints

| Method | Endpoint                         | Description                        |
| ------ | -------------------------------- | ---------------------------------- |
| GET    | `/api/v1/workflows/{id}`         | Load a workflow definition         |
| POST   | `/api/v1/workflows/{id}/execute` | Execute the workflow synchronously |

### Example Usage

#### GET workflow definition

```bash
curl http://localhost:8086/api/v1/workflows/550e8400-e29b-41d4-a716-446655440000
```

#### POST execute workflow

```bash
curl -X POST http://localhost:8086/api/v1/workflows/550e8400-e29b-41d4-a716-446655440000/execute \
     -H "Content-Type: application/json" \
     -d '{}'
```

## 🗄️ Database

- The API uses `api/pkg/db.DefaultConfig()` and reads the URI from `DATABASE_URL`.
- For schema/configuration details, see the main project README or this file's comments.

## 📖 API Documentation

Interactive Swagger UI is available at: **http://localhost:8080/swagger/index.html**

To regenerate docs after modifying endpoints:
```bash
swag init -g main.go -o docs
```

## 🏗️ Architecture Decisions

### 1. Reusable Engine Package (`pkg/engine`)

The workflow execution logic lives in a standalone package with no HTTP or database dependencies.

**Why:** Enables reuse across different applications, easier testing, and cleaner separation of concerns.

**Trade-off:** Requires type conversion between engine and service layers.

### 2. Strategy Pattern for Handlers

Each node type (start, form, weather, condition, email, end) has its own handler implementing a common interface.

**Why:** Adding new node types doesn't require modifying existing code. Each handler is independently testable.

**Trade-off:** More files compared to a single switch statement.

### 3. Graph Traversal with Linear Execution

Workflows are represented as directed graphs but executed one node at a time (no parallel branches).

**Why:** Matches the current UI's single-path model. Simpler state management without concurrency.

**Trade-off:** Cannot run parallel branches simultaneously.

### 4. FormData vs Node Metadata

- **FormData**: Runtime user input (name, email, city)
- **Metadata**: Design-time workflow config (operator, threshold, subject)

**Why:** Separates what users provide at execution time from how the workflow is configured.

### 5. Normalized Database Schema

Workflows, nodes, and edges are stored in separate tables rather than a single JSON blob.

**Why:** Enables querying individual elements, better data integrity with foreign keys, more efficient partial updates.

**Trade-off:** Multiple queries needed to load a complete workflow.

## 🧪 Testing

```bash
# Run all tests
go test ./...

# With coverage
go test ./... -cover
```

## 📁 Project Structure

```
api/
├── pkg/engine/           # Reusable execution engine (no HTTP/DB deps)
│   ├── executor.go       # Graph traversal and execution
│   ├── handler.go        # NodeHandler interface + Registry
│   └── handlers/         # Individual node type implementations
├── pkg/weather/          # Open-Meteo API client
├── services/workflow/    # HTTP handlers + repository
└── docs/                 # Generated Swagger docs
```

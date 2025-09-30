# Go-Fred

A RESTful web application for running one-time or periodic tasks, built with Go and Gin. Supports both synchronous and asynchronous task execution with an event publication layer.

## Features

- **RESTful API**: JSON-based API for task management
- **Synchronous & Asynchronous Execution**: Run tasks immediately or in the background
- **Event Publishing**: Configurable event publisher (no-op or Kafka)
- **Task Types**: Built-in executors for common task patterns
- **Configuration**: YAML-based configuration system
- **Concurrent Execution**: Configurable maximum concurrent tasks

## Quick Start

### Prerequisites

- Go 1.19 or later
- (Optional) Kafka for event publishing

### Installation

1. Clone the repository:

```bash
git clone <repository-url>
cd go-fred
```

2. Install dependencies:

```bash
go mod download
```

3. Build the application:

```bash
go build -o go-fred .
```

4. Run the application:

```bash
./go-fred
```

The server will start on `localhost:8080` by default.

## Configuration

The application uses a YAML configuration file (`config.yaml`):

```yaml
server:
  host: "localhost"
  port: 8080

events:
  publisher: "noop" # "noop" or "kafka"
  kafka:
    brokers: ["localhost:9092"]
    topic: "go-fred-events"

tasks:
  max_concurrent: 10
  timeout_seconds: 300
```

### Configuration Options

- **server**: HTTP server configuration
  - `host`: Server host (default: "localhost")
  - `port`: Server port (default: 8080)

- **events**: Event publishing configuration
  - `publisher`: Event publisher type ("noop" or "kafka")
  - `kafka`: Kafka-specific configuration (required if publisher is "kafka")
    - `brokers`: List of Kafka broker addresses
    - `topic`: Kafka topic for events

- **tasks**: Task execution configuration
  - `max_concurrent`: Maximum number of concurrent tasks (default: 10)
  - `timeout_seconds`: Task timeout in seconds (default: 300)

## API Reference

### Base URL

```
http://localhost:8080/api/v1
```

### Endpoints

#### Health Check

```http
GET /health
```

Returns the health status of the server.

**Response:**

```json
{
  "status": "healthy",
  "service": "go-fred"
}
```

#### Create Task

```http
POST /tasks
```

Creates a new task.

**Request Body:**

```json
{
  "type": "echo",
  "input": {
    "message": "Hello, World!"
  },
  "async": false
}
```

**Response:**

```json
{
  "task": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "type": "echo",
    "status": "pending",
    "input": {
      "message": "Hello, World!"
    },
    "created_at": "2024-01-01T12:00:00Z",
    "is_async": false
  }
}
```

#### List Tasks

```http
GET /tasks
```

Returns all tasks.

**Response:**

```json
{
  "tasks": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "type": "echo",
      "status": "completed",
      "input": {
        "message": "Hello, World!"
      },
      "output": {
        "echo": {
          "message": "Hello, World!"
        },
        "message": "Task executed successfully"
      },
      "created_at": "2024-01-01T12:00:00Z",
      "started_at": "2024-01-01T12:00:01Z",
      "completed_at": "2024-01-01T12:00:02Z",
      "duration_ms": 1000,
      "is_async": false
    }
  ],
  "total": 1
}
```

#### Get Task

```http
GET /tasks/{id}
```

Returns a specific task by ID.

**Response:**

```json
{
  "task": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "type": "echo",
    "status": "completed",
    "input": {
      "message": "Hello, World!"
    },
    "output": {
      "echo": {
        "message": "Hello, World!"
      },
      "message": "Task executed successfully"
    },
    "created_at": "2024-01-01T12:00:00Z",
    "started_at": "2024-01-01T12:00:01Z",
    "completed_at": "2024-01-01T12:00:02Z",
    "duration_ms": 1000,
    "is_async": false
  }
}
```

#### Execute Task (Synchronous)

```http
POST /tasks/{id}/execute
```

Executes a task synchronously and returns the result.

**Response:**

```json
{
  "task": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "type": "echo",
    "status": "completed",
    "input": {
      "message": "Hello, World!"
    },
    "output": {
      "echo": {
        "message": "Hello, World!"
      },
      "message": "Task executed successfully"
    },
    "created_at": "2024-01-01T12:00:00Z",
    "started_at": "2024-01-01T12:00:01Z",
    "completed_at": "2024-01-01T12:00:02Z",
    "duration_ms": 1000,
    "is_async": false
  }
}
```

#### Execute Task (Asynchronous)

```http
POST /tasks/{id}/execute-async
```

Executes a task asynchronously and returns immediately.

**Response:**

```json
{
  "task": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "type": "echo",
    "status": "running",
    "input": {
      "message": "Hello, World!"
    },
    "created_at": "2024-01-01T12:00:00Z",
    "started_at": "2024-01-01T12:00:01Z",
    "is_async": false
  }
}
```

#### Cancel Task

```http
DELETE /tasks/{id}
```

Cancels a running task.

**Response:**

```json
{
  "task": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "type": "echo",
    "status": "cancelled",
    "input": {
      "message": "Hello, World!"
    },
    "created_at": "2024-01-01T12:00:00Z",
    "started_at": "2024-01-01T12:00:01Z",
    "completed_at": "2024-01-01T12:00:02Z",
    "duration_ms": 1000,
    "is_async": false
  }
}
```

#### Get Task Types

```http
GET /task-types
```

Returns all supported task types.

**Response:**

```json
{
  "task_types": ["echo", "sleep", "error", "math"]
}
```

## Built-in Task Types

### Echo

Echoes the input data.

**Input:**

```json
{
  "type": "echo",
  "input": {
    "message": "Hello, World!"
  }
}
```

**Output:**

```json
{
  "echo": {
    "message": "Hello, World!"
  },
  "message": "Task executed successfully"
}
```

### Sleep

Sleeps for a specified duration.

**Input:**

```json
{
  "type": "sleep",
  "input": {
    "duration": 5
  }
}
```

**Output:**

```json
{
  "slept_for_seconds": 5,
  "message": "Sleep completed successfully"
}
```

### Error

Always fails with a custom error message.

**Input:**

```json
{
  "type": "error",
  "input": {
    "message": "Custom error message"
  }
}
```

**Error:**

```json
{
  "error": "Custom error message"
}
```

### Math

Performs basic math operations.

**Input:**

```json
{
  "type": "math",
  "input": {
    "operation": "add",
    "a": 10,
    "b": 5
  }
}
```

**Output:**

```json
{
  "operation": "add",
  "a": 10,
  "b": 5,
  "result": 15
}
```

**Supported operations:** `add`, `subtract`, `multiply`, `divide`

## Task Status

Tasks can have the following statuses:

- `pending`: Task created but not yet started
- `running`: Task is currently executing
- `completed`: Task completed successfully
- `failed`: Task failed with an error
- `cancelled`: Task was cancelled

## Event Publishing

The application publishes events during task lifecycle:

- `task.created`: When a task is created
- `task.started`: When a task starts executing
- `task.completed`: When a task completes successfully
- `task.failed`: When a task fails
- `task.cancelled`: When a task is cancelled

### Event Publisher Types

#### No-op Publisher (Default)

Logs events to stdout. No external dependencies.

#### Kafka Publisher

Publishes events to a Kafka topic. Requires Kafka configuration.

## Usage Examples

### Create and Execute a Task Synchronously

```bash
# Create a task
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "type": "echo",
    "input": {
      "message": "Hello, World!"
    },
    "async": false
  }'

# Execute the task (replace {id} with actual task ID)
curl -X POST http://localhost:8080/api/v1/tasks/{id}/execute
```

### Create and Execute a Task Asynchronously

```bash
# Create a task
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "type": "sleep",
    "input": {
      "duration": 10
    },
    "async": true
  }'

# Execute the task asynchronously (replace {id} with actual task ID)
curl -X POST http://localhost:8080/api/v1/tasks/{id}/execute-async

# Check task status
curl http://localhost:8080/api/v1/tasks/{id}
```

### List All Tasks

```bash
curl http://localhost:8080/api/v1/tasks
```

## Development

### Running Tests

The project includes comprehensive unit tests for all packages. You can run tests using:

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run tests with coverage summary
go test -cover ./...

# Using Make (if available)
make test
make test-verbose
make test-coverage
```

### Adding Custom Task Types

1. Implement the `TaskExecutor` interface:

```go
type CustomExecutor struct{}

func (e *CustomExecutor) Execute(ctx context.Context, task *models.Task) error {
    // Your task logic here
    task.Output = map[string]interface{}{
        "result": "success",
    }
    return nil
}

func (e *CustomExecutor) GetSupportedTypes() []string {
    return []string{"custom"}
}
```

2. Register the executor in the server setup:

```go
registry.Register("custom", &CustomExecutor{})
```

### Running with Kafka

1. Start Kafka (using Docker):

```bash
docker run -d --name kafka -p 9092:9092 apache/kafka:latest
```

2. Update `config.yaml`:

```yaml
events:
  publisher: "kafka"
  kafka:
    brokers: ["localhost:9092"]
    topic: "go-fred-events"
```

3. Run the application:

```bash
./go-fred
```

## License

This project is licensed under the MIT License.

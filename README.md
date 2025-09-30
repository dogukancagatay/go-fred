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

## License

This project is licensed under the MIT License.

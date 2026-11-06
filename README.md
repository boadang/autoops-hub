# autoops-hubAutoOps Hub
![AutoOps Hub CI](https://github.com/boadang/autoops-hub/actions/workflows/ci.yml/badge.svg)

AutoOps Hub is a lightweight, concurrent system automation and infrastructure monitoring tool written in Go. Built with a focus on Clean Architecture, memory safety, and practical engineering principles, it enables parallel execution of system verification tasks while maintaining a reliable SQLite audit trail.

Architectural Overview

The project adheres to Clean Architecture (Separation of Concerns) principles to ensure high maintainability, testability, and clear dependency boundaries:

Command Layer (cmd/): Manages the CLI interface using Cobra. It handles command parsing, user flags, and input validation without containing business logic or storage details.

Engine Layer (internal/engine/): Implements core automation logic, concurrency patterns, process execution, and time management. It operates independently of CLI and storage frameworks.

Storage Layer (internal/storage/): Encapsulates persistence logic, audit logging, and SQLite database connections using standard SQL drivers (modernc.org/sqlite).

Concurrency Model

AutoOps Hub utilizes a Fan-Out / Fan-In Worker Pool design pattern:

Fan-Out: Distributes work targets across a fixed pool of concurrent goroutines via buffered job channels.

Lifecycle & Cancelation: Uses Go's context.Context (context.WithTimeout) to enforce strict task deadlines, prevent goroutine leaks, and ensure non-blocking cleanup.

Fan-In: Aggregates execution results back through a dedicated results channel for logging and persistent audit storage.

Key Technical Features

Goroutine Worker Pool: Executes infrastructure verification tasks concurrently while capping resource usage.

Context Timeout Controls: Guarantees process cleanup and prevents hung execution threads.

Structured Logging (log/slog): Uses Go's standard structured logging framework for key-value logs adaptable to JSON outputs for log aggregators (e.g., Grafana, ELK).

SQLite Audit Trail: Automatically records execution metadata (TaskID, Name, Output, Error, Elapsed) in a persistent database.

Data Race Inspected: Tested and validated with the Go Race Detector (go test -race).

Getting Started

Prerequisites

Go: Version 1.22 or higher installed on your system.

Git: Installed for cloning the repository.

Installation

Clone the repository and build the binary:

git clone https://github.com/boadang/autoops-hub.git
cd autoops-hub/backend
go build -o autoops-hub .


Usage

Run System Scan

Execute the infrastructure scan command with an optional timeout flag:

# Run scan with default parameters
./autoops-hub scan

# Run scan with a custom timeout flag (e.g., 10 seconds)
./autoops-hub scan --timeout=10s


CLI Command Options

Usage:
  autoops-hub scan [flags]

Flags:
  -h, --help              Help for scan
  -t, --timeout duration  Maximum runtime allowed for scanning targets (default 5s)


Verification & Testing

Running Unit Tests

Execute unit tests to verify engine and storage functionality:

go test -v ./internal/engine/...


Checking for Race Conditions

Validate concurrency memory safety using the Go Race Detector:

go test -v -race ./internal/engine/...


Project Structure

autoops-hub/
└── backend/
    ├── cmd/
    │   ├── root.go       # Root Cobra command configuration
    │   └── scan.go       # CLI scan sub-command definition
    ├── internal/
    │   ├── engine/
    │   │   ├── worker.go      # Worker pool logic & process execution
    │   │   └── worker_test.go # Unit tests for worker pool
    │   └── storage/
    │       └── db.go          # SQLite database connection & schema
    ├── go.mod             # Module dependencies
    ├── go.sum             # Module checksums
    └── main.go            # Application entry point


License

This project is open-source and available under the MIT License.
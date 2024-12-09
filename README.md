# Loggerator-Go

Loggerator-Go is a Go-based application for reading and processing log files. It provides an HTTP API to retrieve log file contents, with support for filtering and handling various edge cases.

## Features

- Read log files from a specified directory
- Retrieve log file contents via an HTTP API
- Filter log files by extensions
- Handle non-readable files and empty directories
- Reverse read lines from log files

## Requirements

- Go 1.18 or later
- `github.com/go-chi/chi/v5` - simple, fast, and idiomatic router
- `github.com/sirupsen/logrus` - structured, pluggable logging for Go
- `github.com/stretchr/testify` - testing toolkit

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/leBruchette/loggerator-go.git
    cd loggerator-go
    ```

2. Install dependencies:
    ```sh
    go mod tidy
    ```

## Usage

1. Set the `LOG_DIR` environment variable to the directory containing your log files. If not set, it defaults to `/var/log`.

2. Run the application:
    ```sh
    go run main.go
    ```

3. Access the logs via the HTTP API:
    ```sh
    curl http://localhost:8080/logs
    ```

## API Endpoints

### GET /logs

Retrieve log file contents.

#### Query Parameters

- `excludedFileTypes` (optional): Comma-separated list of file extensions to exclude.

#### Example

```sh
curl "http://localhost:8080/logs?excludedFileTypes=.log,.txt"
```

## Testing
Run the tests using the following command:
```sh
go test ./...
```
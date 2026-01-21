# Mini Content-Addressable Storage

A content-addressable storage (CAS) system inspired by CVMFS. Mini-CAS stores files by their SHA-256 hash rather than their original path, enabling automatic deduplication and efficient storage management.

## Overview

Content-addressable storage eliminates duplicate files automatically by storing each unique file only once, regardless of how many times it appears in your filesystem. Mini-CAS computes a SHA-256 hash of each file's content and uses that hash as the storage key, making identical files share the same storage location.

## Features

- **Automatic deduplication**: Identical files are stored only once
- **Content-based addressing**: Files are stored by their SHA-256 hash
- **Memory-efficient streaming**: Large files are processed without loading entirely into memory
- **Interface-based design**: Follows Go best practices using `io.Reader`/`io.Writer` for flexibility and testability
- **Efficient sharding**: 2-level directory sharding prevents filesystem performance degradation
- **Immutable storage**: Blobs are write-once, read-many with enforced permissions
- **SQLite catalog**: Fast, reliable SQLite database with WAL mode for concurrent access
- **Directory support**: Add entire directory trees with a single command
- **Integrity verification**: Verify stored content against catalog hashes
- **RESTful HTTP API**: Production-ready HTTP server with streaming endpoints
- **ETag caching**: Content-based ETags enable perfect HTTP caching
- **Graceful shutdown**: Handles SIGTERM/SIGINT signals properly
- **Flexible configuration**: Command-line flags and environment variables support
- **Location transparency**: Unified client interface for local and remote storage access
- **Remote operation**: All CLI commands work seamlessly with remote servers via HTTP
- **Thread-safe access**: Safe concurrent operations with read-write mutex protection
- **Docker support**: Multi-stage Docker builds with compose configuration for easy deployment

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/SteliosSpanos/mini-CAS.git
cd mini-CAS

# Build the binary
go build -o cas ./cmd/main
```

### Docker

```bash
# Using Docker Compose (recommended)
docker-compose up --build

# Using Docker directly
docker build -t mini-cas .
docker run -p 8080:8080 -v $(pwd)/cas-data:/data mini-cas
```

The Docker image automatically initializes a CAS repository and starts the HTTP server on port 8080. Data persists in the mounted volume.

## Quick Start

### Local Mode (Default)

```bash
# Initialize a new CAS repository
./cas init

# Add some files
./cas add myfile.txt
./cas add /path/to/directory

# View what's stored
./cas ls

# Check deduplication savings
./cas status

# Retrieve file contents
./cas cat myfile.txt

# Verify storage integrity
./cas verify
```

### Remote Mode

```bash
# Start a server in one terminal
./cas serve --port 8080 --auth-token mysecret

# In another terminal, configure remote access
export CAS_SERVER_URL=http://localhost:8080
export CAS_AUTH_TOKEN=mysecret

# Now all commands work remotely
./cas add myfile.txt
./cas ls
./cas cat myfile.txt
./cas verify
```

## CLI Commands

### init

Initialize a CAS repository in the current directory.

```bash
./cas init
```

Creates a `.cas/` directory structure, similar to how Git creates `.git/`.

### add

Add files or directories to the storage.

```bash
# Add a single file
./cas add myfile.txt

# Add an entire directory
./cas add /path/to/directory
```

When adding directories, Mini-CAS automatically skips the `.cas/` directory to avoid recursion. Each file added displays its short hash (first 8 characters). Files are processed using streaming I/O to handle large files efficiently without loading them entirely into memory.

### ls

List all tracked files in the catalog.

```bash
./cas ls
```

Displays a table with filepath, hash (abbreviated), human-readable size, and modification time.

### cat

Retrieve and display file contents from storage.

```bash
./cas cat <filepath>
```

Retrieves file content based on the catalog entry using streaming I/O. Supports piping to other commands:

```bash
./cas cat config.json | jq .
```

### status

Display repository statistics and deduplication metrics.

```bash
./cas status
```

Shows:
- Total files tracked vs unique blobs stored
- Total logical size vs actual storage used
- Space saved through deduplication (with percentage)

### hash

Compute the SHA-256 hash of any file without adding it to storage.

```bash
./cas hash <filename>
```

This is a standalone utility that does not require an initialized CAS repository. Useful for verifying file integrity or pre-checking what hash a file would receive.

### verify

Verify the integrity of all stored content.

```bash
./cas verify
```

Iterates through all catalog entries and verifies that each stored blob matches its expected SHA-256 hash. Reports:
- OK: Files that pass verification
- MISSING: Files whose blobs are not found in storage
- CORRUPT: Files whose stored content does not match the expected hash

Exits with code 1 if any issues are detected, making it suitable for scripting and automated integrity checks.

### serve

Start an HTTP API server to access the CAS repository over the network.

```bash
./cas serve [options]
```

Options:
- `--port`: HTTP port (default: 8080, env: CAS_PORT)
- `--host`: Bind address (default: 0.0.0.0, env: CAS_HOST)
- `--auth-token`: Bearer token for write operations (optional, env: CAS_AUTH_TOKEN)
- `--cors-origins`: Comma-separated CORS origins (default: *, env: CAS_CORS_ORIGINS)

Examples:

```bash
# Start with defaults (port 8080, no auth)
./cas serve

# Start on custom port with authentication
./cas serve --port 3000 --auth-token mysecret

# Using environment variables
export CAS_PORT=8080
export CAS_AUTH_TOKEN=production-secret-token
export CAS_CORS_ORIGINS="https://app.example.com,https://cdn.example.com"
./cas serve
```

The server runs until interrupted with SIGINT (Ctrl+C) or SIGTERM, performing graceful shutdown with a 30-second timeout.

## Architecture

Mini-CAS is built with a layered architecture:

```
+------------------+------------------+
|  Command Layer   |   HTTP Server    |  CLI and REST API interfaces
+------------------+------------------+
|            Client Layer             |  Unified local/remote access interface
+-------------------------------------+
|         Repository Layer            |  Manages .cas/ directory structure
+-------------------------------------+
|          Catalog Layer              |  Maps file paths to content hashes
+-------------------------------------+
|          Storage Layer              |  Physical blob storage with sharding
+-------------------------------------+
|          Objects Layer              |  Blob types and SHA-256 hashing
+-------------------------------------+
```

- **Objects Layer**: Defines the `Blob` type and SHA-256 hashing
- **Storage Layer**: Manages physical blob storage with 2-level sharding and streaming I/O
- **Catalog Layer**: Maps original file paths to content hashes using SQLite database
- **Repository Layer**: Manages the `.cas/` directory structure
- **Client Layer**: Unified interface for local and remote storage access with thread-safe catalog operations
- **Command Layer**: User-facing CLI commands (init, add, ls, cat, status, hash, verify, serve)
- **HTTP Server**: RESTful API with streaming endpoints, middleware chain, and graceful shutdown

## Storage Structure

Files are stored using a 2-level sharding strategy based on the SHA-256 hash:

```
.cas/
├── storage/
│   └── ab/
│       └── cd/
│           └── abcd1234567890...  (full 64-char SHA-256 hash)
└── catalog.db
```

- **Blob storage**: 2-level sharding using first 4 hash characters scales to millions of files
- **Catalog database**: SQLite with WAL mode for concurrent reads and atomic writes
- **Database features**: Indexed by filepath (primary key) and hash for fast lookups

## Client Library

Mini-CAS provides a unified client abstraction that works with both local and remote CAS repositories. The client library offers a consistent interface regardless of whether you're accessing storage locally or via HTTP.

### Client Interface

The client package defines a `Client` interface that combines blob operations, catalog operations, and lifecycle management:

```go
type Client interface {
    BlobOperations      // Upload, Download, Stat, Exists
    CatalogOperations   // GetCatalog, GetEntry, AddEntry, SaveCatalog
    io.Closer          // Resource cleanup
}
```

### Local Client

The `LocalClient` provides direct access to a local CAS repository with zero overhead. It wraps the existing storage and catalog packages with thread-safe access using `sync.RWMutex`.

**Example usage:**

```go
import "github.com/SteliosSpanos/mini-CAS/pkg/client"

// Create local client
client, err := client.NewLocalClient(".cas")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Upload a file
file, _ := os.Open("example.txt")
hash, err := client.Upload(context.Background(), file)
file.Close()

// Download by hash
reader, err := client.Download(context.Background(), hash)
if err != nil {
    log.Fatal(err)
}
defer reader.Close()
io.Copy(os.Stdout, reader)

// Check if blob exists
exists, err := client.Exists(context.Background(), hash)

// Get catalog entries
entries, err := client.GetCatalog(context.Background())
```

### HTTP Client

The `HTTPClient` provides remote access to a CAS repository via the HTTP API. It handles authentication, connection pooling, and proper error handling. All blob and catalog operations are fully implemented.

**Example usage:**

```go
// Create HTTP client
client := client.NewHTTPClient("http://localhost:8080", "secret-token")
defer client.Close()

// Upload blob
file, _ := os.Open("data.bin")
hash, err := client.Upload(context.Background(), file)
file.Close()

// Download blob
reader, err := client.Download(context.Background(), hash)
if err != nil {
    log.Fatal(err)
}
defer reader.Close()
io.Copy(os.Stdout, reader)

// Check if blob exists
exists, err := client.Exists(context.Background(), hash)

// Get blob metadata
info, err := client.Stat(context.Background(), hash)

// Get catalog entries
entries, err := client.GetCatalog(context.Background())

// Get single entry
entry, err := client.GetEntry(context.Background(), "path/to/file.txt")

// Add catalog entry (blob must exist on server)
entry := catalog.Entry{
    Filepath: "config.json",
    Hash:     hash,
    Filesize: 1024,
    ModTime:  time.Now(),
}
err = client.AddEntry(context.Background(), entry)
```

### Configuration

The client package supports multiple configuration methods:

**1. Direct configuration:**

```go
cfg := client.Config{
    ServerURL: "http://localhost:8080",  // For HTTP client
    AuthToken: "secret-token",           // Optional auth token
    CASDir:    ".cas",                   // For local client
}
client, err := client.NewClient(cfg)
```

**2. Environment variables (12-factor):**

```go
// Reads from environment:
// - CAS_SERVER_URL: HTTP server URL (e.g., "http://localhost:8080")
// - CAS_AUTH_TOKEN: Bearer token for authentication
// - CAS_DIR: Local CAS directory (defaults to ".cas")
client, err := client.NewClientFromEnv()
```

The factory automatically selects the appropriate client type:
- If `ServerURL` is set, creates an `HTTPClient`
- Otherwise, creates a `LocalClient` with the specified `CASDir`

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CAS_SERVER_URL` | HTTP server URL for remote access | (empty, uses local) |
| `CAS_AUTH_TOKEN` | Bearer token for HTTP authentication | (empty, no auth) |
| `CAS_DIR` | Local CAS repository directory | `.cas` |

**Example:**

```bash
# Use remote server
export CAS_SERVER_URL=http://cas.example.com:8080
export CAS_AUTH_TOKEN=production-secret

# Use local repository
export CAS_DIR=/var/lib/cas

# Your application automatically uses the correct client
./my-cas-application
```

### Error Handling

The client package defines custom error types for precise error handling:

```go
import "errors"

// Check for specific errors
reader, err := client.Download(ctx, hash)
if errors.Is(err, client.ErrBlobNotFound) {
    log.Println("Blob does not exist")
}
if errors.Is(err, client.ErrInvalidHash) {
    log.Println("Invalid hash format")
}

// HTTP errors include status codes
var httpErr *client.HTTPError
if errors.As(err, &httpErr) {
    log.Printf("HTTP %d: %s", httpErr.StatusCode, httpErr.Message)
}
```

**Available error types:**
- `ErrBlobNotFound`: Requested blob does not exist
- `ErrEntryNotFound`: Catalog entry not found
- `ErrCatalogNotSupported`: Operation not supported by client type
- `ErrInvalidHash`: Hash format is invalid (must be 64 hex characters)
- `HTTPError`: HTTP request failed with status code and message

### Context Support

All client operations support context cancellation and timeouts:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
hash, err := client.Upload(ctx, reader)

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
go func() {
    <-interruptSignal
    cancel()
}()
reader, err := client.Download(ctx, hash)
```

## HTTP API Server

Mini-CAS includes a production-ready HTTP server that exposes the storage system via a RESTful API. The server uses streaming I/O for all blob operations, ensuring constant memory usage regardless of file size.

### API Endpoints

| Endpoint | Method | Auth Required | Description |
|----------|--------|---------------|-------------|
| `/health` | GET | No | Health check with repository statistics |
| `/blob/{hash}` | GET | No | Download blob by hash (streaming) |
| `/blob/{hash}` | HEAD | No | Check if blob exists (no body) |
| `/blob/{hash}/stat` | GET | No | Get blob metadata (hash, size, exists) |
| `/catalog` | GET | No | Get full catalog as JSON |
| `/catalog?filepath=path` | GET | No | Get single catalog entry by filepath |
| `/blobs` | POST | Yes | Upload blob (streaming, returns hash) |
| `/catalog` | POST | Yes | Add catalog entry (blob must exist) |

### Configuration

The server can be configured via command-line flags or environment variables:

```bash
# Command-line flags
./cas serve --port 8080 --host 0.0.0.0 --auth-token secret123 --cors-origins "*"

# Environment variables (12-factor compatible)
export CAS_PORT=8080
export CAS_HOST=0.0.0.0
export CAS_AUTH_TOKEN=secret123
export CAS_CORS_ORIGINS="https://app.example.com,https://cdn.example.com"
./cas serve
```

**Default values:**
- Port: 8080
- Host: 0.0.0.0 (all interfaces)
- Auth Token: none (optional, only required for POST operations)
- CORS Origins: * (all origins allowed)
- Read Timeout: 30 seconds
- Write Timeout: 30 seconds

### Usage Examples

**Health check:**
```bash
curl http://localhost:8080/health
# {"status":"ok","total_files":42,"unique_blobs":28}
```

**Download a blob:**
```bash
curl http://localhost:8080/blob/e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
# (blob content streamed to stdout)
```

**Check if blob exists (HEAD request):**
```bash
curl -I http://localhost:8080/blob/e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
# HTTP/1.1 200 OK
# Content-Type: application/octet-stream
# ETag: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
# Cache-Control: public, max-age=31536000, immutable
```

**Get blob metadata:**
```bash
curl http://localhost:8080/blob/e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855/stat
# {"hash":"e3b0c44...","size":1024,"exists":true}
```

**Get catalog:**
```bash
curl http://localhost:8080/catalog | jq
# {
#   "entries": {
#     "README.md": {
#       "filepath": "README.md",
#       "hash": "abc123...",
#       "filesize": 2048,
#       "modified": "2024-01-15T10:30:00Z"
#     }
#   }
# }
```

**Get single catalog entry:**
```bash
curl "http://localhost:8080/catalog?filepath=README.md"
# {"filepath":"README.md","hash":"abc123...","filesize":2048,"modified":"..."}
```

**Upload a blob:**
```bash
curl -X POST http://localhost:8080/blobs \
  -H "Authorization: Bearer secret123" \
  -H "Content-Type: application/octet-stream" \
  --data-binary @file.txt
# {"hash":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855","size":1024}
```

**Upload from stdin:**
```bash
echo "Hello, World!" | curl -X POST http://localhost:8080/blobs \
  -H "Authorization: Bearer secret123" \
  --data-binary @-
# {"hash":"dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f","size":14}
```

**Add catalog entry:**
```bash
curl -X POST http://localhost:8080/catalog \
  -H "Authorization: Bearer secret123" \
  -H "Content-Type: application/json" \
  -d '{
    "filepath": "config.json",
    "hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    "size": 1024,
    "modified": "2024-01-15T10:30:00Z"
  }'
# Returns 201 Created with the catalog entry
```

### API Features

- **Streaming I/O**: All blob operations use streaming, supporting arbitrarily large files with constant memory usage
- **Content-based ETags**: The blob's SHA-256 hash IS the ETag, enabling perfect HTTP caching
- **Immutable caching**: Blobs get `Cache-Control: public, max-age=31536000, immutable` headers
- **Authentication**: Bearer token authentication for write operations (POST), reads are public
- **CORS support**: Configurable CORS origins for browser-based applications
- **Graceful shutdown**: Handles SIGTERM/SIGINT with 30-second graceful shutdown timeout
- **Structured logging**: Request logging with method, path, status code, and duration
- **Panic recovery**: Recovery middleware catches panics and returns 500 errors
- **Standard errors**: JSON error responses with consistent structure

### Middleware Chain

The server applies middleware in the following order:

1. **Recovery**: Catches panics and returns 500 errors
2. **Logging**: Logs all requests with timing information
3. **CORS**: Handles CORS preflight and headers
4. **Auth**: Validates Bearer token for write operations (POST)

### Security Considerations

- Authentication is optional but recommended for production deployments
- Read operations (GET, HEAD) are always public
- Write operations (POST) require Bearer token authentication if configured
- CORS can be restricted to specific origins
- Catalog write operations (POST /catalog) validate that blobs exist before accepting entries
- Path traversal is prevented in catalog entries
- Blobs are immutable and stored with read-only permissions
- All hash inputs are validated (must be 64 hex characters)

## Project Structure

```
mini-CAS/
├── cmd/
│   └── main/           # Main entry point
├── commands/           # CLI command implementations
│   ├── init.go         # Repository initialization
│   ├── add.go          # Add files with streaming
│   ├── list.go         # List tracked files
│   ├── cat.go          # Retrieve file contents
│   ├── status.go       # Repository statistics
│   ├── hash.go         # Standalone hash utility
│   ├── verify.go       # Storage integrity verification
│   └── serve.go        # HTTP API server command
├── pkg/
│   ├── objects/        # Blob types and hashing
│   ├── storage/        # Physical storage management
│   ├── catalog/        # Path-to-hash mapping
│   ├── path/           # Repository initialization
│   ├── client/         # Unified local/remote client interface
│   │   ├── client.go   # Client interface definitions
│   │   ├── errors.go   # Custom error types
│   │   ├── config.go   # Configuration and factory functions
│   │   ├── local.go    # Local client implementation
│   │   └── http.go     # HTTP client implementation
│   └── server/         # HTTP server implementation
│       ├── server.go   # Server lifecycle and configuration
│       ├── handlers.go # HTTP request handlers
│       ├── middleware.go # Logging, CORS, auth, recovery
│       ├── response.go # JSON response helpers
│       └── routes.go   # Route registration
└── README.md           # This file
```

## Design Decisions

- **SHA-256 hashing**: Cryptographically secure hash function with excellent collision resistance
- **Go interface patterns**: Follows idiomatic Go design by accepting `io.Reader`/`io.Writer` interfaces rather than concrete types, enabling composition, testing without filesystem, and future extensibility
- **Streaming I/O**: Files are processed using `io.Copy` and `io.MultiWriter` to handle large files without excessive memory usage
- **Write-time deduplication**: Checks for existing blobs before writing to minimize I/O
- **Atomic writes**: New blobs are written to a temp file first, then renamed to final location
- **Immutable blobs**: Files stored with 0444 permissions prevent accidental modification
- **SQLite catalog**: WAL mode enables concurrent reads, indexed queries, and ACID guarantees
- **RESTful API design**: Resource-oriented design following REST principles
- **Content-based ETags**: SHA-256 hash serves as perfect ETag for HTTP caching
- **12-factor configuration**: Supports both command-line flags and environment variables
- **Middleware composition**: Clean separation of cross-cutting concerns (logging, auth, CORS)
- **Multi-stage Docker builds**: Minimal Alpine-based images with CGO-free builds for portability

## Local vs Remote Mode

Mini-CAS provides location transparency through a unified client interface. All CLI commands work identically in both local and remote modes.

### Local Mode

When no environment variables are set, Mini-CAS operates in local mode, accessing the `.cas/` directory directly:

```bash
# Local mode (default)
./cas init
./cas add myfile.txt
./cas ls
```

The client uses `LocalClient` which provides:
- Zero-overhead direct access to storage and catalog
- Thread-safe concurrent operations with `sync.RWMutex`
- Immediate catalog persistence

### Remote Mode

When `CAS_SERVER_URL` is set, Mini-CAS operates in remote mode, communicating with an HTTP server:

```bash
# Terminal 1: Start server
./cas serve --port 8080 --auth-token mysecret

# Terminal 2: Configure remote mode
export CAS_SERVER_URL=http://localhost:8080
export CAS_AUTH_TOKEN=mysecret

# All commands now work remotely
./cas add myfile.txt      # Uploads to server
./cas ls                  # Fetches catalog from server
./cas cat myfile.txt      # Downloads and displays from server
./cas verify              # Verifies against server storage
```

The client uses `HTTPClient` which provides:
- HTTP connection pooling for performance
- Automatic authentication with Bearer tokens
- Streaming uploads and downloads
- Proper error handling with HTTP status codes

### Command Behavior Differences

Most commands work identically in both modes. Key differences:

| Command | Local Mode | Remote Mode |
|---------|-----------|-------------|
| `init` | Creates `.cas/` directory | Not applicable (server manages repo) |
| `add` | Saves catalog after upload | Server saves catalog automatically |
| `ls` | Reads local catalog | Fetches catalog from server |
| `cat` | Reads from local storage | Downloads from server |
| `status` | Analyzes local catalog | Fetches catalog from server |
| `verify` | Opens local blobs | Downloads blobs from server |
| `serve` | Starts HTTP server | Not applicable |
| `hash` | Standalone utility | Standalone utility |

### Testing Remote Mode

To test the complete remote workflow:

```bash
# 1. Initialize and populate local repository
./cas init
./cas add test.txt

# 2. Start server in background
./cas serve --port 8080 --auth-token test123 &
SERVER_PID=$!

# 3. Configure remote access
export CAS_SERVER_URL=http://localhost:8080
export CAS_AUTH_TOKEN=test123

# 4. Add new files remotely
./cas add newfile.txt

# 5. List all files (from server)
./cas ls

# 6. Verify integrity (downloads from server)
./cas verify

# 7. Cleanup
kill $SERVER_PID
unset CAS_SERVER_URL CAS_AUTH_TOKEN
```

## Testing

Mini-CAS includes comprehensive test coverage across all packages. Tests use `t.TempDir()` for isolated environments and follow Go testing best practices.

### Running Tests

```bash
# Run all tests
go test ./pkg/...

# Run with verbose output
go test -v ./pkg/...

# Run specific package
go test ./pkg/objects
go test ./pkg/storage
go test ./pkg/catalog
go test ./pkg/path
go test ./pkg/client
go test ./pkg/server

# Run with race detection
go test -race ./pkg/...

# Coverage report
go test -cover ./pkg/...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./pkg/...
go tool cover -html=coverage.out -o coverage.html
```

### Test Files

| Package | Test File | What It Tests |
|---------|-----------|---------------|
| `pkg/objects` | `blob_test.go` | SHA-256 hashing, empty data, binary data |
| `pkg/storage` | `storage_test.go` | Blob I/O, streaming, sharding, deduplication |
| `pkg/catalog` | `catalog_test.go` | SQLite CRUD, JSON serialization, sorting |
| `pkg/path` | `path_test.go` | Repository init, directory structure |
| `pkg/client` | `local_test.go` | Upload/download, context, sentinel errors |
| `pkg/server` | `handlers_test.go` | HTTP handlers, auth, status codes |

## Development

### Building from Source

```bash
# Build
go build -o cas ./cmd/main
```

### Docker Development

```bash
# Build image
docker-compose build

# Start server with mounted volume
docker-compose up

# View logs
docker-compose logs -f

# Stop server
docker-compose down

# Access container shell
docker-compose exec cas-server sh
```

The `docker-compose.yml` mounts `./docker-data` as the repository directory, allowing you to inspect the SQLite database and blobs directly on your host machine.

## Dependencies

Mini-CAS uses the following key dependencies:

- **modernc.org/sqlite v1.44.2**: Pure Go SQLite implementation (no CGO required)
- **Standard library**: Minimal external dependencies for portability

The pure Go SQLite driver enables CGO-free builds, making the binary fully static and ideal for Docker containers and cross-platform distribution.

## License

This project is intended for educational purposes.

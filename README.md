# Mini Content-Addressable Storage

A content-addressable storage (CAS) system inspired by CVMFS. Mini-CAS stores files by their SHA-256 hash rather than their original path, enabling automatic deduplication and efficient storage management.

## Overview

Content-addressable storage eliminates duplicate files automatically by storing each unique file only once, regardless of how many times it appears in your filesystem. Mini-CAS computes a SHA-256 hash of each file's content and uses that hash as the storage key, making identical files share the same storage location.

## Features

- **Automatic deduplication**: Identical files stored only once by SHA-256 hash
- **Memory-efficient streaming**: Large files processed without loading into memory
- **Interface-based design**: Go best practices with `io.Reader`/`io.Writer`
- **Efficient sharding**: 2-level directory sharding prevents filesystem bottlenecks
- **Immutable storage**: Write-once, read-many blobs with enforced permissions
- **SQLite catalog**: WAL mode for concurrent access with ACID guarantees
- **Directory support**: Add entire directory trees with single command
- **Integrity verification**: Verify stored content against catalog hashes
- **RESTful HTTP API**: Production-ready server with streaming endpoints
- **Content-based caching**: ETags derived from hashes enable perfect HTTP caching
- **Location transparency**: Unified interface for local and remote storage
- **Thread-safe access**: Safe concurrent operations
- **Docker support**: Multi-stage builds with compose configuration
- **Merkle tree support**: Cryptographic proofs for data integrity

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

When adding directories, Mini-CAS automatically skips the `.cas/` directory to avoid recursion. Each file added displays its short hash (first 8 characters).

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

Retrieves file content based on the catalog entry. Supports piping to other commands:

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
- `--tls-cert`: TLS certificate file path (optional, env: CAS_TLS_CERT)
- `--tls-key`: TLS private key file path (optional, env: CAS_TLS_KEY)

Examples:

```bash
# Start with defaults (port 8080, no auth)
./cas serve

# Start on custom port with authentication
./cas serve --port 3000 --auth-token mysecret

# Start with TLS encryption (HTTPS)
./cas serve --port 8443 --tls-cert server.crt --tls-key server.key

# Using environment variables
export CAS_PORT=8080
export CAS_AUTH_TOKEN=production-secret-token
export CAS_CORS_ORIGINS="https://app.example.com,https://cdn.example.com"
./cas serve

# Production HTTPS server with authentication
export CAS_PORT=8443
export CAS_TLS_CERT=/etc/ssl/certs/cas-server.crt
export CAS_TLS_KEY=/etc/ssl/private/cas-server.key
export CAS_AUTH_TOKEN=production-secret-token
./cas serve
```

The server performs graceful shutdown with a 30-second timeout when interrupted.

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
|   Objects Layer   |  Merkle Layer   |  Hashing and Merkle trees
+-------------------------------------+
```

- **Objects Layer**: Defines the `Blob` type and SHA-256 hashing
- **Merkle Layer**: Builds Merkle trees, generates and verifies cryptographic proofs
- **Storage Layer**: Manages physical blob storage with 2-level sharding
- **Catalog Layer**: Maps original file paths to content hashes using SQLite database
- **Repository Layer**: Manages the `.cas/` directory structure
- **Client Layer**: Unified interface for local and remote storage access
- **Command Layer**: User-facing CLI commands (init, add, ls, cat, status, hash, verify, serve)
- **HTTP Server**: RESTful API with middleware chain

## Merkle Trees

Mini-CAS includes a Merkle tree implementation for verifying data integrity through cryptographic proofs. Merkle trees are binary hash trees where each leaf node represents a data hash, and each internal node is the hash of its children. This structure enables efficient verification that specific data exists in a larger dataset without transmitting the entire dataset.

### Features

- **Flexible hash functions**: Use any hash function (SHA-256, Blake3, etc.)
- **Automatic balancing**: Handles odd leaf counts by duplicating the last leaf
- **Proof generation**: Generate compact membership proofs for any leaf
- **Independent verification**: Verify proofs without rebuilding the entire tree
- **Integration with CAS**: Uses `objects.Hash` for consistent hashing

### Basic Usage

```go
import (
    "github.com/SteliosSpanos/mini-CAS/pkg/merkle"
    "github.com/SteliosSpanos/mini-CAS/pkg/objects"
)

// Create tree with SHA-256 hash function
tree := merkle.NewTree(objects.Hash)

// Build tree from blob hashes
leafHashes := []string{
    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    "6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b",
    "d4735e3a265e16eee03f59718b9b5d03019c07d8b6c51f90da3a666eec13ab35",
}
err := tree.Build(leafHashes)

// Get the Merkle root
root, err := tree.RootHash()
fmt.Printf("Merkle root: %s\n", root)

// Generate proof for leaf at index 1
proof, err := tree.GenerateProof(1)

// Verify the proof independently
isValid := proof.Verify(objects.Hash)
fmt.Printf("Proof valid: %v\n", isValid)
```

### Proof Structure

A Merkle proof contains:
- **LeafHash**: The hash of the data being proven
- **LeafIndex**: Position of the leaf in the tree
- **Siblings**: Hashes needed to reconstruct the path to the root
- **RootHash**: Expected root hash for verification

Proofs are compact (log₂N sibling hashes) and can be verified independently without access to the original tree or dataset.

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
- **Catalog database**: SQLite with WAL mode, indexed by filepath (primary key) and hash

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

The `LocalClient` provides direct access to a local CAS repository with zero overhead and thread-safe operations.

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

The `HTTPClient` provides remote access to a CAS repository via the HTTP API with authentication and connection pooling.

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
// Reads from CAS_SERVER_URL, CAS_AUTH_TOKEN, CAS_DIR
client, err := client.NewClientFromEnv()
```

The factory automatically selects `HTTPClient` if `CAS_SERVER_URL` is set, otherwise `LocalClient`.

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CAS_SERVER_URL` | HTTP server URL for remote access | (empty, uses local) |
| `CAS_AUTH_TOKEN` | Bearer token for authentication | (empty, no auth) |
| `CAS_DIR` | Local CAS repository directory | `.cas` |
| `CAS_PORT` | Server port | `8080` |
| `CAS_HOST` | Server bind address | `0.0.0.0` |
| `CAS_CORS_ORIGINS` | Comma-separated CORS origins | `*` |
| `CAS_TLS_CERT` | TLS certificate file path | (empty, HTTP mode) |
| `CAS_TLS_KEY` | TLS private key file path | (empty, HTTP mode) |

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

Mini-CAS includes a production-ready HTTP server that exposes the storage system via a RESTful API.

### API Endpoints

| Endpoint | Method | Auth Required | Description |
|----------|--------|---------------|-------------|
| `/health` | GET | No | Health check with repository statistics |
| `/blobs/{hash}` | GET | No | Download blob by hash (streaming) |
| `/blobs/{hash}` | HEAD | No | Check if blob exists (no body) |
| `/blobs/{hash}/stat` | GET | No | Get blob metadata (hash, size, exists) |
| `/catalog` | GET | No | Get full catalog as JSON |
| `/catalog?filepath=path` | GET | No | Get single catalog entry by filepath |
| `/blobs` | POST | Yes | Upload blob (streaming, returns hash) |
| `/catalog` | POST | Yes | Add catalog entry (blob must exist) |

### Configuration

Configure via command-line flags or environment variables (see Environment Variables section):

```bash
./cas serve --port 8080 --auth-token secret123 --cors-origins "*"
```

### Usage Examples

**Health check:**
```bash
curl http://localhost:8080/health
# {"status":"ok","total_files":42,"unique_blobs":28}
```

**Download a blob:**
```bash
curl http://localhost:8080/blobs/e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
# (blob content streamed to stdout)
```

**Check if blob exists (HEAD request):**
```bash
curl -I http://localhost:8080/blobs/e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
# HTTP/1.1 200 OK
# Content-Type: application/octet-stream
# ETag: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
# Cache-Control: public, max-age=31536000, immutable
```

**Get blob metadata:**
```bash
curl http://localhost:8080/blobs/e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855/stat
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

- **Streaming I/O**: Constant memory usage regardless of file size
- **Content-based ETags**: SHA-256 hash serves as ETag with immutable caching headers
- **Authentication**: Bearer token required for write operations (POST), reads are public
- **CORS support**: Configurable origins for browser applications
- **Structured logging**: Request method, path, status code, and duration
- **Panic recovery**: Middleware catches panics and returns 500 errors

### Middleware Chain

The server applies middleware in the following order:

1. **Recovery**: Catches panics and returns 500 errors
2. **Logging**: Logs all requests with timing information
3. **CORS**: Handles CORS preflight and headers
4. **Auth**: Validates Bearer token for write operations (POST)

### Security Considerations

- Authentication optional but recommended for production
- TLS support for encrypted connections (use `--tls-cert` and `--tls-key` flags)
- Write operations (POST) require Bearer token if configured
- CORS can be restricted to specific origins
- Catalog writes validate blob existence
- Path traversal prevented in catalog entries
- Blobs stored with read-only permissions (0444)
- Hash inputs validated (64 hex characters)

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
│   ├── merkle/         # Merkle tree implementation
│   │   ├── errors.go   # Error definitions
│   │   ├── node.go     # Node type and hash functions
│   │   ├── tree.go     # Tree building
│   │   └── proof.go    # Proof generation and verification
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

- **SHA-256 hashing**: Cryptographically secure with excellent collision resistance
- **Go interface patterns**: Accepts `io.Reader`/`io.Writer` for composition and testability
- **Streaming I/O**: Uses `io.Copy` and `io.MultiWriter` for memory efficiency
- **Write-time deduplication**: Checks existing blobs before writing
- **Atomic writes**: Temp file then rename to final location
- **Immutable blobs**: 0444 permissions prevent modification
- **SQLite catalog**: WAL mode for concurrent reads with ACID guarantees
- **RESTful API design**: Resource-oriented following REST principles
- **12-factor configuration**: Command-line flags and environment variables
- **Middleware composition**: Separation of cross-cutting concerns
- **Multi-stage Docker builds**: Minimal Alpine-based images, CGO-free

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

The client uses `LocalClient` which provides zero-overhead direct access with thread-safe operations.

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

The client uses `HTTPClient` which provides connection pooling, automatic authentication, and streaming operations.

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
| `pkg/merkle` | `merkle_test.go` | Tree building, proof generation, verification |
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

### Makefile Commands

Mini-CAS includes a Makefile for streamlined development workflows. All commands are available at the project root:

```bash
# Building
make build          # Build binary to ./bin/cas
make clean          # Remove ./bin, coverage.out, coverage.html

# Running
make run            # Run with go run (no build)
make serve          # Build and start HTTP server

# Testing
make test           # Run all tests
make test-v         # Verbose test output
make test-race      # Tests with race detector
make test-cover     # Generate HTML coverage report (coverage.html)

# Code Quality
make fmt            # Format code with gofmt
make vet            # Static analysis with go vet
make check          # Run both fmt and vet

# Help
make help           # Show all available commands
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

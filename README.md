# Mini Content-Addressable Storage

A content-addressable storage (CAS) system inspired by CVMFS. Mini-CAS stores files by their SHA-256 hash rather than their original path, enabling automatic deduplication and efficient storage management.

## Overview

Content-addressable storage eliminates duplicate files automatically by storing each unique file only once, regardless of how many times it appears in your filesystem. Mini-CAS computes a SHA-256 hash of each file's content and uses that hash as the storage key, making identical files share the same storage location.

## Features

- **Automatic deduplication**: Identical files are stored only once
- **Content-based addressing**: Files are stored by their SHA-256 hash
- **Memory-efficient streaming**: Large files are processed without loading entirely into memory
- **Efficient sharding**: 2-level directory sharding prevents filesystem performance degradation
- **Immutable storage**: Blobs are write-once, read-many with enforced permissions
- **Human-readable catalog**: JSON-based catalog for easy inspection and debugging
- **Directory support**: Add entire directory trees with a single command

## Installation

```bash
# Clone the repository
git clone https://github.com/SteliosSpanos/mini-CAS.git
cd mini-CAS

# Build the binary
go build -o cas ./cmd/main
```

## Quick Start

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
```

## Commands

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

## Architecture

Mini-CAS is built with a layered architecture:

```
+------------------+
|  Command Layer   |  CLI interface (init, add, ls, cat, status, hash)
+------------------+
|  Repository      |  Manages .cas/ directory structure
+------------------+
|  Catalog         |  Maps file paths to content hashes
+------------------+
|  Storage         |  Physical blob storage with sharding
+------------------+
|  Objects         |  Blob types and SHA-256 hashing
+------------------+
```

- **Objects Layer**: Defines the `Blob` type and SHA-256 hashing
- **Storage Layer**: Manages physical blob storage with 2-level sharding and streaming I/O
- **Catalog Layer**: Maps original file paths to content hashes with JSON persistence
- **Repository Layer**: Manages the `.cas/` directory structure
- **Command Layer**: User-facing CLI commands

## Storage Structure

Files are stored using a 2-level sharding strategy based on the SHA-256 hash:

```
.cas/
├── storage/
│   └── ab/
│       └── cd/
│           └── abcd1234567890...  (full 64-char SHA-256 hash)
└── catalog.json
```

This sharding approach (using the first 4 characters of the hash split into two directory levels) scales to millions of files without filesystem performance degradation.

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
│   └── hash.go         # Standalone hash utility
├── pkg/
│   ├── objects/        # Blob types and hashing
│   ├── storage/        # Physical storage management
│   ├── catalog/        # Path-to-hash mapping
│   └── path/           # Repository initialization
└── README.md           # This file
```

## Design Decisions

- **SHA-256 hashing**: Cryptographically secure hash function with excellent collision resistance
- **Streaming I/O**: Files are processed using `io.Copy` and `io.MultiWriter` to handle large files without excessive memory usage
- **Write-time deduplication**: Checks for existing blobs before writing to minimize I/O
- **Atomic writes**: New blobs are written to a temp file first, then renamed to final location
- **Immutable blobs**: Files stored with 0444 permissions prevent accidental modification
- **JSON catalog**: Human-readable format simplifies debugging and inspection

## Development

```bash
# Build
go build -o cas ./cmd/main

# Run tests
go test ./pkg/...

# Run tests for a specific package
go test ./pkg/storage
```

## License

This project is intended for educational purposes.

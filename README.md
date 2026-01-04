# Mini-CAS

A content-addressable storage (CAS) system inspired by CVMFS. Mini-CAS stores files by their SHA-256 hash rather than their original path, enabling automatic deduplication and efficient storage management.

## Overview

Content-addressable storage eliminates duplicate files automatically by storing each unique file only once, regardless of how many times it appears in your filesystem. Mini-CAS computes a SHA-256 hash of each file's content and uses that hash as the storage key, making identical files share the same storage location.

## Features

- **Automatic deduplication**: Identical files are stored only once
- **Content-based addressing**: Files are stored by their SHA-256 hash
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

## Usage

### Initialize a repository

```bash
./cas init
```

This creates a `.cas/` directory in your current working directory, similar to how Git creates a `.git/` directory.

### Add files

```bash
# Add a single file
./cas add myfile.txt

# Add a directory
./cas add /path/to/directory
```

When adding directories, Mini-CAS automatically skips the `.cas/` directory to avoid recursion.

### List tracked files

```bash
./cas ls
```

This displays all files tracked in the catalog with their hash, size, and modification time.

## Architecture

Mini-CAS is built with a layered architecture:

- **Objects Layer**: Defines blob types and SHA-256 hashing
- **Storage Layer**: Manages physical blob storage with 2-level sharding
- **Catalog Layer**: Maps original file paths to content hashes
- **Repository Layer**: Manages the `.cas/` directory structure
- **Command Layer**: User-facing CLI commands

## Storage Structure

Files are stored using a 2-level sharding strategy:

```
.cas/
├── storage/
│   └── ab/
│       └── cd/
│           └── abcd1234567890...  (full 64-char SHA-256 hash)
└── catalog.json
```

This sharding approach scales to millions of files without filesystem performance issues.

## Project structure

```
mini-CAS/
├── cmd/
│   └── main/           # Main entry point
├── commands/           # CLI command implementations
├── pkg/
│   ├── objects/        # Blob types and hashing
│   ├── storage/        # Physical storage management
│   ├── catalog/        # Path-to-hash mapping
│   └── path/           # Repository initialization
└── CLAUDE.md           # Development guidance
```

## Design Decisions

- **SHA-256 hashing**: Cryptographically secure hash function with excellent collision resistance for content addressing
- **Write-time deduplication**: Checks for existing blobs before writing to minimize I/O
- **Immutable blobs**: Files stored with 0444 permissions prevent accidental modification
- **JSON catalog**: Human-readable format simplifies debugging and inspection

## License

This project is intended for educational purposes.
# buf3pd

Protocol Buffer dependency manager for buf projects.

## Overview

`buf3pd` manages Protocol Buffer dependencies for your buf projects. It downloads proto files from Git repositories and maintains them for use with the buf system.

## Quick Start

1. Create a `buf3pd.yaml` file in your project root:

```yaml
path: gen/buf3pd
deps:
    - type: git
      repo: github.com/googleapis/googleapis
      path: .
      ref: heads/master
      filter:
          - google/{api,longrunning,rpc}/**
```

2. Run the dependency manager:

```bash
# Using the provided script
./scripts/run-buf3pd.sh

# Or directly
go run ./cmd/buf3pd --workdir .
```

3. Your dependencies will be downloaded to the specified path and registered in your buf.yaml

## Features

-   Download proto files from Git repositories
-   Manage dependencies with a lock file for reproducible builds
-   Filter only needed proto files from large repositories
-   Automatically update your buf.yaml modules section
-   Track dependencies with cache and checksums

## Configuration

`buf3pd` can be configured in two ways:

1. **Standalone file (recommended)**: Create a `buf3pd.yaml` file in your project root
2. **Integrated**: Add a `buf3pd` section to your `buf.yaml` file

See the `examples/` directory for configuration examples.

## Scripts

-   `scripts/run-buf3pd.sh`: Runs buf3pd with the standalone configuration
-   `scripts/test-coverage.sh`: Runs tests with coverage reporting

## Development

```bash
# Build
go build -o bin/buf3pd ./cmd/buf3pd

# Test
go test ./...

# Test with coverage
./scripts/test-coverage.sh
```

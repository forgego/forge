---
sidebar_position: 1
description: Install forge and set up your environment.
keywords:
  - install forge
  - forge installation
  - setup forge
  - forge cli
  - forge setup
image: /forge-social-card.svg
---

# Installation

Get forge running locally in a few minutes.

## Prerequisites

- **Go 1.25 or later**
- **PostgreSQL 12 or later**

## Recommended install

```bash
go install github.com/forgego/forge/cli/cmd@latest
forge --version
```

Make sure your Go bin directory is on PATH:

```bash
export PATH="$PATH:$HOME/go/bin"
```

## Create a database

```bash
psql -U postgres -c "CREATE DATABASE myapp_db;"
```

## Alternative installs

### Build from source

```bash
git clone https://github.com/forgego/forge.git
cd forge
go build -o forge ./cli/cmd
```

### Download binary

Pre-built binaries are coming soon.

## Troubleshooting

- **Go version**: run `go version` and ensure it is 1.25+.
- **PostgreSQL**: verify the service is running and the port is 5432.
- **Permissions**: ensure your user can connect to the database.

## Next steps

1. [Quick Start](/docs/getting-started/quickstart/)
2. [Hello World](/docs/getting-started/hello-world/)
3. [Models guide](/docs/guides/models/)

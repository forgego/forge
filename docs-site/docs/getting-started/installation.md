---
sidebar_position: 1
description: Install forge framework on your machine. Step-by-step guide for setting up forge CLI, PostgreSQL, and your development environment.
keywords:
  - install forge
  - forge installation
  - setup forge
  - forge cli
  - forge setup
image: /img/forge-social-card.jpg
---

# Installation

Get forge up and running on your machine in a few minutes.

## Prerequisites

Before installing forge, make sure you have:

- **Go 1.25 or later** - [Download Go](https://go.dev/dl/)
- **PostgreSQL 12 or later** - [Download PostgreSQL](https://www.postgresql.org/download/)
- **Basic knowledge of Go** - Familiarity with Go syntax and concepts

## Install forge CLI

The forge CLI tool is your main interface for creating projects, generating code, running migrations, and starting your development server.

### Option 1: Build from Source (Recommended for development)

Clone the repository and build:

```bash
git clone https://github.com/forgego/forge.git
cd forge
go build -o forge ./cli/cmd
```

Add `forge` to your PATH:
```bash
# Add to your shell profile (~/.bashrc, ~/.zshrc, etc.)
export PATH="$PATH:/path/to/forge"
```

### Option 2: Install via go install

```bash
go install github.com/forgego/forge/cli/cmd@latest
```

Make sure `$GOPATH/bin` or `$HOME/go/bin` is in your PATH.

### Option 3: Download Binary (Coming Soon)

We'll have pre-built binaries available for download soon.

### Verify Installation

Check that forge is installed correctly:

```bash
forge --version
forge help
```

You should see the version number and available commands.

## Set Up PostgreSQL

forge uses PostgreSQL as its primary database. Make sure PostgreSQL is installed and running.

### Create a Database

Create a database for your project:

```bash
# Connect to PostgreSQL
psql -U postgres

# Create database
CREATE DATABASE myapp_db;

# Exit
\q
```

### Test Connection

Verify you can connect to PostgreSQL:

```bash
psql -U postgres -d myapp_db -c "SELECT version();"
```

## Project Structure

When you create a new forge project, you'll get this structure:

```
myapp/
├── main.go              # Application entry point
├── go.mod               # Go module file
├── config/
│   └── config.yaml      # Configuration file
├── models/              # Your model definitions
│   └── example.go       # Example model
└── migrations/          # Database migrations
```

## Next Steps

Now that you have forge installed:

1. [Create your first project](/docs/getting-started/quickstart)
2. [Learn about models](/docs/guides/models)
3. [Explore the examples](../examples/blog)

## Troubleshooting

### Go Version Issues

If you get errors about Go version, make sure you're using Go 1.25 or later:

```bash
go version
```

### PostgreSQL Connection Issues

If you can't connect to PostgreSQL:

1. Make sure PostgreSQL is running:
   ```bash
   # Linux/macOS
   sudo systemctl status postgresql
   
   # macOS (Homebrew)
   brew services list
   ```

2. Check your connection settings in `config/config.yaml`

3. Verify PostgreSQL is listening on the correct port (default: 5432)

### Permission Issues

If you get permission errors:

- Make sure your PostgreSQL user has the necessary permissions
- Check that the database exists and is accessible
- Verify file permissions for your project directory

## Development Tools

Recommended tools for forge development:

- **IDE**: VS Code with Go extension, GoLand, or Vim/Neovim with gopls
- **Database Client**: pgAdmin, DBeaver, or psql
- **API Testing**: Postman, Insomnia, or curl

## What's Next?

Ready to create your first application? Head over to the [Quick Start Guide](/docs/getting-started/quickstart)!


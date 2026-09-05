---
sidebar_position: 3
description: Complete installation guide for Forge framework.
image: /forge-social-card.svg
---

# Installation

This guide covers everything you need to install and set up Forge for development.

## System Requirements

### Go Version

Forge requires Go 1.21 or higher. Check your version:

```bash
go version
```

If you need to install or upgrade Go, download it from [go.dev/dl](https://go.dev/dl/).

### Database

Forge currently supports PostgreSQL 12 or higher. Install PostgreSQL:

- **macOS**: `brew install postgresql`
- **Ubuntu/Debian**: `sudo apt-get install postgresql postgresql-contrib`
- **Windows**: Download from [postgresql.org/download/windows](https://www.postgresql.org/download/windows/)

Verify PostgreSQL is running:

```bash
psql --version
```

### Operating System

Forge works on:
- Linux (any distribution)
- macOS (10.15+)
- Windows (with WSL recommended)

## Install Forge CLI

The Forge CLI provides commands for project creation, code generation, migrations, and more.

### Global Installation

Install the `forge` command globally:

```bash
go install github.com/forgego/forge/cli/cmd@latest
```

This installs the binary to `$GOPATH/bin` or `$GOBIN`.

### Verify Installation

Check that `forge` is available:

```bash
forge --version
```

Expected output:

```
forge version 1.0.0
```

### Add to PATH

If the `forge` command isn't found, add Go's bin directory to your PATH.

**Bash/Zsh** (add to `~/.bashrc` or `~/.zshrc`):

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

**Fish** (add to `~/.config/fish/config.fish`):

```fish
set -gx PATH $PATH (go env GOPATH)/bin
```

Reload your shell:

```bash
source ~/.bashrc  # or ~/.zshrc
```

## Install as a Library

If you want to use Forge in an existing Go project without the CLI:

```bash
go get github.com/forgego/forge@latest
```

Then import packages as needed:

```go
import (
    "github.com/forgego/forge/schema"
    "github.com/forgego/forge/orm"
    "github.com/forgego/forge/server"
)
```

## Database Setup

### Create Database

Create a PostgreSQL database for your project:

```bash
createdb myproject_db
```

Or using SQL:

```sql
CREATE DATABASE myproject_db;
```

### Database User

It's recommended to create a dedicated user for your application:

```bash
psql -c "CREATE USER myapp WITH PASSWORD 'secure_password';"
psql -c "GRANT ALL PRIVILEGES ON DATABASE myproject_db TO myapp;"
```

### Connection String

Your connection details will go in `config/config.yaml`:

```yaml
database:
  host: localhost
  port: 5432
  name: myproject_db
  user: myapp
  password: secure_password
  sslmode: disable
```

## Development Tools (Optional)

### PostgreSQL GUI Tools

- **pgAdmin** - [pgadmin.org](https://www.pgadmin.org/)
- **DBeaver** - [dbeaver.io](https://dbeaver.io/)
- **Postico** (macOS) - [eggerapps.at/postico](https://eggerapps.at/postico/)

### Code Editor Extensions

#### VS Code

Install these extensions for the best experience:

- **Go** by Go Team at Google
- **Go Test Explorer** for test discovery
- **PostgreSQL** for database management
- **YAML** for config file syntax

#### GoLand

GoLand has built-in Go support. Enable these:

- Go modules integration
- Database tools and SQL
- Code completion and refactoring

### Git

Forge projects use Git for version control:

```bash
git --version
```

Install Git if needed: [git-scm.com/downloads](https://git-scm.com/downloads)

## Create Your First Project

Once everything is installed, create a project:

```bash
forge new myproject
cd myproject
```

Configure your database in `config/config.yaml`, then:

```bash
forge generate
forge makemigrations init --auto
forge migrate
forge runserver
```

Visit http://localhost:8000/admin/ to see the admin interface.

## Upgrading Forge

### Upgrade CLI

To upgrade to the latest version:

```bash
go install github.com/forgego/forge/cli/cmd@latest
```

### Upgrade Library

In your project:

```bash
go get -u github.com/forgego/forge@latest
go mod tidy
```

### Check Version

```bash
forge --version
```

## Troubleshooting

### Command not found: forge

**Problem**: Shell can't find the `forge` command.

**Solution**: Make sure `$GOPATH/bin` is in your PATH:

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc
source ~/.bashrc
```

### Cannot connect to database

**Problem**: Database connection errors.

**Solutions**:

1. Check PostgreSQL is running:
   ```bash
   pg_isready
   ```

2. Verify connection details in `config/config.yaml`

3. Test connection manually:
   ```bash
   psql -h localhost -U myapp -d myproject_db
   ```

4. Check PostgreSQL logs:
   ```bash
   tail -f /usr/local/var/log/postgresql.log  # macOS
   sudo tail -f /var/log/postgresql/postgresql-*.log  # Linux
   ```

### go.mod not found

**Problem**: Go module errors.

**Solution**: Make sure you're in the project directory and `go.mod` exists:

```bash
go mod init myproject  # if starting from scratch
go mod tidy
```

### Permission denied

**Problem**: Can't create database or access files.

**Solutions**:

1. For database: Grant permissions to your user
2. For files: Check file ownership and permissions

### Port 8000 already in use

**Problem**: Development server can't start.

**Solution**: Change port in `config/config.yaml` or kill the process:

```bash
# Find process
lsof -ti:8000

# Kill process (macOS/Linux)
kill $(lsof -ti:8000)
```

### Module download issues

**Problem**: Can't download Forge or dependencies.

**Solutions**:

1. Enable Go modules:
   ```bash
   export GO111MODULE=on
   ```

2. Clear module cache:
   ```bash
   go clean -modcache
   ```

3. Set Go proxy:
   ```bash
   export GOPROXY=https://proxy.golang.org,direct
   ```

## Environment Setup

### Development Environment

Set these environment variables for development:

```bash
export FORGE_ENV=development
export FORGE_DEBUG=true
export DATABASE_URL=postgresql://user:pass@localhost/dbname
```

### Production Environment

For production, use:

```bash
export FORGE_ENV=production
export FORGE_DEBUG=false
export DATABASE_URL=postgresql://user:pass@prod-host/dbname
export SECRET_KEY=your-secret-key-here
```

## Docker Setup (Optional)

You can run PostgreSQL in Docker:

```bash
docker run --name postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=myproject_db \
  -p 5432:5432 \
  -d postgres:15
```

Then use these connection details:

```yaml
database:
  host: localhost
  port: 5432
  name: myproject_db
  user: postgres
  password: postgres
```

## Next Steps

Now that Forge is installed:

1. [Quick Start](/docs/quickstart) - Build your first app
2. [Models Guide](/docs/models) - Define your data
3. [Configuration](/docs/config/overview) - Learn about configuration options

## Getting Help

- [Documentation](/docs/) - Full documentation
- [GitHub Issues](https://github.com/hamidrabedi/foreit/issues) - Report issues
- [Examples](https://github.com/hamidrabedi/foreit/tree/main/examples) - Sample projects

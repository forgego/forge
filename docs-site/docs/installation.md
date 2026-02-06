---
sidebar_position: 3
description: Install forge and set up your environment.
image: /forge-social-card.svg
---

# Installation

## Prerequisites

- Go 1.25 or later
- PostgreSQL 12 or later

## Install the CLI

```bash
go install github.com/forgego/forge/cli/cmd@latest
```

Ensure your Go bin directory is on PATH:

```bash
export PATH="$PATH:$HOME/go/bin"
```

## Verify

```bash
forge --version
```

## Next steps

- [Quick Start](/docs/quickstart/)
- [Models](/docs/models/)

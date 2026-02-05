---
sidebar_position: 7
---

# Deployment Guide

This guide covers deploying forge applications to production environments.

## Pre-Deployment Checklist

Before deploying, ensure:

- [ ] All tests pass
- [ ] Migrations are tested
- [ ] Environment variables are configured
- [ ] Database is backed up
- [ ] HTTPS is configured
- [ ] Security settings are reviewed and hardened
- [ ] Logging is configured
- [ ] Monitoring is set up

## Configuration

### Environment Variables

Use environment variables for sensitive configuration:

```bash
export FORGE_DATABASE_HOST=db.example.com
export FORGE_DATABASE_NAME=myapp_prod
export FORGE_DATABASE_USER=myapp_user
export FORGE_DATABASE_PASSWORD=secure_password
export FORGE_SECRET_KEY=your-secret-key
export FORGE_DEBUG=false
```

### Production Config

Create `config/production.yaml`:

```yaml
app:
  name: "My Application"
  env: "production"
  debug: false

server:
  host: "0.0.0.0"
  port: "8080"
  read_timeout: 30s
  write_timeout: 30s

database:
  host: "${FORGE_DATABASE_HOST}"
  port: 5432
  user: "${FORGE_DATABASE_USER}"
  password: "${FORGE_DATABASE_PASSWORD}"
  dbname: "${FORGE_DATABASE_NAME}"
  sslmode: "require"
  max_connections: 25
  max_idle_connections: 5

security:
  secret_key: "${FORGE_SECRET_KEY}"
  csrf_secret_key: "${FORGE_CSRF_SECRET_KEY}"
  session_secret: "${FORGE_SESSION_SECRET}"
  csrf:
    enabled: true
    secure: true
  session:
    secure: true
    http_only: true
    same_site: "strict"

logging:
  level: "info"
  format: "json"
  output: "stdout"
```

## Building for Production

### Build Binary

```bash
# Build for Linux
GOOS=linux GOARCH=amd64 go build -o myapp ./cmd/server

# Or use Makefile
make build
```

### Optimize Build

```bash
# Build with optimizations
go build -ldflags="-s -w" -o myapp ./cmd/server

# Reduce binary size
strip myapp
```

## Database Setup

### Run Migrations

```bash
# Apply all migrations
forge migrate

# Or manually
./myapp migrate
```

### Verify Database

```bash
# Check migration status
forge migrate status

# Verify connection
psql -h db.example.com -U myapp_user -d myapp_prod -c "SELECT version();"
```

## Deployment Options

### Docker Deployment

Create `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o myapp ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/myapp .
COPY --from=builder /app/config ./config
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080
CMD ["./myapp"]
```

Build and run:

```bash
docker build -t myapp:latest .
docker run -d -p 8080:8080 \
  -e FORGE_DATABASE_HOST=db.example.com \
  -e FORGE_SECRET_KEY=your-secret-key \
  myapp:latest
```

### Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - FORGE_DATABASE_HOST=db
      - FORGE_DATABASE_NAME=myapp_prod
      - FORGE_SECRET_KEY=${FORGE_SECRET_KEY}
    depends_on:
      - db
    restart: unless-stopped

  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=myapp_prod
      - POSTGRES_USER=myapp_user
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

volumes:
  postgres_data:
```

Deploy:

```bash
docker-compose up -d
```

### Systemd Service

Create `/etc/systemd/system/myapp.service`:

```ini
[Unit]
Description=My Application
After=network.target

[Service]
Type=simple
User=myapp
WorkingDirectory=/opt/myapp
ExecStart=/opt/myapp/myapp
Restart=always
RestartSec=5
Environment="FORGE_DATABASE_HOST=db.example.com"
Environment="FORGE_SECRET_KEY=your-secret-key"

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable myapp
sudo systemctl start myapp
sudo systemctl status myapp
```

### Kubernetes Deployment

Create `k8s/deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: myapp
        image: myapp:latest
        ports:
        - containerPort: 8080
        env:
        - name: FORGE_DATABASE_HOST
          valueFrom:
            secretKeyRef:
              name: myapp-secrets
              key: database-host
        - name: FORGE_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: myapp-secrets
              key: secret-key
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## Reverse Proxy

### Nginx Configuration

```nginx
server {
    listen 80;
    server_name example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name example.com;

    ssl_certificate /etc/ssl/certs/example.com.crt;
    ssl_certificate_key /etc/ssl/private/example.com.key;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Caddy Configuration

```caddy
example.com {
    reverse_proxy localhost:8080
}
```

## Monitoring

### Health Checks

Add health check endpoint:

```go
router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    // Check database connection
    if err := db.Ping(); err != nil {
        http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
})
```

### Logging

Configure structured logging:

```yaml
logging:
  level: "info"
  format: "json"
  output: "stdout"
  fields:
    service: "myapp"
    environment: "production"
```

### Metrics

Add Prometheus metrics:

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
)
```

## Backup and Recovery

### Database Backups

```bash
# Automated backup script
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
pg_dump -h db.example.com -U myapp_user myapp_prod > backup_$DATE.sql

# Restore
psql -h db.example.com -U myapp_user myapp_prod < backup_20240101_120000.sql
```

### Application Backups

```bash
# Backup application files
tar -czf myapp_backup_$(date +%Y%m%d).tar.gz \
    /opt/myapp/myapp \
    /opt/myapp/config \
    /opt/myapp/migrations
```

## Rollback Procedure

### Application Rollback

```bash
# Stop current version
sudo systemctl stop myapp

# Restore previous binary
cp myapp.backup myapp

# Start service
sudo systemctl start myapp
```

### Database Rollback

```bash
# Rollback migrations
forge migrate down 1

# Or restore from backup
psql -h db.example.com -U myapp_user myapp_prod < backup.sql
```

## Performance Tuning

### Database Connection Pool

```yaml
database:
  max_connections: 25
  max_idle_connections: 5
  max_lifetime: 5m
```

### Application Settings

```yaml
server:
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
  max_header_bytes: 1048576
```

## Security in Production

- Use HTTPS only
- Set secure cookies
- Enable CSRF protection
- Use strong secret keys
- Limit database connections
- Enable rate limiting
- Monitor for attacks
- Regular security updates

## Troubleshooting

### Check Logs

```bash
# Application logs
sudo journalctl -u myapp -f

# Docker logs
docker logs -f myapp

# Nginx logs
tail -f /var/log/nginx/error.log
```

### Common Issues

1. **Database Connection Errors**
   - Check database is running
   - Verify credentials
   - Check network connectivity

2. **Migration Failures**
   - Check migration status
   - Verify database permissions
   - Review migration SQL

3. **Performance Issues**
   - Check database connection pool
   - Review query performance
   - Monitor resource usage

## Next Steps

- [Security Guide](/docs/guides/security) - Production security
- [Monitoring](/docs/advanced/performance) - Performance monitoring
- [Development Guide](/docs/contributing/development) - Contributing

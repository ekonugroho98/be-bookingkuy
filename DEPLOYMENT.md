# Bookingkuy - Production Deployment Guide

Last Updated: 2025-12-26
Phase: 5 - Production Readiness

---

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Local Development Setup](#local-development-setup)
3. [Docker Deployment](#docker-deployment)
4. [Database Setup](#database-setup)
5. [Environment Configuration](#environment-configuration)
6. [Running the Application](#running-the-application)
7. [Health Checks](#health-checks)
8. [Monitoring](#monitoring)
9. [Production Considerations](#production-considerations)
10. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Software:
- **Go:** 1.23 or later
- **PostgreSQL:** 15+
- **Redis:** 7+
- **RabbitMQ:** 3+ (for message queue)
- **Docker & Docker Compose** (optional but recommended)

### External Services:
- **Midtrans Account** - for payment processing
- **SendGrid Account** - for email notifications
- **Hotelbeds API Access** - for hotel inventory

---

## Local Development Setup

### 1. Clone Repository
```bash
git clone <repository-url>
cd be-bookingkuy
```

### 2. Install Dependencies
```bash
go mod download
```

### 3. Configure Environment
```bash
cp .env.example .env
# Edit .env with your configuration
```

### 4. Start Dependencies (without Docker)
```bash
# PostgreSQL
brew services start postgresql  # macOS
# or
sudo systemctl start postgresql  # Linux

# Redis
brew services start redis  # macOS
# or
sudo systemctl start redis  # Linux

# RabbitMQ
brew services start rabbitmq  # macOS
# or
sudo systemctl start rabbitmq  # Linux
```

### 5. Run Database Migrations
```bash
# Using golang-migrate (recommended)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate -path migrations -database "postgres://bookingkuy:bookingkuy_dev_password@localhost:5432/bookingkuy_db?sslmode=disable" up

# Or manually with psql
psql -U bookingkuy -d bookingkuy_db -f migrations/000001_users_schema.up.sql
psql -U bookingkuy -d bookingkuy_db -f migrations/000002_hotels_schema.up.sql
psql -U bookingkuy -d bookingkuy_db -f migrations/000003_payments_schema.up.sql
psql -U bookingkuy -d bookingkuy_db -f migrations/000004_notifications_schema.up.sql
```

### 6. Run Application
```bash
# Development mode
go run ./cmd/api

# Or build and run
go build -o bin/api ./cmd/api
./bin/api
```

### 7. Verify Application
```bash
# Health check
curl http://localhost:8080/health

# Ready check
curl http://localhost:8080/health/ready

# Live check
curl http://localhost:8080/health/live
```

---

## Docker Deployment

### Quick Start with Docker Compose
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f api

# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

### Docker Services:
1. **postgres** - PostgreSQL database (port 5432)
2. **redis** - Redis cache (port 6379)
3. **rabbitmq** - Message queue (ports 5672, 15672)
4. **api** - Bookingkuy API server (port 8080)

### Access Points:
- **API:** http://localhost:8080
- **RabbitMQ Management:** http://localhost:15672 (guest/guest)
- **Metrics:** http://localhost:8080/metrics

---

## Database Setup

### Create Database
```sql
CREATE DATABASE bookingkuy_db;
CREATE USER bookingkuy WITH PASSWORD 'secure_password_here';
GRANT ALL PRIVILEGES ON DATABASE bookingkuy_db TO bookingkuy;
```

### Run Migrations
See step 5 in Local Development Setup.

### Database Connection String
```
host=localhost port=5432 user=bookingkuy password=secure_password_here dbname=bookingkuy_db sslmode=require
```

---

## Environment Configuration

### Required Environment Variables

#### Database
```bash
BOOKINGKUY_DATABASE_HOST=localhost
BOOKINGKUY_DATABASE_PORT=5432
BOOKINGKUY_DATABASE_NAME=bookingkuy_db
BOOKINGKUY_DATABASE_USER=bookingkuy
BOOKINGKUY_DATABASE_PASSWORD=***
BOOKINGKUY_DATABASE_SSLMODE=disable  # Use "require" in production
```

#### Redis
```bash
BOOKINGKUY_REDIS_HOST=localhost
BOOKINGKUY_REDIS_PORT=6379
BOOKINGKUY_REDIS_PASSWORD=  # Optional
```

#### JWT
```bash
BOOKINGKUY_JWT_SECRET=***generate-secure-random-secret***
BOOKINGKUY_JWT_EXPIRATION=24h
```

#### Server
```bash
BOOKINGKUY_SERVER_HOST=0.0.0.0
BOOKINGKUY_SERVER_PORT=8080
BOOKINGKUY_ENVIRONMENT=production
```

#### RabbitMQ
```bash
BOOKINGKUY_RABBITMQ_HOST=localhost
BOOKINGKUY_RABBITMQ_PORT=5672
BOOKINGKUY_RABBITMQ_USER=guest
BOOKINGKUY_RABBITMQ_PASSWORD=***
BOOKINGKUY_RABBITMQ_VHOST=/
```

#### Midtrans Payment
```bash
BOOKINGKUY_MIDTRANS_MERCHANTID=***
BOOKINGKUY_MIDTRANS_CLIENTKEY=***
BOOKINGKUY_MIDTRANS_SERVERKEY=***
BOOKINGKUY_MIDTRANS_ISPRODUCTION=true
```

#### SendGrid Email
```bash
BOOKINGKUY_SENDGRID_APIKEY=***
BOOKINGKUY_SENDGRID_FROMEMAIL=noreply@bookingkuy.com
```

#### Hotelbeds
```bash
BOOKINGKUY_HOTELBEDS_APIKEY=***
BOOKINGKUY_HOTELBEDS_SECRET=***
BOOKINGKUY_HOTELBEDS_BASEURL=https://api.hotelbeds.com
```

---

## Running the Application

### Development Mode
```bash
go run ./cmd/api
```

### Production Mode
```bash
# Build binary
go build -ldflags="-s -w" -o bin/api ./cmd/api

# Run with production config
export BOOKINGKUY_ENVIRONMENT=production
./bin/api
```

### Using Systemd (Linux)
Create `/etc/systemd/system/bookingkuy.service`:
```ini
[Unit]
Description=Bookingkuy API Server
After=network.target postgresql.service redis.service rabbitmq-server.service

[Service]
Type=simple
User=bookingkuy
WorkingDirectory=/opt/bookingkuy
ExecStart=/opt/bookingkuy/bin/api
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable bookingkuy
sudo systemctl start bookingkuy
sudo systemctl status bookingkuy
```

---

## Health Checks

### Endpoints
- `GET /health` - Overall health check
- `GET /health/ready` - Readiness check (dependencies)
- `GET /health/live` - Liveness check

### Example Response
```json
{
  "status": "healthy",
  "timestamp": "2025-12-26T05:00:00Z",
  "dependencies": {
    "database": "healthy",
    "redis": "healthy",
    "rabbitmq": "healthy"
  }
}
```

---

## Monitoring

### Metrics Endpoint
- `GET /metrics` - Application metrics

### Metrics Collected
- Total requests
- Success/failed requests
- Average response time
- Active bookings
- Payment statistics
- Provider call counts

### Example Metrics Response
```json
{
  "total_requests": 15420,
  "success_requests": 15231,
  "failed_requests": 189,
  "active_requests": 3,
  "average_response_time_ms": 45,
  "total_bookings": 523,
  "successful_bookings": 498,
  "failed_bookings": 25,
  "total_payments": 510,
  "successful_payments": 489,
  "failed_payments": 21
}
```

---

## Production Considerations

### Security
1. **HTTPS Only** - Use reverse proxy (nginx) with TLS
2. **JWT Secret** - Use strong random key (32+ chars)
3. **Database Password** - Use strong password, change defaults
4. **Environment Variables** - Never commit `.env` file
5. **Rate Limiting** - Enable and configure appropriately
6. **Input Validation** - Already implemented, keep it updated

### Performance
1. **Connection Pooling** - Configured in `db.go`
2. **Caching** - Use Redis for frequently accessed data
3. **Database Indexes** - Created in migrations
4. **Graceful Shutdown** - Implemented in `server.go`
5. **Worker Processes** - Use multiple instances with load balancer

### Reliability
1. **Health Checks** - For load balancer / orchestrator
2. **Graceful Shutdown** - Handles SIGTERM/SIGINT
3. **Database Migrations** - Version-controlled schema changes
4. **Logging** - Structured logs for debugging
5. **Error Handling** - Proper error responses

### Backup Strategy
1. **Database Backups** - Daily automated backups
   ```bash
   pg_dump -U bookingkuy bookingkuy_db > backup_$(date +%Y%m%d).sql
   ```
2. **Redis Persistence** - Enable RDB/AOF
3. **RabbitMQ** - Queue definitions are not persistent (recreate on restart)

---

## Troubleshooting

### Build Errors
```bash
# Clean build cache
go clean -cache
go mod tidy
go build -a ./cmd/api
```

### Database Connection Issues
```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Check connection
psql -U bookingkuy -d bookingkuy_db -c "SELECT 1"
```

### Port Already in Use
```bash
# Find process
lsof -i :8080

# Kill process
kill -9 <PID>
```

### Redis Connection Issues
```bash
# Check Redis
redis-cli ping

# Check Redis logs
docker-compose logs redis
```

### RabbitMQ Connection Issues
```bash
# Check RabbitMQ
rabbitmq-diagnostics ping

# Check management UI
open http://localhost:15672
```

### Migration Issues
```bash
# Check migration version
psql -U bookingkuy -d bookingkuy_db -c "SELECT version FROM schema_migrations"

# Rollback and re-run
migrate -path migrations -database "postgres://..." down
migrate -path migrations -database "postgres://..." up
```

### Debug Mode
```bash
# Enable debug logging
export BOOKINGKUY_ENVIRONMENT=development

# Run with verbose output
go run ./cmd/api
```

---

## Scaling the Application

### Horizontal Scaling
```bash
# Run multiple instances
INSTANCE_ID=1 ./bin/api &
INSTANCE_ID=2 ./bin/api &
INSTANCE_ID=3 ./bin/api &

# Use nginx as load balancer
upstream bookingkuy {
    server localhost:8081;
    server localhost:8082;
    server localhost:8083;
}
```

### Database Scaling
- Use read replicas for read-heavy operations
- Implement connection pooling (already done)
- Consider database partitioning for large datasets

---

## Logging

### Log Location
- **Development:** Console (stdout)
- **Production:** `/var/log/bookingkuy/app.log`

### Log Rotation (Production)
Create `/etc/logrotate.d/bookingkuy`:
```
/var/log/bookingkuy/*.log {
    daily
    rotate 30
    compress
    delaycompress
    notifempty
    create 0640 bookingkuy bookingkuy
    sharedscripts
    postrotate
        systemctl reload bookingkuy > /dev/null 2>&1 || true
    endscript
}
```

---

## Support

For issues or questions:
1. Check logs: `docker-compose logs -f api`
2. Check health: `curl http://localhost:8080/health`
3. Review this guide
4. Check PHASE_REVIEW.md for architecture details

---

**Deployment Readiness: ✅ VERIFIED**
**Date:** 2025-12-26
**Phase:** 5 - Production Readiness (COMPLETE)

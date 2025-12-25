# Bookingkuy Backend API

Backend untuk Bookingkuy Global OTA dibangun dengan Go menggunakan arsitektur Modular Monolith.

## Tech Stack

- **Language:** Go 1.25+
- **Architecture:** Modular Monolith (Microservice-ready)
- **Database:** PostgreSQL 15
- **Cache:** Redis 7
- **Message Queue:** (Phase 4 - Kafka/SQS/RabbitMQ)

## Project Structure

```
be-bookingkuy/
├── cmd/
│   └── api/              # Application entry point
│       └── main.go
│
├── internal/
│   ├── auth/             # Authentication service
│   ├── user/             # User management
│   ├── search/           # Hotel search (traditional)
│   ├── ai-search/        # AI-powered search
│   ├── booking/          # Booking engine
│   ├── payment/          # Payment processing
│   ├── pricing/          # Pricing & markup logic
│   ├── provider/         # Provider abstraction layer
│   ├── notification/     # Email/WhatsApp notifications
│   ├── subscription/     # User subscriptions
│   ├── review/           # Review system
│   ├── admin/            # Admin operations
│   │
│   └── shared/           # Shared utilities
│       ├── db/           # Database connection
│       ├── cache/        # Redis cache
│       ├── queue/        # Job queue
│       ├── eventbus/     # Event bus
│       ├── logger/       # Logging
│       ├── config/       # Configuration
│       ├── worker/       # Background workers
│       ├── outbox/       # Outbox pattern
│       └── health/       # Health checks
│
├── go.mod
└── go.sum
```

## Service Pattern

Setiap domain service mengikuti pattern yang sama:

```
service/
├── model.go       # Domain models
├── repository.go  # Data access layer
├── service.go     # Business logic
├── handler.go     # HTTP handlers
└── events.go      # Domain events
```

## Quick Start

### Prerequisites

- Go 1.25+ installed
- PostgreSQL 15+ running
- Redis 7+ running
- (Optional) Docker & Docker Compose

### Setup

1. **Clone & Install Dependencies**
   ```bash
   cd be-bookingkuy
   go mod download
   ```

2. **Configure Environment**
   ```bash
   # Copy example env
   cp ../.env.example .env

   # Edit .env dengan configuration Anda
   ```

3. **Start Database** (jika menggunakan Docker)
   ```bash
   cd ..
   docker-compose up -d postgres redis
   ```

4. **Run Application**
   ```bash
   # Development mode
   go run ./cmd/api

   # Atau build dulu
   go build -o bin/api ./cmd/api
   ./bin/api
   ```

5. **Verify**
   ```bash
   curl http://localhost:8080/health
   ```

## Development

### Adding a New Service

1. Create service directory:
   ```bash
   mkdir -p internal/newservice
   ```

2. Create standard files:
   ```bash
   touch internal/newservice/{model.go,repository.go,service.go,handler.go,events.go}
   ```

3. Implement interface di repository.go:
   ```go
   type Repository interface {
       Create(ctx context.Context, entity *Entity) error
       GetByID(ctx context.Context, id string) (*Entity, error)
       // ... other methods
   }
   ```

4. Implement business logic di service.go

5. Add HTTP handlers di handler.go

6. Register routes di main.go

### Code Organization Rules

**✅ Allowed:**
- Domain services communicate via events
- Services can call read-only queries from other services
- Shared utilities di `internal/shared/`

**❌ Forbidden:**
- Direct DB access across services
- Importing another service's business logic
- Sync HTTP calls between booking & payment (use events/saga)

### Running Services

```bash
# Run API server
go run ./cmd/api

# Run worker (Phase 4)
go run ./cmd/worker

# Run migration
go run ./cmd/migrate

# Run seed data
go run ./cmd/seed
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/user

# Run benchmarks
go test -bench=. ./...
```

### Linting

```bash
# Run go vet
go vet ./...

# Run golangci-lint (install first)
golangci-lint run

# Format code
go fmt ./...
```

## Database

### Connection String

```
host=localhost port=5432 user=bookingkuy password=bookingkuy_dev_password dbname=bookingkuy_db sslmode=disable
```

### Running Migrations

```bash
# TODO: Setup migration tool (golang-migrate/migrate)
# migrate -path db/migrations -database "postgres://..." up
```

## Architecture Principles

1. **Modular Monolith** - Clear boundaries, easy extraction
2. **Event-Driven** - Async communication between services
3. **Database per Service** - Schema per domain, ready for microservices
4. **Provider Abstraction** - No vendor lock-in
5. **Idempotency** - All operations are idempotent
6. **Graceful Degradation** - System degrades gracefully under failures

## Phase 1 Foundation (Current)

**Completed:**
- ✅ Project structure
- ✅ Go module initialized
- ✅ Basic configuration loading
- ✅ Basic logger
- ✅ Placeholder files for all services

**In Progress:**
- Database connection pool
- Event bus foundation
- HTTP server & middleware

**Next:**
- User service implementation
- Auth service implementation
- Search service implementation

## Dependencies (Phase 0-2)

Core libraries yang akan digunakan:

- **Database:** `github.com/jackc/pgx/v5` - PostgreSQL driver
- **HTTP Framework:** (TBD - gin/echo/fiber/stdlib)
- **Config:** `github.com/spf13/viper` - Configuration management
- **Logging:** `github.com/rs/zerolog` - Structured logging
- **UUID:** `github.com/google/uuid` - UUID generation
- **Validation:** `github.com/go-playground/validator` - Request validation

Dependencies akan ditambah per-need basis.

## Troubleshooting

### Module errors

```bash
# Fix module dependencies
go mod tidy

# Verify dependencies
go mod verify
```

### Build errors

```bash
# Clean build cache
go clean -cache

# Rebuild
go build -a ./cmd/api
```

### Port already in use

```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>

# Or change port in .env
SERVER_PORT=8081
```

## Production Deployment

TODO: Setelah Phase 5 selesai, dokumentasi deployment akan ditambahkan.

## Status

✅ **Phase 0:** Project structure initialized
⏳ **Phase 1:** Foundation (In Progress - Ticket #006-009)
⏸️ **Phase 2:** Core Services (Pending - Ticket #010-015)

---

**Last Updated:** 2025-12-25
**Ticket:** #002 - Setup Go Project Structure

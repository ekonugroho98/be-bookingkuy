# BookingKuy — Backend API

> Travel booking platform backend that sells third-party hotel inventory: real-time supplier search, local markup & payments, and consistent booking state.

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-4169E1?style=flat-square&logo=postgresql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat-square&logo=redis&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=flat-square&logo=docker&logoColor=white)

## Overview

BookingKuy needs to sell hotel inventory it doesn't own. That means pulling availability and pricing from external suppliers in real time, applying the business's own markup, taking payment locally, and confirming the booking back with the supplier — while staying fast enough for a search page and reliable enough that money and reservations never drift out of sync.

This service is built as a **modular monolith (microservice-ready)**: a single deployable binary split internally into self-contained domain modules, each following a strict `handler → service → repository → model` layering so business logic never touches HTTP or SQL directly.

## Architecture

The core design decision is the **provider abstraction layer**. External suppliers (Hotelbeds, HotelPlanner, …) sit behind a single `Provider` interface, registered in a `registry` that handles selection and **failover**. The rest of the system talks only to canonical, supplier-agnostic models — so onboarding a new supplier means writing one adapter, not touching search, pricing, or booking.

```
be-bookingkuy/
├── cmd/
│   ├── api/        # HTTP API entry point (+ Swagger docs)
│   └── sync/       # Background worker (supplier sync)
├── internal/
│   ├── auth/       user/        search/      ai-search/
│   ├── booking/    payment/     pricing/     provider/   # interface, registry, failover
│   ├── hotelbeds/  midtrans/    sendgrid/    notification/
│   ├── review/     subscription/ webhook/    queue/
│   └── shared/     testutil/
├── migrations/     # 22 SQL migrations
├── docker-compose.yml
└── Dockerfile
```

## Tech Stack

- **Language:** Go 1.25
- **Architecture:** Modular monolith, microservice-ready
- **Database:** PostgreSQL 15 (source of truth, 22 migrations)
- **Cache:** Redis 7 (hot search results, rate limiting)
- **Auth:** JWT
- **Payments:** Midtrans
- **Supplier:** Hotelbeds (via provider abstraction + failover)
- **Email:** SendGrid (transactional)
- **Async:** background worker (`cmd/sync`) + webhook handlers
- **Docs & Ops:** Swagger / OpenAPI, Docker, integration tests

## Key Features

- Provider interface + registry with **failover** between hotel suppliers
- Real-time hotel search (traditional + AI-powered)
- Booking engine with pricing/markup logic
- Payment processing with consistent payment ↔ booking state
- Asynchronous supplier sync and payment webhooks
- JWT authentication, subscription handling, reviews
- Swagger-documented API + integration test suite

## Getting Started

### Prerequisites
- Go 1.25+
- Docker & Docker Compose

### Run

```bash
# clone
git clone https://github.com/ekonugroho98/be-bookingkuy.git
cd be-bookingkuy

# configure
cp .env.example .env   # fill in DB, Redis, Midtrans, Hotelbeds, SendGrid keys

# start the full stack (API + PostgreSQL + Redis)
docker-compose up -d

# run migrations
# (see DEPLOYMENT.md for the migration command)

# API server
go run ./cmd/api

# background worker
go run ./cmd/sync
```

API documentation is available via Swagger once the server is running (see `API_DOCUMENTATION.md`).

## Testing

```bash
go test ./...                 # all tests
go test -cover ./...          # with coverage
go test -bench=. ./...        # benchmarks
```

## Documentation

Additional docs live in `docs/` — API contracts, frontend integration, notifications, and admin API.

---

*Built by [Eko Nugroho](https://github.com/ekonugroho98) — Backend Engineer specializing in Go & system architecture.*

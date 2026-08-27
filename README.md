# Log Monitor & Alerting Service

Real-time log ingestion and threshold-based alerting service built in Go.

## Features
- JWT auth (REST + WebSocket)
- Async worker-pool log ingestion (10 goroutines, buffered channel)
- Redis-based sliding-window threshold counting for alerts
- Real-time alert broadcast via WebSocket
- Retry logic with graceful degradation on DB failures
- Graceful shutdown (SIGINT/SIGTERM)

## Performance
- 2,860 req/s at p99 latency 53ms (benchmarked with `hey`, n=1000, c=50)
- Systematically identified PostgreSQL — not the Go service layer — as the throughput bottleneck (CPU-bound at 100%+ under load) via `docker stats` profiling across multiple worker-pool/connection-pool configurations
- Verified race-free with Go's race detector

## Tech Stack
Go, Gin, PostgreSQL, Redis, WebSocket (gorilla/websocket), JWT

## Architecture
```
Client → JwtAuth middleware → CreateLog handler
                                     ↓ (enqueue, respond 202 immediately)
                              Buffered channel
                                     ↓
                        Worker pool (10 goroutines)
                                     ↓
                    Insert to Postgres + check alert_rules
                                     ↓ (if threshold exceeded)
                          Redis INCR/EXPIRE → Hub.Broadcast
                                     ↓
                          WebSocket → Dashboard clients
```

## Running locally

**Prerequisites:** Go 1.2x+, Docker, [golang-migrate](https://github.com/golang-migrate/migrate)

1. Clone and set up environment variables:
```bash
   git clone <repo-url>
   cd log-monitors
   cp .env.example .env  # fill in DB/Redis/JWT config
```

2. Start PostgreSQL and Redis:
```bash
   docker compose up -d
```

3. Run migrations:
```bash
   migrate -path pkg/databases/migrations \
     -database "postgres://logmon:123456@localhost:5555/logmon_db?sslmode=disable" up
```

4. Run the server (with hot-reload):
```bash
   air
```
   Or without hot-reload:
```bash
   go run ./cmd
```

5. Server runs at `http://localhost:3000`. See [log-dashboard](../log-dashboard) for the frontend.

**Running tests:**
```bash
go test ./...
go test -race ./internal/hub/...
```

## Testing
- Unit tests for auth (password hashing, JWT generation/verification) and handler request validation
- Concurrency safety verified with `go test -race` on the WebSocket hub

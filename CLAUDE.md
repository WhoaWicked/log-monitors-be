# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project context

This is a Go learning project (log monitoring/alerting backend), built incrementally by a Go beginner with a strong Node.js/Express/Next.js/TypeScript background. See `log-monitor-session-handoff.md` for the full learning history, locked-in decisions, and remaining TODOs — read it before making architectural suggestions, since several alternatives (Fiber, Supabase/Upstash, full ORM, Kafka/RabbitMQ, RBAC, cloud deploy) have already been deliberately rejected.

When teaching or explaining code in this repo: explain the concept first, then have the user write the code themselves, correcting piece by piece — don't hand over a full solution immediately.

## Commands

Requires Postgres + Redis running via Docker first:
```
docker-compose up -d
```

Run the app (with live reload via air):
```
air
```

Run the app directly:
```
go run ./cmd
```

Build:
```
go build -o ./tmp/main.exe ./cmd
```

Run all tests:
```
go test ./...
```

Run a single test:
```
go test ./internal/hub -run TestHubConcurrentAccess
```

Race-condition testing (important for this project — see below):
```
go test -race ./...
```

Database migrations (uses the `migrate` CLI, migration files in `pkg/databases/migrations/`):
```
migrate -path pkg/databases/migrations -database "postgres://<user>:<password>@<host>:<port>/<dbname>?sslmode=disable" up
migrate -path pkg/databases/migrations -database "postgres://<user>:<password>@<host>:<port>/<dbname>?sslmode=disable" down
```
Connection values come from `.env` (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`); default docker-compose maps Postgres to host port `5555` and Redis to `6380`.

## Architecture

Domain-vertical layout: `cmd/main.go` wires everything together and starts the Gin server on `:3000`.

**Request/event flow:**
1. `internal/config` loads `.env` + environment into a typed `Config` (fails fast if required vars are missing).
2. `internal/db` and `internal/cache` open the Postgres (`sqlx` + `pgx`) and Redis (`go-redis`) connections.
3. `internal/hub` is a `map[*websocket.Conn]bool` + `sync.Mutex` broadcast hub for pushing alerts to connected WebSocket clients (`GET /ws`).
4. `internal/ingest` owns a buffered channel (`jobs chan *models.Log`) and a worker pool (`Start(n)` spawns `n` goroutines). Each worker: inserts the log row, looks up a matching enabled `alert_rules` row for that `service`+`level`, increments a Redis counter keyed `alert:count:<rule_id>` with a sliding window (`Expire` set only when the count hits 1), and broadcasts an alert via the hub once the count reaches the rule's threshold. This is the core async pipeline — `POST /logs` just calls `ingest.Enqueue` and returns `202 Accepted` immediately (see `internal/handler/logs.go`).
5. `internal/handler` holds the Gin `Server` struct (holds `db`, `redis`, `hub`, `cfg`, `ingest`) and all route handlers. Routes are registered in `internal/handler/route.go`.
6. `internal/auth` implements JWT issuing/verification (`golang-jwt/jwt`) and bcrypt password comparison by hand — no external auth service. `internal/handler/middleware.go`'s `JwtAuth()` guards protected routes and sets `user_id`/`email` in the Gin context.
7. `internal/models` are plain structs with `db:"..."` tags for `sqlx` — no ORM; all SQL is hand-written in the handler/ingest layer.

**Auth model:** JWT-only, no signup endpoint — the admin user is seeded directly via SQL migration (`000004_seed_admin_user`). Login is `POST /login`; all other routes except `/health`, `/login`, and `/ws` require `Authorization: Bearer <token>`.

**Concurrency note:** this project is explicitly used to demonstrate real goroutine/channel/worker-pool patterns and `go test -race` cleanliness for resume purposes — don't simplify the worker pool into synchronous code, and keep any new concurrent code race-clean.

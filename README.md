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
- Verified race-free with Go's race detector

## Tech Stack
Go, Gin, PostgreSQL, Redis, WebSocket (gorilla/websocket), JWT

## Architecture
[ใส่ diagram ที่เคยทำไว้ตอนอธิบาย flow — ถ้าจำได้ export ออกมาเป็นรูปได้]

## Running locally
[docker-compose up + migrate + air ตามที่ทำมาตลอด]
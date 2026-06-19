# FitFlow Backend

Backend API for the FitFlow meal, workout, and body progress tracker.

## Stack

- Go (Chi Router)
- PostgreSQL
- Redis

## Local Setup

1. Copy `.env.example` to `.env`.
2. Start infrastructure: `docker compose up -d`.
3. Start the API server: `go run ./cmd/api`.
4. Verify the server is running: `GET http://localhost:8080/health`.

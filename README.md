# FitFlow Backend

Backend API for the FitFlow habit and workout tracker. The frontend is intentionally kept in a separate folder/repository.

## Stack

- Go
- Chi router
- PostgreSQL
- Redis
- Structured JSON logging with `log/slog`
- Docker Compose for local infrastructure
- Elasticsearch and Kibana for later observability work

## Local Setup

1. Copy `.env.example` to `.env` if you want shell-based environment loading.
2. Start infrastructure with `docker compose up -d`.
3. Apply SQL files from `migrations/` using your preferred migration tool.
4. Run the API with `go run ./cmd/api`.
5. Check `GET http://localhost:8080/health`.

## Using Supabase Postgres

Supabase can be used as the managed PostgreSQL database for this API. The backend only needs `DATABASE_URL` to point at your Supabase database.

1. Create a Supabase project.
2. Open Project Settings, then Database.
3. Copy a PostgreSQL connection string.
4. Set `DATABASE_URL` in your environment.
5. Add `sslmode=require` to the connection URL.
6. Apply the SQL files from `migrations/` in order.
7. Run the API with `go run ./cmd/api`.

Direct database connection example:

```txt
DATABASE_URL=postgresql://postgres:YOUR_DATABASE_PASSWORD@db.YOUR_PROJECT_REF.supabase.co:5432/postgres?sslmode=require
```

Pooler connection example for IPv4-only networks:

```txt
DATABASE_URL=postgresql://postgres.YOUR_PROJECT_REF:YOUR_DATABASE_PASSWORD@YOUR_POOLER_HOST.supabase.com:5432/postgres?sslmode=require&connect_timeout=15
```

For local development, keep Redis in Docker with `docker compose up -d redis`, or use a managed Redis provider later.

## API Scope

Planned MVP modules:

- Auth: register, login, refresh, logout, current user
- Habits: CRUD, completion logs, streaks, weekly completion
- Workouts: plans, exercises, sessions, set logs
- Progress: dashboard aggregates, body weight logs, chart data

Implemented endpoint docs are in `API.md`.

## Environment Variables

See `.env.example` for defaults.

Auth password encryption requires an RSA keypair:

- Backend: `AUTH_PASSWORD_PRIVATE_KEY` contains the PEM PKCS8 RSA private key.
- Frontend: `NEXT_PUBLIC_AUTH_PASSWORD_PUBLIC_KEY` contains the matching PEM public key.
- Register/login payloads must send `password_encrypted` and `password_alg: "RSA-OAEP-SHA256"`; raw `password` is not accepted.

## Notes

- PostgreSQL is the source of truth.
- Redis should be used first for refresh/session support, then optional caching.
- Elasticsearch should not block core API features; add it after the MVP routes are stable.
- Do not log passwords, token values, authorization headers, or sensitive health details.

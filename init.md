# FitFlow — Habit & Workout Tracker Full-Stack Project Plan

## 1. Project Overview

**FitFlow** is a full-stack Habit + Workout Tracker built with **Go** and **Next.js**.

The app helps users create daily habits, plan workout routines, log workout sessions, track progress, and observe system activity through an ELK-based logging infrastructure.

The goal of this project is to build a portfolio-worthy full-stack application with:

- Authentication
- Habit tracking
- Workout planning
- Workout logging
- Progress analytics
- PostgreSQL persistence
- Redis caching/session support
- Elasticsearch and Kibana observability

---

## 2. Core Product Idea

FitFlow helps users answer:

- What workout should I do today?
- Did I complete my habits today?
- Am I consistent this week?
- Which exercises am I progressing on?
- How many workouts have I completed this month?
- What does my activity trend look like?

The app should start simple, then grow into a production-style application.

---

## 3. Recommended Tech Stack

### Frontend

- Next.js
- TypeScript
- Tailwind CSS
- Shadcn UI
- TanStack Query
- React Hook Form
- Zod
- Recharts
- Zustand

### Backend

- Go
- Chi router
- PostgreSQL
- Redis
- JWT authentication
- Refresh token rotation
- Structured logging with zap

### Infrastructure

- Docker Compose
- PostgreSQL
- Redis
- Elasticsearch
- Kibana
- Optional Filebeat
- Optional Logstash

---

## 4. High-Level Architecture

```txt
User Browser
   ↓
Next.js Frontend
   ↓
Go REST API
   ↓
PostgreSQL

Go REST API
   ↓
Redis

Go API Logs / Docker Logs
   ↓
Filebeat or Logstash
   ↓
Elasticsearch
   ↓
Kibana
```

Recommended architecture:

```txt
FitFlow Core App:
Next.js → Go API → PostgreSQL

Session and Cache:
Go API → Redis

Observability:
Go API JSON logs → Filebeat → Elasticsearch → Kibana

Optional Search:
Go API → Elasticsearch
```

---

## 5. Main Modules

### 5.1 Authentication

Users can register, log in, refresh sessions, and log out.

Features:

- Register
- Login
- Logout
- Refresh token rotation
- Protected routes
- User profile
- Password hashing
- Auth middleware

API routes:

```txt
POST   /api/auth/register
POST   /api/auth/login
POST   /api/auth/refresh
POST   /api/auth/logout
GET    /api/auth/me
```

---

### 5.2 Habit Tracker

Users can create daily habits and mark them as completed.

Example habits:

- Workout
- Drink 2L water
- Eat enough protein
- Sleep before 11 PM
- Walk 5,000 steps
- No junk food

Features:

- Create habit
- Edit habit
- Delete habit
- Mark habit complete for today
- View daily habit checklist
- View streak count
- View weekly completion percentage
- PWA friendly
- responsive UI

API routes:

```txt
GET    /api/habits
POST   /api/habits
GET    /api/habits/:id
PUT    /api/habits/:id
DELETE /api/habits/:id

POST   /api/habits/:id/complete
DELETE /api/habits/:id/complete
GET    /api/habits/logs?from=2026-06-01&to=2026-06-07
```

---

### 5.3 Workout Routine Planner

Users can create workout plans and add exercises to each plan.

Example workout plan:

```txt
Monday — Pull Day
- One-arm dumbbell row
- Dumbbell curl
- Hammer curl
- Rear delt raise
- Plank
```

Features:

- Create workout plan
- Add exercises to workout plan
- Set reps, sets, weight, and rest time
- Assign workout to scheduled day
- View today's workout

API routes:

```txt
GET    /api/workout-plans
POST   /api/workout-plans
GET    /api/workout-plans/:id
PUT    /api/workout-plans/:id
DELETE /api/workout-plans/:id

POST   /api/workout-plans/:id/exercises
PUT    /api/exercises/:id
DELETE /api/exercises/:id
```

---

### 5.4 Workout Logging

Users can start a workout session, log sets/reps/weight, and finish the session.

Example log:

```txt
Exercise: Dumbbell Curl
Set 1: 10 reps x 10kg
Set 2: 9 reps x 10kg
Set 3: 8 reps x 10kg
```

Features:

- Start workout session
- Log sets
- Log reps
- Log weight
- Add notes
- Finish workout
- View workout history

API routes:

```txt
POST   /api/workout-sessions/start
GET    /api/workout-sessions
GET    /api/workout-sessions/:id
POST   /api/workout-sessions/:id/sets
PUT    /api/workout-sessions/:id/finish
DELETE /api/workout-sessions/:id
```

---

### 5.5 Progress Dashboard

The dashboard gives users a summary of habits, workouts, and body progress.

Dashboard cards:

- Habits completed today
- Current workout streak
- Weekly habit completion
- Total workouts this month
- Latest body weight
- Most trained muscle group

Charts:

- Body weight over time
- Workouts per week
- Habit completion rate
- Exercise progression

API routes:

```txt
GET    /api/dashboard
GET    /api/progress/habits
GET    /api/progress/workouts
GET    /api/progress/body-weight
POST   /api/progress/body-weight
```

Example dashboard response:

```json
{
  "date": "2026-06-08",
  "habits": {
    "total": 5,
    "completed": 3,
    "completion_rate": 60
  },
  "workouts": {
    "completed_this_week": 4,
    "current_streak": 3,
    "scheduled_today": {
      "id": "workout-plan-id",
      "name": "Pull Day"
    }
  },
  "body": {
    "latest_weight_kg": 79,
    "weight_change_30_days": -1.5
  }
}
```

---

## 6. MVP Scope

The recommended MVP should focus only on habit and workout tracking.

### Phase 1 — Foundation

- Auth
- User profile
- Dashboard layout
- Protected routes
- Basic app shell

### Phase 2 — Habits

- Create habit
- List habits
- Mark habit complete today
- Show current streak
- Show weekly completion

### Phase 3 — Workout Planner

- Create workout plan
- Add exercises
- Assign workout day
- View today's workout

### Phase 4 — Workout Logging

- Start workout session
- Log sets, reps, and weight
- Finish workout
- View workout history

### Phase 5 — Progress

- Body weight log
- Workout count chart
- Habit completion chart
- Exercise progress chart

---

## 7. Version 2 Features

After finishing the MVP, add:

- Meal logging
- Calendar view
- Workout reminders
- Prebuilt workout templates
- Advanced analytics
- Export to CSV/PDF
- PWA mobile support

---

## 8. Version 3 Features

Advanced features for a stronger portfolio project:

- XP and level system
- Badges and achievements
- AI-like workout suggestions
- Smart workout adjustment based on missed days
- Public progress sharing
- Admin exercise library
- Elasticsearch-powered search

---

## 9. Database Design

### 9.1 users

```sql
CREATE TABLE users (
  id UUID PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  fitness_goal VARCHAR(50),
  height_cm INT,
  weight_kg DECIMAL(5,2),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);
```

---

### 9.2 habits

```sql
CREATE TABLE habits (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name VARCHAR(100) NOT NULL,
  description TEXT,
  frequency VARCHAR(30) DEFAULT 'daily',
  target_count INT DEFAULT 1,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);
```

---

### 9.3 habit_logs

```sql
CREATE TABLE habit_logs (
  id UUID PRIMARY KEY,
  habit_id UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  log_date DATE NOT NULL,
  completed_count INT DEFAULT 1,
  is_completed BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT NOW(),

  UNIQUE(habit_id, log_date)
);
```

---

### 9.4 workout_plans

```sql
CREATE TABLE workout_plans (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name VARCHAR(100) NOT NULL,
  description TEXT,
  scheduled_day VARCHAR(20),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);
```

---

### 9.5 exercises

```sql
CREATE TABLE exercises (
  id UUID PRIMARY KEY,
  workout_plan_id UUID NOT NULL REFERENCES workout_plans(id) ON DELETE CASCADE,
  name VARCHAR(100) NOT NULL,
  muscle_group VARCHAR(50),
  target_sets INT,
  target_reps INT,
  target_weight_kg DECIMAL(5,2),
  rest_seconds INT DEFAULT 60,
  order_index INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT NOW()
);
```

---

### 9.6 workout_sessions

```sql
CREATE TABLE workout_sessions (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  workout_plan_id UUID REFERENCES workout_plans(id) ON DELETE SET NULL,
  started_at TIMESTAMP DEFAULT NOW(),
  finished_at TIMESTAMP,
  status VARCHAR(30) DEFAULT 'in_progress',
  notes TEXT
);
```

---

### 9.7 workout_set_logs

```sql
CREATE TABLE workout_set_logs (
  id UUID PRIMARY KEY,
  workout_session_id UUID NOT NULL REFERENCES workout_sessions(id) ON DELETE CASCADE,
  exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
  set_number INT NOT NULL,
  reps INT NOT NULL,
  weight_kg DECIMAL(5,2),
  completed BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT NOW()
);
```

---

### 9.8 body_weight_logs

```sql
CREATE TABLE body_weight_logs (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  weight_kg DECIMAL(5,2) NOT NULL,
  log_date DATE NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),

  UNIQUE(user_id, log_date)
);
```

---

## 10. Backend Folder Structure

```txt
backend/
  cmd/
    api/
      main.go

  internal/
    config/
      config.go

    database/
      postgres.go
      redis.go

    middleware/
      auth.go
      cors.go
      logger.go

    modules/
      auth/
        handler.go
        service.go
        repository.go
        dto.go

      user/
        handler.go
        service.go
        repository.go
        dto.go

      habit/
        handler.go
        service.go
        repository.go
        dto.go

      workout/
        handler.go
        service.go
        repository.go
        dto.go

      progress/
        handler.go
        service.go
        repository.go
        dto.go

    utils/
      jwt.go
      password.go
      response.go
      validator.go

  migrations/
    001_create_users.sql
    002_create_habits.sql
    003_create_workouts.sql

  go.mod
  Dockerfile
```

---

## 11. Frontend Folder Structure

```txt
frontend/
  app/
    auth/
      login/
        page.tsx
      register/
        page.tsx

    dashboard/
      page.tsx

    habits/
      page.tsx
      new/
        page.tsx
      [id]/
        page.tsx

    workouts/
      page.tsx
      new/
        page.tsx
      [id]/
        page.tsx
      session/
        [id]/
          page.tsx

    progress/
      page.tsx

    settings/
      page.tsx

  components/
    ui/
    layout/
    habits/
    workouts/
    charts/

  features/
    auth/
      api.ts
      hooks.ts
      schema.ts
      types.ts

    habits/
      api.ts
      hooks.ts
      schema.ts
      types.ts

    workouts/
      api.ts
      hooks.ts
      schema.ts
      types.ts

  lib/
    api-client.ts
    auth.ts
    utils.ts

  stores/
    auth-store.ts

  middleware.ts
```

---

## 12. ELK Infrastructure Plan

ELK should be used for observability, logs, analytics, and optional search.

Do not use Elasticsearch as the main database. PostgreSQL should remain the source of truth.

### ELK Usage

Use Elasticsearch and Kibana for:

- Centralized logging
- API monitoring
- Error tracking
- Auth event tracking
- User activity analytics
- Optional search

---

## 13. What to Log

### Technical Logs

Examples:

- HTTP request received
- HTTP request completed
- API error occurred
- Database error occurred
- Unauthorized request
- Slow endpoint detected

Example:

```json
{
  "service": "fitflow-api",
  "level": "info",
  "event": "http_request",
  "method": "POST",
  "path": "/api/workout-sessions/start",
  "status": 201,
  "duration_ms": 42,
  "user_id": "user-id",
  "timestamp": "2026-06-08T10:15:00Z"
}
```

### Business Events

Examples:

- User registered
- Habit created
- Habit completed
- Workout session started
- Workout session finished
- Workout set logged
- Body weight logged

Example:

```json
{
  "event": "habit_completed",
  "user_id": "user-id",
  "habit_id": "habit-id",
  "habit_name": "Drink water",
  "log_date": "2026-06-08",
  "timestamp": "2026-06-08T08:00:00Z"
}
```

### Auth Events

Examples:

- Login success
- Login failed
- Logout
- Refresh token rotated
- Unauthorized access attempt

Example:

```json
{
  "event": "login_failed",
  "user_identifier_hash": "hashed-value",
  "ip": "127.0.0.1",
  "reason": "invalid_credentials"
}
```

---

## 14. What Not to Log

Avoid logging sensitive data:

- Password
- Password hash
- Access token
- Refresh token
- Authorization header
- Raw personal health details
- Full email address in failure logs, unless necessary

Safer fields:

- User ID
- Request ID
- Event type
- Status code
- Endpoint path
- Duration
- Generic error reason

---

## 15. Elasticsearch Index Design

Recommended indices:

```txt
fitflow-api-logs-YYYY.MM.DD
fitflow-business-events-YYYY.MM.DD
fitflow-auth-events-YYYY.MM.DD
```

Examples:

```txt
fitflow-api-logs-2026.06.08
fitflow-business-events-2026.06.08
fitflow-auth-events-2026.06.08
```

---

## 16. Kibana Dashboard Ideas

### API Health Dashboard

- Requests per minute
- Average response time
- Error rate
- Top 10 slow endpoints
- Status code distribution

### Auth Dashboard

- Successful logins
- Failed logins
- Refresh token rotations
- Logout count
- Unauthorized attempts

### Product Usage Dashboard

- Habits completed per day
- Workouts completed per day
- Most popular habit names
- Most logged exercises
- Average workout duration

### Error Dashboard

- Errors by endpoint
- Errors by service
- 500 errors over time
- Most common error messages

---

## 17. Docker Compose Infrastructure

```yaml
version: "3.9"

services:
  postgres:
    image: postgres:16
    container_name: fitflow_postgres
    environment:
      POSTGRES_USER: fitflow
      POSTGRES_PASSWORD: fitflow_password
      POSTGRES_DB: fitflow_db
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7
    container_name: fitflow_redis
    ports:
      - "6379:6379"

  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.17.0
    container_name: fitflow_elasticsearch
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
      - ES_JAVA_OPTS=-Xms512m -Xmx512m
    ports:
      - "9200:9200"
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data

  kibana:
    image: docker.elastic.co/kibana/kibana:8.17.0
    container_name: fitflow_kibana
    environment:
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
    ports:
      - "5601:5601"
    depends_on:
      - elasticsearch

volumes:
  postgres_data:
  elasticsearch_data:
```

For local development, disabling Elastic security is convenient. For production, enable:

- Authentication
- TLS
- Network restrictions
- Strong passwords
- Role-based access

---

## 18. Go Request Logging Middleware Idea

```go
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := NewResponseWriter(w)
		next.ServeHTTP(ww, r)

		duration := time.Since(start)

		slog.Info("http_request",
			"service", "fitflow-api",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"duration_ms", duration.Milliseconds(),
			"request_id", r.Header.Get("X-Request-ID"),
		)
	})
}
```

---

## 19. Recommended Development Roadmap

### Step 1 — Project Setup

- Create monorepo
- Add frontend app
- Add backend app
- Add Docker Compose
- Add PostgreSQL, Redis, Elasticsearch, and Kibana

### Step 2 — Authentication

- User registration
- Login
- JWT access token
- Refresh token rotation
- Protected API middleware
- Protected frontend routes

### Step 3 — Habit Tracker

- Habit CRUD
- Habit completion logs
- Daily checklist
- Habit streak logic

### Step 4 — Workout Planner

- Workout plan CRUD
- Exercise CRUD
- Scheduled workout day
- Today's workout view

### Step 5 — Workout Session Logging

- Start workout
- Add set logs
- Finish workout
- Workout history

### Step 6 — Dashboard and Progress

- Dashboard summary
- Weekly habit completion
- Workout count
- Body weight logs
- Charts

### Step 7 — ELK Observability

- Add structured JSON logs
- Ship logs to Elasticsearch
- Create Kibana index patterns
- Build API health dashboard
- Build product usage dashboard

---

## 20. Final MVP Definition

The final MVP should include:

1. Register/login
2. Dashboard
3. Create daily habits
4. Mark habits complete
5. Create workout plans
6. Add exercises
7. Start workout session
8. Log sets, reps, and weight
9. Finish workout
10. View weekly progress
11. Structured backend logs
12. Elasticsearch log storage
13. Kibana dashboards

---

## 21. Portfolio Description

You can describe this project as:

> FitFlow is a full-stack habit and workout tracking system built with Go and Next.js. It includes JWT authentication, refresh token rotation, PostgreSQL persistence, Redis session support, structured logging, Elasticsearch indexing, and Kibana dashboards for observability and product analytics.

This project demonstrates:

- Full-stack app architecture
- REST API design
- Authentication and authorization
- PostgreSQL schema design
- Redis integration
- Structured backend logging
- Docker-based local infrastructure
- Elasticsearch and Kibana observability
- Dashboard and analytics UI

# FitFlow Backend API

Base URL for local development:

```txt
http://localhost:8080
```

All JSON responses use one of these shapes:

```json
{ "data": {} }
```

```json
{ "error": { "code": "error_code", "message": "Human readable message" } }
```

## Health

### GET /health

Returns API health.

### GET /api/health

Returns API health under the API prefix.

## Auth

### POST /api/auth/register

Request:

```json
{
  "name": "Brian",
  "email": "brian@example.com",
  "password_encrypted": "base64-rsa-oaep-ciphertext",
  "password_alg": "RSA-OAEP-SHA256"
}
```

Response `201`:

```json
{
  "data": {
    "access_token": "jwt",
    "refresh_token": "refresh-token",
    "user": {
      "id": "uuid",
      "name": "Brian",
      "email": "brian@example.com",
      "created_at": "2026-06-11T00:00:00Z"
    }
  }
}
```

### POST /api/auth/login

Request:

```json
{
  "email": "brian@example.com",
  "password_encrypted": "base64-rsa-oaep-ciphertext",
  "password_alg": "RSA-OAEP-SHA256"
}
```

Response `200` has the same shape as register.

The frontend must encrypt passwords with Web Crypto using RSA-OAEP and SHA-256. The frontend public key must match the backend private key configured in `AUTH_PASSWORD_PRIVATE_KEY`.

Frontend environment variable:

```txt
NEXT_PUBLIC_AUTH_PASSWORD_PUBLIC_KEY=-----BEGIN PUBLIC KEY-----...
```

Backend environment variable:

```txt
AUTH_PASSWORD_PRIVATE_KEY=-----BEGIN PRIVATE KEY-----...
```

The backend never accepts raw `password` in register/login payloads.

### POST /api/auth/refresh

Request:

```json
{ "refresh_token": "refresh-token" }
```

Response `200` returns a new access token and rotated refresh token.

### POST /api/auth/logout

Request:

```json
{ "refresh_token": "refresh-token" }
```

Response `200`:

```json
{ "data": { "success": true } }
```

### GET /api/auth/me

Headers:

```txt
Authorization: Bearer ACCESS_TOKEN
```

Response `200`:

```json
{
  "data": {
    "id": "uuid",
    "name": "Brian",
    "email": "brian@example.com",
    "created_at": "2026-06-11T00:00:00Z"
  }
}
```

## Habits

All habit endpoints require:

```txt
Authorization: Bearer ACCESS_TOKEN
```

### GET /api/habits

Response `200`:

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Drink water",
      "description": "2L per day",
      "frequency": "daily",
      "target_count": 1,
      "is_active": true,
      "completed_today": false,
      "current_streak": 0,
      "created_at": "2026-06-11T00:00:00Z",
      "updated_at": "2026-06-11T00:00:00Z"
    }
  ]
}
```

### POST /api/habits

Request:

```json
{
  "name": "Drink water",
  "description": "2L per day",
  "frequency": "daily",
  "target_count": 1
}
```

Response `201` returns the created habit.

### POST /api/habits/{id}/complete

Marks the habit complete for today.

Response `200`:

```json
{ "data": { "success": true } }
```

### DELETE /api/habits/{id}/complete

Removes today's completion log.

Response `200`:

```json
{ "data": { "success": true } }
```

## Planned Next Endpoints

These are still planned from `init.md` and not implemented yet:

- Workout plan CRUD
- Exercise CRUD
- Workout session logging
- Dashboard aggregates
- Body measurement logs API

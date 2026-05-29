# Implementation Plan: JWT Authentication + Account Entity

## Context & Corrections

This plan replaces the old API Key authentication with JWT-based authentication.

**Key business rules (corrected):**
- The system has **2 roles only**: `admin` and `driver`
- **Customers do NOT have system access** — orders are created on their behalf by admins
- When an order is created, it is **assigned to a driver** (by `driver_id`)
- Demo account to seed: **1 admin** and **1 driver** for testing

---

## Current State

| What exists | Detail |
|---|---|
| Auth middleware | Static API Key via `X-API-KEY` header, 3 hardcoded env vars |
| Roles in use | `admin`, `driver`, `customer` (customer must be removed) |
| No user table | No database table for accounts |
| Orders table | Has `user_info` JSONB (customer info), but no `driver_id` column |

---

## Phase 1 — Add JWT Library

Install `github.com/golang-jwt/jwt/v5` via `go get`.
`golang.org/x/crypto` (for `bcrypt`) is already a transitive dependency — just import it directly.

---

## Phase 2 — Database Migration: `accounts` Table + `driver_id` on Orders

### Migration `000006_add_accounts_and_driver_assignment`

**Up:**
```sql
-- System accounts (admin and driver roles only)
CREATE TABLE accounts (
    id         BIGSERIAL PRIMARY KEY,
    username   VARCHAR(100) NOT NULL UNIQUE,
    email      VARCHAR(255) NOT NULL UNIQUE,
    password   VARCHAR(255) NOT NULL,          -- bcrypt hash
    role       VARCHAR(20)  NOT NULL,          -- 'admin' or 'driver'
    created_at TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Assign each order to a driver
ALTER TABLE orders ADD COLUMN driver_id BIGINT REFERENCES accounts(id) ON DELETE SET NULL;
```

**Down:**
```sql
ALTER TABLE orders DROP COLUMN IF EXISTS driver_id;
DROP TABLE IF EXISTS accounts;
```

---

## Phase 3 — Account Model & Repository

### `internal/models/account.go`

```go
type Account struct {
    ID        int64     `gorm:"primaryKey;column:id"`
    Username  string    `gorm:"column:username;uniqueIndex;not null"`
    Email     string    `gorm:"column:email;uniqueIndex;not null"`
    Password  string    `gorm:"column:password;not null"`  // bcrypt hash
    Role      string    `gorm:"column:role;not null"`      // "admin" | "driver"
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Update `internal/models/order.go`

Add `DriverID *int64` field to the `Order` struct (nullable, references `accounts.id`).

### `internal/repositories/account.go`

```go
type AccountRepository interface {
    FindByEmail(email string) (*models.Account, error)
    Create(account models.Account) (*models.Account, error)
}
```

---

## Phase 4 — Auth DTOs

### `internal/dto/auth.go`

```go
type LoginRequest struct {
    Email    string `json:"email"    validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
    AccessToken string `json:"access_token"`
    TokenType   string `json:"token_type"`  // "Bearer"
    ExpiresIn   int    `json:"expires_in"`  // seconds, e.g. 86400
    Role        string `json:"role"`        // "admin" or "driver"
}
```

---

## Phase 5 — Auth Service

### `internal/services/auth.go`

- `Login(email, password string) (*dto.LoginResponse, error)`
  1. Fetch account by email from DB
  2. Compare password with `bcrypt.CompareHashAndPassword`
  3. Return `ERR_UNAUTHENTICATED` if not found or wrong password
  4. Sign JWT with claims: `sub` (account ID), `email`, `role`, `exp`
  5. Return `LoginResponse`

- JWT secret from env: `JWT_SECRET`
- Token expiry from env: `JWT_EXPIRE_HOURS` (default: `24`)
- Algorithm: **HS256**

---

## Phase 6 — Auth Handler & Router

### `internal/handlers/auth.go`

Endpoint: `POST /api/v1/auth/login`
- Bind + validate `LoginRequest`
- Call `authService.Login()`
- Return `LoginResponse` (200) or 401 on bad credentials

### `internal/routers/v1/auth.go`

Register `POST /api/v1/auth/login` with **no** auth middleware.

---

## Phase 7 — Rewrite Auth Middleware (JWT)

### `internal/middlewares/auth.go`

**`Authenticate()` — new logic:**
```
1. Read header: Authorization: Bearer <token>
2. Parse JWT, verify signature with JWT_SECRET
3. Validate expiry (jwt library does this automatically)
4. Extract claims: role, sub (account_id)
5. Set c.Locals("role", role)
6. Set c.Locals("account_id", sub)
7. Return 401 if token is missing, malformed, expired, or invalid signature
```

**`Authorize(roles []string)` — unchanged**, reads `c.Locals("role")` exactly as before.

---

## Phase 8 — Update Router Role Guards

Remove `"customer"` from all `Authorize()` calls:

| Router | Before | After |
|---|---|---|
| `GET /api/v1/orders` | `["admin","customer","driver"]` | `["admin","driver"]` |
| `GET /api/v1/orders/:id` | `["admin","customer","driver"]` | `["admin","driver"]` |
| `POST /api/v1/orders` | `["customer","admin"]` | `["admin"]` |
| `PATCH /api/v1/orders/:id/status` | `["admin","customer","driver"]` | `["admin","driver"]` |
| `POST /api/v1/order-events/import` | `["admin"]` | `["admin"]` (unchanged) |
| `GET/POST /api/v1/reports/daily` | `["admin"]` | `["admin"]` (unchanged) |

---

## Phase 9 — Update `main.go`

```go
// Wire auth
accountRepo    := repositories.NewAccountRepository(db)
authService    := services.NewAuthService(accountRepo)
authHandler    := handlers.NewAuthHandler(authService)
routers.SetupAuthRouter(app, authHandler)

// Update Swagger annotation
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter: Bearer <token>
```

---

## Phase 10 — Environment Variables

**Add to `.env`:**
```env
JWT_SECRET=your-super-secret-key-change-in-production
JWT_EXPIRE_HOURS=24
```

**Remove from `.env`:**
```env
# CUSTOMER_API_KEY=...   ← remove
# DRIVER_API_KEY=...     ← remove
# ADMIN_API_KEY=...      ← remove
```

**Update `docker-compose.yml`** environment section accordingly.

---

## Phase 11 — Seed Demo Accounts

Add to **`db/seed.sql`** (or migration `000007_seed_demo_accounts`):

```sql
-- password: Admin@123  (bcrypt hash)
INSERT INTO accounts (username, email, password, role)
VALUES ('admin_demo', 'admin@demo.com', '$2a$10$<hash>', 'admin');

-- password: Driver@123  (bcrypt hash)
INSERT INTO accounts (username, email, password, role)
VALUES ('driver_demo', 'driver@demo.com', '$2a$10$<hash>', 'driver');
```

> The actual bcrypt hashes will be pre-generated during implementation using `bcrypt.GenerateFromPassword`.

**Demo credentials:**

| Role | Email | Password |
|---|---|---|
| Admin | `admin@demo.com` | `Admin@123` |
| Driver | `driver@demo.com` | `Driver@123` |

---

## Phase 12 — Update Swagger Docs

In `cmd/api/main.go`, replace:
```go
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-KEY
```
With:
```go
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.
```

In all handler godoc comments, replace:
```
// @Slack Security ApiKeyAuth
```
With:
```
// @Slack Security BearerAuth
```

---

## File Change Summary

| Action | File | Description |
|---|---|---|
| NEW | `db/migrations/000006_add_accounts_and_driver_assignment.up.sql` | Create `accounts` table, add `driver_id` to orders |
| NEW | `db/migrations/000006_add_accounts_and_driver_assignment.down.sql` | Rollback |
| NEW | `internal/models/account.go` | GORM Account model |
| MODIFIED | `internal/models/order.go` | Add `DriverID *int64` field |
| NEW | `internal/repositories/account.go` | AccountRepository interface + impl |
| NEW | `internal/dto/auth.go` | LoginRequest / LoginResponse DTOs |
| NEW | `internal/services/auth.go` | AuthService (login + JWT generation) |
| NEW | `internal/handlers/auth.go` | Auth handler for login endpoint |
| NEW | `internal/routers/v1/auth.go` | Auth route setup |
| MODIFIED | `internal/middlewares/auth.go` | Replace API Key → JWT validation |
| MODIFIED | `internal/routers/v1/order.go` | Remove "customer" from all Authorize() calls |
| MODIFIED | `cmd/api/main.go` | Wire auth layer, update Swagger security def |
| MODIFIED | `internal/handlers/order.go` | `@Security BearerAuth` annotation |
| MODIFIED | `internal/handlers/order_event.go` | `@Security BearerAuth` annotation |
| MODIFIED | `internal/handlers/report.go` | `@Security BearerAuth` annotation |
| MODIFIED | `db/seed.sql` | Add demo admin + driver accounts |
| MODIFIED | `.env` | Add JWT vars, remove API key vars |

---

## Auth Flow Summary

```
1. Login
   POST /api/v1/auth/login
   Body: { "email": "admin@demo.com", "password": "Admin@123" }
   Response: { "access_token": "eyJ...", "token_type": "Bearer", "expires_in": 86400, "role": "admin" }

2. Use token on protected endpoints
   GET /api/v1/orders
   Header: Authorization: Bearer eyJ...
   Response: 200 OK  (admin or driver token accepted)

   POST /api/v1/orders
   Header: Authorization: Bearer eyJ...   (must be admin role)
   Response: 201 Created
```

---

## Role Permission Matrix (Final)

| Endpoint | admin | driver |
|---|---|---|
| `POST /api/v1/auth/login` | :white_check_mark: (public) | :white_check_mark: (public) |
| `GET /api/v1/orders` | :white_check_mark: | :white_check_mark: |
| `GET /api/v1/orders/:id` | :white_check_mark: | :white_check_mark: |
| `POST /api/v1/orders` | :white_check_mark: | :x: |
| `PATCH /api/v1/orders/:id/status` | :white_check_mark: | :white_check_mark: |
| `POST /api/v1/order-events/import` | :white_check_mark: | :x: |
| `GET /api/v1/reports/daily` | :white_check_mark: | :x: |
| `POST /api/v1/reports/daily` | :white_check_mark: | :x: |
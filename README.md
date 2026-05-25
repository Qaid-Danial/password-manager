# SecureVault — Password Manager

**CCS6344 Database & Cloud Security | Group 17**

A web-based password manager built to demonstrate SQL database security, AES-256-GCM encryption, JWT authentication, audit logging, and role-based access control.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | React 18 + Vite + TailwindCSS |
| Backend | Go 1.21 + Gin |
| Database | PostgreSQL 16 |
| Authentication | JWT (HS256) + bcrypt (cost 12) |
| Encryption | AES-256-GCM (application layer) |
| Dev Environment | Docker Compose |

---

## Project Structure

```
password-manager/
├── backend/          Go API server
│   ├── cmd/          Entry point (main.go)
│   ├── config/       Environment config loader
│   ├── database/     PostgreSQL connection pool
│   ├── handlers/     HTTP request handlers
│   ├── middleware/   JWT auth, RBAC, rate limit, security headers
│   ├── models/       Request/response structs
│   ├── routes/       Route registration
│   ├── services/     Business logic (auth, vault, audit, admin)
│   └── utils/        Crypto, JWT, password helpers
├── frontend/         React SPA
│   └── src/
│       ├── components/  Navbar, PasswordGenerator, PasswordStrength, etc.
│       ├── pages/       Login, Register, Dashboard, Add/Edit, AuditLogs
│       ├── services/    API call functions
│       └── utils/       Axios instance, password generator
├── database/
│   ├── schema.sql    Table definitions + triggers
│   ├── seed.sql      Dev seed accounts
│   └── security.sql  Least-privilege DB role + indexes
├── docs/             Assignment documentation
└── docker-compose.yml
```

---

## Quick Start (Docker — recommended)

### 1. Clone and enter the project

```bash
git clone https://github.com/Qaid-Danial/password-manager.git
cd password-manager
```

### 2. Create the backend `.env` file

```bash
cp backend/.env.example backend/.env
```

Edit `backend/.env` with real values:

```env
DATABASE_URL=postgres://pmuser:pmpassword@db:5432/passwordmanager?sslmode=disable
JWT_SECRET=<run: openssl rand -base64 32>
ENCRYPTION_KEY=<run: openssl rand -base64 32>
PORT=8080
ENVIRONMENT=development
RATE_LIMIT_REQUESTS=5
RATE_LIMIT_WINDOW=60
```

> **Important:** `ENCRYPTION_KEY` must decode to exactly 32 bytes.
> `openssl rand -base64 32` produces the correct length.

### 3. Start all services

```bash
docker compose up --build
```

| Service | URL |
|---|---|
| Frontend | http://localhost |
| Backend API | http://localhost:8080 |
| PostgreSQL | localhost:5432 |

---

## Manual Setup (without Docker)

### Prerequisites
- Go 1.21+
- Node.js 20+
- PostgreSQL 16 running locally

### Backend

```bash
cd backend
cp .env.example .env
# Edit .env — set DATABASE_URL host to "localhost" instead of "db"
# Set JWT_SECRET and ENCRYPTION_KEY

go mod tidy
go run ./cmd/main.go
```

### Frontend

```bash
cd frontend
npm install
npm run dev
# Open http://localhost:5173
```

The Vite dev server proxies all `/api/` requests to the backend on port 8080.

### Database (local PostgreSQL)

```bash
psql -U postgres -c "CREATE DATABASE passwordmanager;"
psql -U postgres -d passwordmanager -f database/schema.sql
psql -U postgres -d passwordmanager -f database/security.sql
psql -U postgres -d passwordmanager -f database/seed.sql
```

---

## Default Seed Accounts

> For development only. Never use these credentials in production.

| Role | Username | Email | Password |
|---|---|---|---|
| Admin | `admin` | `admin@example.com` | `AdminPass123!` |
| User | `testuser` | `test@example.com` | `TestPass123!` |

The admin account can access the **Audit Logs** page at `/admin/audit-logs`.

---

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `JWT_SECRET` | Yes | JWT signing secret (minimum 32 characters) |
| `ENCRYPTION_KEY` | Yes | AES-256 key, base64-encoded (must decode to exactly 32 bytes) |
| `PORT` | No | HTTP listen port (default: `8080`) |
| `ENVIRONMENT` | No | `development` or `production` |
| `RATE_LIMIT_REQUESTS` | No | Max login attempts per window per IP (default: `5`) |
| `RATE_LIMIT_WINDOW` | No | Rate limit window in seconds (default: `60`) |

---

## API Endpoints

### Public

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/health` | Health check — DB ping |
| `POST` | `/api/auth/register` | Register a new user |
| `POST` | `/api/auth/login` | Login — rate limited (5 req/60s per IP) |

### Protected (requires `Authorization: Bearer <token>`)

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/vault` | List all vault entries (decrypted) |
| `POST` | `/api/vault` | Create a credential |
| `GET` | `/api/vault/:id` | Get one credential (decrypted + audited) |
| `PUT` | `/api/vault/:id` | Update a credential |
| `DELETE` | `/api/vault/:id` | Delete a credential |

### Admin only (requires `role: admin`)

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/admin/audit-logs` | Paginated audit log (`?page=1&page_size=20`) |

---

## Security Features

| Feature | Implementation |
|---|---|
| Password hashing | bcrypt cost 12 — plaintext never stored or logged |
| Vault encryption | AES-256-GCM at application layer; DB stores only ciphertext |
| SQL injection prevention | Parameterized queries (`$1`, `$2`, …) throughout — no string-concatenated SQL |
| JWT authentication | HS256 with explicit algorithm validation (prevents alg:none attack) |
| IDOR prevention | Every vault query includes `AND user_id = $N` ownership check |
| Brute-force mitigation | Token-bucket rate limit on login route (5 req / 60s per IP) |
| Timing attack prevention | Dummy bcrypt runs when email not found (prevents user enumeration) |
| Audit logging | 7 action types logged with IP and user agent; errors never block the primary action |
| Role-based access control | `user` vs `admin` enforced in middleware chain |
| HTTP security headers | X-Frame-Options, CSP, X-Content-Type-Options, Referrer-Policy, Permissions-Policy |
| CORS | Restricted to localhost origins only |
| Least-privilege DB | `pm_app_user` role — per-table GRANT; UPDATE and DELETE revoked on `audit_logs` |
| Audit immutability | Application role cannot modify or delete audit log rows |
| Non-root container | Backend Docker image runs as unprivileged `appuser` |
| Minimal container image | Alpine runtime — no build tools or source in final image |

---

## Database Schema Summary

```sql
users (
  id UUID PK, username VARCHAR(50) UNIQUE,
  email VARCHAR(255) UNIQUE, password_hash TEXT,  -- bcrypt
  role VARCHAR(20) CHECK ('user','admin'), created_at TIMESTAMPTZ
)

vault_entries (
  id UUID PK, user_id UUID FK→users CASCADE,
  site_name VARCHAR(255), site_url VARCHAR(500),
  vault_username TEXT,
  encrypted_password TEXT,  -- AES-256-GCM ciphertext
  encrypted_notes TEXT,     -- AES-256-GCM ciphertext
  created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ
)

audit_logs (
  id UUID PK, user_id UUID FK→users SET NULL,  -- preserved on account deletion
  action VARCHAR(50), ip_address VARCHAR(45),
  user_agent TEXT, created_at TIMESTAMPTZ
)
```

---

## Verifying Security in the Database

```bash
# Connect to the running PostgreSQL container
docker compose exec db psql -U pmuser -d passwordmanager
```

```sql
-- Passwords stored as bcrypt hashes, never plaintext
SELECT username, LEFT(password_hash, 30) || '...' AS hash FROM users;

-- Vault credentials stored as AES-256-GCM ciphertext
SELECT site_name, LEFT(encrypted_password, 40) || '...' AS ciphertext FROM vault_entries;

-- Audit trail
SELECT action, ip_address, created_at FROM audit_logs ORDER BY created_at DESC LIMIT 10;

-- Confirm pm_app_user cannot delete audit logs (will return permission denied)
SET ROLE pm_app_user;
DELETE FROM audit_logs;  -- should fail
RESET ROLE;
```

---

## GitHub Repository

https://github.com/Qaid-Danial/password-manager

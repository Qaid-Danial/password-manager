# API Reference — SecureVault

**Base URL (local dev):** `http://localhost:8080`  
**Content-Type:** `application/json`  
**Authentication:** `Authorization: Bearer <jwt_token>`

---

## Health Check

### `GET /api/health`

No authentication required.

**Response 200:**
```json
{ "status": "ok", "database": "connected" }
```

**Response 503:**
```json
{ "status": "degraded", "database": "disconnected" }
```

---

## Authentication

### `POST /api/auth/register`

Register a new user account.

**Request body:**
```json
{
  "username": "alice",
  "email": "alice@example.com",
  "password": "SecurePass123!"
}
```

| Field | Type | Constraints |
|---|---|---|
| `username` | string | Required, 3–50 characters |
| `email` | string | Required, valid email format |
| `password` | string | Required, minimum 8 characters |

**Response 201:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "alice",
    "email": "alice@example.com",
    "role": "user",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Response 400** — validation failure:
```json
{ "error": "invalid request" }
```

**Response 409** — duplicate username or email:
```json
{ "error": "username or email already exists" }
```

**curl example:**
```bash
curl -s -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@example.com","password":"SecurePass123!"}'
```

---

### `POST /api/auth/login`

Authenticate and receive a JWT. **Rate limited: 5 requests per 60 seconds per IP.**

**Request body:**
```json
{
  "email": "alice@example.com",
  "password": "SecurePass123!"
}
```

**Response 200:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "alice",
    "email": "alice@example.com",
    "role": "user",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Response 401** — wrong credentials (same message for wrong email or wrong password):
```json
{ "error": "invalid credentials" }
```

**Response 429** — rate limit exceeded:
```json
{ "error": "too many requests, please try again later" }
```
Headers: `Retry-After: 60`

**curl example:**
```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"SecurePass123!"}' | jq -r .token)
```

---

## Vault (Protected — requires Bearer token)

### `GET /api/vault`

List all vault entries for the authenticated user. Passwords are decrypted before returning.

**Response 200:**
```json
{
  "entries": [
    {
      "id": "661e8400-e29b-41d4-a716-446655440001",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "site_name": "GitHub",
      "site_url": "https://github.com",
      "vault_username": "alice",
      "password": "gh_secret_abc123",
      "notes": "Work account",
      "created_at": "2024-01-15T11:00:00Z",
      "updated_at": "2024-01-15T11:00:00Z"
    }
  ]
}
```

**curl example:**
```bash
curl -s http://localhost:8080/api/vault \
  -H "Authorization: Bearer $TOKEN" | jq .
```

---

### `POST /api/vault`

Create a new vault entry. Password and notes are AES-256-GCM encrypted before storage.

**Request body:**
```json
{
  "site_name": "GitHub",
  "site_url": "https://github.com",
  "vault_username": "alice",
  "password": "gh_secret_abc123",
  "notes": "Work account"
}
```

| Field | Type | Constraints |
|---|---|---|
| `site_name` | string | Required, max 255 characters |
| `site_url` | string | Optional, max 500 characters |
| `vault_username` | string | Required |
| `password` | string | Required |
| `notes` | string | Optional |

**Response 201:** Returns the created entry (decrypted, same shape as GET response).

**curl example:**
```bash
curl -s -X POST http://localhost:8080/api/vault \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"site_name":"GitHub","site_url":"https://github.com","vault_username":"alice","password":"gh_secret_abc123","notes":"Work account"}' | jq .
```

---

### `GET /api/vault/:id`

Fetch and decrypt a single vault entry. Triggers an audit log entry (`view_credential`).

Returns **404** if the entry does not exist OR belongs to another user — this prevents ID enumeration.

**Response 200:** Same shape as a single entry in the list response.

**curl example:**
```bash
curl -s http://localhost:8080/api/vault/661e8400-e29b-41d4-a716-446655440001 \
  -H "Authorization: Bearer $TOKEN" | jq .
```

---

### `PUT /api/vault/:id`

Update a vault entry. Encrypts the new password and notes before storage.

**Request body:** Same fields as `POST /api/vault` (all required).

**Response 200:** Returns the updated entry.

**Response 404:** Entry not found or belongs to another user.

**curl example:**
```bash
curl -s -X PUT http://localhost:8080/api/vault/661e8400-e29b-41d4-a716-446655440001 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"site_name":"GitHub","site_url":"https://github.com","vault_username":"alice","password":"new_secret_456","notes":"Updated notes"}' | jq .
```

---

### `DELETE /api/vault/:id`

Delete a vault entry. Triggers an audit log entry (`delete_credential`).

**Response 204:** No content — deletion successful.

**Response 404:** Entry not found or belongs to another user.

**curl example:**
```bash
curl -s -X DELETE http://localhost:8080/api/vault/661e8400-e29b-41d4-a716-446655440001 \
  -H "Authorization: Bearer $TOKEN" -v
```

---

## Admin (requires `role: admin`)

### `GET /api/admin/audit-logs`

Returns paginated audit log entries, newest first.

**Query parameters:**

| Parameter | Default | Max | Description |
|---|---|---|---|
| `page` | `1` | — | Page number (1-indexed) |
| `page_size` | `20` | `100` | Results per page |

**Response 200:**
```json
{
  "logs": [
    {
      "id": "772e8400-e29b-41d4-a716-446655440002",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "action": "login_success",
      "ip_address": "127.0.0.1",
      "user_agent": "curl/8.1.2",
      "created_at": "2024-01-15T11:05:00Z"
    },
    {
      "id": "883e8400-e29b-41d4-a716-446655440003",
      "user_id": null,
      "action": "login_failure",
      "ip_address": "192.168.1.10",
      "user_agent": "Mozilla/5.0...",
      "created_at": "2024-01-15T10:59:00Z"
    }
  ],
  "total": 47,
  "page": 1,
  "page_size": 20
}
```

Note: `user_id` is `null` for unauthenticated events (e.g., login failures before identity is established).

**Response 403** — non-admin user:
```json
{ "error": "insufficient permissions" }
```

**curl example:**
```bash
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"AdminPass123!"}' | jq -r .token)

curl -s "http://localhost:8080/api/admin/audit-logs?page=1&page_size=20" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

---

## Common Error Responses

| Status | Meaning | Body |
|---|---|---|
| 400 | Invalid request body or validation failure | `{ "error": "invalid request" }` |
| 401 | Missing, expired, or invalid JWT | `{ "error": "invalid or expired token" }` |
| 403 | Authenticated but insufficient role | `{ "error": "insufficient permissions" }` |
| 404 | Resource not found (or belongs to another user) | `{ "error": "entry not found" }` |
| 409 | Conflict — e.g., duplicate username/email | `{ "error": "username or email already exists" }` |
| 429 | Rate limit exceeded | `{ "error": "too many requests, please try again later" }` |
| 500 | Internal server error | `{ "error": "<generic message>" }` |

> Error responses deliberately omit database details, stack traces, or field names to prevent information leakage.

---

## Security Headers on All Responses

```
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: camera=(), microphone=(), geolocation=()
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'
```

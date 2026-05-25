# Report Notes — CCS6344 Database & Cloud Security

**Assignment: Database Security Implementation**
**Application: SecureVault Password Manager**
**Group: 17**

These notes are structured to match the assignment rubric sections. Use them as the basis for the written report.

---

## 1. Introduction and Objectives

SecureVault is a web-based password manager developed as the Group 17 project for CCS6344 Database & Cloud Security. The application allows users to securely store, retrieve, and manage their login credentials for various online services.

**Primary objectives:**
- Demonstrate SQL database security principles in a working application
- Implement encryption at the application layer to protect sensitive data at rest
- Apply secure authentication and role-based access control
- Maintain an immutable audit trail of all security-relevant actions
- Illustrate internal and external attack protections through code and database configuration

**Scope — included features:**
- User registration and login with JWT session management
- Encrypted vault for storing website credentials (username + password + notes)
- Password generator using cryptographically secure randomness
- Password strength indicator
- Admin audit log viewer

**Scope — excluded (out of scope):**
Browser extension, autofill, mobile app, MFA, email verification, password sharing, cloud sync.

---

## 2. System Architecture

SecureVault uses a three-tier architecture:

```
Presentation Tier   →   Logic Tier      →   Data Tier
React + Vite            Go + Gin             PostgreSQL 16
(Port 5173 dev /        (Port 8080)          (Port 5432)
 Port 80 prod)
```

The frontend is a Single-Page Application (SPA) that communicates with the backend exclusively through a REST API. The backend enforces all security logic before any database interaction. The database stores only hashed passwords and ciphertext — it never holds plaintext secrets.

**Component responsibilities:**

| Component | Responsibility |
|---|---|
| React SPA | User interface, client-side search, password generation |
| Vite dev proxy / Nginx | Forwards `/api/` requests to backend |
| Gin (HTTP framework) | Request routing, middleware chain, JSON serialization |
| Config package | Loads and validates all environment variables at startup |
| Auth service | bcrypt hashing, JWT creation, dummy comparison (anti-enumeration) |
| Vault service | AES-256-GCM encrypt/decrypt, ownership enforcement |
| Audit service | Append-only writes to `audit_logs` table |
| Admin service | Paginated audit log reads |
| PostgreSQL | Persistent storage with enforced schema constraints |

---

## 3. Technology Stack and Justification

| Technology | Version | Justification |
|---|---|---|
| Go + Gin | 1.21 / 1.9 | Compiled, statically typed, fast; excellent standard library for crypto |
| PostgreSQL | 16 | Mature RDBMS with UUID support, row-level security, fine-grained GRANT/REVOKE |
| React + Vite | 18 / 5 | Component-based UI; Vite provides fast dev proxy for local API calls |
| TailwindCSS | 3 | Utility-first CSS; no external UI library dependencies |
| JWT (HS256) | — | Stateless auth; tokens signed with server secret |
| bcrypt | cost 12 | Standard password hashing; cost 12 balances security and UX latency |
| AES-256-GCM | — | Authenticated encryption; provides both confidentiality and integrity |
| Docker Compose | — | Reproducible local environment; one command to run all services |

**Why `database/sql` instead of an ORM:**
Using raw SQL with explicit `$1`/`$2` parameterized placeholders makes the SQL injection prevention immediately visible and auditable. An ORM hides query construction and makes it harder to verify the security property.

**Why AES-256-GCM at the application layer instead of `pgcrypto`:**
Encrypting in the Go process means the PostgreSQL server never receives the plaintext or the encryption key. In a multi-tier deployment, the database server could be on a different host — a DB dump alone yields only ciphertext. GCM mode also provides an authentication tag, detecting any tampering.

---

## 4. Database Design

### Entity-Relationship Summary

```
users 1────< vault_entries
users 1────< audit_logs
```

### Table Definitions

**users**
```sql
id            UUID PRIMARY KEY DEFAULT uuid_generate_v4()
username      VARCHAR(50) UNIQUE NOT NULL
email         VARCHAR(255) UNIQUE NOT NULL
password_hash TEXT NOT NULL                    -- bcrypt hash
role          VARCHAR(20) NOT NULL DEFAULT 'user'
              CHECK (role IN ('user', 'admin'))
created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

**vault_entries**
```sql
id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4()
user_id            UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
site_name          VARCHAR(255) NOT NULL
site_url           VARCHAR(500)
vault_username     TEXT NOT NULL
encrypted_password TEXT NOT NULL    -- AES-256-GCM ciphertext
encrypted_notes    TEXT            -- AES-256-GCM ciphertext (nullable)
created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

**audit_logs**
```sql
id         UUID PRIMARY KEY DEFAULT uuid_generate_v4()
user_id    UUID REFERENCES users(id) ON DELETE SET NULL  -- preserved on deletion
action     VARCHAR(50) NOT NULL
ip_address VARCHAR(45) NOT NULL
user_agent TEXT
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

### Key Design Decisions

- `ON DELETE CASCADE` on `vault_entries`: removing a user removes all their credentials.
- `ON DELETE SET NULL` on `audit_logs.user_id`: audit history is preserved even after the user account is deleted — critical for forensic investigations.
- `role CHECK (role IN ('user', 'admin'))`: domain constraint enforced at the database layer, not just the application layer.
- `updated_at` trigger: automatically updates the timestamp on any `vault_entries` row modification without relying on the application to set it.

---

## 5. Authentication and Authorization

### Registration Flow

1. Client submits `{ username, email, password }` to `POST /api/auth/register`
2. Backend validates input (min length, email format) using Gin binding tags
3. `bcrypt.GenerateFromPassword(password, cost=12)` produces a hash
4. Hash is stored in `users.password_hash`; plaintext is never retained
5. JWT is generated and returned to the client

### Login Flow

1. Client submits `{ email, password }` to `POST /api/auth/login`
2. Backend queries `users` by email using a parameterized query
3. **If email not found:** a dummy bcrypt comparison runs to maintain constant response time (prevents timing-based email enumeration), then returns `401 invalid credentials`
4. **If email found:** `bcrypt.CompareHashAndPassword` runs
5. On success: JWT signed with HS256 is returned; `login_success` is logged
6. On failure: same `401 invalid credentials` regardless of failure reason; `login_failure` is logged

### JWT Structure

```json
{
  "user_id": "<uuid>",
  "role": "user|admin",
  "exp": "<24 hours from issue>",
  "iat": "<issue time>",
  "iss": "password-manager"
}
```

The `ValidateToken` function explicitly checks `token.Method.(*jwt.SigningMethodHMAC)` before accepting the token, blocking the algorithm-confusion attack where an attacker sends a token signed with `alg: none`.

### Role-Based Access Control

```
Public route     → no middleware
Protected route  → AuthRequired middleware (validates JWT)
Admin route      → AuthRequired + AdminRequired (checks role = "admin")
```

The middleware chain in `routes/routes.go` ensures these are applied in the correct order.

---

## 6. Encryption Implementation

### Algorithm: AES-256-GCM

- **AES-256**: 256-bit key, 128-bit block cipher — NIST-approved symmetric cipher
- **GCM (Galois/Counter Mode)**: Provides authenticated encryption — confidentiality + integrity in one pass
- **Authentication tag**: Any bit-flip in the ciphertext causes decryption to fail, detecting tampering

### Encryption Format

```
DB column contains: base64( nonce[12 bytes] || ciphertext || gcm_tag[16 bytes] )
```

The 12-byte nonce is generated fresh for every encryption call using `crypto/rand.Reader`. Nonce reuse with the same key would break GCM's security guarantee; generating a fresh nonce per call prevents this.

### Key Management

- `ENCRYPTION_KEY` is stored as a base64 string in the `.env` file
- On startup, `config.Load()` decodes it to `[]byte` and validates it is exactly 32 bytes
- The key is passed as a function parameter through the service layer — never hardcoded or logged
- The database never receives the key

### What Is Encrypted vs. Plaintext

| Field | Stored as |
|---|---|
| `vault_entries.encrypted_password` | AES-256-GCM ciphertext |
| `vault_entries.encrypted_notes` | AES-256-GCM ciphertext |
| `users.password_hash` | bcrypt hash (one-way) |
| `vault_entries.site_name` | Plaintext |
| `vault_entries.site_url` | Plaintext |
| `vault_entries.vault_username` | Plaintext |

In a production hardening step, all vault fields would be encrypted.

---

## 7. Database Security Measures

### 7.1 SQL Injection Prevention

All database interactions use parameterized queries with positional placeholders (`$1`, `$2`, …). The `lib/pq` PostgreSQL driver sends the query and parameters separately — the server never performs string concatenation.

**Example (vault ownership check):**
```go
s.db.QueryRow(
    `SELECT id, encrypted_password FROM vault_entries
     WHERE id = $1 AND user_id = $2`,
    id, userID,
)
```

There is no code path in the application where user input is interpolated into a query string. This completely eliminates SQL injection risk.

### 7.2 Least-Privilege Database User

`security.sql` creates `pm_app_user` and grants only the minimum necessary privileges:

| Table | Granted | Denied |
|---|---|---|
| `users` | SELECT, INSERT, UPDATE | DELETE |
| `vault_entries` | SELECT, INSERT, UPDATE, DELETE | — |
| `audit_logs` | SELECT, INSERT | **UPDATE, DELETE (explicitly revoked)** |

The `audit_logs` immutability is enforced at the database layer: even if the application is compromised, the attacker cannot delete log entries using the application's database credentials.

### 7.3 IDOR Prevention (Insecure Direct Object Reference)

Every SELECT, UPDATE, and DELETE on `vault_entries` includes `AND user_id = $N`. This means a user who knows another user's entry UUID cannot access, modify, or delete it. The API returns `404` (not `403`) when an entry is not found or belongs to another user — this prevents ID enumeration.

### 7.4 Indexes

```sql
CREATE INDEX idx_vault_entries_user_id ON vault_entries(user_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE UNIQUE INDEX idx_users_email_lower ON users(LOWER(email));
```

- The `vault_entries` index ensures O(log n) lookup per authenticated request instead of a full table scan.
- The case-insensitive email index supports `WHERE LOWER(email) = LOWER($1)` efficiently.

### 7.5 Row-Level Security (RLS)

`security.sql` documents how to enable PostgreSQL Row-Level Security on `vault_entries`:

```sql
ALTER TABLE vault_entries ENABLE ROW LEVEL SECURITY;
CREATE POLICY vault_owner_policy ON vault_entries
  USING (user_id = current_setting('app.current_user_id')::uuid);
```

This would enforce ownership at the database layer in addition to the application layer, providing defense-in-depth. Not enabled in the current implementation (ownership is enforced in the service layer).

### 7.6 External Attack Protection

| Attack | Protection |
|---|---|
| SQL injection | Parameterized queries |
| Brute-force | Rate limiting (5 req/60s per IP), bcrypt cost 12 |
| Session hijacking | JWT expiry (24h), HTTPS in production |
| Clickjacking | `X-Frame-Options: DENY` header |
| MIME sniffing | `X-Content-Type-Options: nosniff` |
| XSS | Content-Security-Policy header |
| CSRF | No cookies used; JWT in Authorization header |
| IDOR | `AND user_id = $N` on all vault queries |

### 7.7 Internal Attack Protection

| Threat | Protection |
|---|---|
| DBA reads vault passwords | AES-256-GCM encryption — ciphertext only in DB |
| DBA reads user passwords | bcrypt one-way hash |
| Insider deletes audit logs | `pm_app_user` has no DELETE on `audit_logs` |
| Insider modifies audit logs | `pm_app_user` has no UPDATE on `audit_logs` |
| Rogue admin reads vault | Vault entries are per-user; admin role only grants audit log access |

---

## 8. Audit Logging

All security-relevant actions are recorded in `audit_logs` with the user ID (nullable for unauthenticated events), action string, client IP, and user agent.

**Logged actions:**

| Action Constant | Trigger |
|---|---|
| `register` | Successful user registration |
| `login_success` | Successful login |
| `login_failure` | Failed login (wrong password or unknown email) |
| `add_credential` | New vault entry created |
| `update_credential` | Vault entry updated |
| `delete_credential` | Vault entry deleted |
| `view_credential` | Single vault entry fetched (GET /api/vault/:id) |

**Design decisions:**
- Audit errors are logged server-side but never fail the primary operation — a DB write error on the audit table must not block a legitimate login.
- `user_id` is nullable: failed logins before identity is established record NULL user_id but still capture IP and user agent.
- The admin can view all audit logs paginated at `GET /api/admin/audit-logs`.

---

## 9. Implementation Steps (for Screenshot Section)

Follow these steps in order to produce the report screenshots:

1. Run `docker compose up --build` — verify all containers start successfully
2. Open `http://localhost` in Chrome with DevTools open (Network tab)
3. Navigate to `/register` — register a new user (e.g., `alice@example.com`)
4. Check Network tab: 201 response, JWT token in response body
5. Check Application tab → Local Storage: `token` and `user` keys set
6. Navigate to `/dashboard` — empty vault
7. Click **Add Credential** → fill form → click **Show password generator** → generate a password → click **Use this password** → click **Save**
8. Return to dashboard — new card appears
9. Click **View** on the card — detail page shows; click **Show** to reveal password; click **Copy**
10. Click **Edit** — modify site name or notes — click **Update Credential**
11. Click **Delete** on the card — confirm deletion modal — confirm — card disappears
12. Click **Add Credential** again — add two more credentials (for database testing below)
13. Log out
14. Login as `admin@example.com` / `AdminPass123!`
15. Click **Audit Logs** in navbar — view full log table
16. Run `docker compose exec db psql -U pmuser -d passwordmanager`
17. Run `SELECT site_name, LEFT(encrypted_password, 50) FROM vault_entries;` — show ciphertext
18. Run `SELECT username, LEFT(password_hash, 30) FROM users;` — show bcrypt hashes
19. In a new terminal, attempt 6 consecutive login requests — 6th returns `429 Too Many Requests`
20. In browser DevTools, inspect any API response headers — show security headers

---

## 10. CRUD Testing Steps

To satisfy the "insert/delete/insert testing" rubric requirement:

### Insert (Create)

```bash
# Register and get token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"testuser@example.com","password":"TestPass123!"}' | jq -r .token)

# Insert credential 1
curl -s -X POST http://localhost:8080/api/vault \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"site_name":"GitHub","site_url":"https://github.com","vault_username":"alice","password":"gh_abc123!"}' | jq .

# Insert credential 2
curl -s -X POST http://localhost:8080/api/vault \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"site_name":"Gmail","site_url":"https://gmail.com","vault_username":"alice@gmail.com","password":"gmail_xyz789!","notes":"Personal email"}' | jq .
```

Verify in database:
```sql
SELECT site_name, LEFT(encrypted_password, 40) || '...' AS ciphertext FROM vault_entries;
```

### Delete

```bash
# Get the ID of the first entry
ENTRY_ID=$(curl -s http://localhost:8080/api/vault \
  -H "Authorization: Bearer $TOKEN" | jq -r '.entries[0].id')

# Delete it
curl -s -X DELETE http://localhost:8080/api/vault/$ENTRY_ID \
  -H "Authorization: Bearer $TOKEN" -v
# Expect: HTTP 204 No Content
```

Verify in database:
```sql
SELECT COUNT(*) FROM vault_entries;  -- should be 1 now
SELECT action FROM audit_logs WHERE action = 'delete_credential';
```

### Insert again

```bash
curl -s -X POST http://localhost:8080/api/vault \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"site_name":"LinkedIn","vault_username":"alice@work.com","password":"li_secure999!"}' | jq .
```

Verify in database:
```sql
SELECT site_name, created_at FROM vault_entries ORDER BY created_at;
```

---

## 11. STRIDE and DREAD Summary

See `docs/threat-model.md` for the full tables.

**Highest-risk threats identified:**

| Threat | DREAD Total | Mitigation |
|---|---|---|
| IDOR — cross-user vault access | 37 (High) | `AND user_id = $N` on all vault queries |
| Brute-force login | 35 (Medium-High) | Rate limiting + bcrypt cost 12 |
| XSS → token theft | 34 (Medium) | CSP header; documented residual risk |
| Login DoS | 34 (Medium) | IP rate limiter with Retry-After |

The overall attack surface is well-contained: the application has no file upload, no shell execution, no external service calls, and all SQL is parameterized.

---

## 12. PDPA 2010 Notes (Malaysia Personal Data Protection Act)

The Personal Data Protection Act 2010 (PDPA 2010) governs the processing of personal data in Malaysia. The following principles apply to SecureVault:

### Data Collected

| Data Type | Stored | Classification |
|---|---|---|
| Username | Yes | Personal data |
| Email address | Yes | Personal data |
| Password hash | Yes | Not plaintext; protected |
| Website credentials | Yes (encrypted) | Sensitive personal data |
| IP address | Yes (audit logs) | Personal data |
| User agent | Yes (audit logs) | Personal data |

### PDPA Principle Compliance

**General Principle (Section 6):**
Data is collected only for the purpose of providing the password management service. No data is shared with third parties.

**Security Principle (Section 9):**
- Passwords hashed with bcrypt (one-way, salted)
- Vault passwords encrypted with AES-256-GCM (authenticated encryption)
- Access controlled by JWT authentication
- Role-based access control limits admin visibility
- Audit logs maintained for accountability

**Retention Principle (Section 10):**
Data is retained as long as the user account exists. Deleting the account cascades to delete all vault entries. Audit logs retain the action but set `user_id` to NULL — the data subject can no longer be identified.

**Integrity Principle (Section 11):**
Parameterized queries prevent SQL injection from corrupting the database. GCM authentication tags detect tampering with stored ciphertext.

**Data Subject Rights:**
Users can view, update, and delete their own credentials through the application. Account deletion (not implemented in this version) would cascade to remove all personal data except anonymous audit log entries.

### Recommendations for Full PDPA Compliance

1. Implement account deletion with full data erasure
2. Provide a data export function (right of access)
3. Display a privacy notice on the registration page
4. Log all data export and deletion events
5. Encrypt all vault fields (not just password and notes)
6. Store audit logs on a separate system with restricted access

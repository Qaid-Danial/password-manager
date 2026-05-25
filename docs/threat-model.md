# Threat Model — SecureVault Password Manager

**CCS6344 Database & Cloud Security | Group 17**

---

## System Data Flow

```
  Browser (React SPA)
       |
       |  HTTPS (TLS in production)
       |  JWT Bearer token in Authorization header
       v
  Nginx / Vite Dev Server
       |
       |  HTTP proxy  /api/*
       v
  Go + Gin Backend (Port 8080)
       |
       |  [1] JWT validated — AuthRequired middleware
       |  [2] Ownership enforced — AND user_id = $N
       |  [3] AES-256-GCM decrypt/encrypt
       |  [4] Parameterized SQL query
       v
  PostgreSQL (Port 5432)
       |
       |  Stores bcrypt hashes (users.password_hash)
       |  Stores AES-256-GCM ciphertext (vault_entries.encrypted_password)
       |  Stores plaintext metadata only (site_name, site_url, vault_username)
       v
  pgdata volume (encrypted disk recommended in production)
```

**Key security boundaries:**
- Encryption key lives in the Go process only — PostgreSQL never sees plaintext vault passwords.
- JWT secret lives in the Go process only — clients cannot forge tokens.
- `pm_app_user` DB role has no DELETE on `audit_logs` — even a compromised app cannot erase its tracks.

---

## STRIDE Threat Table

| # | Category | Threat Description | Target Asset | Mitigation Implemented |
|---|---|---|---|---|
| S1 | **Spoofing** | Attacker forges a JWT to impersonate another user | `/api/vault`, `/api/admin` | JWT signed with HS256; explicit signing-method check in `ValidateToken` prevents alg:none attack; 24-hour expiry |
| S2 | **Spoofing** | Attacker guesses/brute-forces another user's login password | `POST /api/auth/login` | bcrypt cost 12 (~250 ms/attempt); IP-based rate limit (5 req/60 s); generic error message prevents user enumeration |
| T1 | **Tampering** | Authenticated user reads or modifies another user's vault entry (IDOR) | `GET/PUT/DELETE /api/vault/:id` | Every vault query appends `AND user_id = $N`; returns 404 (not 403) to avoid confirming entry existence |
| T2 | **Tampering** | Attacker modifies ciphertext in the database to corrupt stored passwords | `vault_entries.encrypted_password` | AES-256-GCM authentication tag detects any bit-flip; decryption fails and returns error |
| R1 | **Repudiation** | Malicious insider denies performing a login or credential operation | Audit trail | `audit_logs` records every significant action with IP and user agent; app role cannot UPDATE or DELETE audit rows |
| I1 | **Information Disclosure** | Database server is fully compromised — attacker dumps all tables | `vault_entries`, `users` | `password_hash` is bcrypt (one-way); `encrypted_password` and `encrypted_notes` are AES-256-GCM ciphertext; key is never stored in the DB |
| I2 | **Information Disclosure** | Error messages leak database schema, table names, or query structure | All API error responses | Generic error messages returned to clients; detailed errors logged server-side only |
| I3 | **Information Disclosure** | JWT stored in localStorage is read by malicious JavaScript (XSS) | `localStorage` | Content-Security-Policy header restricts script sources; documented residual risk (see below) |
| D1 | **Denial of Service** | Attacker floods the login endpoint to lock out legitimate users | `POST /api/auth/login` | Token-bucket rate limiter per client IP; returns 429 with `Retry-After: 60` header |
| E1 | **Elevation of Privilege** | Regular user accesses admin-only audit log endpoint | `GET /api/admin/audit-logs` | `AdminRequired` middleware reads role from JWT claims; returns 403 if role ≠ "admin" |
| E2 | **Elevation of Privilege** | SQL injection in a vault field escalates DB privileges | All parameterized queries | `database/sql` with `$1`/`$2` placeholders; pq driver never interpolates user input into query strings |

---

## DREAD Scoring Table

Scores are 1 (lowest) to 10 (highest).  
**Risk Level:** Low ≤ 25 | Medium 26–35 | High ≥ 36

| # | Threat | Damage (D) | Reproducibility (R) | Exploitability (E) | Affected Users (A) | Discoverability (Di) | Total /50 | Risk |
|---|---|:---:|:---:|:---:|:---:|:---:|:---:|---|
| S1 | JWT forgery / alg:none | 9 | 3 | 4 | 10 | 5 | **31** | Medium |
| S2 | Brute-force login | 7 | 8 | 7 | 8 | 5 | **35** | Medium |
| T1 | IDOR — cross-user vault access | 8 | 7 | 6 | 10 | 6 | **37** | **High** |
| T2 | GCM ciphertext tampering | 6 | 4 | 3 | 10 | 3 | **26** | Medium |
| R1 | Repudiation — deny an action | 5 | 7 | 8 | 10 | 3 | **33** | Medium |
| I1 | Database dump — password exposure | 10 | 2 | 3 | 10 | 2 | **27** | Medium |
| I2 | Error message information leakage | 4 | 8 | 8 | 5 | 7 | **32** | Medium |
| I3 | XSS → localStorage token theft | 8 | 5 | 5 | 10 | 6 | **34** | Medium |
| D1 | Login endpoint DoS / flooding | 6 | 8 | 7 | 8 | 5 | **34** | Medium |
| E1 | Privilege escalation to admin | 9 | 3 | 4 | 10 | 5 | **31** | Medium |
| E2 | SQL injection | 10 | 2 | 2 | 10 | 4 | **28** | Medium |

### Scoring Justifications

**T1 — IDOR (Highest risk, 37)**
Rated High because it is reproducible (any authenticated user can try another user's UUID), easy to exploit with a simple curl command, and affects all users' stored credentials. Mitigated by `AND user_id = $N` on every vault query.

**S2 — Brute-force login (35)**
High reproducibility because automation tools can send thousands of requests. Mitigated by bcrypt cost 12 (makes each guess expensive) and the rate limiter.

**I3 — XSS to token theft (34)**
Medium-High because a successful XSS steals the JWT from localStorage and grants full account access. Mitigated by a strict CSP that blocks inline scripts and unknown sources. Residual risk documented below.

**D1 — Login DoS (34)**
High reproducibility from botnets. Rate limiter reduces but does not eliminate the risk (in-memory store resets on restart).

**I1 — DB dump (27)**
Damage is catastrophic (10) but exploitability is low (3) because it requires database-level access, not just application-level. AES-256-GCM encryption makes a raw dump useless without the application key.

**E2 — SQL injection (28)**
Damage would be catastrophic but exploitability is very low (2) because all queries use parameterized placeholders — there is no code path where user input is concatenated into a query string.

---

## Residual Risks

These risks are acknowledged but not fully mitigated in the current implementation:

| Risk | Reason Not Fully Mitigated | Recommended Hardening |
|---|---|---|
| JWT in localStorage (XSS-readable) | httpOnly cookies would prevent JS access but require same-site CSRF tokens | Use httpOnly + SameSite=Strict cookies in production |
| In-memory rate limiter resets on restart | Restart clears all IP counters | Use Redis-backed rate limiter for multi-instance deployments |
| No JWT revocation | Logout only removes the client-side token; server cannot invalidate it | Implement a token blocklist (Redis set of revoked JIDs) |
| Plaintext metadata in DB | `site_name`, `site_url`, and `vault_username` are not encrypted | Encrypt all vault fields, not just password and notes |
| No TLS in dev Docker setup | Traffic between browser and backend is HTTP | Terminate TLS at the nginx reverse proxy in production |

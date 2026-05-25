# Screenshots Checklist — CCS6344 Report

**Group 17 | SecureVault Password Manager**

Take these screenshots in order. Each item describes the exact application state required.

---

## Setup & Infrastructure

- [ ] **SS01 — Docker Compose running**
  - Command: `docker compose up --build`
  - Show the terminal output with all three containers healthy
  - Optional: run `docker ps` to show container names and ports in a clean table

- [ ] **SS02 — Database tables created**
  - Command: `docker compose exec db psql -U pmuser -d passwordmanager`
  - Inside psql, run `\dt` to show all tables
  - Expected: `audit_logs`, `users`, `vault_entries` listed

- [ ] **SS03 — Schema details**
  - Inside psql, run `\d users` and `\d vault_entries`
  - Show column names and types; highlight `password_hash TEXT` and `encrypted_password TEXT`

---

## User Registration

- [ ] **SS04 — Register page**
  - Navigate to `http://localhost:5173/register`
  - Fill in username, email, and a weak password (e.g., `abc`)
  - Show the password strength bar at **Very Weak** (red)

- [ ] **SS05 — Password strength — Strong**
  - Same register form
  - Type a strong password like `Tr0ub4dor&3`
  - Show the strength bar at **Strong** or **Very Strong** (green)

- [ ] **SS06 — Successful registration (Network tab)**
  - Submit the registration form
  - Open DevTools → Network → click the `/api/auth/register` request
  - Show: `201 Created` status and the JSON response with `token` and `user` fields

---

## Login

- [ ] **SS07 — Login page**
  - Navigate to `http://localhost:5173/login`
  - Show the success banner from the redirect: "Account created — please sign in."

- [ ] **SS08 — Successful login (LocalStorage)**
  - Log in with the registered account
  - Open DevTools → Application → Local Storage → `http://localhost:5173`
  - Show both `token` (JWT string) and `user` (JSON object) keys present

- [ ] **SS09 — Login failure**
  - Enter a correct email but wrong password
  - Show the red error banner: "invalid credentials"
  - Open DevTools → Network → show `401 Unauthorized` response

---

## Vault Dashboard

- [ ] **SS10 — Empty dashboard**
  - After first login, show the vault dashboard with "No credentials yet" state
  - Show the navbar with username and Logout button

- [ ] **SS11 — Add credential form**
  - Click **+ Add Credential**
  - Fill in: Site Name, Site URL, Username
  - Click **Show password generator** — show the generator panel open with a generated password
  - Show the length slider and charset checkboxes

- [ ] **SS12 — Password generator in use**
  - Click **Use this password** — show the password field populated
  - Show the **Password Strength** bar under the password field

- [ ] **SS13 — Dashboard with entries**
  - After saving 3+ credentials, show the vault dashboard grid
  - Each card shows: site name, username, masked password (••••••••), Copy / View / Edit / Delete buttons

- [ ] **SS14 — Client-side search**
  - Type part of a site name in the search box
  - Show the grid filtered to only matching cards

---

## Credential Detail

- [ ] **SS15 — Credential detail — password hidden**
  - Click **View** on any vault card
  - Show the detail page with password displayed as `••••••••••••`

- [ ] **SS16 — Credential detail — password revealed**
  - Click the **Show** button
  - Show the actual plaintext password visible

- [ ] **SS17 — Delete confirmation modal**
  - Click **Delete** on a credential
  - Show the confirmation modal: "Delete credential? This action cannot be undone."

---

## Edit Credential

- [ ] **SS18 — Edit form pre-populated**
  - Click **Edit** on any credential
  - Show the form with all fields already filled in from the database

---

## Database Security Evidence

- [ ] **SS19 — Vault passwords stored as ciphertext**
  - In psql: `SELECT site_name, LEFT(encrypted_password, 60) || '...' AS ciphertext FROM vault_entries;`
  - Show that `encrypted_password` contains base64-encoded ciphertext, not the real password

- [ ] **SS20 — User passwords stored as bcrypt hashes**
  - In psql: `SELECT username, email, LEFT(password_hash, 40) || '...' AS hash FROM users;`
  - Show the `$2a$12$...` bcrypt format — not plaintext

---

## Audit Logging

- [ ] **SS21 — Audit logs page**
  - Login as `admin@example.com` / `AdminPass123!`
  - Navigate to **Audit Logs** in the navbar
  - Show the table with multiple log entries in colour-coded badges

- [ ] **SS22 — Login failure in audit log**
  - Scroll or filter to find a `login_failure` entry (red badge)
  - Show it has an IP address but `user_id` may be `—` (NULL) for unknown emails

- [ ] **SS23 — Audit logs in psql**
  - In psql: `SELECT action, ip_address, created_at FROM audit_logs ORDER BY created_at DESC LIMIT 10;`
  - Show the raw data to confirm logs are persisted in the database

---

## Security Features

- [ ] **SS24 — Rate limiting (429)**
  - Run 6 rapid login attempts in the terminal:
    ```bash
    for i in {1..6}; do
      curl -s -X POST http://localhost:8080/api/auth/login \
        -H "Content-Type: application/json" \
        -d '{"email":"x@x.com","password":"wrong"}' | jq .error
    done
    ```
  - Show the 6th response: `"too many requests, please try again later"`

- [ ] **SS25 — Security headers**
  - Open DevTools → Network → click any API response (e.g., `/api/vault`)
  - Show the Response Headers panel
  - Highlight: `X-Frame-Options`, `X-Content-Type-Options`, `Content-Security-Policy`

- [ ] **SS26 — IDOR prevention**
  - Login as `testuser` and create a credential; note its `id`
  - Login as `alice` (different account) and attempt:
    ```bash
    curl -s http://localhost:8080/api/vault/<testuser-entry-id> \
      -H "Authorization: Bearer $ALICE_TOKEN"
    ```
  - Show `404 Not Found` — entry invisible to the other user

- [ ] **SS27 — security.sql — least-privilege user**
  - In psql (as superuser): show `pm_app_user` privileges:
    ```sql
    \dp users
    \dp vault_entries
    \dp audit_logs
    ```
  - Or show the output of running `database/security.sql`

- [ ] **SS28 — Parameterized query evidence**
  - Open `backend/services/vault.go` in the editor
  - Highlight the `WHERE id = $1 AND user_id = $2` line
  - Screenshot the code to show parameterized query usage

---

## Summary Table for Report

| Screenshot | Section in Report | Purpose |
|---|---|---|
| SS01–SS03 | Setup / System Design | Infrastructure running |
| SS04–SS06 | Implementation Steps | Registration flow |
| SS07–SS09 | Implementation Steps | Login flow |
| SS10–SS18 | Implementation Steps | CRUD operations |
| SS19–SS20 | Database Security | Encryption at rest |
| SS21–SS23 | Audit Logging | Security audit trail |
| SS24–SS28 | Security Measures | Attack mitigations |

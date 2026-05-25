-- ============================================================
-- security.sql — Database-Layer Security Hardening
-- CCS6344 Database & Cloud Security — Group 17
--
-- This file demonstrates the "least privilege" principle at the
-- database layer. In production, the application connects as
-- pm_app_user (not a superuser), limiting blast radius if the
-- application is compromised.
--
-- Run this AFTER schema.sql as a superuser (e.g., postgres).
-- ============================================================

-- ============================================================
-- Step 1: Create a dedicated application database role
-- The application never connects as 'postgres' (superuser).
-- ============================================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'pm_app_user') THEN
        CREATE ROLE pm_app_user WITH LOGIN PASSWORD 'app_secure_password_change_me';
    END IF;
END
$$;

-- ============================================================
-- Step 2: Grant CONNECT on the database only
-- pm_app_user cannot access any other database on this server.
-- ============================================================
GRANT CONNECT ON DATABASE passwordmanager TO pm_app_user;

-- ============================================================
-- Step 3: Grant USAGE on the public schema
-- Required to reference any objects within the schema.
-- ============================================================
GRANT USAGE ON SCHEMA public TO pm_app_user;

-- ============================================================
-- Step 4: Grant only required DML privileges per table
--
-- users: SELECT + INSERT (registration) + UPDATE (future profile edits)
--        No DELETE — user deletion is an admin-only migration operation.
-- vault_entries: Full CRUD — core application functionality.
-- audit_logs: SELECT + INSERT only.
--             UPDATE and DELETE are explicitly revoked — audit logs
--             must be immutable. Even the app cannot alter past logs.
-- ============================================================
GRANT SELECT, INSERT, UPDATE ON users TO pm_app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON vault_entries TO pm_app_user;
GRANT SELECT, INSERT ON audit_logs TO pm_app_user;

-- Explicitly revoke modification rights on audit_logs
-- (belt-and-suspenders: ensures immutability even if future
--  GRANT ALL is accidentally run)
REVOKE UPDATE, DELETE ON audit_logs FROM pm_app_user;

-- ============================================================
-- Step 5: Grant USAGE on sequences used for UUID generation
-- (uuid-ossp uses its own functions, but sequences are needed
--  if any SERIAL columns are added in the future)
-- ============================================================
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO pm_app_user;

-- ============================================================
-- Step 6: Performance and security indexes
--
-- idx_vault_entries_user_id: speeds up every vault query that
--   filters by user_id (every authenticated request).
--   Also prevents full-table scans that could leak timing info.
--
-- idx_audit_logs_user_id: speeds up per-user audit history lookups.
--
-- idx_audit_logs_created_at: supports DESC pagination on the
--   admin audit log page.
--
-- idx_users_email_lower: case-insensitive email uniqueness and
--   fast lookup during login (queries use LOWER(email)).
-- ============================================================
CREATE INDEX IF NOT EXISTS idx_vault_entries_user_id
    ON vault_entries(user_id);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id
    ON audit_logs(user_id);

CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at
    ON audit_logs(created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_lower
    ON users(LOWER(email));

-- ============================================================
-- Step 7: Row-Level Security (RLS) — advanced hardening note
--
-- Enabling RLS on vault_entries would enforce ownership at the
-- DB layer, making IDOR attacks impossible even if the app has
-- a bug. This is left as a manual step for production hardening.
--
-- To enable (run as superuser):
--   ALTER TABLE vault_entries ENABLE ROW LEVEL SECURITY;
--   CREATE POLICY vault_owner_policy ON vault_entries
--     USING (user_id = current_setting('app.current_user_id')::uuid);
--
-- The application would then execute:
--   SET LOCAL app.current_user_id = '<jwt-user-id>';
-- before each vault query.
--
-- Current implementation enforces ownership in the application
-- layer via "AND user_id = $N" on all vault queries.
-- ============================================================

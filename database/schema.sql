-- ============================================================
-- schema.sql — Password Manager Database Schema
-- CCS6344 Database & Cloud Security — Group 17
-- ============================================================

-- Extensions required by the schema
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";  -- UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";   -- bcrypt for seed data

-- ============================================================
-- Table: users
-- Stores authenticated user accounts.
-- password_hash is a bcrypt hash (cost 12). Plaintext passwords
-- are NEVER stored — the application hashes before INSERT.
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    username      VARCHAR(50)  UNIQUE NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT         NOT NULL,
    role          VARCHAR(20)  NOT NULL DEFAULT 'user'
                               CHECK (role IN ('user', 'admin')),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ============================================================
-- Table: vault_entries
-- Stores encrypted credentials per user.
-- encrypted_password and encrypted_notes contain AES-256-GCM
-- ciphertext produced by the application layer before INSERT.
-- A full DB dump reveals only ciphertext — no plaintext secrets.
-- ON DELETE CASCADE removes all entries when the user is deleted.
-- ============================================================
CREATE TABLE IF NOT EXISTS vault_entries (
    id                 UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id            UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    site_name          VARCHAR(255) NOT NULL,
    site_url           VARCHAR(500),
    vault_username     TEXT         NOT NULL,
    encrypted_password TEXT         NOT NULL,
    encrypted_notes    TEXT,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ============================================================
-- Table: audit_logs
-- Append-only record of all security-relevant actions.
-- ON DELETE SET NULL preserves log history even after the user
-- account is removed — critical for forensic investigations.
-- ============================================================
CREATE TABLE IF NOT EXISTS audit_logs (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID        REFERENCES users(id) ON DELETE SET NULL,
    action     VARCHAR(50) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- Trigger: auto-update updated_at on vault_entries
-- ============================================================
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS vault_entries_updated_at ON vault_entries;
CREATE TRIGGER vault_entries_updated_at
    BEFORE UPDATE ON vault_entries
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

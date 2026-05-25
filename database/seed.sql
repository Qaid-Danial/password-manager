-- ============================================================
-- seed.sql — Development Seed Data
-- CCS6344 Database & Cloud Security — Group 17
-- WARNING: FOR DEVELOPMENT ONLY. Never use in production.
-- ============================================================
-- Passwords are hashed using pgcrypto's bcrypt implementation
-- (blowfish, cost 12) which is interoperable with Go's
-- x/crypto/bcrypt for ASCII passwords.
--
-- Default accounts:
--   admin    / admin@example.com  / AdminPass123!   (role: admin)
--   testuser / test@example.com   / TestPass123!    (role: user)
-- ============================================================

INSERT INTO users (id, username, email, password_hash, role)
VALUES (
    uuid_generate_v4(),
    'admin',
    'admin@example.com',
    crypt('AdminPass123!', gen_salt('bf', 12)),
    'admin'
) ON CONFLICT (username) DO NOTHING;

INSERT INTO users (id, username, email, password_hash, role)
VALUES (
    uuid_generate_v4(),
    'testuser',
    'test@example.com',
    crypt('TestPass123!', gen_salt('bf', 12)),
    'user'
) ON CONFLICT (username) DO NOTHING;

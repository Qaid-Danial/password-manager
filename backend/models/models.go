package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User mirrors the users table. PasswordHash is tagged json:"-" so it is
// never serialized into any API response.
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

// VaultEntry mirrors the vault_entries table (DB-layer struct).
// encrypted_password and encrypted_notes hold AES-256-GCM ciphertext.
type VaultEntry struct {
	ID                string
	UserID            string
	SiteName          string
	SiteURL           string
	VaultUsername     string
	EncryptedPassword string
	EncryptedNotes    string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// VaultEntryResponse is the decrypted view returned to the API client.
// It must never be written to the database.
type VaultEntryResponse struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	SiteName      string    `json:"site_name"`
	SiteURL       string    `json:"site_url"`
	VaultUsername string    `json:"vault_username"`
	Password      string    `json:"password"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// AuditLog mirrors the audit_logs table.
type AuditLog struct {
	ID        string    `json:"id"`
	UserID    *string   `json:"user_id"`
	Action    string    `json:"action"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

// --- Request / response body structs ---

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type VaultEntryRequest struct {
	SiteName      string `json:"site_name"      binding:"required,max=255"`
	SiteURL       string `json:"site_url"       binding:"omitempty,max=500"`
	VaultUsername string `json:"vault_username" binding:"required"`
	Password      string `json:"password"       binding:"required"`
	Notes         string `json:"notes"`
}

// JWTClaims extends the standard JWT claims with application-specific fields.
type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

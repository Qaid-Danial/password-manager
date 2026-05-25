package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Qaid-Danial/password-manager/backend/config"
	"github.com/Qaid-Danial/password-manager/backend/models"
	"github.com/Qaid-Danial/password-manager/backend/utils"
)

type VaultService struct {
	db       *sql.DB
	config   *config.Config
	auditSvc *AuditService
}

func NewVaultService(db *sql.DB, cfg *config.Config, auditSvc *AuditService) *VaultService {
	return &VaultService{db: db, config: cfg, auditSvc: auditSvc}
}

func (s *VaultService) List(userID string) ([]models.VaultEntryResponse, error) {
	rows, err := s.db.Query(
		`SELECT id, user_id, site_name, site_url, vault_username,
		        encrypted_password, encrypted_notes, created_at, updated_at
		 FROM vault_entries
		 WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch vault entries")
	}
	defer rows.Close()

	entries := []models.VaultEntryResponse{}
	for rows.Next() {
		var (
			id, uid, siteName, vaultUsername, encPassword string
			siteURL, encNotes                             sql.NullString
			createdAt, updatedAt                          time.Time
		)
		if err := rows.Scan(&id, &uid, &siteName, &siteURL, &vaultUsername,
			&encPassword, &encNotes, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to read vault entries")
		}

		password, err := utils.Decrypt(encPassword, s.config.EncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt entry")
		}

		notes := ""
		if encNotes.Valid && encNotes.String != "" {
			notes, err = utils.Decrypt(encNotes.String, s.config.EncryptionKey)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt entry")
			}
		}

		entries = append(entries, models.VaultEntryResponse{
			ID:            id,
			UserID:        uid,
			SiteName:      siteName,
			SiteURL:       siteURL.String,
			VaultUsername: vaultUsername,
			Password:      password,
			Notes:         notes,
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
		})
	}

	return entries, rows.Err()
}

func (s *VaultService) GetByID(id, userID, ip, userAgent string) (*models.VaultEntryResponse, error) {
	var (
		rid, uid, siteName, vaultUsername, encPassword string
		siteURL, encNotes                              sql.NullString
		createdAt, updatedAt                           time.Time
	)

	// AND user_id = $2 is the critical IDOR prevention — a user cannot read
	// another user's entry even if they know the UUID.
	err := s.db.QueryRow(
		`SELECT id, user_id, site_name, site_url, vault_username,
		        encrypted_password, encrypted_notes, created_at, updated_at
		 FROM vault_entries
		 WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&rid, &uid, &siteName, &siteURL, &vaultUsername,
		&encPassword, &encNotes, &createdAt, &updatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // caller returns 404
		}
		return nil, fmt.Errorf("failed to fetch vault entry")
	}

	password, err := utils.Decrypt(encPassword, s.config.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt entry")
	}

	notes := ""
	if encNotes.Valid && encNotes.String != "" {
		notes, err = utils.Decrypt(encNotes.String, s.config.EncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt entry")
		}
	}

	s.auditSvc.Log(&userID, ActionViewCredential, ip, userAgent)

	return &models.VaultEntryResponse{
		ID:            rid,
		UserID:        uid,
		SiteName:      siteName,
		SiteURL:       siteURL.String,
		VaultUsername: vaultUsername,
		Password:      password,
		Notes:         notes,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}, nil
}

func (s *VaultService) Create(userID string, req models.VaultEntryRequest, ip, userAgent string) (*models.VaultEntryResponse, error) {
	encPassword, err := utils.Encrypt(req.Password, s.config.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password")
	}

	var encNotes *string
	if req.Notes != "" {
		n, err := utils.Encrypt(req.Notes, s.config.EncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt notes")
		}
		encNotes = &n
	}

	var (
		id, uid, siteName, vaultUsername string
		siteURL                          sql.NullString
		createdAt, updatedAt             time.Time
	)

	err = s.db.QueryRow(
		`INSERT INTO vault_entries
		    (id, user_id, site_name, site_url, vault_username, encrypted_password, encrypted_notes)
		 VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6)
		 RETURNING id, user_id, site_name, site_url, vault_username, created_at, updated_at`,
		userID, req.SiteName, nullableString(req.SiteURL), req.VaultUsername, encPassword, encNotes,
	).Scan(&id, &uid, &siteName, &siteURL, &vaultUsername, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault entry")
	}

	s.auditSvc.Log(&userID, ActionAddCredential, ip, userAgent)

	return &models.VaultEntryResponse{
		ID:            id,
		UserID:        uid,
		SiteName:      siteName,
		SiteURL:       siteURL.String,
		VaultUsername: vaultUsername,
		Password:      req.Password,
		Notes:         req.Notes,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}, nil
}

func (s *VaultService) Update(id, userID string, req models.VaultEntryRequest, ip, userAgent string) (*models.VaultEntryResponse, error) {
	encPassword, err := utils.Encrypt(req.Password, s.config.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password")
	}

	var encNotes *string
	if req.Notes != "" {
		n, err := utils.Encrypt(req.Notes, s.config.EncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt notes")
		}
		encNotes = &n
	}

	var (
		rid, uid, siteName, vaultUsername string
		siteURL                           sql.NullString
		createdAt, updatedAt              time.Time
	)

	// AND user_id = $2 prevents a user from updating another user's entry.
	err = s.db.QueryRow(
		`UPDATE vault_entries
		 SET site_name = $3, site_url = $4, vault_username = $5,
		     encrypted_password = $6, encrypted_notes = $7
		 WHERE id = $1 AND user_id = $2
		 RETURNING id, user_id, site_name, site_url, vault_username, created_at, updated_at`,
		id, userID, req.SiteName, nullableString(req.SiteURL), req.VaultUsername, encPassword, encNotes,
	).Scan(&rid, &uid, &siteName, &siteURL, &vaultUsername, &createdAt, &updatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // caller returns 404
		}
		return nil, fmt.Errorf("failed to update vault entry")
	}

	s.auditSvc.Log(&userID, ActionUpdateCredential, ip, userAgent)

	return &models.VaultEntryResponse{
		ID:            rid,
		UserID:        uid,
		SiteName:      siteName,
		SiteURL:       siteURL.String,
		VaultUsername: vaultUsername,
		Password:      req.Password,
		Notes:         req.Notes,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}, nil
}

func (s *VaultService) Delete(id, userID, ip, userAgent string) (bool, error) {
	// AND user_id = $2 prevents deleting another user's entry.
	result, err := s.db.Exec(
		`DELETE FROM vault_entries WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return false, fmt.Errorf("failed to delete vault entry")
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return false, nil
	}

	s.auditSvc.Log(&userID, ActionDeleteCredential, ip, userAgent)
	return true, nil
}

// nullableString converts an empty Go string to nil so the database driver
// inserts NULL rather than an empty string for optional text columns.
func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

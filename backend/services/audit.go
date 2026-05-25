package services

import (
	"database/sql"
	"log"
)

// Action constants for audit log entries.
const (
	ActionRegister         = "register"
	ActionLoginSuccess     = "login_success"
	ActionLoginFailure     = "login_failure"
	ActionAddCredential    = "add_credential"
	ActionUpdateCredential = "update_credential"
	ActionDeleteCredential = "delete_credential"
	ActionViewCredential   = "view_credential"
)

type AuditService struct {
	db *sql.DB
}

func NewAuditService(db *sql.DB) *AuditService {
	return &AuditService{db: db}
}

// Log writes an audit entry. userID is nil for unauthenticated events
// (e.g., login failures before identity is established).
//
// Errors are logged server-side but never propagated — an audit write
// failure must not block the primary operation or alert the caller.
func (s *AuditService) Log(userID *string, action, ipAddress, userAgent string) {
	_, err := s.db.Exec(
		`INSERT INTO audit_logs (id, user_id, action, ip_address, user_agent)
		 VALUES (uuid_generate_v4(), $1, $2, $3, $4)`,
		userID, action, ipAddress, userAgent,
	)
	if err != nil {
		log.Printf("audit log error [%s]: %v", action, err)
	}
}

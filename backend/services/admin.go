package services

import (
	"database/sql"
	"fmt"

	"github.com/Qaid-Danial/password-manager/backend/models"
)

type AdminService struct {
	db *sql.DB
}

func NewAdminService(db *sql.DB) *AdminService {
	return &AdminService{db: db}
}

// GetAuditLogs returns a paginated list of audit log entries newest-first,
// along with the total count for pagination metadata.
func (s *AdminService) GetAuditLogs(page, pageSize int) ([]models.AuditLog, int, error) {
	var total int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM audit_logs`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs")
	}

	offset := (page - 1) * pageSize
	rows, err := s.db.Query(
		`SELECT id, user_id, action, ip_address, user_agent, created_at
		 FROM audit_logs
		 ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`,
		pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch audit logs")
	}
	defer rows.Close()

	logs := []models.AuditLog{}
	for rows.Next() {
		var l models.AuditLog
		var userID, userAgent sql.NullString
		if err := rows.Scan(&l.ID, &userID, &l.Action, &l.IPAddress, &userAgent, &l.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to read audit logs")
		}
		if userID.Valid {
			l.UserID = &userID.String
		}
		if userAgent.Valid {
			l.UserAgent = userAgent.String
		}
		logs = append(logs, l)
	}

	return logs, total, rows.Err()
}

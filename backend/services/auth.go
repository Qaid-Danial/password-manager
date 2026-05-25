package services

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/Qaid-Danial/password-manager/backend/config"
	"github.com/Qaid-Danial/password-manager/backend/models"
	"github.com/Qaid-Danial/password-manager/backend/utils"
)

type AuthService struct {
	db       *sql.DB
	config   *config.Config
	auditSvc *AuditService
}

func NewAuthService(db *sql.DB, cfg *config.Config, auditSvc *AuditService) *AuthService {
	return &AuthService{db: db, config: cfg, auditSvc: auditSvc}
}

func (s *AuthService) Register(req models.RegisterRequest, ip, userAgent string) (string, *models.User, error) {
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return "", nil, fmt.Errorf("registration failed")
	}

	var user models.User
	err = s.db.QueryRow(
		`INSERT INTO users (id, username, email, password_hash, role)
		 VALUES (uuid_generate_v4(), $1, $2, $3, 'user')
		 RETURNING id, username, email, role, created_at`,
		req.Username, req.Email, hash,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return "", nil, errors.New("username or email already exists")
		}
		return "", nil, fmt.Errorf("registration failed")
	}

	token, err := utils.GenerateToken(user.ID, user.Role, s.config.JWTSecret)
	if err != nil {
		return "", nil, fmt.Errorf("registration failed")
	}

	s.auditSvc.Log(&user.ID, ActionRegister, ip, userAgent)
	return token, &user, nil
}

func (s *AuthService) Login(req models.LoginRequest, ip, userAgent string) (string, *models.User, error) {
	var user models.User
	err := s.db.QueryRow(
		`SELECT id, username, email, password_hash, role, created_at
		 FROM users WHERE LOWER(email) = LOWER($1)`,
		req.Email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Run dummy bcrypt to prevent timing-based email enumeration.
			// Without this, a missing user returns faster than a wrong password,
			// revealing which emails are registered.
			utils.DummyPasswordCompare(req.Password)
			s.auditSvc.Log(nil, ActionLoginFailure, ip, userAgent)
			return "", nil, errors.New("invalid credentials")
		}
		return "", nil, fmt.Errorf("login failed")
	}

	if err := utils.CheckPassword(req.Password, user.PasswordHash); err != nil {
		s.auditSvc.Log(&user.ID, ActionLoginFailure, ip, userAgent)
		return "", nil, errors.New("invalid credentials")
	}

	token, err := utils.GenerateToken(user.ID, user.Role, s.config.JWTSecret)
	if err != nil {
		return "", nil, fmt.Errorf("login failed")
	}

	s.auditSvc.Log(&user.ID, ActionLoginSuccess, ip, userAgent)
	return token, &user, nil
}

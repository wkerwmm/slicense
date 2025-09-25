package service

import (
	"database/sql"
	"errors"
	"license-server/database"
	"license-server/utils"
)

type AuthService struct {
	db *sql.DB
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Register(username, email, password, passwordRepeat string) error {
	if password != passwordRepeat {
		return errors.New("şifreler uyuşmuyor")
	}

	hashed, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		INSERT INTO Accounts (username, email, password_hash)
		VALUES (?, ?, ?)`,
		username, email, hashed)
	return err
}

func (s *AuthService) Login(email, password string) (*database.Account, error) {
	row := s.db.QueryRow(`
		SELECT id, username, email, password_hash, created_at, last_login, last_login_ip
		FROM Accounts WHERE email = ?`, email)

	var acc database.Account
	err := row.Scan(&acc.ID, &acc.Username, &acc.Email, &acc.PasswordHash, &acc.CreatedAt, &acc.LastLogin, &acc.LastLoginIP)
	if err != nil {
		return nil, errors.New("e-posta veya şifre hatalı")
	}

	if !utils.CheckPasswordHash(password, acc.PasswordHash) {
		return nil, errors.New("e-posta veya şifre hatalı")
	}

	_, _ = s.db.Exec(`UPDATE Accounts SET last_login = NOW() WHERE id = ?`, acc.ID)

	return &acc, nil
}

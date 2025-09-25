package license

import (
	"errors"
	"license-server/database"
	"regexp"
	"time"
)

type Service struct {
	db *database.Database
}

func NewService(db *database.Database) *Service {
	return &Service{db: db}
}

func (s *Service) AddLicense(key, product, ownerEmail, ownerName string, expiresAt *time.Time) error {
	if !isValidLicenseKey(key) {
		return ErrInvalidLicenseKey
	}
	if !isValidEmail(ownerEmail) {
		return ErrInvalidEmail
	}
	return s.db.AddLicense(key, product, expiresAt, ownerEmail, ownerName)
}

func (s *Service) GetLicense(key, product string) (*database.License, error) {
	return s.db.GetLicense(key, product)
}

func (s *Service) DeleteLicense(key, product string) error {
	return s.db.DeleteLicense(key, product)
}

func (s *Service) ListLicenses(product string) ([]database.License, error) {
	return s.db.ListLicenses(product)
}

func (s *Service) GetAuditLogs(limit int) ([]database.AuditLog, error) {
	return s.db.GetAuditLogs(limit)
}

func isValidLicenseKey(key string) bool {
	match, _ := regexp.MatchString(`^[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$`, key)
	return match
}

func isValidEmail(email string) bool {
	match, _ := regexp.MatchString(`^[^@]+@[^@]+\.[^@]+$`, email)
	return match
}

var (
	ErrInvalidLicenseKey = errors.New("invalid license key format")
	ErrInvalidEmail     = errors.New("invalid email format")
)
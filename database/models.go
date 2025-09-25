package database

import "time"

type License struct {
	ID          int
	Key         string
	Product     string
	ExpiresAt   *time.Time
	OwnerEmail  string
	OwnerName   string
	IsActivated bool
}

type AuditLog struct {
	Action     string
	LicenseKey string
	Product    string
	ChangedAt  time.Time
	Details    string
}

type Account struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    string
	LastLogin    *string
	LastLoginIP  *string
}

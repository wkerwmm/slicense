package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Database struct {
	db *sql.DB
}

var (
	ErrLicenseNotFound = errors.New("license not found")
	ErrDuplicateKey    = errors.New("duplicate license key")
)

func New(dsn string) (*Database, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql connection failed: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("mysql ping failed: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("table creation failed: %w", err)
	}

	return &Database{db: db}, nil
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS licenses (
			id INT AUTO_INCREMENT PRIMARY KEY,
			license_key VARCHAR(255) NOT NULL,
			product VARCHAR(255) NOT NULL,
			expires_at DATETIME NULL,
			owner_email VARCHAR(255),
			owner_name VARCHAR(255),
			is_activated BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_license_product (license_key, product)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	if err != nil {
		return fmt.Errorf("licenses table creation failed: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS Accounts (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(50) NOT NULL UNIQUE,
			email VARCHAR(100) NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_login DATETIME NULL,
			last_login_ip VARCHAR(45) NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
		if err != nil {
			return fmt.Errorf("Accounts table creation failed: %w", err)
		}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS audit_log (
			id INT AUTO_INCREMENT PRIMARY KEY,
			action VARCHAR(50) NOT NULL,
			license_key VARCHAR(255) NOT NULL,
			product VARCHAR(255) NOT NULL,
			changed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			details TEXT,
			KEY idx_license_key (license_key),
			KEY idx_changed_at (changed_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	if err != nil {
		return fmt.Errorf("audit_log table creation failed: %w", err)
	}

	return nil
}

func (d *Database) AddLicense(key, product string, expiresAt *time.Time, ownerEmail, ownerName string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("transaction begin failed: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(
		`INSERT INTO licenses 
		(license_key, product, expires_at, owner_email, owner_name) 
		VALUES (?, ?, ?, ?, ?)`,
		key, product, expiresAt, ownerEmail, ownerName,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrDuplicateKey
		}
		return fmt.Errorf("license insert failed: %w", err)
	}

	_, err = tx.Exec(
		`INSERT INTO audit_log 
		(action, license_key, product, details) 
		VALUES (?, ?, ?, ?)`,
		"ADD", key, product, fmt.Sprintf("Owner: %s (%s)", ownerName, ownerEmail),
	)
	if err != nil {
		return fmt.Errorf("audit log insert failed: %w", err)
	}

	return tx.Commit()
}

func (d *Database) GetLicense(key, product string) (*License, error) {
	row := d.db.QueryRow(
		`SELECT 
			id, license_key, product, expires_at, 
			owner_email, owner_name, is_activated 
		FROM licenses 
		WHERE license_key = ? AND product = ?`,
		key, product,
	)

	var lic License
	var expiresAt sql.NullTime
	err := row.Scan(
		&lic.ID, &lic.Key, &lic.Product, &expiresAt,
		&lic.OwnerEmail, &lic.OwnerName, &lic.IsActivated,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrLicenseNotFound
		}
		return nil, fmt.Errorf("license query failed: %w", err)
	}

	if expiresAt.Valid {
		lic.ExpiresAt = &expiresAt.Time
	}

	return &lic, nil
}

func (d *Database) DeleteLicense(key, product string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("transaction begin failed: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	result, err := tx.Exec(
		`DELETE FROM licenses 
		WHERE license_key = ? AND product = ?`,
		key, product,
	)
	if err != nil {
		return fmt.Errorf("license delete failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected check failed: %w", err)
	}
	if rowsAffected == 0 {
		return ErrLicenseNotFound
	}

	_, err = tx.Exec(
		`INSERT INTO audit_log 
		(action, license_key, product) 
		VALUES (?, ?, ?)`,
		"DELETE", key, product,
	)
	if err != nil {
		return fmt.Errorf("audit log insert failed: %w", err)
	}

	return tx.Commit()
}

func (d *Database) ListLicenses(product string) ([]License, error) {
	rows, err := d.db.Query(
		`SELECT 
			license_key, expires_at, 
			owner_email, owner_name, is_activated 
		FROM licenses 
		WHERE product = ?`,
		product,
	)
	if err != nil {
		return nil, fmt.Errorf("license query failed: %w", err)
	}
	defer rows.Close()

	var licenses []License
	for rows.Next() {
		var lic License
		var expiresAt sql.NullTime
		if err := rows.Scan(
			&lic.Key, &expiresAt,
			&lic.OwnerEmail, &lic.OwnerName, &lic.IsActivated,
		); err != nil {
			return nil, fmt.Errorf("license scan failed: %w", err)
		}
		if expiresAt.Valid {
			lic.ExpiresAt = &expiresAt.Time
		}
		licenses = append(licenses, lic)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return licenses, nil
}

func (d *Database) GetAuditLogs(limit int) ([]AuditLog, error) {
	rows, err := d.db.Query(
		`SELECT 
			action, license_key, product, 
			changed_at, details 
		FROM audit_log 
		ORDER BY changed_at DESC 
		LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("audit log query failed: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		if err := rows.Scan(
			&log.Action, &log.LicenseKey, &log.Product,
			&log.ChangedAt, &log.Details,
		); err != nil {
			return nil, fmt.Errorf("audit log scan failed: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return logs, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "Error 1062: Duplicate entry"
}
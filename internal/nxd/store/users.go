package store

import (
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// LogAudit insere um registro na tabela de auditoria.
func LogAudit(action, apiKey, deviceID, status, message, ipAddress string) {
	db := NXDDB()
	_, err := db.Exec(`
		INSERT INTO nxd.audit_log (action, api_key, device_id, status, message, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, action, apiKey, deviceID, status, message, ipAddress)

	if err != nil {
		// Loga o erro de auditoria no console, mas não trava a aplicação
		log.Printf("CRITICAL: Failed to write to audit log: %v", err)
	}
}

type User struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func CreateUser(db *sql.DB, name, email, password string) (uuid.UUID, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, err
	}

	var id uuid.UUID
	err = db.QueryRow(
		`INSERT INTO nxd.users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id`,
		name, email, string(passwordHash),
	).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	var u User
	err := db.QueryRow(
		`SELECT id, name, email, password_hash, created_at, updated_at FROM nxd.users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByID(db *sql.DB, id uuid.UUID) (*User, error) {
	var u User
	err := db.QueryRow(
		`SELECT id, name, email, password_hash, created_at, updated_at FROM nxd.users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

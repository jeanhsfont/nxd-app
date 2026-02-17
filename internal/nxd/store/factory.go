package store

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// FactoryRow represents a row from nxd.factories (for list responses).
type FactoryRow struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CNPJ      string    `json:"cnpj,omitempty"`
	Location  string    `json:"location,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// GetMemberRole returns the role for (factory_id, user_id) in factory_members. user_id is typically email.
func GetMemberRole(db *sql.DB, factoryID uuid.UUID, userID string) (role string, err error) {
	err = db.QueryRow(
		`SELECT role FROM nxd.factory_members WHERE factory_id = $1 AND user_id = $2`,
		factoryID, userID,
	).Scan(&role)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return role, err
}

// AddFactoryMember inserts or updates factory_members.
func AddFactoryMember(db *sql.DB, factoryID uuid.UUID, userID, role string) error {
	_, err := db.Exec(
		`INSERT INTO nxd.factory_members (factory_id, user_id, role) VALUES ($1, $2, $3)
		 ON CONFLICT (factory_id, user_id) DO UPDATE SET role = EXCLUDED.role`,
		factoryID, userID, role,
	)
	return err
}

// ListFactoriesForUser returns factories the user is a member of.
func ListFactoriesForUser(db *sql.DB, userID string) ([]FactoryRow, error) {
	rows, err := db.Query(
		`SELECT f.id, f.name, f.cnpj, f.location, f.created_at
		 FROM nxd.factories f
		 JOIN nxd.factory_members m ON m.factory_id = f.id AND m.user_id = $1
		 ORDER BY f.name`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []FactoryRow
	for rows.Next() {
		var r FactoryRow
		var cnpj, loc sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&r.ID, &r.Name, &cnpj, &loc, &createdAt); err != nil {
			return nil, err
		}
		r.CreatedAt = createdAt
		if cnpj.Valid {
			r.CNPJ = cnpj.String
		}
		if loc.Valid {
			r.Location = loc.String
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

// CreateFactory inserts a factory and adds the owner to factory_members. Returns the new factory ID.
func CreateFactory(db *sql.DB, name, cnpj, location, ownerUserID string) (uuid.UUID, error) {
	id := uuid.New()
	tx, err := db.Begin()
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback()
	_, err = tx.Exec(
		`INSERT INTO nxd.factories (id, name, cnpj, location) VALUES ($1, $2, NULLIF($3,''), NULLIF($4,''))`,
		id, name, cnpj, location,
	)
	if err != nil {
		return uuid.Nil, err
	}
	_, err = tx.Exec(
		`INSERT INTO nxd.factory_members (factory_id, user_id, role) VALUES ($1, $2, 'OWNER')`,
		id, ownerUserID,
	)
	if err != nil {
		return uuid.Nil, err
	}
	if err = tx.Commit(); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// GetFactoryByGatewayKeyHash returns a factory whose gateway_key_hash equals the given hash (for ingest auth).
func GetFactoryByGatewayKeyHash(db *sql.DB, hash string) (*FactoryRow, error) {
	var r FactoryRow
	var cnpj, loc sql.NullString
	var createdAt time.Time
	err := db.QueryRow(
		`SELECT id, name, cnpj, location, created_at FROM nxd.factories WHERE gateway_key_hash = $1`,
		hash,
	).Scan(&r.ID, &r.Name, &cnpj, &loc, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.CreatedAt = createdAt
	if cnpj.Valid {
		r.CNPJ = cnpj.String
	}
	if loc.Valid {
		r.Location = loc.String
	}
	return &r, nil
}

// SetFactoryGatewayKeyHash stores the hash of the gateway key (key shown once, never stored in plain).
func SetFactoryGatewayKeyHash(db *sql.DB, factoryID uuid.UUID, hash string) error {
	_, err := db.Exec(
		`UPDATE nxd.factories SET gateway_key_hash = $1, gateway_key_salt = $2, updated_at = NOW() WHERE id = $3`,
		hash, "", factoryID,
	)
	return err
}

// GetFactoryByID loads a factory by id. Returns nil if not found.
func GetFactoryByID(db *sql.DB, id uuid.UUID) (*FactoryRow, error) {
	var r FactoryRow
	var cnpj, loc sql.NullString
	var createdAt time.Time
	err := db.QueryRow(
		`SELECT id, name, cnpj, location, created_at FROM nxd.factories WHERE id = $1`,
		id,
	).Scan(&r.ID, &r.Name, &cnpj, &loc, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.CreatedAt = createdAt
	if cnpj.Valid {
		r.CNPJ = cnpj.String
	}
	if loc.Valid {
		r.Location = loc.String
	}
	return &r, nil
}

// AuditWrite writes an entry to nxd.audit_log (for cross-factory denial, etc.).
func AuditWrite(db *sql.DB, actorUserID string, factoryID *uuid.UUID, action, entityType, entityID, correlationID string, metadata interface{}) error {
	var fid interface{}
	if factoryID != nil {
		fid = *factoryID
	}
	_, err := db.Exec(
		`INSERT INTO nxd.audit_log (actor_user_id, factory_id, action, entity_type, entity_id, correlation_id, metadata)
		 VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)`,
		actorUserID, fid, action, entityType, entityID, correlationID, toJSON(metadata),
	)
	return err
}

func toJSON(v interface{}) string {
	if v == nil {
		return "{}"
	}
	b, _ := json.Marshal(v)
	if len(b) == 0 {
		return "{}"
	}
	return string(b)
}

// RoleAllowed returns true if role is in the allowed list.
func RoleAllowed(role string, allowed []string) bool {
	for _, a := range allowed {
		if role == a {
			return true
		}
	}
	return false
}

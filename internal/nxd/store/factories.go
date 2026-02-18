package store

import (
	"database/sql"
	"hubsystem/core"

	"github.com/google/uuid"
)

// CreateFactoryForUser cria uma fábrica para um usuário específico.
func CreateFactoryForUser(db *sql.DB, name, apiKey string, userID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := db.QueryRow(
		`INSERT INTO nxd.factories (name, api_key, user_id) VALUES ($1, $2, $3) RETURNING id`,
		name, apiKey, userID,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// GetFactoryByUserID busca a fábrica associada a um usuário.
func GetFactoryByUserID(db *sql.DB, userID uuid.UUID) (*core.Factory, error) {
	var f core.Factory
	var factoryID uuid.UUID
	var fUserID uuid.UUID

	err := db.QueryRow(
		`SELECT id, user_id, name, api_key, created_at, is_active FROM nxd.factories WHERE user_id = $1 LIMIT 1`,
		userID,
	).Scan(&factoryID, &fUserID, &f.Name, &f.APIKey, &f.CreatedAt, &f.IsActive)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Nenhum erro, apenas nenhuma fábrica encontrada
		}
		return nil, err
	}

	f.ID = factoryID.String() // Converte UUID para string para o core.Factory
	f.UserID = fUserID.String()

	return &f, nil
}

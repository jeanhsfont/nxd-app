package store

import (
	"database/sql"
	"hubsystem/core"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// GetFactoryByAPIKey busca uma fábrica pela API Key.
// ATENÇÃO: Esta função é ineficiente e não deve ser usada em produção com muitas fábricas.
// Ela busca todas as fábricas e compara o hash da chave uma por uma.
// TODO: Refatorar para um método de busca mais eficiente (ex: usando um prefixo de chave).
func GetFactoryByAPIKey(db *sql.DB, apiKey string) (*core.Factory, error) {
	rows, err := db.Query(`SELECT id, user_id, name, api_key_hash, created_at, is_active FROM nxd.factories`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var f core.Factory
		var factoryID, userID uuid.UUID
		var apiKeyHash []byte

		if err := rows.Scan(&factoryID, &userID, &f.Name, &apiKeyHash, &f.CreatedAt, &f.IsActive); err != nil {
			return nil, err
		}

		// Compara a chave fornecida com o hash armazenado
		if err := bcrypt.CompareHashAndPassword(apiKeyHash, []byte(apiKey)); err == nil {
			f.ID = factoryID.String()
			f.UserID = userID.String()
			// Não retornamos o hash, mas a chave original para consistência (embora não esteja no struct)
			f.APIKey = apiKey
			return &f, nil
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return nil, nil // Nenhuma fábrica encontrada
}

// RegenerateAPIKey gera uma nova API key para a fábrica e salva seu hash.
func RegenerateAPIKey(db *sql.DB, factoryID uuid.UUID) (string, error) {
	newAPIKey, hash, err := core.GenerateAndHashAPIKey()
	if err != nil {
		return "", err
	}
	_, err = db.Exec(
		`UPDATE nxd.factories SET api_key_hash = $1, updated_at = NOW() WHERE id = $2`,
		hash, factoryID,
	)
	if err != nil {
		return "", err
	}
	return newAPIKey, nil
}

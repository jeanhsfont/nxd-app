package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// SectorRow represents a row from the sectors table.
type SectorRow struct {
	ID          uuid.UUID `json:"id"`
	FactoryID   uuid.UUID `json:"factory_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// ListSectors returns all sectors for a given factory.
func ListSectors(db *sql.DB, factoryID uuid.UUID) ([]SectorRow, error) {
	rows, err := db.Query(
		`SELECT id, factory_id, name, COALESCE(description, ''), created_at FROM sectors WHERE factory_id = $1 ORDER BY name`,
		factoryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []SectorRow
	for rows.Next() {
		var r SectorRow
		if err := rows.Scan(&r.ID, &r.FactoryID, &r.Name, &r.Description, &r.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

// CreateSector inserts a new sector into the database.
func CreateSector(db *sql.DB, factoryID uuid.UUID, name string, description string) (uuid.UUID, error) {
	id := uuid.New()
	_, err := db.Exec(
		`INSERT INTO sectors (id, factory_id, name, description) VALUES ($1, $2, $3, $4)`,
		id, factoryID, name, description,
	)
	return id, err
}

// UpdateSector updates a sector's name and description.
func UpdateSector(db *sql.DB, sectorID uuid.UUID, factoryID uuid.UUID, name string, description string) error {
	_, err := db.Exec(
		`UPDATE sectors SET name = $1, description = $2 WHERE id = $3 AND factory_id = $4`,
		name, description, sectorID, factoryID,
	)
	return err
}

// DeleteSector removes a sector from the database.
func DeleteSector(db *sql.DB, sectorID uuid.UUID, factoryID uuid.UUID) error {
	// Antes de deletar, precisamos garantir que nenhum ativo está usando este setor.
	// O ideal é desassociar os ativos ou impedir a exclusão se houver algum associado.
	// Por enquanto, vamos desassociar.
	_, err := db.Exec(`UPDATE assets SET sector_id = NULL WHERE sector_id = $1 AND factory_id = $2`, sectorID, factoryID)
	if err != nil {
		return err
	}

	_, err = db.Exec(`DELETE FROM sectors WHERE id = $1 AND factory_id = $2`, sectorID, factoryID)
	return err
}

// AssetTelemetryRow represents a row from the asset_telemetry table.
type AssetTelemetryRow struct {
	ID        int64     `json:"id"`
	AssetID   uuid.UUID `json:"asset_id"`
	Timestamp time.Time `json:"timestamp"`
	Payload   string    `json:"payload"`
}

// ListAssetsBySector returns all assets for a given sector.
func ListAssetsBySector(db *sql.DB, sectorID uuid.UUID) ([]AssetRow, error) {
	rows, err := db.Query(
		`SELECT id, factory_id, group_id, source_tag_id, display_name, description, COALESCE(annotations::text,'{}'), created_at FROM nxd.assets WHERE group_id = $1`,
		sectorID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AssetRow
	for rows.Next() {
		var r AssetRow
		var gid sql.NullString
		var desc sql.NullString
		var ann []byte
		if err := rows.Scan(&r.ID, &r.FactoryID, &gid, &r.SourceTagID, &r.DisplayName, &desc, &ann, &r.CreatedAt); err != nil {
			return nil, err
		}
		if gid.Valid {
			u, _ := uuid.Parse(gid.String)
			r.GroupID = &u
		}
		if desc.Valid {
			r.Description = desc.String
		}
		r.Annotations = ann
		list = append(list, r)
	}
	return list, rows.Err()
}

// ListTelemetryByAssets returns all telemetry data for a given list of asset IDs.
func ListTelemetryByAssets(db *sql.DB, assetIDs []uuid.UUID) ([]AssetTelemetryRow, error) {
	// Se não houver IDs de ativos, não há nada a fazer.
	if len(assetIDs) == 0 {
		return []AssetTelemetryRow{}, nil
	}

	// Construir a query com placeholders para a lista de IDs.
	// NOTA: A forma de fazer isso pode variar um pouco dependendo do driver do banco de dados (PostgreSQL vs SQLite).
	// Esta abordagem é mais genérica, mas para PostgreSQL, poderíamos usar `ANY($1)`.
	query := `SELECT id, asset_id, timestamp, payload FROM asset_telemetry WHERE asset_id IN (`
	args := make([]interface{}, len(assetIDs))
	for i, id := range assetIDs {
		if i > 0 {
			query += ","
		}
		query += "?" // Placeholder
		args[i] = id
	}
	query += `) ORDER BY timestamp DESC`

	// TODO: O placeholder "?" é para SQLite. Se estiver usando PostgreSQL, precisa ser "$1, $2, ...".
	// Uma biblioteca como sqlx ou gorm abstrai isso, mas com `database/sql` puro, precisamos de mais lógica
	// para suportar ambos os bancos de dados de forma limpa. Por enquanto, focando em SQLite.

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AssetTelemetryRow
	for rows.Next() {
		var r AssetTelemetryRow
		if err := rows.Scan(&r.ID, &r.AssetID, &r.Timestamp, &r.Payload); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

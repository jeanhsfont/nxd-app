package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func tableSectors() string {
	if Driver() == "postgres" {
		return "nxd.sectors"
	}
	return "sectors"
}

func tableAssets() string {
	if Driver() == "postgres" {
		return "nxd.assets"
	}
	return "assets"
}

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
	t := tableSectors()
	rows, err := db.Query(
		fmt.Sprintf("SELECT id, factory_id, name, COALESCE(description, ''), created_at FROM %s WHERE factory_id = $1 ORDER BY name", t),
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
	t := tableSectors()
	_, err := db.Exec(
		fmt.Sprintf("INSERT INTO %s (id, factory_id, name, description) VALUES ($1, $2, $3, $4)", t),
		id, factoryID, name, description,
	)
	return id, err
}

// UpdateSector updates a sector's name and description.
func UpdateSector(db *sql.DB, sectorID uuid.UUID, factoryID uuid.UUID, name string, description string) error {
	t := tableSectors()
	_, err := db.Exec(
		fmt.Sprintf("UPDATE %s SET name = $1, description = $2 WHERE id = $3 AND factory_id = $4", t),
		name, description, sectorID, factoryID,
	)
	return err
}

// DeleteSector removes a sector from the database.
func DeleteSector(db *sql.DB, sectorID uuid.UUID, factoryID uuid.UUID) error {
	ts := tableSectors()
	if Driver() == "postgres" {
		ta := tableAssets()
		_, _ = db.Exec(fmt.Sprintf("UPDATE %s SET group_id = NULL WHERE group_id = $1 AND factory_id = $2", ta), sectorID, factoryID)
	}
	_, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = $1 AND factory_id = $2", ts), sectorID, factoryID)
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

// NXDTelemetryRow represents a condensed telemetry entry from nxd.telemetry_log,
// used by ReportIAHandler to build the IA analysis prompt.
type NXDTelemetryRow struct {
	AssetID     uuid.UUID `json:"asset_id"`
	AssetName   string    `json:"asset_name"`
	MetricKey   string    `json:"metric_key"`
	MetricValue float64   `json:"metric_value"`
	Timestamp   time.Time `json:"timestamp"`
	Status      string    `json:"status"`
}

// ListTelemetryByAssets returns the latest telemetry data for a list of asset IDs
// from nxd.telemetry_log (PostgreSQL, schema nxd). Uses $1,$2,... placeholders.
// Returns up to 500 rows ordered by timestamp DESC for IA prompt construction.
func ListTelemetryByAssets(db *sql.DB, assetIDs []uuid.UUID) ([]NXDTelemetryRow, error) {
	if len(assetIDs) == 0 {
		return []NXDTelemetryRow{}, nil
	}

	// Build PostgreSQL $1,$2,... placeholders for the IN clause.
	// We pass them starting at $1; asset_ids occupy $1..$N.
	args := make([]interface{}, len(assetIDs))
	placeholders := make([]string, len(assetIDs))
	for i, id := range assetIDs {
		args[i] = id
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
		SELECT
			tl.asset_id,
			COALESCE(a.display_name, a.source_tag_id, tl.asset_id::text) AS asset_name,
			tl.metric_key,
			tl.metric_value,
			tl.ts,
			COALESCE(tl.status, 'OK') AS status
		FROM nxd.telemetry_log tl
		LEFT JOIN nxd.assets a ON a.id = tl.asset_id
		WHERE tl.asset_id IN (%s)
		ORDER BY tl.ts DESC
		LIMIT 500
	`, strings.Join(placeholders, ","))

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []NXDTelemetryRow
	for rows.Next() {
		var r NXDTelemetryRow
		if err := rows.Scan(&r.AssetID, &r.AssetName, &r.MetricKey, &r.MetricValue, &r.Timestamp, &r.Status); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

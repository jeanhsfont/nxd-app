package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AssetRow represents a row from nxd.assets.
type AssetRow struct {
	ID          uuid.UUID  `json:"id"`
	FactoryID   uuid.UUID  `json:"factory_id"`
	GroupID     *uuid.UUID `json:"group_id,omitempty"`
	SourceTagID string     `json:"source_tag_id"`
	DisplayName string     `json:"display_name"`
	Description string     `json:"description,omitempty"`
	Annotations []byte     `json:"-"` // JSONB; expose as object in API
	CreatedAt   time.Time  `json:"created_at"`
}

// AnnotationsMap returns annotations as map for JSON response.
func (a *AssetRow) AnnotationsMap() map[string]interface{} {
	if len(a.Annotations) == 0 {
		return nil
	}
	var m map[string]interface{}
	_ = json.Unmarshal(a.Annotations, &m)
	return m
}

// ListAssets returns assets for a factory. If ungroupedOnly, only group_id IS NULL. If search != "", filter by display_name/source_tag_id/description.
func ListAssets(db *sql.DB, factoryID uuid.UUID, ungroupedOnly bool, search string) ([]AssetRow, error) {
	query := `SELECT id, factory_id, group_id, source_tag_id, display_name, description, COALESCE(annotations::text,'{}'), created_at FROM nxd.assets WHERE factory_id = $1`
	args := []interface{}{factoryID}
	n := 2
	if ungroupedOnly {
		query += ` AND group_id IS NULL`
	}
	if search != "" {
		query += fmt.Sprintf(` AND (display_name ILIKE $%d OR source_tag_id ILIKE $%d OR description ILIKE $%d)`, n, n, n)
		args = append(args, "%"+search+"%")
		n++
	}
	query += ` ORDER BY display_name`
	rows, err := db.Query(query, args...)
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

// ListAssetsByGroup returns assets in a given group.
func ListAssetsByGroup(db *sql.DB, factoryID, groupID uuid.UUID) ([]AssetRow, error) {
	rows, err := db.Query(
		`SELECT id, factory_id, group_id, source_tag_id, display_name, description, COALESCE(annotations::text,'{}'), created_at FROM nxd.assets WHERE factory_id = $1 AND group_id = $2 ORDER BY display_name`,
		factoryID, groupID,
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

// CreateAsset inserts an asset (auto-discovery or manual). display_name defaults to source_tag_id. On conflict updates and returns existing id.
func CreateAsset(db *sql.DB, factoryID uuid.UUID, groupID *uuid.UUID, sourceTagID, displayName, description string, annotations map[string]interface{}) (uuid.UUID, error) {
	if displayName == "" {
		displayName = sourceTagID
	}
	id := uuid.New()
	annJSON := "{}"
	if len(annotations) > 0 {
		b, _ := json.Marshal(annotations)
		annJSON = string(b)
	}
	var out uuid.UUID
	err := db.QueryRow(
		`INSERT INTO nxd.assets (id, factory_id, group_id, source_tag_id, display_name, description, annotations)
		 VALUES ($1, $2, $3, $4, $5, NULLIF($6,''), $7::jsonb)
		 ON CONFLICT (factory_id, source_tag_id) DO UPDATE SET group_id = EXCLUDED.group_id, display_name = EXCLUDED.display_name, description = EXCLUDED.description, annotations = EXCLUDED.annotations, updated_at = NOW()
		 RETURNING id`,
		id, factoryID, groupID, sourceTagID, displayName, description, annJSON,
	).Scan(&out)
	if err != nil {
		return uuid.Nil, err
	}
	return out, nil
}

// UpdateAsset updates display_name, description, annotations, group_id.
func UpdateAsset(db *sql.DB, id, factoryID uuid.UUID, displayName, description string, annotations map[string]interface{}) error {
	annJSON := "{}"
	if len(annotations) > 0 {
		b, _ := json.Marshal(annotations)
		annJSON = string(b)
	}
	_, err := db.Exec(
		`UPDATE nxd.assets SET display_name = COALESCE(NULLIF($1,''), display_name), description = $2, annotations = $3::jsonb, updated_at = NOW() WHERE id = $4 AND factory_id = $5`,
		displayName, description, annJSON, id, factoryID,
	)
	return err
}

// MoveAsset sets group_id (nil = ungroup).
func MoveAsset(db *sql.DB, id, factoryID uuid.UUID, groupID *uuid.UUID) error {
	_, err := db.Exec(
		`UPDATE nxd.assets SET group_id = $1, updated_at = NOW() WHERE id = $2 AND factory_id = $3`,
		groupID, id, factoryID,
	)
	return err
}

// GetAssetByID returns an asset by id if it belongs to the factory.
func GetAssetByID(db *sql.DB, id, factoryID uuid.UUID) (*AssetRow, error) {
	var r AssetRow
	var gid sql.NullString
	var desc sql.NullString
	var ann []byte
	err := db.QueryRow(
		`SELECT id, factory_id, group_id, source_tag_id, display_name, description, COALESCE(annotations::text,'{}'), created_at FROM nxd.assets WHERE id = $1 AND factory_id = $2`,
		id, factoryID,
	).Scan(&r.ID, &r.FactoryID, &gid, &r.SourceTagID, &r.DisplayName, &desc, &ann, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
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
	return &r, nil
}

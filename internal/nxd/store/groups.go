package store

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// GroupRow represents a row from nxd.groups.
type GroupRow struct {
	ID        uuid.UUID  `json:"id"`
	FactoryID uuid.UUID  `json:"factory_id"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	Name      string     `json:"name"`
	Path      string     `json:"path,omitempty"`
	Metadata  []byte     `json:"-"`
	CreatedAt time.Time  `json:"created_at"`
}

// MetadataMap returns metadata as map for JSON response.
func (g *GroupRow) MetadataMap() map[string]interface{} {
	if len(g.Metadata) == 0 {
		return nil
	}
	var m map[string]interface{}
	_ = json.Unmarshal(g.Metadata, &m)
	return m
}

// ListGroups returns groups for a factory (optionally under a parent).
func ListGroups(db *sql.DB, factoryID uuid.UUID, parentID *uuid.UUID) ([]GroupRow, error) {
	var rows *sql.Rows
	var err error
	if parentID == nil {
		rows, err = db.Query(
			`SELECT id, factory_id, parent_id, name, path, COALESCE(metadata::text,'{}'), created_at FROM nxd.groups WHERE factory_id = $1 AND parent_id IS NULL ORDER BY name`,
			factoryID,
		)
	} else {
		rows, err = db.Query(
			`SELECT id, factory_id, parent_id, name, path, COALESCE(metadata::text,'{}'), created_at FROM nxd.groups WHERE factory_id = $1 AND parent_id = $2 ORDER BY name`,
			factoryID, *parentID,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGroupRows(rows)
}

// ListGroupsRecursive returns all groups for a factory (flat list).
func ListGroupsRecursive(db *sql.DB, factoryID uuid.UUID) ([]GroupRow, error) {
	rows, err := db.Query(
		`SELECT id, factory_id, parent_id, name, path, COALESCE(metadata::text,'{}'), created_at FROM nxd.groups WHERE factory_id = $1 ORDER BY path NULLS FIRST, name`,
		factoryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGroupRows(rows)
}

func scanGroupRows(rows *sql.Rows) ([]GroupRow, error) {
	var list []GroupRow
	for rows.Next() {
		var r GroupRow
		var pid sql.NullString
		var path sql.NullString
		var meta []byte
		if err := rows.Scan(&r.ID, &r.FactoryID, &pid, &r.Name, &path, &meta, &r.CreatedAt); err != nil {
			return nil, err
		}
		if pid.Valid {
			u, _ := uuid.Parse(pid.String)
			r.ParentID = &u
		}
		if path.Valid {
			r.Path = path.String
		}
		r.Metadata = meta
		list = append(list, r)
	}
	return list, rows.Err()
}

// CreateGroup inserts a group.
func CreateGroup(db *sql.DB, factoryID uuid.UUID, parentID *uuid.UUID, name string, metadata map[string]interface{}) (uuid.UUID, error) {
	id := uuid.New()
	path := name
	if parentID != nil {
		var parentPath sql.NullString
		_ = db.QueryRow(`SELECT path FROM nxd.groups WHERE id = $1`, *parentID).Scan(&parentPath)
		if parentPath.Valid && parentPath.String != "" {
			path = parentPath.String + "/" + name
		}
	}
	metaJSON := "{}"
	if len(metadata) > 0 {
		b, _ := json.Marshal(metadata)
		metaJSON = string(b)
	}
	_, err := db.Exec(
		`INSERT INTO nxd.groups (id, factory_id, parent_id, name, path, metadata) VALUES ($1, $2, $3, $4, NULLIF($5,''), $6::jsonb)`,
		id, factoryID, parentID, name, path, metaJSON,
	)
	return id, err
}

// UpdateGroup updates name and metadata.
func UpdateGroup(db *sql.DB, id, factoryID uuid.UUID, name string, metadata map[string]interface{}) error {
	metaJSON := "{}"
	if len(metadata) > 0 {
		b, _ := json.Marshal(metadata)
		metaJSON = string(b)
	}
	_, err := db.Exec(
		`UPDATE nxd.groups SET name = $1, metadata = $2::jsonb, updated_at = NOW() WHERE id = $3 AND factory_id = $4`,
		name, metaJSON, id, factoryID,
	)
	return err
}

// GetGroupByID returns a group by id if it belongs to the factory.
func GetGroupByID(db *sql.DB, id, factoryID uuid.UUID) (*GroupRow, error) {
	var r GroupRow
	var pid sql.NullString
	var path sql.NullString
	var meta []byte
	err := db.QueryRow(
		`SELECT id, factory_id, parent_id, name, path, COALESCE(metadata::text,'{}'), created_at FROM nxd.groups WHERE id = $1 AND factory_id = $2`,
		id, factoryID,
	).Scan(&r.ID, &r.FactoryID, &pid, &r.Name, &path, &meta, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if pid.Valid {
		u, _ := uuid.Parse(pid.String)
		r.ParentID = &u
	}
	if path.Valid {
		r.Path = path.String
	}
	r.Metadata = meta
	return &r, nil
}

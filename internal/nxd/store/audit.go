package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// AuditRow is one row from nxd.audit_log for list response.
type AuditRow struct {
	ID        uuid.UUID `json:"id"`
	Ts        time.Time `json:"ts"`
	ActorUser string   `json:"actor_user_id,omitempty"`
	Action    string   `json:"action"`
	EntityType string  `json:"entity_type,omitempty"`
	EntityID  string   `json:"entity_id,omitempty"`
}

// ListAuditLog returns the most recent audit entries for the factory.
func ListAuditLog(db *sql.DB, factoryID uuid.UUID, limit int) ([]AuditRow, error) {
	rows, err := db.Query(
		`SELECT id, ts, actor_user_id, action, entity_type, entity_id FROM nxd.audit_log WHERE factory_id = $1 ORDER BY ts DESC LIMIT $2`,
		factoryID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []AuditRow
	for rows.Next() {
		var r AuditRow
		var actor, etype, eid sql.NullString
		if err := rows.Scan(&r.ID, &r.Ts, &actor, &r.Action, &etype, &eid); err != nil {
			return nil, err
		}
		if actor.Valid {
			r.ActorUser = actor.String
		}
		if etype.Valid {
			r.EntityType = etype.String
		}
		if eid.Valid {
			r.EntityID = eid.String
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

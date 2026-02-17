package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// AlertRuleRow is a row from nxd.alert_rules.
type AlertRuleRow struct {
	ID           uuid.UUID `json:"id"`
	FactoryID    uuid.UUID `json:"factory_id"`
	ScopeType    string    `json:"scope_type"`
	ScopeID      uuid.UUID `json:"scope_id"`
	ConditionType string   `json:"condition_type"`
	Threshold    float64   `json:"threshold,omitempty"`
	Channel      string    `json:"channel,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// AlertRow is a row from nxd.alerts.
type AlertRow struct {
	ID             uuid.UUID  `json:"id"`
	Ts             time.Time  `json:"ts"`
	RuleID         uuid.UUID  `json:"rule_id"`
	AssetID        *uuid.UUID `json:"asset_id,omitempty"`
	GroupID        *uuid.UUID `json:"group_id,omitempty"`
	Severity       string    `json:"severity"`
	Message        string    `json:"message"`
	AcknowledgedBy string    `json:"acknowledged_by,omitempty"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
}

// ListAlertRules returns rules for the factory.
func ListAlertRules(db *sql.DB, factoryID uuid.UUID) ([]AlertRuleRow, error) {
	rows, err := db.Query(
		`SELECT id, factory_id, scope_type, scope_id, condition_type, threshold, channel, created_at FROM nxd.alert_rules WHERE factory_id = $1 ORDER BY created_at`,
		factoryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []AlertRuleRow
	for rows.Next() {
		var r AlertRuleRow
		var thresh sql.NullFloat64
		var ch sql.NullString
		if err := rows.Scan(&r.ID, &r.FactoryID, &r.ScopeType, &r.ScopeID, &r.ConditionType, &thresh, &ch, &r.CreatedAt); err != nil {
			return nil, err
		}
		if thresh.Valid {
			r.Threshold = thresh.Float64
		}
		if ch.Valid {
			r.Channel = ch.String
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

// CreateAlertRule inserts a rule.
func CreateAlertRule(db *sql.DB, factoryID uuid.UUID, scopeType, scopeID, conditionType string, threshold float64, channel string) (uuid.UUID, error) {
	id := uuid.New()
	scopeUUID, _ := uuid.Parse(scopeID)
	_, err := db.Exec(
		`INSERT INTO nxd.alert_rules (id, factory_id, scope_type, scope_id, condition_type, threshold, channel) VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7,''))`,
		id, factoryID, scopeType, scopeUUID, conditionType, threshold, channel,
	)
	return id, err
}

// ListAlerts returns alerts for the factory (optional: only unacknowledged).
func ListAlerts(db *sql.DB, factoryID uuid.UUID, unackOnly bool) ([]AlertRow, error) {
	query := `SELECT a.id, a.ts, a.rule_id, a.asset_id, a.group_id, a.severity, a.message, a.acknowledged_by, a.acknowledged_at
		  FROM nxd.alerts a
		  JOIN nxd.alert_rules r ON r.id = a.rule_id AND r.factory_id = $1`
	if unackOnly {
		query += ` WHERE a.acknowledged_by IS NULL`
	}
	query += ` ORDER BY a.ts DESC LIMIT 100`
	rows, err := db.Query(query, factoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []AlertRow
	for rows.Next() {
		var r AlertRow
		var aid, gid sql.NullString
		var ackBy sql.NullString
		var ackAt sql.NullTime
		if err := rows.Scan(&r.ID, &r.Ts, &r.RuleID, &aid, &gid, &r.Severity, &r.Message, &ackBy, &ackAt); err != nil {
			return nil, err
		}
		if aid.Valid {
			u, _ := uuid.Parse(aid.String)
			r.AssetID = &u
		}
		if gid.Valid {
			u, _ := uuid.Parse(gid.String)
			r.GroupID = &u
		}
		if ackBy.Valid {
			r.AcknowledgedBy = ackBy.String
		}
		if ackAt.Valid {
			r.AcknowledgedAt = &ackAt.Time
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

// CreateAlert inserts an alert.
func CreateAlert(db *sql.DB, ruleID uuid.UUID, assetID, groupID *uuid.UUID, severity, message string) (uuid.UUID, error) {
	id := uuid.New()
	_, err := db.Exec(
		`INSERT INTO nxd.alerts (id, rule_id, asset_id, group_id, severity, message) VALUES ($1, $2, $3, $4, $5, $6)`,
		id, ruleID, assetID, groupID, severity, message,
	)
	return id, err
}

// AckAlert sets acknowledged_by and acknowledged_at.
func AckAlert(db *sql.DB, alertID, factoryID uuid.UUID, userID string) error {
	_, err := db.Exec(
		`UPDATE nxd.alerts SET acknowledged_by = $1, acknowledged_at = NOW() WHERE id = $2 AND rule_id IN (SELECT id FROM nxd.alert_rules WHERE factory_id = $3)`,
		userID, alertID, factoryID,
	)
	return err
}

package store

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// InsertTelemetryBatch inserts multiple rows into telemetry_log. Asset must exist (caller ensures via CreateAsset).
func InsertTelemetryBatch(db *sql.DB, factoryID, assetID uuid.UUID, correlationID string, rows []TelemetryRow) error {
	for _, row := range rows {
		rawJSON := "null"
		if len(row.Raw) > 0 {
			rawJSON = string(row.Raw)
		}
		_, err := db.Exec(
			`INSERT INTO nxd.telemetry_log (ts, factory_id, asset_id, metric_key, metric_value, status, raw, correlation_id)
			 VALUES ($1, $2, $3, $4, $5, COALESCE(NULLIF($6,''), 'OK'), $7::jsonb, $8)`,
			row.Ts, factoryID, assetID, row.MetricKey, row.MetricValue, row.Status, rawJSON, correlationID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// TelemetryRow is one row for telemetry_log.
type TelemetryRow struct {
	Ts         time.Time
	MetricKey  string
	MetricValue float64
	Status     string
	Raw        []byte
}

// UpsertAssetMetricCatalog records that this asset has this metric (first_seen/last_seen).
func UpsertAssetMetricCatalog(db *sql.DB, factoryID, assetID uuid.UUID, metricKey string, ts time.Time) error {
	_, err := db.Exec(
		`INSERT INTO nxd.asset_metric_catalog (factory_id, asset_id, metric_key, first_seen, last_seen)
		 VALUES ($1, $2, $3, $4, $4)
		 ON CONFLICT (factory_id, asset_id, metric_key) DO UPDATE SET last_seen = $4`,
		factoryID, assetID, metricKey, ts,
	)
	return err
}

// LastTelemetryTs returns the latest ts for the factory (for health "último ts").
func LastTelemetryTs(db *sql.DB, factoryID uuid.UUID) (time.Time, error) {
	var t time.Time
	err := db.QueryRow(
		`SELECT COALESCE(MAX(ts), '1970-01-01'::timestamptz) FROM nxd.telemetry_log WHERE factory_id = $1`,
		factoryID,
	).Scan(&t)
	return t, err
}

// TelemetryIngestPayload is the validated contract for POST /nxd/telemetry/ingest.
type TelemetryIngestPayload struct {
	GatewayKey  string           `json:"gateway_key"`
	SourceTagID string           `json:"source_tag_id"`
	Ts          time.Time        `json:"ts"`
	Metrics     []MetricEntry    `json:"metrics"`
	Raw         *json.RawMessage `json:"raw,omitempty"`
}

// MetricEntry is one metric in the ingest payload.
type MetricEntry struct {
	MetricKey  string   `json:"metric_key"`
	Value      float64  `json:"value"`
	Status     string   `json:"status,omitempty"`
	Unit       string   `json:"unit,omitempty"`
	Quality    string   `json:"quality,omitempty"`
}

var _ = sql.ErrNoRows

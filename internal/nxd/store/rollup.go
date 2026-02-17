package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// RunRollup aggregates telemetry_log into telemetry_rollup_1m for the given time window (e.g. [now-120d, now-90d]).
func RunRollup(db *sql.DB, from, to time.Time) (inserted int64, err error) {
	res, err := db.Exec(`
		INSERT INTO nxd.telemetry_rollup_1m (bucket_ts, factory_id, asset_id, metric_key, avg_value, min_value, max_value, samples, status_counts)
		SELECT
			date_trunc('minute', ts),
			factory_id,
			asset_id,
			metric_key,
			AVG(metric_value),
			MIN(metric_value),
			MAX(metric_value),
			COUNT(*)::bigint,
			'{}'::jsonb
		FROM nxd.telemetry_log
		WHERE ts >= $1 AND ts < $2
		GROUP BY date_trunc('minute', ts), factory_id, asset_id, metric_key
		ON CONFLICT (bucket_ts, factory_id, asset_id, metric_key) DO UPDATE SET
			avg_value = EXCLUDED.avg_value,
			min_value = EXCLUDED.min_value,
			max_value = EXCLUDED.max_value,
			samples = nxd.telemetry_rollup_1m.samples + EXCLUDED.samples
	`, from, to)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// MarkRollupDone records that a range was compacted (optional: use to avoid re-aggregating).
func MarkRollupDone(db *sql.DB, factoryID, assetID uuid.UUID, metricKey string, bucketStart, bucketEnd time.Time) error {
	_, err := db.Exec(
		`INSERT INTO nxd.rollup_ranges (factory_id, asset_id, metric_key, bucket_start, bucket_end) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`,
		factoryID, assetID, metricKey, bucketStart, bucketEnd,
	)
	return err
}

// QueryRollup returns aggregated data from telemetry_rollup_1m for the given window (for COLD read).
func QueryRollup(db *sql.DB, factoryID uuid.UUID, assetID *uuid.UUID, from, to time.Time) ([]RollupRow, error) {
	query := `SELECT bucket_ts, factory_id, asset_id, metric_key, avg_value, min_value, max_value, samples FROM nxd.telemetry_rollup_1m WHERE factory_id = $1 AND bucket_ts >= $2 AND bucket_ts < $3`
	args := []interface{}{factoryID, from, to}
	if assetID != nil {
		query += ` AND asset_id = $4`
		args = append(args, *assetID)
	}
	query += ` ORDER BY bucket_ts`
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []RollupRow
	for rows.Next() {
		var r RollupRow
		if err := rows.Scan(&r.BucketTs, &r.FactoryID, &r.AssetID, &r.MetricKey, &r.AvgValue, &r.MinValue, &r.MaxValue, &r.Samples); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

// RollupRow is one row from telemetry_rollup_1m.
type RollupRow struct {
	BucketTs  time.Time
	FactoryID uuid.UUID
	AssetID   uuid.UUID
	MetricKey string
	AvgValue  float64
	MinValue  float64
	MaxValue  float64
	Samples   int64
}

var _ = sql.ErrNoRows

//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:NXD_Prod_2026!@127.0.0.1:5433/postgres?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping: %v", err)
	}
	fmt.Println("✓ Connected to database")

	indexes := []struct {
		name string
		sql  string
	}{
		{
			"idx_telemetry_log_asset_ts",
			`CREATE INDEX IF NOT EXISTS idx_telemetry_log_asset_ts
				ON nxd.telemetry_log (asset_id, ts DESC)
				WHERE asset_id IS NOT NULL`,
		},
		{
			"idx_telemetry_log_factory_ts",
			`CREATE INDEX IF NOT EXISTS idx_telemetry_log_factory_ts
				ON nxd.telemetry_log (factory_id, ts DESC)
				WHERE factory_id IS NOT NULL`,
		},
		{
			"idx_telemetry_log_asset_metric_ts",
			`CREATE INDEX IF NOT EXISTS idx_telemetry_log_asset_metric_ts
				ON nxd.telemetry_log (asset_id, metric_key, ts DESC)
				WHERE asset_id IS NOT NULL AND metric_key IS NOT NULL`,
		},
	}

	for _, idx := range indexes {
		fmt.Printf("Creating index %s ... ", idx.name)
		start := time.Now()
		if _, err := db.Exec(idx.sql); err != nil {
			log.Fatalf("\nFailed: %v", err)
		}
		fmt.Printf("done in %s\n", time.Since(start).Round(time.Second))
	}

	fmt.Println("\n✓ All indexes created successfully!")
}

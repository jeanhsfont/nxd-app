package store

import (
	"database/sql"
	"fmt"
)

func RunMigrations(db *sql.DB, driver string) error {
	var migrations []string
	if driver == "postgres" {
		migrations = postgresMigrations
	} else {
		migrations = sqliteMigrations
	}

	for i, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration %d (%s): %w", i+1, driver, err)
		}
	}
	return nil
}

var postgresMigrations = []string{
	`CREATE SCHEMA IF NOT EXISTS nxd`,
	`CREATE TABLE IF NOT EXISTS nxd.users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	)`,
	`CREATE TABLE IF NOT EXISTS nxd.sectors (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name TEXT NOT NULL,
		description TEXT,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	)`,
	`CREATE TABLE IF NOT EXISTS nxd.assets (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		sector_id UUID REFERENCES nxd.sectors(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		description TEXT,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	)`,
	`CREATE TABLE IF NOT EXISTS nxd.asset_telemetry (
		ts TIMESTAMPTZ NOT NULL,
		asset_id UUID NOT NULL REFERENCES nxd.assets(id) ON DELETE CASCADE,
		metric_key TEXT NOT NULL,
		metric_value DOUBLE PRECISION NOT NULL
	)`,
	`SELECT create_hypertable('nxd.asset_telemetry', 'ts', if_not_exists => TRUE)`,
}

var sqliteMigrations = []string{
	`CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS sectors (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS assets (
		id TEXT PRIMARY KEY,
		sector_id TEXT NOT NULL REFERENCES sectors(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS asset_telemetry (
		ts DATETIME NOT NULL,
		asset_id TEXT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
		metric_key TEXT NOT NULL,
		metric_value REAL NOT NULL
	)`,
}

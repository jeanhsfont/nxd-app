package store

import (
	"database/sql"
	"fmt"
	"log"
)

// optionalMigrations are executed but failures are only logged, not fatal.
// Used for extensions that may not be available (e.g. TimescaleDB).
var optionalMigrations = map[string]bool{
	`SELECT create_hypertable('nxd.asset_telemetry', 'ts', if_not_exists => TRUE)`:  true,
	`SELECT create_hypertable('nxd.telemetry_log', 'ts', if_not_exists => TRUE)`:    true,
}

func RunMigrations(db *sql.DB, driver string) error {
	var migrations []string
	if driver == "postgres" {
		migrations = postgresMigrations
	} else {
		migrations = sqliteMigrations
	}

	for i, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			if optionalMigrations[m] {
				log.Printf("⚠️  Migration opcional %d ignorada (%s): %v", i+1, driver, err)
				continue
			}
			return fmt.Errorf("migration %d (%s): %w", i+1, driver, err)
		}
	}
	return nil
}

var postgresMigrations = []string{
	// Schema
	`CREATE SCHEMA IF NOT EXISTS nxd`,

	// Users
	`CREATE TABLE IF NOT EXISTS nxd.users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	)`,

	// Factories
	`CREATE TABLE IF NOT EXISTS nxd.factories (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID REFERENCES nxd.users(id) ON DELETE CASCADE,
		name TEXT NOT NULL DEFAULT 'Minha Fábrica',
		api_key TEXT,
		api_key_hash BYTEA,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	)`,

	// Sectors (groups of assets inside a factory)
	`CREATE TABLE IF NOT EXISTS nxd.sectors (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		factory_id UUID REFERENCES nxd.factories(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		description TEXT,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	)`,

	// Drop old assets table if it has wrong schema (missing factory_id)
	// We use a DO block so it's idempotent
	`DO $$
	BEGIN
		IF EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'nxd' AND table_name = 'assets'
		) AND NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'nxd' AND table_name = 'assets' AND column_name = 'factory_id'
		) THEN
			DROP TABLE IF EXISTS nxd.asset_telemetry CASCADE;
			DROP TABLE IF EXISTS nxd.assets CASCADE;
		END IF;
	END $$`,

	// Assets (devices/CLPs reporting telemetry)
	`CREATE TABLE IF NOT EXISTS nxd.assets (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		factory_id UUID NOT NULL REFERENCES nxd.factories(id) ON DELETE CASCADE,
		group_id UUID REFERENCES nxd.sectors(id) ON DELETE SET NULL,
		source_tag_id TEXT NOT NULL,
		display_name TEXT NOT NULL,
		description TEXT,
		annotations JSONB,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW(),
		UNIQUE(factory_id, source_tag_id)
	)`,

	// Raw telemetry (time-series)
	`CREATE TABLE IF NOT EXISTS nxd.asset_telemetry (
		ts TIMESTAMPTZ NOT NULL,
		asset_id UUID NOT NULL REFERENCES nxd.assets(id) ON DELETE CASCADE,
		metric_key TEXT NOT NULL,
		metric_value DOUBLE PRECISION NOT NULL
	)`,
	`SELECT create_hypertable('nxd.asset_telemetry', 'ts', if_not_exists => TRUE)`,

	// Audit log
	`CREATE TABLE IF NOT EXISTS nxd.audit_log (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		ts TIMESTAMPTZ DEFAULT NOW(),
		factory_id UUID REFERENCES nxd.factories(id) ON DELETE CASCADE,
		actor_user_id UUID,
		action TEXT NOT NULL,
		entity_type TEXT,
		entity_id TEXT,
		api_key TEXT,
		device_id TEXT,
		status TEXT,
		message TEXT,
		ip_address TEXT
	)`,

	// Asset metric catalog (tracks which metrics each asset reports)
	`CREATE TABLE IF NOT EXISTS nxd.asset_metric_catalog (
		factory_id UUID NOT NULL REFERENCES nxd.factories(id) ON DELETE CASCADE,
		asset_id UUID NOT NULL REFERENCES nxd.assets(id) ON DELETE CASCADE,
		metric_key TEXT NOT NULL,
		first_seen TIMESTAMPTZ DEFAULT NOW(),
		last_seen TIMESTAMPTZ DEFAULT NOW(),
		PRIMARY KEY (factory_id, asset_id, metric_key)
	)`,

	// Telemetry log (enriched ingest log with correlation IDs)
	`CREATE TABLE IF NOT EXISTS nxd.telemetry_log (
		ts TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		factory_id UUID REFERENCES nxd.factories(id) ON DELETE CASCADE,
		asset_id UUID REFERENCES nxd.assets(id) ON DELETE CASCADE,
		metric_key TEXT,
		metric_value DOUBLE PRECISION,
		status TEXT,
		raw JSONB,
		correlation_id TEXT
	)`,
	`SELECT create_hypertable('nxd.telemetry_log', 'ts', if_not_exists => TRUE)`,

	// ─── Performance indexes on telemetry_log ───────────────────────────────
	// These are critical for dashboard queries on large datasets.
	// CREATE INDEX CONCURRENTLY is not allowed inside transactions, so we use
	// a regular CREATE INDEX with IF NOT EXISTS (idempotent).
	`CREATE INDEX IF NOT EXISTS idx_telemetry_log_asset_ts
		ON nxd.telemetry_log (asset_id, ts DESC)
		WHERE asset_id IS NOT NULL`,
	`CREATE INDEX IF NOT EXISTS idx_telemetry_log_factory_ts
		ON nxd.telemetry_log (factory_id, ts DESC)
		WHERE factory_id IS NOT NULL`,
	`CREATE INDEX IF NOT EXISTS idx_telemetry_log_asset_metric_ts
		ON nxd.telemetry_log (asset_id, metric_key, ts DESC)
		WHERE asset_id IS NOT NULL AND metric_key IS NOT NULL`,

	// ─── api_key_prefix for O(1) ingest authentication ───────────────────────
	// Add api_key_prefix column to nxd.factories for fast API key lookup.
	// Uses ALTER TABLE ... ADD COLUMN IF NOT EXISTS (PostgreSQL 9.6+) — idempotent.
	// The prefix is the first 16 chars of the plaintext key (NXD_xxxxxxxxxxxx).
	// It is NOT secret: it cannot be used to reconstruct the key, and bcrypt
	// comparison still happens as the second verification step.
	`ALTER TABLE nxd.factories ADD COLUMN IF NOT EXISTS api_key_prefix TEXT`,
	`CREATE INDEX IF NOT EXISTS idx_factories_api_key_prefix ON nxd.factories (api_key_prefix) WHERE api_key_prefix IS NOT NULL`,
	// Backfill prefix from existing plaintext api_key column (safe: non-destructive).
	`UPDATE nxd.factories SET api_key_prefix = LEFT(api_key, 16) WHERE api_key IS NOT NULL AND api_key_prefix IS NULL`,

	// ─── Fix nxd.sectors schema ──────────────────────────────────────────────
	// The original CREATE TABLE IF NOT EXISTS for nxd.sectors may have been executed
	// without the factory_id column if the table existed from a prior schema version.
	// These ALTER TABLE ... ADD COLUMN IF NOT EXISTS statements bring it up to spec.
	// Idempotent: IF NOT EXISTS means they are safe to run multiple times.
	`ALTER TABLE nxd.sectors ADD COLUMN IF NOT EXISTS factory_id UUID REFERENCES nxd.factories(id) ON DELETE CASCADE`,
	`CREATE INDEX IF NOT EXISTS idx_sectors_factory_id ON nxd.sectors (factory_id) WHERE factory_id IS NOT NULL`,

	// Alert rules
	`CREATE TABLE IF NOT EXISTS nxd.alert_rules (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		factory_id UUID NOT NULL REFERENCES nxd.factories(id) ON DELETE CASCADE,
		scope_type TEXT NOT NULL,
		scope_id UUID,
		condition_type TEXT NOT NULL,
		threshold DOUBLE PRECISION,
		channel TEXT,
		created_at TIMESTAMPTZ DEFAULT NOW()
	)`,

	// Alerts fired
	`CREATE TABLE IF NOT EXISTS nxd.alerts (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		ts TIMESTAMPTZ DEFAULT NOW(),
		rule_id UUID REFERENCES nxd.alert_rules(id) ON DELETE CASCADE,
		asset_id UUID REFERENCES nxd.assets(id) ON DELETE CASCADE,
		group_id UUID,
		severity TEXT,
		message TEXT,
		acknowledged_by UUID,
		acknowledged_at TIMESTAMPTZ
	)`,

	// Telemetry rollup 1-minute buckets
	`CREATE TABLE IF NOT EXISTS nxd.telemetry_rollup_1m (
		bucket_ts TIMESTAMPTZ NOT NULL,
		factory_id UUID NOT NULL,
		asset_id UUID NOT NULL,
		metric_key TEXT NOT NULL,
		avg_value DOUBLE PRECISION,
		min_value DOUBLE PRECISION,
		max_value DOUBLE PRECISION,
		samples INT,
		status_counts JSONB,
		PRIMARY KEY (bucket_ts, factory_id, asset_id, metric_key)
	)`,

	// Report templates
	`CREATE TABLE IF NOT EXISTS nxd.report_templates (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		category TEXT,
		name TEXT NOT NULL,
		description TEXT,
		default_filters JSONB,
		prompt_instructions TEXT,
		output_schema_version TEXT,
		created_at TIMESTAMPTZ DEFAULT NOW()
	)`,

	// Report runs
	`CREATE TABLE IF NOT EXISTS nxd.report_runs (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		factory_id UUID REFERENCES nxd.factories(id) ON DELETE CASCADE,
		requested_by UUID,
		filters JSONB,
		prompt_contract TEXT,
		status TEXT DEFAULT 'pending',
		result_json JSONB,
		export_url TEXT,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	)`,

	// ─── Historical import jobs (base for "download longo" feature) ──────────
	// Tracks background imports of historical DX data.
	// status lifecycle: pending → running → done | failed | cancelled
	// rows_total/rows_done enable real progress percentage in the UI.
	// batch_size controls how many rows are inserted per DB transaction
	// (default 1000 — balances memory vs. commit overhead).
	// source_config stores DX connection details (endpoint, auth) as JSONB
	// so the import worker can reconnect to the DX autonomously.
	`CREATE TABLE IF NOT EXISTS nxd.import_jobs (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		factory_id UUID NOT NULL REFERENCES nxd.factories(id) ON DELETE CASCADE,
		asset_id UUID REFERENCES nxd.assets(id) ON DELETE SET NULL,
		requested_by UUID,
		status TEXT NOT NULL DEFAULT 'pending',
		period_start TIMESTAMPTZ,
		period_end TIMESTAMPTZ,
		rows_total BIGINT DEFAULT 0,
		rows_done BIGINT DEFAULT 0,
		batch_size INT DEFAULT 1000,
		source_type TEXT DEFAULT 'dx_http',
		source_config JSONB,
		error_message TEXT,
		started_at TIMESTAMPTZ,
		finished_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	)`,
	`CREATE INDEX IF NOT EXISTS idx_import_jobs_factory_status
		ON nxd.import_jobs (factory_id, status)
		WHERE status IN ('pending', 'running')`,

	// ─── P6: expected_interval_s — base for offline DX detection ────────────
	// Nullable integer (seconds). NULL = no expectation set.
	// When set, a future monitoring job can flag assets that have not reported
	// within expected_interval_s * some_multiplier (e.g. 3x).
	// NOT implementing the detection UI now — just preparing the schema column.
	`ALTER TABLE nxd.assets ADD COLUMN IF NOT EXISTS expected_interval_s INT`,

	// ─── P4: Idempotency index for historical import ─────────────────────────
	// Partial index on telemetry_log to speed up the range-check deduplication
	// query used by the import worker before inserting each batch.
	// Strategy: RANGE CHECK (see importer.go checkRangeExists).
	// We do NOT use a UNIQUE constraint because:
	//   1. TimescaleDB hypertables require the partition key (ts) in any unique
	//      index, making (asset_id, metric_key, ts) the minimum — this is a
	//      225-byte index on a write-heavy table, prohibitively expensive.
	//   2. Duplicate protection at ingest time would double write latency.
	//   3. The operational contract for historical import is: same job ID is
	//      never resubmitted for the same asset + time range. Enforced in the
	//      worker via job-level range tracking.
	// This index makes the range-check query (COUNT(*) WHERE asset_id=X AND
	// ts BETWEEN start AND end) O(log n) instead of O(n).
	`CREATE INDEX IF NOT EXISTS idx_telemetry_log_import_range
		ON nxd.telemetry_log (asset_id, ts)
		WHERE asset_id IS NOT NULL`,

	// ─── MVP Indicadores Financeiros: Configuração de negócio por setor ─────
	// Unidade de cálculo: Setor (ou linha = conjunto de ativos no setor).
	// Parâmetros: valor_venda_ok (R$/un), custo_refugo_un (R$/un), custo_parada_h (R$/h).
	`CREATE TABLE IF NOT EXISTS nxd.business_config (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		factory_id UUID NOT NULL REFERENCES nxd.factories(id) ON DELETE CASCADE,
		sector_id UUID REFERENCES nxd.sectors(id) ON DELETE CASCADE,
		valor_venda_ok NUMERIC(18,4) NOT NULL DEFAULT 0,
		custo_refugo_un NUMERIC(18,4) NOT NULL DEFAULT 0,
		custo_parada_h NUMERIC(18,4) NOT NULL DEFAULT 0,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW(),
		UNIQUE(factory_id, sector_id)
	)`,
	`CREATE INDEX IF NOT EXISTS idx_business_config_factory ON nxd.business_config (factory_id)`,
	// Apenas uma config "padrão" (sector_id NULL) por fábrica.
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_business_config_factory_default ON nxd.business_config (factory_id) WHERE sector_id IS NULL`,

	// Mapeamento de tags do CLP por ativo (linha/máquina): OK, NOK, Status.
	// reading_rule: 'delta' = usar variação por período; 'absolute' = usar valor absoluto.
	`CREATE TABLE IF NOT EXISTS nxd.tag_mapping (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		asset_id UUID NOT NULL REFERENCES nxd.assets(id) ON DELETE CASCADE UNIQUE,
		tag_ok TEXT,
		tag_nok TEXT,
		tag_status TEXT,
		reading_rule TEXT NOT NULL DEFAULT 'delta',
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	)`,
	`CREATE INDEX IF NOT EXISTS idx_tag_mapping_asset ON nxd.tag_mapping (asset_id)`,
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
		factory_id TEXT NOT NULL,
		source_tag_id TEXT NOT NULL,
		display_name TEXT NOT NULL,
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
	`CREATE TABLE IF NOT EXISTS audit_log (
		id TEXT PRIMARY KEY,
		ts DATETIME DEFAULT CURRENT_TIMESTAMP,
		action TEXT NOT NULL,
		api_key TEXT,
		device_id TEXT,
		status TEXT,
		message TEXT,
		ip_address TEXT
	)`,
}

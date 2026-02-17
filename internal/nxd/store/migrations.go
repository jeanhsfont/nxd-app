package store

import (
	"database/sql"
	"fmt"
)

// RunMigrations creates schema nxd and all tables. Idempotent (CREATE IF NOT EXISTS / CREATE TABLE IF NOT EXISTS).
func RunMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE SCHEMA IF NOT EXISTS nxd`,
		`CREATE TABLE IF NOT EXISTS nxd.factories (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL,
			cnpj TEXT,
			location TEXT,
			gateway_key_hash TEXT,
			gateway_key_salt TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_nxd_factories_name ON nxd.factories(name)`,
		`CREATE TABLE IF NOT EXISTS nxd.factory_members (
			factory_id UUID NOT NULL REFERENCES nxd.factories(id) ON DELETE CASCADE,
			user_id TEXT NOT NULL,
			role TEXT NOT NULL CHECK (role IN ('OWNER','ADMIN','MANAGER','VIEWER')),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			PRIMARY KEY (factory_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS nxd.groups (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			factory_id UUID NOT NULL REFERENCES nxd.factories(id) ON DELETE CASCADE,
			parent_id UUID REFERENCES nxd.groups(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			path TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_nxd_groups_factory ON nxd.groups(factory_id)`,
		`CREATE INDEX IF NOT EXISTS idx_nxd_groups_parent ON nxd.groups(parent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_nxd_groups_factory_path ON nxd.groups(factory_id, path)`,
		`CREATE TABLE IF NOT EXISTS nxd.assets (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			factory_id UUID NOT NULL REFERENCES nxd.factories(id) ON DELETE CASCADE,
			group_id UUID REFERENCES nxd.groups(id) ON DELETE SET NULL,
			source_tag_id TEXT NOT NULL,
			display_name TEXT NOT NULL,
			description TEXT,
			annotations JSONB DEFAULT '{}',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(factory_id, source_tag_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_nxd_assets_factory ON nxd.assets(factory_id)`,
		`CREATE INDEX IF NOT EXISTS idx_nxd_assets_group ON nxd.assets(group_id)`,
		`CREATE INDEX IF NOT EXISTS idx_nxd_assets_factory_source ON nxd.assets(factory_id, source_tag_id)`,
		`CREATE INDEX IF NOT EXISTS idx_nxd_assets_annotations_gin ON nxd.assets USING GIN (annotations jsonb_path_ops)`,
		// telemetry_log: no FK to assets to allow insert before asset exists (auto-discovery)
		`CREATE TABLE IF NOT EXISTS nxd.telemetry_log (
			ts TIMESTAMPTZ NOT NULL,
			factory_id UUID NOT NULL,
			asset_id UUID NOT NULL,
			metric_key TEXT NOT NULL,
			metric_value DOUBLE PRECISION,
			status TEXT CHECK (status IN ('OK','WARN','ALERT','OFFLINE')) DEFAULT 'OK',
			raw JSONB,
			correlation_id TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_nxd_telemetry_asset_ts ON nxd.telemetry_log(asset_id, ts DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_nxd_telemetry_factory_ts ON nxd.telemetry_log(factory_id, ts DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_nxd_telemetry_metric_ts ON nxd.telemetry_log(metric_key, ts DESC)`,
		`CREATE TABLE IF NOT EXISTS nxd.audit_log (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			ts TIMESTAMPTZ DEFAULT NOW(),
			actor_user_id TEXT,
			factory_id UUID,
			action TEXT NOT NULL,
			entity_type TEXT,
			entity_id TEXT,
			metadata JSONB DEFAULT '{}',
			correlation_id TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_nxd_audit_factory_ts ON nxd.audit_log(factory_id, ts DESC)`,
		`CREATE TABLE IF NOT EXISTS nxd.report_runs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			factory_id UUID NOT NULL REFERENCES nxd.factories(id) ON DELETE CASCADE,
			requested_by TEXT,
			filters JSONB DEFAULT '{}',
			prompt_contract JSONB,
			status TEXT NOT NULL CHECK (status IN ('PENDING','RUNNING','DONE','FAILED')) DEFAULT 'PENDING',
			result_json JSONB,
			export_url TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_nxd_report_runs_factory ON nxd.report_runs(factory_id)`,
		`CREATE TABLE IF NOT EXISTS nxd.report_templates (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			category TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			default_filters JSONB DEFAULT '{}',
			prompt_instructions TEXT,
			output_schema_version TEXT DEFAULT '1',
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS nxd.telemetry_rollup_1m (
			bucket_ts TIMESTAMPTZ NOT NULL,
			factory_id UUID NOT NULL,
			asset_id UUID NOT NULL,
			metric_key TEXT NOT NULL,
			avg_value DOUBLE PRECISION,
			min_value DOUBLE PRECISION,
			max_value DOUBLE PRECISION,
			samples BIGINT DEFAULT 0,
			status_counts JSONB DEFAULT '{}',
			PRIMARY KEY (bucket_ts, factory_id, asset_id, metric_key)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_nxd_rollup_asset_ts ON nxd.telemetry_rollup_1m(asset_id, bucket_ts DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_nxd_rollup_factory_ts ON nxd.telemetry_rollup_1m(factory_id, bucket_ts DESC)`,
		`CREATE TABLE IF NOT EXISTS nxd.rollup_ranges (
			factory_id UUID NOT NULL,
			asset_id UUID NOT NULL,
			metric_key TEXT NOT NULL,
			bucket_start TIMESTAMPTZ NOT NULL,
			bucket_end TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			PRIMARY KEY (factory_id, asset_id, metric_key, bucket_start)
		)`,
		`CREATE TABLE IF NOT EXISTS nxd.alert_rules (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			factory_id UUID NOT NULL REFERENCES nxd.factories(id) ON DELETE CASCADE,
			scope_type TEXT NOT NULL CHECK (scope_type IN ('group','asset')),
			scope_id UUID NOT NULL,
			condition_type TEXT NOT NULL,
			threshold DOUBLE PRECISION,
			channel TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_nxd_alert_rules_factory ON nxd.alert_rules(factory_id)`,
		`CREATE TABLE IF NOT EXISTS nxd.alerts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			ts TIMESTAMPTZ DEFAULT NOW(),
			rule_id UUID NOT NULL REFERENCES nxd.alert_rules(id) ON DELETE CASCADE,
			asset_id UUID,
			group_id UUID,
			severity TEXT NOT NULL,
			message TEXT NOT NULL,
			acknowledged_by TEXT,
			acknowledged_at TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS idx_nxd_alerts_rule ON nxd.alerts(rule_id)`,
		`CREATE INDEX IF NOT EXISTS idx_nxd_alerts_ts ON nxd.alerts(ts DESC)`,
		// Catálogo de sinais: métricas vistas por asset
		`CREATE TABLE IF NOT EXISTS nxd.asset_metric_catalog (
			factory_id UUID NOT NULL,
			asset_id UUID NOT NULL,
			metric_key TEXT NOT NULL,
			first_seen TIMESTAMPTZ DEFAULT NOW(),
			last_seen TIMESTAMPTZ DEFAULT NOW(),
			PRIMARY KEY (factory_id, asset_id, metric_key)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_nxd_metric_catalog_factory ON nxd.asset_metric_catalog(factory_id)`,
		`ALTER TABLE nxd.groups ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}'`,
	}
	for i, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration %d: %w", i+1, err)
		}
	}
	return nil
}

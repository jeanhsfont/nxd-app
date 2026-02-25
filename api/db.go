package api

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

var db *sql.DB

func InitDB() error {
	var err error
	databaseURL := os.Getenv("DATABASE_URL")
	driverName := "sqlite"
	connStr := "./hubsystem.db"

	if databaseURL != "" {
		driverName = "postgres"
		connStr = databaseURL
		log.Println("✓ Usando banco de dados PostgreSQL (produção).")
	} else {
		log.Println("✓ Usando banco de dados SQLite (desenvolvimento).")
	}

	db, err = sql.Open(driverName, connStr)
	if err != nil {
		return err
	}

	if err = db.Ping(); err != nil {
		return err
	}

	if driverName == "postgres" {
		if _, err := db.Exec("SET search_path TO public"); err != nil {
			log.Printf("⚠️  SET search_path falhou (ignorando): %v", err)
		}
	}

	log.Println("✓ Conexão com o banco de dados estabelecida com sucesso.")
	if err := runMigrations(driverName); err != nil {
		db.Close()
		db = nil
		return err
	}
	return nil
}

func GetDB() *sql.DB {
	return db
}

// EnsureAuthTables cria public.users e public.factories se não existirem (Postgres).
// Chamado em todo Register/Login para que, mesmo se InitDB falhou no cold start, a primeira
// requisição que conseguir conectar crie as tabelas. Sem sync.Once para permitir retry até dar certo.
func EnsureAuthTables() {
	if db == nil || os.Getenv("DATABASE_URL") == "" {
		return
	}
	for _, q := range []string{
		`CREATE TABLE IF NOT EXISTS public.users (
			id SERIAL PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			full_name TEXT,
			cpf TEXT,
			two_factor_enabled BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS public.factories (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL UNIQUE,
			name TEXT NOT NULL,
			cnpj TEXT,
			address TEXT,
			api_key_hash TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES public.users(id)
		)`,
		`ALTER TABLE public.factories ADD COLUMN IF NOT EXISTS subscription_plan TEXT DEFAULT 'free'`,
		`ALTER TABLE public.factories ADD COLUMN IF NOT EXISTS next_billing_date DATE`,
		`ALTER TABLE public.factories ADD COLUMN IF NOT EXISTS subscription_status TEXT DEFAULT 'active'`,
		`ALTER TABLE public.factories ADD COLUMN IF NOT EXISTS gateway_subscription_id TEXT`,
		`ALTER TABLE public.factories ADD COLUMN IF NOT EXISTS trial_ends_at TIMESTAMP WITH TIME ZONE`,
	} {
		if _, err := db.Exec(q); err != nil {
			log.Printf("[EnsureAuthTables] %v", err)
			return
		}
	}
}

func runMigrations(driverName string) error {
	var userTableSQL string
	if driverName == "postgres" {
		userTableSQL = `
		CREATE TABLE IF NOT EXISTS public.users (
			id SERIAL PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			full_name TEXT,
			cpf TEXT,
			two_factor_enabled BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`
	} else {
		userTableSQL = `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			full_name TEXT,
			cpf TEXT,
			two_factor_enabled BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`
	}

	if _, err := db.Exec(userTableSQL); err != nil {
		return err
	}

	log.Println("✓ Migração da tabela 'users' executada.")
	return createFactoryTable(driverName)
}

func createFactoryTable(driverName string) error {
	var factoryTableSQL string
	if driverName == "postgres" {
		factoryTableSQL = `
		CREATE TABLE IF NOT EXISTS public.factories (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL UNIQUE,
			name TEXT NOT NULL,
			cnpj TEXT,
			address TEXT,
			api_key_hash TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES public.users(id)
		);`
	} else {
		factoryTableSQL = `
		CREATE TABLE IF NOT EXISTS factories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			cnpj TEXT,
			address TEXT,
			api_key_hash TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id)
		);`
	}

	if _, err := db.Exec(factoryTableSQL); err != nil {
		return err
	}

	log.Println("✓ Migração da tabela 'factories' executada.")
	return createAssetAndTelemetryTables(driverName)
}

func createAssetAndTelemetryTables(driverName string) error {
	var sectorsTableSQL, assetsTableSQL, telemetryTableSQL string

	if driverName == "postgres" {
		sectorsTableSQL = `
		CREATE TABLE IF NOT EXISTS public.sectors (
			id SERIAL PRIMARY KEY,
			factory_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(factory_id) REFERENCES public.factories(id)
		);`

		assetsTableSQL = `
		CREATE TABLE IF NOT EXISTS public.assets (
			id SERIAL PRIMARY KEY,
			factory_id INTEGER NOT NULL,
			sector_id INTEGER,
			name TEXT NOT NULL,
			description TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(factory_id) REFERENCES public.factories(id),
			FOREIGN KEY(sector_id) REFERENCES public.sectors(id)
		);`

		telemetryTableSQL = `
		CREATE TABLE IF NOT EXISTS public.asset_telemetry (
			id BIGSERIAL PRIMARY KEY,
			asset_id INTEGER NOT NULL,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			payload JSONB NOT NULL,
			FOREIGN KEY(asset_id) REFERENCES public.assets(id)
		);`

	} else { // SQLite
		sectorsTableSQL = `
		CREATE TABLE IF NOT EXISTS sectors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			factory_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(factory_id) REFERENCES factories(id)
		);`

		assetsTableSQL = `
		CREATE TABLE IF NOT EXISTS assets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			factory_id INTEGER NOT NULL,
			sector_id INTEGER,
			name TEXT NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(factory_id) REFERENCES factories(id),
			FOREIGN KEY(sector_id) REFERENCES sectors(id)
		);`

		telemetryTableSQL = `
		CREATE TABLE IF NOT EXISTS asset_telemetry (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			asset_id INTEGER NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			payload TEXT NOT NULL, -- SQLite não tem JSON nativo, usamos TEXT
			FOREIGN KEY(asset_id) REFERENCES assets(id)
		);`
	}

	if _, err := db.Exec(sectorsTableSQL); err != nil {
		return err
	}
	log.Println("✓ Migração da tabela 'sectors' executada.")

	if _, err := db.Exec(assetsTableSQL); err != nil {
		return err
	}
	log.Println("✓ Migração da tabela 'assets' executada.")

	if _, err := db.Exec(telemetryTableSQL); err != nil {
		return err
	}
	log.Println("✓ Migração da tabela 'asset_telemetry' executada.")

	return ensureBillingAndSupportTables(driverName)
}

func ensureBillingAndSupportTables(driverName string) error {
	if driverName == "postgres" {
		for _, q := range []string{
			`ALTER TABLE public.factories ADD COLUMN IF NOT EXISTS subscription_plan TEXT DEFAULT 'free'`,
			`ALTER TABLE public.factories ADD COLUMN IF NOT EXISTS next_billing_date DATE`,
			`ALTER TABLE public.factories ADD COLUMN IF NOT EXISTS subscription_status TEXT DEFAULT 'active'`,
			`ALTER TABLE public.factories ADD COLUMN IF NOT EXISTS gateway_subscription_id TEXT`,
			`ALTER TABLE public.factories ADD COLUMN IF NOT EXISTS trial_ends_at TIMESTAMP WITH TIME ZONE`,
		} {
			if _, err := db.Exec(q); err != nil {
				log.Printf("⚠️  Billing migration (ignorável se já existe): %v", err)
			}
		}
		supportSQL := `
		CREATE TABLE IF NOT EXISTS public.support_tickets (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES public.users(id),
			email TEXT NOT NULL,
			subject TEXT NOT NULL,
			message TEXT NOT NULL,
			status TEXT DEFAULT 'open',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`
		if _, err := db.Exec(supportSQL); err != nil {
			return err
		}
	} else {
		// SQLite: ADD COLUMN não tem IF NOT EXISTS
		_, _ = db.Exec(`ALTER TABLE factories ADD COLUMN subscription_plan TEXT DEFAULT 'free'`)
		_, _ = db.Exec(`ALTER TABLE factories ADD COLUMN next_billing_date TEXT`)
		supportSQL := `
		CREATE TABLE IF NOT EXISTS support_tickets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id),
			email TEXT NOT NULL,
			subject TEXT NOT NULL,
			message TEXT NOT NULL,
			status TEXT DEFAULT 'open',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`
		if _, err := db.Exec(supportSQL); err != nil {
			return err
		}
	}
	log.Println("✓ Migração cobrança e suporte executada.")
	return ensureAuditLogAndRoleTables(driverName)
}

func ensureAuditLogAndRoleTables(driverName string) error {
	if driverName == "postgres" {
		for _, q := range []string{
			`CREATE TABLE IF NOT EXISTS public.audit_log (
				id BIGSERIAL PRIMARY KEY,
				user_id BIGINT NOT NULL,
				action TEXT NOT NULL,
				entity_type TEXT NOT NULL,
				entity_id TEXT,
				old_value TEXT,
				new_value TEXT,
				ip TEXT,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			)`,
			`ALTER TABLE public.users ADD COLUMN IF NOT EXISTS role TEXT DEFAULT 'operador'`,
		} {
			if _, err := db.Exec(q); err != nil {
				log.Printf("⚠️  audit/role migration: %v", err)
			}
		}
	} else {
		_, _ = db.Exec(`ALTER TABLE users ADD COLUMN role TEXT DEFAULT 'operador'`)
		_, _ = db.Exec(`
		CREATE TABLE IF NOT EXISTS audit_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			action TEXT NOT NULL,
			entity_type TEXT NOT NULL,
			entity_id TEXT,
			old_value TEXT,
			new_value TEXT,
			ip TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`)
	}
	log.Println("✓ Migração audit_log e role executada.")
	return ensureIAReportsTable(driverName)
}

func ensureIAReportsTable(driverName string) error {
	if driverName == "postgres" {
		_, _ = db.Exec(`
		CREATE TABLE IF NOT EXISTS public.ia_reports (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			factory_id TEXT,
			title TEXT NOT NULL,
			text_content TEXT NOT NULL,
			sources_json TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`)
	} else {
		_, _ = db.Exec(`
		CREATE TABLE IF NOT EXISTS ia_reports (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			factory_id TEXT,
			title TEXT NOT NULL,
			text_content TEXT NOT NULL,
			sources_json TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`)
	}
	log.Println("✓ Migração ia_reports executada.")
	return nil
}

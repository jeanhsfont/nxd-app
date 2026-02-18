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

	log.Println("✓ Conexão com o banco de dados estabelecida com sucesso.")
	return runMigrations(driverName)
}

func GetDB() *sql.DB {
	return db
}

func runMigrations(driverName string) error {
	var userTableSQL string
	if driverName == "postgres" {
		userTableSQL = `
		CREATE TABLE IF NOT EXISTS users (
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
		CREATE TABLE IF NOT EXISTS factories (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL UNIQUE,
			name TEXT NOT NULL,
			cnpj TEXT,
			address TEXT,
			api_key_hash TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id)
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
		CREATE TABLE IF NOT EXISTS sectors (
			id SERIAL PRIMARY KEY,
			factory_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(factory_id) REFERENCES factories(id)
		);`

		assetsTableSQL = `
		CREATE TABLE IF NOT EXISTS assets (
			id SERIAL PRIMARY KEY,
			factory_id INTEGER NOT NULL,
			sector_id INTEGER, -- Pode ser nulo se o ativo ainda não foi categorizado
			name TEXT NOT NULL,
			description TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(factory_id) REFERENCES factories(id),
			FOREIGN KEY(sector_id) REFERENCES sectors(id)
		);`

		telemetryTableSQL = `
		CREATE TABLE IF NOT EXISTS asset_telemetry (
			id BIGSERIAL PRIMARY KEY,
			asset_id INTEGER NOT NULL,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			payload JSONB NOT NULL, -- Usamos JSONB para flexibilidade nos dados de telemetria
			FOREIGN KEY(asset_id) REFERENCES assets(id)
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

	return nil
}

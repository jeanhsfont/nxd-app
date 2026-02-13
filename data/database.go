package data

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDatabase inicializa o banco de dados SQLite
func InitDatabase() error {
	// Garante que o diretório existe
	dbDir := filepath.Dir("./data/hubsystem.db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório do banco: %w", err)
	}

	var err error
	DB, err = sql.Open("sqlite3", "./data/hubsystem.db?cache=shared&mode=rwc")
	if err != nil {
		return fmt.Errorf("erro ao abrir banco de dados: %w", err)
	}

	// Configurações de performance
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	// Testa a conexão
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("erro ao conectar no banco: %w", err)
	}

	log.Println("✓ Banco de dados inicializado com sucesso")
	return nil
}

// RunMigrations executa as migrações do banco
func RunMigrations() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS factories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			api_key TEXT UNIQUE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			is_active BOOLEAN DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS machines (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			factory_id INTEGER NOT NULL,
			device_id TEXT NOT NULL,
			name TEXT,
			brand TEXT,
			protocol TEXT,
			last_seen DATETIME,
			status TEXT DEFAULT 'offline',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (factory_id) REFERENCES factories(id),
			UNIQUE(factory_id, device_id)
		)`,
		`CREATE TABLE IF NOT EXISTS tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			machine_id INTEGER NOT NULL,
			tag_name TEXT NOT NULL,
			tag_type TEXT NOT NULL,
			unit TEXT,
			min_value REAL,
			max_value REAL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (machine_id) REFERENCES machines(id),
			UNIQUE(machine_id, tag_name)
		)`,
		`CREATE TABLE IF NOT EXISTS data_points (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			machine_id INTEGER NOT NULL,
			tag_id INTEGER NOT NULL,
			value TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			quality TEXT DEFAULT 'good',
			FOREIGN KEY (machine_id) REFERENCES machines(id),
			FOREIGN KEY (tag_id) REFERENCES tags(id)
		)`,
		`CREATE TABLE IF NOT EXISTS alerts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tag_id INTEGER NOT NULL,
			condition TEXT NOT NULL,
			threshold REAL NOT NULL,
			message TEXT,
			is_active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (tag_id) REFERENCES tags(id)
		)`,
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			action TEXT NOT NULL,
			api_key TEXT,
			device_id TEXT,
			status TEXT NOT NULL,
			message TEXT,
			ip_address TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_data_points_timestamp ON data_points(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_data_points_machine ON data_points(machine_id)`,
		`CREATE INDEX IF NOT EXISTS idx_machines_factory ON machines(factory_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tags_machine ON tags(machine_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_logs(timestamp)`,
	}

	for i, migration := range migrations {
		if _, err := DB.Exec(migration); err != nil {
			return fmt.Errorf("erro na migração %d: %w", i+1, err)
		}
	}

	log.Println("✓ Migrações executadas com sucesso")
	return nil
}

// Close fecha a conexão com o banco
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

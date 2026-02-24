package store

import (
	"database/sql"
	"log"
	"os"
	"sync"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	nxdDB    *sql.DB
	nxdOnce  sync.Once
	dbDriver string
)

func NXDDB() *sql.DB {
	return nxdDB
}

// Driver retorna o driver do banco ("postgres" ou "sqlite3"). Usado para queries que dependem do schema (ex.: nxd.sectors no Postgres).
func Driver() string {
	return dbDriver
}

func InitNXDDB() error {
	var err error
	nxdOnce.Do(func() {
		// Tenta NXD_DATABASE_URL primeiro, depois DATABASE_URL (mesmo Postgres da API legada)
		connURL := os.Getenv("NXD_DATABASE_URL")
		if connURL == "" {
			connURL = os.Getenv("DATABASE_URL")
		}
		if connURL != "" {
			// Modo Produção: PostgreSQL
			dbDriver = "postgres"
			log.Println("✓ NXD store: usando banco de dados PostgreSQL.")
			nxdDB, err = sql.Open(dbDriver, connURL)
		} else {
			// Modo Desenvolvimento: SQLite
			dbDriver = "sqlite3"
			log.Println("✓ NXD store: usando banco de dados SQLite (desenvolvimento).")
			nxdDB, err = sql.Open(dbDriver, "./nxd.db")
		}

		if err != nil {
			return
		}

		if err = nxdDB.Ping(); err != nil {
			log.Printf("❌ Erro ao conectar com o banco de dados: %v", err)
			nxdDB.Close()
			nxdDB = nil
			return
		}
		log.Println("✓ Conexão com o banco de dados estabelecida com sucesso.")

		if err = RunMigrations(nxdDB, dbDriver); err != nil {
			log.Printf("❌ Erro ao executar migrações: %v", err)
			nxdDB.Close()
			nxdDB = nil
			return
		}
	})
	return err
}

func CloseNXDDB() {
	if nxdDB != nil {
		nxdDB.Close()
		nxdDB = nil
	}
}



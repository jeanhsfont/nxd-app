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

func InitNXDDB() error {
	var err error
	nxdOnce.Do(func() {
		connURL := os.Getenv("NXD_DATABASE_URL")
		if connURL != "" {
			// Modo Produção: PostgreSQL
			dbDriver = "postgres"
			log.Println("✓ Usando banco de dados PostgreSQL (produção).")
			nxdDB, err = sql.Open(dbDriver, connURL)
		} else {
			// Modo Desenvolvimento: SQLite
			dbDriver = "sqlite3"
			log.Println("✓ Usando banco de dados SQLite (desenvolvimento).")
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



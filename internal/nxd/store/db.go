package store

import (
	"database/sql"
	"log"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

var (
	nxdDB   *sql.DB
	nxdOnce sync.Once
)

// NXDDB returns the Postgres connection for NXD. Returns nil if NXD_DATABASE_URL is not set.
func NXDDB() *sql.DB {
	return nxdDB
}

// InitNXDDB opens Postgres and runs migrations when NXD_DATABASE_URL is set.
// Safe to call multiple times; only the first call does work.
func InitNXDDB() error {
	var err error
	nxdOnce.Do(func() {
		connURL := os.Getenv("NXD_DATABASE_URL")
		if connURL == "" {
			log.Println("NXD: NXD_DATABASE_URL not set; /nxd/* stack disabled")
			return
		}
		nxdDB, err = sql.Open("postgres", connURL)
		if err != nil {
			return
		}
		if err = nxdDB.Ping(); err != nil {
			nxdDB.Close()
			nxdDB = nil
			return
		}
		if err = RunMigrations(nxdDB); err != nil {
			nxdDB.Close()
			nxdDB = nil
			return
		}
		if err = SeedReportTemplates(nxdDB); err != nil {
			log.Printf("NXD: seed report_templates warning: %v", err)
		}
		log.Println("NXD: Postgres connected and migrations applied")
	})
	return err
}

// CloseNXDDB closes the NXD Postgres connection. Safe to call if not initialized.
func CloseNXDDB() {
	if nxdDB != nil {
		nxdDB.Close()
		nxdDB = nil
	}
}

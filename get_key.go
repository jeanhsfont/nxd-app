//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./data/nxd.db")
	if err != nil {
		log.Fatalf("Erro ao abrir o banco de dados: %v", err)
	}
	defer db.Close()

	var apiKey string
	err = db.QueryRow("SELECT api_key FROM factories LIMIT 1").Scan(&apiKey)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Nenhuma fábrica encontrada no banco de dados.")
			return
		}
		log.Fatalf("Erro ao consultar a api_key: %v", err)
	}

	fmt.Println(apiKey)
}

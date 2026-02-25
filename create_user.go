
//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	db, err := sql.Open("sqlite3", "./nxd.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	password := "password"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	userID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
	email := "test@test.com"
	now := time.Now().UTC()

	_, err = db.Exec(`
		INSERT INTO users (id, email, password_hash, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(email) DO NOTHING;
	`, userID, email, string(hashedPassword), now, now)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Usuário de teste criado (ou já existente).")
}

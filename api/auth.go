package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret []byte

func InitAuth() {
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("super-secret-key-for-local-dev")
	}
}

type User struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
}

type jwtClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Corpo da requisição inválido"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		http.Error(w, `{"error":"Nome, email e senha são obrigatórios"}`, http.StatusBadRequest)
		return
	}

	_, err := GetUserByEmail(req.Email)
	if err == nil {
		http.Error(w, `{"error":"Este email já está em uso"}`, http.StatusConflict)
		return
	}
	if err != sql.ErrNoRows {
		http.Error(w, `{"error":"Erro ao verificar usuário"}`, http.StatusInternalServerError)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error":"Falha ao processar senha"}`, http.StatusInternalServerError)
		return
	}

	if _, err := CreateUser(req.Email, string(passwordHash), req.Name); err != nil {
		http.Error(w, `{"error":"Falha ao criar usuário"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Usuário criado com sucesso"})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Corpo da requisição inválido"}`, http.StatusBadRequest)
		return
	}

	user, err := GetUserByEmail(req.Email)
	if err != nil {
		http.Error(w, `{"error":"Credenciais inválidas"}`, http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, `{"error":"Credenciais inválidas"}`, http.StatusUnauthorized)
		return
	}

	claims := jwtClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := jwtToken.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, `{"error":"Falha ao gerar token"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenStr})
}

func GetUserByEmail(email string) (*User, error) {
	db := GetDB()
	row := db.QueryRow("SELECT id, email, password_hash FROM users WHERE email = $1", email)
	user := &User{}
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func CreateUser(email, passwordHash, name string) (int64, error) {
	db := GetDB()
	var id int64
	err := db.QueryRow(
		"INSERT INTO users (email, password_hash, full_name) VALUES ($1, $2, $3) RETURNING id",
		email, passwordHash, name,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

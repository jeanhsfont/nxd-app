package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret []byte

func InitAuth() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Em produção (DATABASE_URL definida = Cloud Run), recusar o default inseguro.
		// Em desenvolvimento local (sem DATABASE_URL), aceitar o default para facilitar setup.
		if os.Getenv("DATABASE_URL") != "" {
			// Produção detectada — sem JWT_SECRET é fatal.
			panic("FATAL: JWT_SECRET não definida em produção. Configure a variável de ambiente JWT_SECRET no Cloud Run.")
		}
		secret = "super-secret-key-for-local-dev"
		log.Printf("⚠️  JWT_SECRET não definida — usando valor default APENAS para desenvolvimento local. NUNCA use isso em produção.")
	}
	if len(secret) < 32 {
		log.Printf("⚠️  JWT_SECRET parece curta (%d chars). Recomendado: 64+ chars hex aleatórios.", len(secret))
	}
	jwtSecret = []byte(secret)
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

// UserFromRequest extracts the authenticated User from the request context.
// Returns nil if the user is not authenticated or the user ID is not set.
// Used by rbac.go middleware.
func UserFromRequest(r *http.Request) *User {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok || userID == 0 {
		return nil
	}
	db := GetDB()
	if db == nil {
		return nil
	}
	user := &User{}
	err := db.QueryRow("SELECT id, email, password_hash FROM users WHERE id = $1", userID).
		Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil
	}
	return user
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

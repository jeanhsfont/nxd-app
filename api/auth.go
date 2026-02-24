package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrDBNotReady = errors.New("database not initialized")

var jwtSecret []byte

func InitAuth() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		if os.Getenv("DATABASE_URL") != "" {
			// Produção: não fazer panic para permitir que o servidor suba e os logs apareçam; usar placeholder.
			log.Printf("⚠️  JWT_SECRET não definida em produção — usando placeholder. Configure JWT_SECRET no Cloud Run.")
			secret = "placeholder-configure-jwt-secret-in-cloud-run-" + time.Now().Format("20060102")
		} else {
			secret = "super-secret-key-for-local-dev"
			log.Printf("⚠️  JWT_SECRET não definida — usando valor default APENAS para desenvolvimento local.")
		}
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
	EnsureAuthTables()
	log.Printf("[Register] request start")
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[Register] decode error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Corpo da requisição inválido"}`))
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Nome, email e senha são obrigatórios"}`))
		return
	}

	_, err := GetUserByEmail(req.Email)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error":"Este email já está em uso"}`))
		return
	}
	if !errors.Is(err, sql.ErrNoRows) {
		if errors.Is(err, ErrDBNotReady) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"Sistema temporariamente indisponível. Tente em instantes."}`))
			return
		}
		log.Printf("[Register] GetUserByEmail falhou: %v", err)
		errStr := err.Error()
		if strings.Contains(errStr, "does not exist") || strings.Contains(errStr, "não existe") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"Banco ainda inicializando. Aguarde 30 segundos e tente novamente."}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Erro ao processar cadastro. Tente novamente em instantes."}`))
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Falha ao processar senha"}`))
		return
	}

	if _, err := CreateUser(req.Email, string(passwordHash), req.Name); err != nil {
		if errors.Is(err, ErrDBNotReady) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"Sistema temporariamente indisponível. Tente em instantes."}`))
			return
		}
		log.Printf("[Register] CreateUser falhou: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Falha ao criar usuário"}`))
		return
	}

	log.Printf("[Register] usuário criado: %s", req.Email)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Usuário criado com sucesso"})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	EnsureAuthTables()
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Corpo da requisição inválido"}`))
		return
	}

	user, err := GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, ErrDBNotReady) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"Sistema temporariamente indisponível. Tente em instantes."}`))
			return
		}
		errStr := err.Error()
		if strings.Contains(errStr, "does not exist") || strings.Contains(errStr, "não existe") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"Banco ainda inicializando. Aguarde 30 segundos e tente novamente."}`))
			return
		}
		// Usuário não encontrado (sql.ErrNoRows) ou outro erro → credenciais inválidas
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Credenciais inválidas"}`))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Credenciais inválidas"}`))
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Falha ao gerar token"}`))
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
	err := db.QueryRow("SELECT id, email, password_hash FROM public.users WHERE id = $1", userID).
		Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil
	}
	return user
}

func GetUserByEmail(email string) (*User, error) {
	db := GetDB()
	if db == nil {
		return nil, ErrDBNotReady
	}
	row := db.QueryRow("SELECT id, email, password_hash FROM public.users WHERE email = $1", email)
	user := &User{}
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func CreateUser(email, passwordHash, name string) (int64, error) {
	db := GetDB()
	if db == nil {
		return 0, ErrDBNotReady
	}
	var id int64
	err := db.QueryRow(
		"INSERT INTO public.users (email, password_hash, full_name) VALUES ($1, $2, $3) RETURNING id",
		email, passwordHash, name,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

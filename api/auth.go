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
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

var ErrDBNotReady = errors.New("database not initialized")

var jwtSecret []byte

func InitAuth() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		if os.Getenv("DATABASE_URL") != "" {
			log.Fatal("JWT_SECRET é obrigatório em produção. Configure a variável de ambiente no Cloud Run (ou onde o serviço roda).")
		}
		secret = "super-secret-key-for-local-dev"
		log.Printf("⚠️  JWT_SECRET não definida — usando valor default APENAS para desenvolvimento local.")
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
	UserID  int64  `json:"user_id"`
	Purpose string `json:"purpose,omitempty"` // "2fa_pending" = token temporário para concluir login 2FA
	jwt.RegisteredClaims
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	db := GetDB()
	if db == nil {
		log.Printf("[Register] banco da API indisponível (db nil)")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error":"Sistema temporariamente indisponível. Tente em instantes."}`))
		return
	}
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
		errStr := err.Error()
		log.Printf("[Register] GetUserByEmail falhou: %s", errStr)
		if strings.Contains(errStr, "does not exist") || strings.Contains(errStr, "não existe") || strings.Contains(errStr, "not exist") {
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
	if GetDB() == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error":"Sistema temporariamente indisponível. Tente em instantes."}`))
		return
	}
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

	// Se 2FA está ativo, retornar token temporário para conclusão do login
	db := GetDB()
	if db != nil {
		var totpEnabled bool
		_ = db.QueryRow("SELECT COALESCE(totp_enabled, FALSE) FROM users WHERE id = $1", user.ID).Scan(&totpEnabled)
		if totpEnabled {
			claims := jwtClaims{
				UserID:  user.ID,
				Purpose: "2fa_pending",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
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
			json.NewEncoder(w).Encode(map[string]interface{}{
				"requires_2fa": true,
				"temp_token":   tokenStr,
			})
			return
		}
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

// Login2FAConfirmHandler — POST /api/login/2fa — troca temp_token + código TOTP por JWT definitivo.
func Login2FAConfirmHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Token temporário necessário"}`))
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) { return jwtSecret, nil })
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Token inválido ou expirado"}`))
		return
	}
	claims, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid || claims.Purpose != "2fa_pending" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Token inválido ou expirado"}`))
		return
	}
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Código do app autenticador é obrigatório"}`))
		return
	}
	db := GetDB()
	if db == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error":"Sistema indisponível"}`))
		return
	}
	var secret string
	if err := db.QueryRow("SELECT totp_secret FROM users WHERE id = $1 AND totp_enabled = TRUE", claims.UserID).Scan(&secret); err != nil || secret == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"2FA não está ativo para este usuário"}`))
		return
	}
	if !totp.Validate(req.Code, secret) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Código inválido. Tente novamente."}`))
		return
	}
	// Emite JWT definitivo
	fullClaims := jwtClaims{
		UserID: claims.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, fullClaims)
	fullToken, err := jwtToken.SignedString(jwtSecret)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Falha ao gerar token"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": fullToken})
}

// userHasRole retorna true se o usuário tem a role indicada (ex: "admin").
func userHasRole(userID int64, role string) bool {
	db := GetDB()
	if db == nil {
		return false
	}
	var r string
	err := db.QueryRow("SELECT COALESCE(role, 'operador') FROM users WHERE id = $1", userID).Scan(&r)
	if err != nil {
		return false
	}
	return r == role
}

// MeHandler — GET /api/me — retorna dados do usuário autenticado (incluindo role para RBAC no frontend).
func MeHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	db := GetDB()
	if db == nil {
		http.Error(w, "Indisponível", http.StatusServiceUnavailable)
		return
	}
	var email, fullName, role string
	err := db.QueryRow("SELECT email, COALESCE(full_name,''), COALESCE(role,'operador') FROM users WHERE id = $1", userID).Scan(&email, &fullName, &role)
	if err != nil {
		http.Error(w, "Usuário não encontrado", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":   userID,
		"email":     email,
		"full_name": fullName,
		"role":      role,
	})
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
	if db == nil {
		return nil, ErrDBNotReady
	}
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
	if db == nil {
		return 0, ErrDBNotReady
	}
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

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"hubsystem/core"
	"hubsystem/data"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	cookieName   = "nxd_session"
	cookieMaxAge = 7 * 24 * 3600 // 7 dias
)

type contextKey string

const userContextKey contextKey = "user"

var (
	googleOAuthConfig *oauth2.Config
	jwtSecret         []byte
	baseURL           string
)

func init() {
	initAuth()
}

func initAuth() {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return
	}
	baseURL = os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("nxd-dev-secret-change-in-production")
	}
	googleOAuthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  baseURL + "/api/auth/google/callback",
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

// Claims do JWT
type jwtClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// GoogleLoginHandler redireciona para o Google OAuth
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	if googleOAuthConfig == nil {
		http.Error(w, "Google Auth não configurado (GOOGLE_CLIENT_ID/SECRET)", http.StatusServiceUnavailable)
		return
	}
	state, _ := core.GenerateAPIKey()
	if state == "" {
		state = "nxd_oauth_state"
	}
	// Em produção guardar state em cookie/cache e validar no callback
	url := googleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "select_account"))
	http.Redirect(w, r, url, http.StatusFound)
}

// GoogleCallbackHandler troca o code por token e cria sessão
func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if googleOAuthConfig == nil {
		http.Error(w, "Google Auth não configurado", http.StatusServiceUnavailable)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, "/#login?error=no_code", http.StatusFound)
		return
	}
	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Redirect(w, r, "/#login?error=exchange", http.StatusFound)
		return
	}
	client := googleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Redirect(w, r, "/#login?error=userinfo", http.StatusFound)
		return
	}
	defer resp.Body.Close()
	var info struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		http.Redirect(w, r, "/#login?error=decode", http.StatusFound)
		return
	}
	if info.Email == "" {
		info.Email = info.ID + "@google"
	}
	if info.Name == "" {
		info.Name = info.Email
	}

	user, err := data.GetUserByGoogleUID(info.ID)
	if err != nil {
		http.Redirect(w, r, "/#login?error=db", http.StatusFound)
		return
	}
	if user == nil {
		userID, err := data.CreateUser(info.Email, info.Name, info.ID, nil)
		if err != nil {
			http.Redirect(w, r, "/#login?error=create_user", http.StatusFound)
			return
		}
		user = &core.User{ID: userID, Email: info.Email, Name: info.Name, GoogleUID: info.ID, CreatedAt: time.Now()}
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
		http.Redirect(w, r, "/#login?error=jwt", http.StatusFound)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    tokenStr,
		Path:     "/",
		MaxAge:   cookieMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   baseURL != "" && baseURL[0:5] == "https",
	})
	// Redirecionar para o app (frontend)
	frontURL := os.Getenv("FRONTEND_URL")
	if frontURL == "" {
		frontURL = baseURL
	}
	http.Redirect(w, r, frontURL+"/#/dashboard", http.StatusFound)
}

// AuthMeHandler retorna o usuário atual e a fábrica (se houver)
func AuthMeHandler(w http.ResponseWriter, r *http.Request) {
	user := UserFromRequest(r)
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "não autenticado"})
		return
	}
	factory, _ := data.GetFactoryByUserID(user.ID)
	res := map[string]interface{}{
		"user":    user,
		"factory": nil,
	}
	if factory != nil {
		// Não enviar api_key em listagem
		f := *factory
		f.APIKey = ""
		res["factory"] = &f
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// UserFromRequest extrai o user do JWT no cookie ou Authorization
func UserFromRequest(r *http.Request) *core.User {
	tokenStr := ""
	if c, _ := r.Cookie(cookieName); c != nil {
		tokenStr = c.Value
	}
	if tokenStr == "" {
		if b := r.Header.Get("Authorization"); len(b) > 7 && b[:7] == "Bearer " {
			tokenStr = b[7:]
		}
	}
	if tokenStr == "" {
		return nil
	}
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil
	}
	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		return nil
	}
	user, _ := data.GetUserByID(claims.UserID)
	return user
}

// RequireAuth é um middleware que exige usuário logado
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if UserFromRequest(r) == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "faça login"})
			return
		}
		next(w, r)
	}
}


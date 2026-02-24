package api

import (
	"encoding/json"
	"image/png"
	"net/http"

	"github.com/pquerna/otp/totp"
)

// SetupTOTPHandler gera um segredo TOTP e retorna o QR code como PNG
func SetupTOTPHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	// Busca email do usuário para label do QR code
	email := "usuario@nxd"
	row := db.QueryRow("SELECT email FROM users WHERE id = $1", userID)
	row.Scan(&email)

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "NXD HubSystem",
		AccountName: email,
	})
	if err != nil {
		http.Error(w, "Erro ao gerar segredo TOTP", http.StatusInternalServerError)
		return
	}

	// Salva o segredo temporariamente (pending confirmação)
	_, err = db.Exec("UPDATE users SET totp_secret_pending = $1 WHERE id = $2", key.Secret(), userID)
	if err != nil {
		// Coluna pode não existir ainda — tenta criar
		db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_secret_pending TEXT")
		db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_secret TEXT")
		db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_enabled BOOLEAN DEFAULT FALSE")
		db.Exec("UPDATE users SET totp_secret_pending = $1 WHERE id = $2", key.Secret(), userID)
	}

	if r.URL.Query().Get("format") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"secret": key.Secret(),
			"url":    key.URL(),
		})
		return
	}

	// Retorna QR code como imagem PNG
	img, err := key.Image(256, 256)
	if err != nil {
		http.Error(w, "Erro ao gerar QR code", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	png.Encode(w, img)
}

// ConfirmTOTPHandler confirma e ativa o 2FA com um código do app autenticador
func ConfirmTOTPHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		http.Error(w, "Código inválido", http.StatusBadRequest)
		return
	}

	// Busca o segredo pendente
	var secret string
	err := db.QueryRow("SELECT totp_secret_pending FROM users WHERE id = $1", userID).Scan(&secret)
	if err != nil || secret == "" {
		http.Error(w, "Nenhum segredo 2FA pendente. Configure primeiro.", http.StatusBadRequest)
		return
	}

	if !totp.Validate(req.Code, secret) {
		http.Error(w, "Código inválido. Tente novamente.", http.StatusUnauthorized)
		return
	}

	// Ativa o 2FA
	_, err = db.Exec("UPDATE users SET totp_secret = $1, totp_enabled = TRUE, totp_secret_pending = NULL WHERE id = $2", secret, userID)
	if err != nil {
		http.Error(w, "Erro ao ativar 2FA", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "2FA ativado com sucesso!"})
}

// DisableTOTPHandler desativa o 2FA
func DisableTOTPHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	var secret string
	var enabled bool
	db.QueryRow("SELECT totp_secret, totp_enabled FROM users WHERE id = $1", userID).Scan(&secret, &enabled)

	if enabled && secret != "" {
		if !totp.Validate(req.Code, secret) {
			http.Error(w, "Código inválido", http.StatusUnauthorized)
			return
		}
	}

	db.Exec("UPDATE users SET totp_secret = NULL, totp_enabled = FALSE WHERE id = $1", userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "2FA desativado."})
}

// TOTPStatusHandler retorna se o 2FA está ativo para o usuário
func TOTPStatusHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	var enabled bool
	db.QueryRow("SELECT COALESCE(totp_enabled, FALSE) FROM users WHERE id = $1", userID).Scan(&enabled)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"enabled": enabled})
}

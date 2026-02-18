package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func GenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	// Extrai o ID do usuário do contexto da requisição, injetado pelo middleware
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "ID de usuário inválido no token", http.StatusUnauthorized)
		return
	}

	// Gerar uma chave de API aleatória
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		http.Error(w, "Falha ao gerar a chave de API", http.StatusInternalServerError)
		return
	}
	apiKey := hex.EncodeToString(bytes)

	// Gerar o hash da chave de API para armazenamento seguro
	apiKeyHash, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Falha ao proteger a chave de API", http.StatusInternalServerError)
		return
	}

	// Atualiza a api_key_hash para o usuário/fábrica correspondente
	_, err = db.Exec("UPDATE factories SET api_key_hash = ? WHERE user_id = ?", string(apiKeyHash), userID)
	if err != nil {
		http.Error(w, "Falha ao salvar a chave de API", http.StatusInternalServerError)
		return
	}

	// Retornar a chave de API original para o usuário (APENAS DESTA VEZ)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"apiKey": apiKey})
}

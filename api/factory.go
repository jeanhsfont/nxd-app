package api

import (
	"encoding/json"
	"hubsystem/core"
	"hubsystem/internal/nxd/store"
	"net/http"
)

func GenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	// Extrai o ID do usuário do contexto da requisição, injetado pelo middleware
	_, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "ID de usuário inválido no token", http.StatusUnauthorized)
		return
	}

	// Gera chave no formato NXD_ (68 chars) + hash bcrypt — formato exigido pelo IngestHandler
	apiKey, apiKeyHash, err := core.GenerateAndHashAPIKey()
	if err != nil {
		http.Error(w, "Falha ao gerar a chave de API", http.StatusInternalServerError)
		return
	}

	// Salva na nxd.factories — tabela consultada pelo IngestHandler para autenticar o DX Simulator
	nxdDB := store.NXDDB()
	_, err = nxdDB.Exec(`
		INSERT INTO nxd.factories (name, api_key_hash)
		VALUES ('Minha Fábrica', $1)
	`, apiKeyHash)
	if err != nil {
		http.Error(w, "Falha ao salvar a chave de API", http.StatusInternalServerError)
		return
	}

	// Retornar a chave original para o usuário (APENAS DESTA VEZ)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"apiKey": apiKey})
}

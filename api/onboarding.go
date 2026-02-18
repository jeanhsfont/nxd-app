package api

import (
	"encoding/json"
	"hubsystem/core"
	"net/http"
)

type OnboardingData struct {
	PersonalData struct {
		FullName string `json:"fullName"`
		CPF      string `json:"cpf"`
	} `json:"personalData"`
	FactoryData struct {
		Name    string `json:"name"`
		CNPJ    string `json:"cnpj"`
		Address string `json:"address"`
	} `json:"factoryData"`
	TwoFactorEnabled bool `json:"twoFactorEnabled"`
}

func OnboardingHandler(w http.ResponseWriter, r *http.Request) {
	// Extrai o ID do usuário do contexto da requisição, injetado pelo middleware
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "ID de usuário inválido no token", http.StatusUnauthorized)
		return
	}

	var data OnboardingData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	// Iniciar uma transação para garantir a atomicidade
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Falha ao iniciar a operação no banco de dados", http.StatusInternalServerError)
		return
	}

	// Atualizar a tabela 'users'
	_, err = tx.Exec("UPDATE users SET full_name = $1, cpf = $2, two_factor_enabled = $3 WHERE id = $4",
		data.PersonalData.FullName, data.PersonalData.CPF, data.TwoFactorEnabled, userID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Falha ao salvar dados pessoais", http.StatusInternalServerError)
		return
	}

	// Gerar a chave de API
	apiKey, apiKeyHash, err := core.GenerateAndHashAPIKey()
	if err != nil {
		tx.Rollback()
		http.Error(w, "Falha ao gerar a chave de API", http.StatusInternalServerError)
		return
	}

	// Inserir ou atualizar a tabela 'factories' com a chave
	_, err = tx.Exec(`
		INSERT INTO factories (user_id, name, cnpj, address, api_key_hash) 
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT(user_id) DO UPDATE SET
		name = EXCLUDED.name,
		cnpj = EXCLUDED.cnpj,
		address = EXCLUDED.address,
		api_key_hash = EXCLUDED.api_key_hash
	`, userID, data.FactoryData.Name, data.FactoryData.CNPJ, data.FactoryData.Address, string(apiKeyHash))
	if err != nil {
		tx.Rollback()
		http.Error(w, "Falha ao salvar dados da fábrica", http.StatusInternalServerError)
		return
	}

	// Se tudo deu certo, comitar a transação
	if err := tx.Commit(); err != nil {
		http.Error(w, "Falha ao finalizar a operação", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Cadastro finalizado com sucesso!",
		"apiKey":  apiKey,
	})
}

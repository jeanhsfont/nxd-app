package api

import (
	"encoding/json"
	"hubsystem/core"
	"hubsystem/internal/nxd/store"
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
	// Campos simplificados para onboarding rápido (frontend pode enviar só factory_name)
	FactoryName string `json:"factory_name"`
	FullName    string `json:"full_name"`
}

func OnboardingHandler(w http.ResponseWriter, r *http.Request) {
	// Extrai o ID do usuário do contexto da requisição, injetado pelo middleware
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "ID de usuário inválido no token", http.StatusUnauthorized)
		return
	}

	var data OnboardingData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	// Normaliza campos simplificados para a struct aninhada
	if data.FactoryName != "" && data.FactoryData.Name == "" {
		data.FactoryData.Name = data.FactoryName
	}
	if data.FullName != "" && data.PersonalData.FullName == "" {
		data.PersonalData.FullName = data.FullName
	}
	// Garante nome padrão se nada foi enviado
	if data.FactoryData.Name == "" {
		data.FactoryData.Name = "Minha Fábrica"
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

	// Também salva na nxd.factories vinculando ao nxd.user pelo email,
	// para que IngestHandler, getFactoryIDForUser e CreateSectorHandler funcionem.
	if nxdDB := store.NXDDB(); nxdDB != nil {
		// Buscar email do usuário legado
		var email string
		db.QueryRow("SELECT email FROM users WHERE id = $1", userID).Scan(&email)

		if email != "" {
			// Garantir que o nxd.user existe
			var nxdUserID string
			err := nxdDB.QueryRow(
				`INSERT INTO nxd.users (name, email, password_hash)
				 VALUES ($1, $2, 'onboarding-migrated')
				 ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name
				 RETURNING id`,
				data.PersonalData.FullName, email,
			).Scan(&nxdUserID)
			if err == nil && nxdUserID != "" {
				// Verificar se já existe factory para este user_id
			var existingFactoryID string
			nxdDB.QueryRow(`SELECT id FROM nxd.factories WHERE user_id = $1 LIMIT 1`, nxdUserID).Scan(&existingFactoryID)
			if existingFactoryID != "" {
				// Atualizar factory existente
				nxdDB.Exec(`
					UPDATE nxd.factories SET name = $1, api_key_hash = $2, updated_at = NOW()
					WHERE id = $3
				`, data.FactoryData.Name, string(apiKeyHash), existingFactoryID)
			} else {
				// Inserir nova factory vinculada ao nxd.user
				nxdDB.Exec(`
					INSERT INTO nxd.factories (user_id, name, api_key_hash)
					VALUES ($1, $2, $3)
				`, nxdUserID, data.FactoryData.Name, string(apiKeyHash))
			}
			}
		} else {
			// Fallback sem email: insert sem user_id (comportamento anterior)
			nxdDB.Exec(`
				INSERT INTO nxd.factories (name, api_key_hash)
				VALUES ($1, $2)
			`, data.FactoryData.Name, string(apiKeyHash))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Cadastro finalizado com sucesso!",
		"apiKey":  apiKey,
	})
}

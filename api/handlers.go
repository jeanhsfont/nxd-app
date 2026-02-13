package api

import (
	"encoding/json"
	"fmt"
	"hubsystem/core"
	"hubsystem/data"
	"hubsystem/services"
	"log"
	"net/http"
	"strings"
	"time"
)

// IngestHandler recebe dados do DX (endpoint principal)
func IngestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Pega IP do cliente
	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = strings.Split(forwarded, ",")[0]
	}

	// Decodifica payload
	var payload core.IngestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		services.LogError("INGEST", "", "", "Payload inválido", ipAddress)
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}

	// Valida API Key
	if !core.ValidateAPIKey(payload.APIKey) {
		services.LogError("INGEST", payload.APIKey, payload.DeviceID, "API Key inválida", ipAddress)
		http.Error(w, "API Key inválida", http.StatusUnauthorized)
		return
	}

	// Busca fábrica
	factory, err := data.GetFactoryByAPIKey(payload.APIKey)
	if err != nil {
		services.LogError("INGEST", payload.APIKey, payload.DeviceID, "Erro ao buscar fábrica", ipAddress)
		http.Error(w, "Erro interno", http.StatusInternalServerError)
		return
	}
	if factory == nil || !factory.IsActive {
		services.LogError("INGEST", payload.APIKey, payload.DeviceID, "Fábrica não encontrada ou inativa", ipAddress)
		http.Error(w, "Fábrica não autorizada", http.StatusUnauthorized)
		return
	}

	// Sanitiza device ID
	deviceID := core.SanitizeDeviceID(payload.DeviceID)

	// Auto-discovery: Cria ou busca máquina
	machine, err := data.GetOrCreateMachine(factory.ID, deviceID, payload.Brand, payload.Protocol)
	if err != nil {
		services.LogError("INGEST", payload.APIKey, deviceID, fmt.Sprintf("Erro ao criar/buscar máquina: %v", err), ipAddress)
		http.Error(w, "Erro ao processar máquina", http.StatusInternalServerError)
		return
	}

	log.Printf("📥 [INGEST] Fábrica: %s | Máquina: %s (%s) | Tags: %d", 
		factory.Name, machine.Name, machine.Brand, len(payload.Tags))

	// Processa cada tag (auto-discovery)
	timestamp := payload.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	for tagName, tagValue := range payload.Tags {
		// Detecta tipo do valor
		tagType := detectType(tagValue)

		// Auto-discovery: Cria ou busca tag
		tag, err := data.GetOrCreateTag(machine.ID, tagName, tagType)
		if err != nil {
			log.Printf("❌ Erro ao criar/buscar tag %s: %v", tagName, err)
			continue
		}

		// Converte valor para string
		valueStr := fmt.Sprintf("%v", tagValue)

		// Insere data point
		if err := data.InsertDataPoint(machine.ID, tag.ID, valueStr, timestamp); err != nil {
			log.Printf("❌ Erro ao inserir data point para tag %s: %v", tagName, err)
			continue
		}

		log.Printf("  ✓ Tag: %s = %v (%s)", tagName, tagValue, tagType)
	}

	// Log de auditoria
	services.LogSuccess("INGEST", payload.APIKey, deviceID, 
		fmt.Sprintf("Processadas %d tags", len(payload.Tags)), ipAddress)

	// Notifica WebSocket
	services.BroadcastUpdate(machine.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"machine_id": machine.ID,
		"tags_count": len(payload.Tags),
	})
}

// CreateFactoryHandler cria uma nova fábrica
func CreateFactoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Nome é obrigatório", http.StatusBadRequest)
		return
	}

	// Gera API Key
	apiKey, err := core.GenerateAPIKey()
	if err != nil {
		http.Error(w, "Erro ao gerar API Key", http.StatusInternalServerError)
		return
	}

	// Cria fábrica
	factoryID, err := data.CreateFactory(req.Name, apiKey)
	if err != nil {
		http.Error(w, "Erro ao criar fábrica", http.StatusInternalServerError)
		return
	}

	log.Printf("🏭 Nova fábrica criada: %s (ID: %d)", req.Name, factoryID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      factoryID,
		"name":    req.Name,
		"api_key": apiKey,
	})
}

// GetDashboardHandler retorna dados do dashboard
func GetDashboardHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := r.URL.Query().Get("api_key")
	if apiKey == "" {
		http.Error(w, "API Key não fornecida", http.StatusBadRequest)
		return
	}

	// Busca fábrica
	factory, err := data.GetFactoryByAPIKey(apiKey)
	if err != nil || factory == nil {
		http.Error(w, "Fábrica não encontrada", http.StatusNotFound)
		return
	}

	// Busca máquinas
	machines, err := data.GetMachinesByFactory(factory.ID)
	if err != nil {
		http.Error(w, "Erro ao buscar máquinas", http.StatusInternalServerError)
		return
	}

	// Para cada máquina, busca tags e últimos valores
	type MachineData struct {
		core.Machine
		Tags   []core.Tag        `json:"tags"`
		Values map[string]string `json:"values"`
	}

	var result []MachineData
	for _, machine := range machines {
		tags, _ := data.GetTagsByMachine(machine.ID)
		values, _ := data.GetLatestDataPoints(machine.ID)
		
		result = append(result, MachineData{
			Machine: machine,
			Tags:    tags,
			Values:  values,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"factory":  factory,
		"machines": result,
	})
}

// HealthHandler verifica se o servidor está rodando
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "online",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// detectType detecta o tipo de um valor
func detectType(value interface{}) string {
	switch value.(type) {
	case bool:
		return "bool"
	case float64, float32:
		return "float"
	case int, int32, int64:
		return "int"
	default:
		return "string"
	}
}

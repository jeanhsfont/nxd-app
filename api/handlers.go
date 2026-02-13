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

// AnalyticsHandler retorna métricas financeiras e de eficiência
func AnalyticsHandler(w http.ResponseWriter, r *http.Request) {
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

	type MachineAnalytics struct {
		ID              int     `json:"id"`
		Name            string  `json:"name"`
		Brand           string  `json:"brand"`
		Status          bool    `json:"status"`
		TotalPecas      int     `json:"total_pecas"`
		ConsumoEnergia  float64 `json:"consumo_energia"`
		Eficiencia      float64 `json:"eficiencia"` // peças/kWh
		HealthScore     float64 `json:"health_score"`
		CustoHoraParada float64 `json:"custo_hora_parada"`
		TempoParado     float64 `json:"tempo_parado_min"`
		LucroCessante   float64 `json:"lucro_cessante"`
		Temperatura     float64 `json:"temperatura"`
		UltimoUpdate    string  `json:"ultimo_update"`
	}

	var analytics []MachineAnalytics
	totalLucroCessante := 0.0
	totalPecas := 0

	for _, machine := range machines {
		values, _ := data.GetLatestDataPoints(machine.ID)
		
		// Extrai valores das tags
		status := parseBool(values["Status_Producao"])
		totalPecasM := parseInt(values["Total_Pecas"])
		consumoEnergia := parseFloat(values["Consumo_Energia_kWh"])
		healthScore := parseFloat(values["Health_Score"])
		custoHora := parseFloat(values["Custo_Hora_Parada"])
		temperatura := parseFloat(values["Temperatura_Molde"])

		// Calcula eficiência
		eficiencia := 0.0
		if consumoEnergia > 0 {
			eficiencia = float64(totalPecasM) / consumoEnergia
		}

		// Calcula lucro cessante (tempo parado estimado baseado no status)
		tempoParado := 0.0
		lucroCessante := 0.0
		if !status && custoHora > 0 {
			// Se está parado agora, estima 5 minutos de parada
			tempoParado = 5.0
			lucroCessante = (tempoParado / 60.0) * custoHora
		}

		// Acumula totais
		totalLucroCessante += lucroCessante
		totalPecas += totalPecasM

		analytics = append(analytics, MachineAnalytics{
			ID:              machine.ID,
			Name:            machine.Name,
			Brand:           machine.Brand,
			Status:          status,
			TotalPecas:      totalPecasM,
			ConsumoEnergia:  consumoEnergia,
			Eficiencia:      eficiencia,
			HealthScore:     healthScore,
			CustoHoraParada: custoHora,
			TempoParado:     tempoParado,
			LucroCessante:   lucroCessante,
			Temperatura:     temperatura,
			UltimoUpdate:    machine.LastSeen.Format("15:04:05"),
		})
	}

	// Calcula comparativo de eficiência
	var maisEficiente string
	var menorEficiencia float64 = 999999
	var maiorEficiencia float64 = 0
	for _, a := range analytics {
		if a.Eficiencia > maiorEficiencia {
			maiorEficiencia = a.Eficiencia
			maisEficiente = a.Name
		}
		if a.Eficiencia < menorEficiencia && a.Eficiencia > 0 {
			menorEficiencia = a.Eficiencia
		}
	}

	ganhoEficiencia := 0.0
	if menorEficiencia > 0 && menorEficiencia < 999999 {
		ganhoEficiencia = ((maiorEficiencia / menorEficiencia) - 1) * 100
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"factory":            factory.Name,
		"machines":           analytics,
		"total_pecas":        totalPecas,
		"total_lucro_cessante": totalLucroCessante,
		"mais_eficiente":     maisEficiente,
		"ganho_eficiencia":   ganhoEficiencia,
		"timestamp":          time.Now().Format(time.RFC3339),
	})
}

// DeleteMachineHandler remove uma máquina específica
func DeleteMachineHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	apiKey := r.URL.Query().Get("api_key")
	machineIDStr := r.URL.Query().Get("machine_id")

	if apiKey == "" || machineIDStr == "" {
		http.Error(w, "API Key e machine_id são obrigatórios", http.StatusBadRequest)
		return
	}

	var machineID int
	fmt.Sscanf(machineIDStr, "%d", &machineID)

	// Valida API Key
	factory, err := data.GetFactoryByAPIKey(apiKey)
	if err != nil || factory == nil {
		http.Error(w, "Fábrica não encontrada", http.StatusUnauthorized)
		return
	}

	// Deleta a máquina
	if err := data.DeleteMachine(machineID); err != nil {
		http.Error(w, "Erro ao deletar máquina", http.StatusInternalServerError)
		return
	}

	log.Printf("🗑️  Máquina ID %d removida da fábrica %s", machineID, factory.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Máquina %d removida com sucesso", machineID),
	})
}

// Funções auxiliares de parsing
func parseBool(s string) bool {
	return s == "true" || s == "1"
}

func parseInt(s string) int {
	var v int
	fmt.Sscanf(s, "%d", &v)
	return v
}

func parseFloat(s string) float64 {
	var v float64
	fmt.Sscanf(s, "%f", &v)
	return v
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

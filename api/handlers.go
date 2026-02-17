package api

import (
	"encoding/json"
	"fmt"
	"hubsystem/core"
	"hubsystem/data"
	"hubsystem/services"
	"log"
	"net/http"
	"os"
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

// CreateFactoryAuthHandler cria fábrica para o usuário logado; retorna api_key uma vez
func CreateFactoryAuthHandler(w http.ResponseWriter, r *http.Request) {
	user := UserFromRequest(r)
	if user == nil {
		http.Error(w, "faça login", http.StatusUnauthorized)
		return
	}
	existing, _ := data.GetFactoryByUserID(user.ID)
	if existing != nil {
		http.Error(w, "você já possui uma fábrica cadastrada", http.StatusConflict)
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, "Nome da fábrica é obrigatório", http.StatusBadRequest)
		return
	}
	apiKey, err := core.GenerateAPIKey()
	if err != nil {
		http.Error(w, "Erro ao gerar API Key", http.StatusInternalServerError)
		return
	}
	factoryID, err := data.CreateFactoryForUser(req.Name, apiKey, user.ID)
	if err != nil {
		http.Error(w, "Erro ao criar fábrica", http.StatusInternalServerError)
		return
	}
	log.Printf("🏭 Fábrica criada para user %d: %s (ID: %d)", user.ID, req.Name, factoryID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": factoryID, "name": req.Name, "api_key": apiKey,
	})
}

// RegenerateAPIKeyHandler gera nova API Key para a fábrica do usuário
func RegenerateAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	user := UserFromRequest(r)
	if user == nil {
		http.Error(w, "faça login", http.StatusUnauthorized)
		return
	}
	factory, err := data.GetFactoryByUserID(user.ID)
	if err != nil || factory == nil {
		http.Error(w, "fábrica não encontrada", http.StatusNotFound)
		return
	}
	newKey, err := data.RegenerateAPIKey(factory.ID)
	if err != nil {
		http.Error(w, "Erro ao regenerar API Key", http.StatusInternalServerError)
		return
	}
	log.Printf("🔑 API Key regenerada para fábrica %s (user %d)", factory.Name, user.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"api_key": newKey})
}

// GetSectorsHandler lista setores (baias) da fábrica
func GetSectorsHandler(w http.ResponseWriter, r *http.Request) {
	factory, err := getFactoryFromRequest(r)
	if err != nil || factory == nil {
		http.Error(w, "API Key não fornecida ou faça login", http.StatusBadRequest)
		return
	}
	list, err := data.GetSectorsByFactory(factory.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"sectors": list})
}

// CreateSectorHandler cria um setor (baia)
func CreateSectorHandler(w http.ResponseWriter, r *http.Request) {
	factory, err := getFactoryFromRequest(r)
	if err != nil || factory == nil {
		http.Error(w, "API Key não fornecida ou faça login", http.StatusBadRequest)
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, "Nome do setor é obrigatório", http.StatusBadRequest)
		return
	}
	id, err := data.CreateSector(factory.ID, strings.TrimSpace(req.Name))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "name": req.Name})
}

// UpdateMachineAssetHandler atualiza display_name, notes e setor da máquina
func UpdateMachineAssetHandler(w http.ResponseWriter, r *http.Request) {
	factory, err := getFactoryFromRequest(r)
	if err != nil || factory == nil {
		http.Error(w, "API Key não fornecida ou faça login", http.StatusBadRequest)
		return
	}
	var req struct {
		MachineID   int    `json:"machine_id"`
		DisplayName string `json:"display_name"`
		Notes       string `json:"notes"`
		SectorID    int    `json:"sector_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}
	if req.MachineID <= 0 {
		http.Error(w, "machine_id obrigatório", http.StatusBadRequest)
		return
	}
	machines, _ := data.GetMachinesByFactory(factory.ID)
	var found bool
	for _, m := range machines {
		if m.ID == req.MachineID {
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "Máquina não encontrada", http.StatusNotFound)
		return
	}
	if req.DisplayName != "" {
		_ = data.UpdateMachineDisplayName(req.MachineID, strings.TrimSpace(req.DisplayName))
	}
	_ = data.UpdateMachineNotes(req.MachineID, strings.TrimSpace(req.Notes))
	_ = data.SetMachineSector(req.MachineID, req.SectorID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// DashboardSummaryHandler retorna resumo para o "Dashboard de Alívio" (cards: ON/OFF, alertas, produção)
func DashboardSummaryHandler(w http.ResponseWriter, r *http.Request) {
	factory, err := getFactoryFromRequest(r)
	if err != nil || factory == nil {
		http.Error(w, "API Key não fornecida ou faça login", http.StatusBadRequest)
		return
	}
	statuses, err := services.GetMachineHealthStatus(factory.ID)
	if err != nil {
		http.Error(w, "Erro ao buscar status", http.StatusInternalServerError)
		return
	}
	online, offline, critical := 0, 0, 0
	for _, s := range statuses {
		switch s.Status {
		case "online": online++
		case "offline": offline++
		case "critical": critical++
		}
	}
	machines, _ := data.GetMachinesByFactory(factory.ID)
	totalPecas := 0
	totalLucroCessante := 0.0
	for _, machine := range machines {
		values, _ := data.GetLatestDataPoints(machine.ID)
		totalPecas += parseInt(values["Total_Pecas"])
		status := parseBool(values["Status_Producao"])
		custoHora := parseFloat(values["Custo_Hora_Parada"])
		if !status && custoHora > 0 {
			totalLucroCessante += (5.0 / 60.0) * custoHora
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"factory":               factory.Name,
		"online":                online,
		"offline":               offline,
		"critical":              critical,
		"total_machines":         len(statuses),
		"total_pecas":           totalPecas,
		"lucro_cessante":        totalLucroCessante,
		"message":               getOverallHealthMessage(online, offline, critical),
		"timestamp":             time.Now().Format(time.RFC3339),
	})
}

// getFactoryFromRequest retorna a fábrica via api_key (query) ou via usuário logado
func getFactoryFromRequest(r *http.Request) (*core.Factory, error) {
	if apiKey := r.URL.Query().Get("api_key"); apiKey != "" {
		return data.GetFactoryByAPIKey(apiKey)
	}
	user := UserFromRequest(r)
	if user == nil {
		return nil, nil
	}
	return data.GetFactoryByUserID(user.ID)
}

// GetDashboardHandler retorna dados do dashboard (api_key na query ou sessão)
func GetDashboardHandler(w http.ResponseWriter, r *http.Request) {
	factory, err := getFactoryFromRequest(r)
	if err != nil || factory == nil {
		http.Error(w, "API Key não fornecida ou faça login", http.StatusBadRequest)
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

// SystemParamsHandler retorna parâmetros do sistema (ex.: se IA está operando 100%).
// GET /api/system
func SystemParamsHandler(w http.ResponseWriter, r *http.Request) {
	iaProvider := os.Getenv("NXD_IA_PROVIDER")   // "vertex" | "gemini" | ""
	geminiKey := os.Getenv("GEMINI_API_KEY")     // legado / API Gemini direta
	vertexProject := os.Getenv("VERTEX_AI_PROJECT")
	iaConfigurado := iaProvider != "" || geminiKey != "" || vertexProject != ""
	// NXD relatórios usam Vertex quando NXD_IA_PROVIDER=vertex e VERTEX_AI_PROJECT definido.
	iaOperando100 := iaProvider == "vertex" && vertexProject != ""
	iaParametro := "NXD_IA_PROVIDER"
	if geminiKey != "" {
		iaParametro = "GEMINI_API_KEY"
	}
	if vertexProject != "" {
		iaParametro = "VERTEX_AI_PROJECT"
	}
	var iaFalta []string
	if !iaConfigurado {
		iaFalta = append(iaFalta, "Definir NXD_IA_PROVIDER=gemini ou NXD_IA_PROVIDER=vertex (e VERTEX_AI_PROJECT), ou GEMINI_API_KEY para API Gemini.")
	}
	if iaConfigurado && !iaOperando100 {
		iaFalta = append(iaFalta, "Definir NXD_IA_PROVIDER=vertex e VERTEX_AI_PROJECT para relatórios NXD; implementar modelo em POST /api/report/ia (legado) se desejar.")
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ia_operando_100": iaOperando100,
		"ia_parametro":    iaParametro,
		"ia_configurado":  iaConfigurado,
		"ia_falta":        iaFalta,
	})
}

// AnalyticsHandler retorna métricas financeiras e de eficiência (api_key ou sessão)
func AnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	factory, err := getFactoryFromRequest(r)
	if err != nil || factory == nil {
		http.Error(w, "API Key não fornecida ou faça login", http.StatusBadRequest)
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

// HealthStatusHandler retorna o status de conexão das máquinas (api_key ou sessão)
func HealthStatusHandler(w http.ResponseWriter, r *http.Request) {
	factory, err := getFactoryFromRequest(r)
	if err != nil || factory == nil {
		http.Error(w, "API Key não fornecida ou faça login", http.StatusBadRequest)
		return
	}

	statuses, err := services.GetMachineHealthStatus(factory.ID)
	if err != nil {
		http.Error(w, "Erro ao buscar status", http.StatusInternalServerError)
		return
	}

	// Conta status
	online := 0
	offline := 0
	critical := 0
	for _, s := range statuses {
		switch s.Status {
		case "online":
			online++
		case "offline":
			offline++
		case "critical":
			critical++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"factory":      factory.Name,
		"total":        len(statuses),
		"online":       online,
		"offline":      offline,
		"critical":     critical,
		"machines":     statuses,
		"timestamp":    time.Now().Format(time.RFC3339),
		"message":      getOverallHealthMessage(online, offline, critical),
		"nxd_status":   "operational", // NXD sempre operacional se responder
		"responsibility": map[string]string{
			"nxd":    "Servidor NXD operacional. Dados que chegam são processados corretamente.",
			"dx":     "Se máquina está offline, verifique: energia do DX, sinal 4G, configuração.",
			"notice": "O NXD só pode processar dados que chegam na API. Problemas de conexão são responsabilidade do ambiente da fábrica.",
		},
	})
}

func getOverallHealthMessage(online, offline, critical int) string {
	if critical > 0 {
		return "🚨 ATENÇÃO: Há máquinas sem comunicação há muito tempo. Verifique os módulos DX."
	}
	if offline > 0 {
		return "⚠️ Algumas máquinas estão offline. Pode ser problema no DX ou na rede da fábrica."
	}
	return "✅ Todas as máquinas comunicando normalmente."
}

// ConnectionLogsHandler retorna os logs de conexão/desconexão (api_key ou sessão)
func ConnectionLogsHandler(w http.ResponseWriter, r *http.Request) {
	factory, err := getFactoryFromRequest(r)
	if err != nil || factory == nil {
		http.Error(w, "API Key não fornecida ou faça login", http.StatusBadRequest)
		return
	}

	// Pega limite (default 100)
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	logs := services.GetConnectionLogs(limit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"factory":   factory.Name,
		"logs":      logs,
		"count":     len(logs),
		"timestamp": time.Now().Format(time.RFC3339),
		"notice":    "Estes logs mostram quando máquinas conectaram/desconectaram. Use para diagnóstico de problemas de rede.",
	})
}

// ReportIARequest corpo do POST /api/report/ia
type ReportIARequest struct {
	SectorID    int      `json:"sector_id"`
	MachineIDs  []int    `json:"machine_ids"`
	Period      string   `json:"period"`
	Shift       string   `json:"shift"`
	DetailLevel string   `json:"detail_level"`
	Niches      []string `json:"niches"`
}

// RastreabilidadeItem um item do "Caminho da Verdade"
type RastreabilidadeItem struct {
	Metrica       string `json:"metrica"`
	Fonte          string `json:"fonte"`
	Origem         string `json:"origem"`
	Processamento  string `json:"processamento"`
	DataHora       string `json:"data_hora"`
}

// ReportIAHandler gera relatório com filtros e rastreabilidade (Central de IA)
func ReportIAHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	factory, err := getFactoryFromRequest(r)
	if err != nil || factory == nil {
		http.Error(w, "API Key não fornecida ou faça login", http.StatusBadRequest)
		return
	}
	var req ReportIARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}
	machines, err := data.GetMachinesByFactory(factory.ID)
	if err != nil {
		http.Error(w, "Erro ao buscar máquinas", http.StatusInternalServerError)
		return
	}
	if len(req.MachineIDs) > 0 {
		m := make(map[int]bool)
		for _, id := range req.MachineIDs {
			m[id] = true
		}
		var filtered []core.Machine
		for _, mc := range machines {
			if m[mc.ID] {
				filtered = append(filtered, mc)
			}
		}
		machines = filtered
	} else if req.SectorID > 0 {
		var filtered []core.Machine
		for _, mc := range machines {
			if mc.SectorID != nil && *mc.SectorID == req.SectorID {
				filtered = append(filtered, mc)
			}
		}
		machines = filtered
	}
	now := time.Now().Format("02/01/2006 15:04:05")
	var rastreabilidade []RastreabilidadeItem
	totalPecas := 0
	totalLucroCessante := 0.0
	online, offline := 0, 0
	var reportLines []string
	reportLines = append(reportLines, fmt.Sprintf("Relatório NXD — %s — Fábrica: %s", now, factory.Name))
	reportLines = append(reportLines, fmt.Sprintf("Período: %s | Turno: %s | Detalhamento: %s", req.Period, req.Shift, req.DetailLevel))
	reportLines = append(reportLines, "")
	for _, machine := range machines {
		values, _ := data.GetLatestDataPoints(machine.ID)
		status := parseBool(values["Status_Producao"])
		if status {
			online++
		} else {
			offline++
		}
		pecas := parseInt(values["Total_Pecas"])
		totalPecas += pecas
		custoHora := parseFloat(values["Custo_Hora_Parada"])
		lucro := 0.0
		if !status && custoHora > 0 {
			lucro = (5.0 / 60.0) * custoHora
			totalLucroCessante += lucro
		}
		name := machine.DisplayName
		if name == "" {
			name = machine.Name
		}
		rastreabilidade = append(rastreabilidade, RastreabilidadeItem{
			Metrica:       "Total_Pecas",
			Fonte:         "Última leitura DX (data_points)",
			Origem:        fmt.Sprintf("Máquina %s, Tag Total_Pecas", name),
			Processamento: "Leitura direta do CLP",
			DataHora:      now,
		})
		rastreabilidade = append(rastreabilidade, RastreabilidadeItem{
			Metrica:       "Status_Producao",
			Fonte:         "Última leitura DX",
			Origem:        fmt.Sprintf("Máquina %s, Tag Status_Producao", name),
			Processamento: "Leitura direta (bool)",
			DataHora:      now,
		})
		if custoHora > 0 {
			rastreabilidade = append(rastreabilidade, RastreabilidadeItem{
				Metrica:       "Lucro cessante (estimado)",
				Fonte:         "Agregado NXD",
				Origem:        fmt.Sprintf("Máquina %s, Custo_Hora_Parada", name),
				Processamento: "(5/60) * Custo_Hora_Parada quando Status_Producao = false",
				DataHora:      now,
			})
		}
		if req.DetailLevel == "medio" || req.DetailLevel == "detalhado" {
			reportLines = append(reportLines, fmt.Sprintf("• %s: %d peças, operando=%v, lucro cessante est.= R$ %.2f", name, pecas, status, lucro))
		}
	}
	reportLines = append(reportLines, "")
	reportLines = append(reportLines, fmt.Sprintf("Resumo: %d máquina(s) | Produção total (última leitura): %d peças | Lucro cessante estimado: R$ %.2f", len(machines), totalPecas, totalLucroCessante))
	reportLines = append(reportLines, "Todos os valores são rastreáveis na seção Rastreabilidade abaixo.")
	rastreabilidade = append(rastreabilidade, RastreabilidadeItem{
		Metrica:       "Produção total (agregado)",
		Fonte:         "Soma das leituras Total_Pecas por máquina",
		Origem:        "NXD dashboard aggregate",
		Processamento: "Soma(Total_Pecas) por máquina no escopo",
		DataHora:      now,
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"report":         strings.Join(reportLines, "\n"),
		"rastreabilidade": rastreabilidade,
		"timestamp":      time.Now().Format(time.RFC3339),
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

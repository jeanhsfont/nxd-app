package api

import (
	"context"
	"encoding/json"
	"fmt"
	"hubsystem/core"
	"hubsystem/internal/nxd/store"
	"hubsystem/services"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"github.com/google/uuid"
)

// IngestHandler recebe dados do DX (endpoint principal)
func IngestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = strings.Split(forwarded, ",")[0]
	}

	var payload core.IngestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		services.LogError("INGEST", "", "", "Payload inválido", ipAddress)
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}

	if !core.ValidateAPIKey(payload.APIKey) {
		services.LogError("INGEST", payload.APIKey, payload.DeviceID, "API Key inválida", ipAddress)
		http.Error(w, "API Key inválida", http.StatusUnauthorized)
		return
	}

	db := store.NXDDB()
	factory, err := store.GetFactoryByAPIKey(db, payload.APIKey)
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

	factoryID, _ := uuid.Parse(factory.ID)
	deviceID := core.SanitizeDeviceID(payload.DeviceID)

	assetID, err := store.CreateAsset(db, factoryID, nil, deviceID, payload.Brand, "", nil)
	if err != nil {
		services.LogError("INGEST", payload.APIKey, deviceID, fmt.Sprintf("Erro ao criar/buscar asset: %v", err), ipAddress)
		http.Error(w, "Erro ao processar asset", http.StatusInternalServerError)
		return
	}

	log.Printf("📥 [INGEST] Fábrica: %s | Asset: %s (%s) | Tags: %d",
		factory.Name, deviceID, payload.Brand, len(payload.Tags))

	timestamp := payload.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	var telemetryRows []store.TelemetryRow
	for tagName, tagValue := range payload.Tags {
		tagType := detectType(tagValue)
		value, ok := tagValue.(float64)
		if !ok {
			log.Printf("  ⚠️ Tag: %s com valor não numérico %v (%T)", tagName, tagValue, tagValue)
			continue
		}

		if err := store.UpsertAssetMetricCatalog(db, factoryID, assetID, tagName, timestamp); err != nil {
			log.Printf("❌ Erro ao criar/buscar métrica %s: %v", tagName, err)
			continue
		}

		telemetryRows = append(telemetryRows, store.TelemetryRow{
			Ts:          timestamp,
			MetricKey:   tagName,
			MetricValue: value,
		})
		log.Printf("  ✓ Tag: %s = %v (%s)", tagName, value, tagType)
	}

	if len(telemetryRows) > 0 {
		correlationID := uuid.New().String()
		if err := store.InsertTelemetryBatch(db, factoryID, assetID, correlationID, telemetryRows); err != nil {
			log.Printf("❌ Erro ao inserir data points: %v", err)
		}
	}

	services.LogSuccess("INGEST", payload.APIKey, deviceID,
		fmt.Sprintf("Processadas %d tags", len(payload.Tags)), ipAddress)

	// services.BroadcastUpdate(assetID.String())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"machine_id": assetID,
		"tags_count": len(payload.Tags),
	})
}

func detectType(value interface{}) string {
	if value == nil {
		return "nil"
	}
	return reflect.TypeOf(value).String()
}

func CreateFactoryHandler(w http.ResponseWriter, r *http.Request) {
	// Implementação futura se necessário
}

func CreateFactoryAuthHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*store.User)

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "O nome da fábrica é obrigatório", http.StatusBadRequest)
		return
	}

	apiKey, err := core.GenerateAPIKey()
	if err != nil {
		http.Error(w, "Erro ao gerar API Key", http.StatusInternalServerError)
		return
	}
	db := store.NXDDB()

	factoryID, err := store.CreateFactoryForUser(db, req.Name, apiKey, user.ID)
	if err != nil {
		http.Error(w, "Erro ao criar fábrica", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":      factoryID.String(),
		"name":    req.Name,
		"api_key": apiKey,
	})
}

func GetFactoryDetailsHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*store.User)
	db := store.NXDDB()

	factory, err := store.GetFactoryByUserID(db, user.ID)
	if err != nil {
		http.Error(w, "Erro ao buscar fábrica", http.StatusInternalServerError)
		return
	}

	if factory == nil {
		http.Error(w, "Nenhuma fábrica encontrada para este usuário", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(factory)
}

func RegenerateAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*store.User)
	db := store.NXDDB()

	factory, err := store.GetFactoryByUserID(db, user.ID)
	if err != nil || factory == nil {
		http.Error(w, "Fábrica não encontrada", http.StatusNotFound)
		return
	}

	factoryID, _ := uuid.Parse(factory.ID)
	newAPIKey, err := store.RegenerateAPIKey(db, factoryID)
	if err != nil {
		http.Error(w, "Erro ao gerar nova chave", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"api_key": newAPIKey})
}

func GetSectorsHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*store.User)
	db := store.NXDDB()
	factory, err := store.GetFactoryByUserID(db, user.ID)
	if err != nil || factory == nil {
		http.Error(w, "Fábrica não encontrada", http.StatusNotFound)
		return
	}
	factoryID, _ := uuid.Parse(factory.ID)
	sectors, err := store.ListSectors(db, factoryID)
	if err != nil {
		http.Error(w, "Erro ao buscar setores", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sectors)
}

func CreateSectorHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*store.User)
	db := store.NXDDB()
	factory, err := store.GetFactoryByUserID(db, user.ID)
	if err != nil || factory == nil {
		http.Error(w, "Fábrica não encontrada", http.StatusNotFound)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "O nome do setor é obrigatório", http.StatusBadRequest)
		return
	}

	factoryID, _ := uuid.Parse(factory.ID)
	sectorID, err := store.CreateSector(db, factoryID, req.Name, req.Description)
	if err != nil {
		http.Error(w, "Erro ao criar setor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": sectorID.String()})
}

func UpdateSectorHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*store.User)
	db := store.NXDDB()
	factory, err := store.GetFactoryByUserID(db, user.ID)
	if err != nil || factory == nil {
		http.Error(w, "Fábrica não encontrada", http.StatusNotFound)
		return
	}

	// Extrair o ID do setor da URL, por exemplo /api/sectors/{id}
	parts := strings.Split(r.URL.Path, "/")
	sectorIDStr := parts[len(parts)-1]
	sectorID, err := uuid.Parse(sectorIDStr)
	if err != nil {
		http.Error(w, "ID de setor inválido", http.StatusBadRequest)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "O nome do setor é obrigatório", http.StatusBadRequest)
		return
	}

	factoryID, _ := uuid.Parse(factory.ID)
	err = store.UpdateSector(db, sectorID, factoryID, req.Name, req.Description)
	if err != nil {
		http.Error(w, "Erro ao atualizar setor", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteSectorHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*store.User)
	db := store.NXDDB()
	factory, err := store.GetFactoryByUserID(db, user.ID)
	if err != nil || factory == nil {
		http.Error(w, "Fábrica não encontrada", http.StatusNotFound)
		return
	}

	// Extrair o ID do setor da URL
	parts := strings.Split(r.URL.Path, "/")
	sectorIDStr := parts[len(parts)-1]
	sectorID, err := uuid.Parse(sectorIDStr)
	if err != nil {
		http.Error(w, "ID de setor inválido", http.StatusBadRequest)
		return
	}

	factoryID, _ := uuid.Parse(factory.ID)
	err = store.DeleteSector(db, sectorID, factoryID)
	if err != nil {
		http.Error(w, "Erro ao deletar setor", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func UpdateMachineAssetHandler(w http.ResponseWriter, r *http.Request) {
	// Implementação futura
}
func GetDashboardHandler(w http.ResponseWriter, r *http.Request)     {}
func DashboardSummaryHandler(w http.ResponseWriter, r *http.Request) {}
func ReportIAHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	projectID := "slideflow-prod"     // Substituir pelo seu ID de projeto do Google Cloud
	location := "us-central1"         // Substituir pela sua região
	modelName := "gemini-1.0-pro-001" // O modelo que você deseja usar

	sectorIDStr := r.URL.Query().Get("sector_id")
	if sectorIDStr == "" {
		http.Error(w, "O 'sector_id' é obrigatório", http.StatusBadRequest)
		return
	}

	sectorID, err := uuid.Parse(sectorIDStr)
	if err != nil {
		http.Error(w, "ID de setor inválido", http.StatusBadRequest)
		return
	}

	db := store.NXDDB()
	assets, err := store.ListAssetsBySector(db, sectorID)
	if err != nil {
		http.Error(w, "Erro ao buscar ativos para o setor", http.StatusInternalServerError)
		return
	}

	if len(assets) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Nenhum ativo encontrado para este setor.",
			"data":    nil,
		})
		return
	}

	assetIDs := make([]uuid.UUID, len(assets))
	for i, asset := range assets {
		assetIDs[i] = asset.ID
	}

	telemetry, err := store.ListTelemetryByAssets(db, assetIDs)
	if err != nil {
		http.Error(w, "Erro ao buscar dados de telemetria", http.StatusInternalServerError)
		return
	}

	// Construir o prompt para a IA
	var promptBuilder strings.Builder
	promptBuilder.WriteString("Analise os seguintes dados de telemetria de uma fábrica e forneça um resumo sobre a saúde operacional, identificando possíveis anomalias ou pontos de atenção.\n\n")
	promptBuilder.WriteString(fmt.Sprintf("Dados do Setor ID: %s\n", sectorID.String()))
	promptBuilder.WriteString(fmt.Sprintf("Total de Ativos: %d\n\n", len(assets)))

	for _, t := range telemetry {

		// O payload é um JSON, então precisamos decodificá-lo.
		// A estrutura de telemetria mudou para usar um campo Payload (JSON) e Timestamp.
		var telemetryData map[string]interface{}
		if err := json.Unmarshal([]byte(t.Payload), &telemetryData); err != nil {
			log.Printf("Erro ao decodificar payload de telemetria: %v", err)
			continue
		}

		// Agora podemos construir a string a partir do mapa decodificado
		promptBuilder.WriteString(fmt.Sprintf("Timestamp: %s, Dados: %v\n", t.Timestamp.Format(time.RFC3339), telemetryData))
	}

	// Inicializar o cliente do Vertex AI
	client, err := genai.NewClient(ctx, projectID, location)
	if err != nil {
		log.Printf("Erro ao criar cliente do Vertex AI: %v", err)
		http.Error(w, "Erro ao conectar com o serviço de IA", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	model := client.GenerativeModel(modelName)
	resp, err := model.GenerateContent(ctx, genai.Text(promptBuilder.String()))
	if err != nil {
		log.Printf("Erro ao gerar conteúdo do Vertex AI: %v", err)
		http.Error(w, "Erro ao processar a análise de IA", http.StatusInternalServerError)
		return
	}

	// Extrair e formatar a resposta da IA
	var aiResponse strings.Builder
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if txt, ok := part.(genai.Text); ok {
					aiResponse.WriteString(string(txt))
				}
			}
		}
	}

	responseData := map[string]interface{}{
		"sector_id":         sectorID,
		"assets_count":      len(assets),
		"telemetry_records": len(telemetry),
		"analysis":          aiResponse.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
func AnalyticsHandler(w http.ResponseWriter, r *http.Request)      {}
func DeleteMachineHandler(w http.ResponseWriter, r *http.Request)  {}
func HealthStatusHandler(w http.ResponseWriter, r *http.Request)   {}
func ConnectionLogsHandler(w http.ResponseWriter, r *http.Request) {}

func HealthHandler(w http.ResponseWriter, r *http.Request)       {}
func SystemParamsHandler(w http.ResponseWriter, r *http.Request) {}

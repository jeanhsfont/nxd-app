package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hubsystem/core"
	"hubsystem/internal/nxd/store"
	"hubsystem/services"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// ─── Dashboard in-memory cache ────────────────────────────────────────────────
// Avoids hammering the DB on every 10s frontend poll.
// TTL = 8s so at most 1 real DB query per frontend refresh cycle.
type dashboardCacheEntry struct {
	payload   []byte
	expiresAt time.Time
}

var (
	dashboardCacheMu sync.Mutex
	dashboardCache   = map[string]dashboardCacheEntry{}
)

const dashboardCacheTTL = 8 * time.Second

// ─── Ingest rate limiter ──────────────────────────────────────────────────────
// Simple token-bucket per API key: max ingestRateLimit requests per ingestRateWindow.
// Uses in-memory map — no Redis dependency. Resets on Cloud Run instance restart,
// which is acceptable for MVP: state loss only on cold starts, not between requests.
//
// Limits are intentionally lenient (30 req/min per key) — enough to catch runaway
// DX bugs without blocking legitimate fast-sampling devices.

const (
	ingestRateLimit  = 30              // max requests per window per API key
	ingestRateWindow = 1 * time.Minute // sliding window
	ingestMaxBodyKB  = 256             // max request body in KB
	ingestMaxTags    = 200             // max metrics per single ingest request
)

type rateBucket struct {
	count     int
	windowEnd time.Time
}

var (
	ingestRateMu      sync.Mutex
	ingestRateBuckets = map[string]*rateBucket{}
)

// checkIngestRate returns true if the request is within rate limits.
func checkIngestRate(apiKey string) bool {
	ingestRateMu.Lock()
	defer ingestRateMu.Unlock()
	now := time.Now()
	b, ok := ingestRateBuckets[apiKey]
	if !ok || now.After(b.windowEnd) {
		// New window
		ingestRateBuckets[apiKey] = &rateBucket{count: 1, windowEnd: now.Add(ingestRateWindow)}
		return true
	}
	b.count++
	return b.count <= ingestRateLimit
}

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

	// ── Payload size guard ────────────────────────────────────────────────
	// Prevents memory exhaustion from oversized payloads. 256 KB is generous
	// for any real industrial ingest (typical: <2 KB per request).
	r.Body = http.MaxBytesReader(w, r.Body, ingestMaxBodyKB*1024)

	var payload core.IngestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		if strings.Contains(err.Error(), "http: request body too large") {
			log.Printf("⚠️ [INGEST] Payload muito grande (>%dKB) de %s", ingestMaxBodyKB, ipAddress)
			http.Error(w, fmt.Sprintf("Payload excede limite de %dKB", ingestMaxBodyKB), http.StatusRequestEntityTooLarge)
			return
		}
		log.Printf("⚠️ [INGEST] JSON inválido de %s: %v", ipAddress, err)
		services.LogError("INGEST", "", "", fmt.Sprintf("JSON inválido: %v", err), ipAddress)
		http.Error(w, "Payload inválido: JSON malformado", http.StatusBadRequest)
		return
	}

	// ── API key format validation ─────────────────────────────────────────
	if !core.ValidateAPIKey(payload.APIKey) {
		log.Printf("⚠️ [INGEST] API Key inválida (formato) de %s — key: %.12s...", ipAddress, payload.APIKey)
		services.LogError("INGEST", payload.APIKey, payload.DeviceID, "API Key inválida (formato)", ipAddress)
		http.Error(w, "API Key inválida", http.StatusUnauthorized)
		return
	}

	// ── Device ID validation ─────────────────────────────────────────────
	if strings.TrimSpace(payload.DeviceID) == "" {
		log.Printf("⚠️ [INGEST] device_id ausente de %s", ipAddress)
		http.Error(w, "Campo device_id é obrigatório", http.StatusBadRequest)
		return
	}

	// ── Rate limiting per API key ─────────────────────────────────────────
	// Applied after format check so invalid keys don't pollute the rate table.
	if !checkIngestRate(payload.APIKey) {
		log.Printf("⚠️ [INGEST] Rate limit excedido para key %.12s... de %s", payload.APIKey, ipAddress)
		services.LogError("INGEST", payload.APIKey, payload.DeviceID,
			fmt.Sprintf("Rate limit excedido (%d req/min)", ingestRateLimit), ipAddress)
		w.Header().Set("Retry-After", "60")
		http.Error(w, fmt.Sprintf("Rate limit excedido: máximo %d requisições por minuto", ingestRateLimit), http.StatusTooManyRequests)
		return
	}

	// ── Tags count guard ─────────────────────────────────────────────────
	if len(payload.Tags) > ingestMaxTags {
		log.Printf("⚠️ [INGEST] Excesso de tags (%d > %d) de device %s", len(payload.Tags), ingestMaxTags, payload.DeviceID)
		http.Error(w, fmt.Sprintf("Número de tags excede limite de %d por request", ingestMaxTags), http.StatusBadRequest)
		return
	}

	db := store.NXDDB()
	factory, err := store.GetFactoryByAPIKey(db, payload.APIKey)
	if err != nil {
		log.Printf("❌ [INGEST] Erro DB ao buscar fábrica para key %.12s...: %v", payload.APIKey, err)
		services.LogError("INGEST", payload.APIKey, payload.DeviceID, fmt.Sprintf("Erro DB: %v", err), ipAddress)
		http.Error(w, "Erro interno ao autenticar", http.StatusInternalServerError)
		return
	}
	if factory == nil {
		log.Printf("⚠️ [INGEST] API Key não encontrada: %.12s... de %s", payload.APIKey, ipAddress)
		services.LogError("INGEST", payload.APIKey, payload.DeviceID, "API Key não encontrada", ipAddress)
		http.Error(w, "API Key não autorizada", http.StatusUnauthorized)
		return
	}
	if !factory.IsActive {
		log.Printf("⚠️ [INGEST] Fábrica inativa para key %.12s... de %s", payload.APIKey, ipAddress)
		services.LogError("INGEST", payload.APIKey, payload.DeviceID, "Fábrica inativa", ipAddress)
		http.Error(w, "Fábrica desativada — contate o suporte", http.StatusForbidden)
		return
	}

	factoryID, _ := uuid.Parse(factory.ID)
	deviceID := core.SanitizeDeviceID(payload.DeviceID)

	if deviceID == "" {
		log.Printf("⚠️ [INGEST] device_id resultou vazio após sanitização: original=%q", payload.DeviceID)
		http.Error(w, "device_id inválido: use apenas letras, números, hífens e underscores", http.StatusBadRequest)
		return
	}

	assetID, err := store.CreateAsset(db, factoryID, nil, deviceID, payload.Brand, "", nil)
	if err != nil {
		log.Printf("❌ [INGEST] Erro ao criar/buscar asset %s: %v", deviceID, err)
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
	skippedTags := 0
	for tagName, tagValue := range payload.Tags {
		value, ok := tagValue.(float64)
		if !ok {
			// Tenta converter string numérica — tolerância para DX que serializa números como string
			if strVal, isStr := tagValue.(string); isStr {
				if parsed, parseErr := fmt.Sscanf(strVal, "%f", &value); parseErr != nil || parsed != 1 {
					log.Printf("  ⚠️ [INGEST] Tag ignorada: %s = %v (%T) — não numérico", tagName, tagValue, tagValue)
					skippedTags++
					continue
				}
			} else {
				log.Printf("  ⚠️ [INGEST] Tag ignorada: %s = %v (%T) — não numérico", tagName, tagValue, tagValue)
				skippedTags++
				continue
			}
		}

		if err := store.UpsertAssetMetricCatalog(db, factoryID, assetID, tagName, timestamp); err != nil {
			log.Printf("❌ [INGEST] Erro ao atualizar catálogo de métrica %s: %v", tagName, err)
			continue
		}

		telemetryRows = append(telemetryRows, store.TelemetryRow{
			Ts:          timestamp,
			MetricKey:   tagName,
			MetricValue: value,
		})
	}

	// P5: Log when all tags were skipped — helps diagnose DX misconfiguration.
	if len(telemetryRows) == 0 && len(payload.Tags) > 0 {
		log.Printf("⚠️  [INGEST] Todas as %d tags ignoradas para device %s (fábrica=%s) — verifique tipos de dados enviados pelo DX",
			len(payload.Tags), deviceID, factory.Name)
	} else if len(payload.Tags) == 0 {
		log.Printf("⚠️  [INGEST] Payload sem tags para device %s (fábrica=%s) — possível DX sem dados configurados",
			deviceID, factory.Name)
	}

	if len(telemetryRows) > 0 {
		correlationID := uuid.New().String()
		if err := store.InsertTelemetryBatch(db, factoryID, assetID, correlationID, telemetryRows); err != nil {
			log.Printf("❌ [INGEST] Erro ao inserir telemetria para %s: %v", deviceID, err)
		}
	}

	services.LogSuccess("INGEST", payload.APIKey, deviceID,
		fmt.Sprintf("Processadas %d tags, ignoradas %d", len(telemetryRows), skippedTags), ipAddress)

	// services.BroadcastUpdate(assetID.String())

	// P5: Richer response — includes diagnostic info without changing contract.
	// Fields "tags_count" and "tags_skipped" were already present; adding
	// "warnings" array for actionable DX debugging info.
	var warnings []string
	if skippedTags > 0 {
		warnings = append(warnings, fmt.Sprintf("%d tag(s) ignorada(s) por tipo não-numérico", skippedTags))
	}
	if len(payload.Tags) == 0 {
		warnings = append(warnings, "payload enviado sem tags")
	}

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"status":       "success",
		"machine_id":   assetID,
		"tags_count":   len(telemetryRows),
		"tags_skipped": skippedTags,
	}
	if len(warnings) > 0 {
		resp["warnings"] = warnings
	}
	json.NewEncoder(w).Encode(resp)
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

// getNXDFactoryForUser busca a fábrica na nxd.factories pelo userID legado (int64)
// buscando via tabela pública e retornando o UUID da nxd.factories
func getNXDFactoryIDForUser(userID int64) (uuid.UUID, error) {
	nxdDB := store.NXDDB()
	if nxdDB == nil {
		return uuid.Nil, fmt.Errorf("NXD DB não inicializado")
	}
	// Busca o email do usuário legado
	var email string
	if err := db.QueryRow("SELECT email FROM users WHERE id = $1", userID).Scan(&email); err != nil {
		return uuid.Nil, fmt.Errorf("usuário não encontrado: %w", err)
	}
	// Busca o user NXD pelo email
	var nxdUserID uuid.UUID
	err := nxdDB.QueryRow("SELECT id FROM nxd.users WHERE email = $1", email).Scan(&nxdUserID)
	if err != nil {
		// Usuário não tem conta NXD ainda — retorna nil para sinalizar
		return uuid.Nil, nil
	}
	// Busca a factory pelo nxd user_id
	var factoryID uuid.UUID
	err = nxdDB.QueryRow("SELECT id FROM nxd.factories WHERE user_id = $1 LIMIT 1", nxdUserID).Scan(&factoryID)
	if err != nil {
		// Sem factory NXD — retorna nil
		return uuid.Nil, nil
	}
	return factoryID, nil
}

// getOrCreateNXDFactory busca ou cria factory NXD para o usuário legado
func getOrCreateNXDFactory(userID int64) (uuid.UUID, error) {
	factoryID, err := getNXDFactoryIDForUser(userID)
	if err != nil {
		return uuid.Nil, err
	}
	if factoryID != uuid.Nil {
		return factoryID, nil
	}

	nxdDB := store.NXDDB()
	if nxdDB == nil {
		return uuid.Nil, fmt.Errorf("NXD DB não inicializado")
	}

	// Busca email e nome da fábrica legada
	var email, factoryName string
	db.QueryRow("SELECT email FROM users WHERE id = $1", userID).Scan(&email)
	if err := db.QueryRow("SELECT name FROM factories WHERE user_id = $1 LIMIT 1", userID).Scan(&factoryName); err != nil {
		factoryName = "Minha Fábrica"
	}

	// Garante que o usuário NXD exista
	var nxdUserID uuid.UUID
	err = nxdDB.QueryRow("SELECT id FROM nxd.users WHERE email = $1", email).Scan(&nxdUserID)
	if err != nil {
		// Cria usuário NXD
		nxdDB.QueryRow(
			"INSERT INTO nxd.users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id",
			"Usuário", email, "migrated",
		).Scan(&nxdUserID)
	}

	// Cria factory NXD
	var newFactoryID uuid.UUID
	err = nxdDB.QueryRow(
		"INSERT INTO nxd.factories (user_id, name) VALUES ($1, $2) RETURNING id",
		nxdUserID, factoryName,
	).Scan(&newFactoryID)
	if err != nil {
		return uuid.Nil, err
	}
	return newFactoryID, nil
}

func CreateFactoryAuthHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
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
		http.Error(w, "O nome da fábrica é obrigatório", http.StatusBadRequest)
		return
	}

	apiKey, apiKeyHash, err := core.GenerateAndHashAPIKey()
	if err != nil {
		http.Error(w, "Erro ao gerar API Key", http.StatusInternalServerError)
		return
	}

	nxdDB := store.NXDDB()
	var factoryID uuid.UUID
	nxdDB.QueryRow(
		"INSERT INTO nxd.factories (name, api_key_hash) VALUES ($1, $2) RETURNING id",
		req.Name, apiKeyHash,
	).Scan(&factoryID)

	// Também salva no legado
	db.Exec(
		"INSERT INTO factories (user_id, name, api_key_hash) VALUES ($1, $2, $3) ON CONFLICT (user_id) DO UPDATE SET name=$2, api_key_hash=$3",
		userID, req.Name, string(apiKeyHash),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":      factoryID.String(),
		"name":    req.Name,
		"api_key": apiKey,
	})
}

func GetFactoryDetailsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	var name, cnpj, address string
	var apiKeyHash *string
	err := db.QueryRow(
		"SELECT name, COALESCE(cnpj,''), COALESCE(address,''), api_key_hash FROM factories WHERE user_id = $1 LIMIT 1",
		userID,
	).Scan(&name, &cnpj, &address, &apiKeyHash)
	if err != nil {
		http.Error(w, "Nenhuma fábrica encontrada", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":         name,
		"cnpj":         cnpj,
		"address":      address,
		"has_api_key":  apiKeyHash != nil && *apiKeyHash != "",
	})
}

func RegenerateAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	apiKey, apiKeyHash, err := core.GenerateAndHashAPIKey()
	if err != nil {
		http.Error(w, "Erro ao gerar chave", http.StatusInternalServerError)
		return
	}

	// Atualiza no legado
	db.Exec("UPDATE factories SET api_key_hash = $1 WHERE user_id = $2", string(apiKeyHash), userID)

	// Também insere/atualiza na nxd.factories (para o IngestHandler)
	if nxdDB := store.NXDDB(); nxdDB != nil {
		nxdDB.Exec("INSERT INTO nxd.factories (name, api_key_hash) VALUES ('Minha Fábrica', $1)", apiKeyHash)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"api_key": apiKey})
}

// getFactoryIDForUser retorna o factoryID NXD para um userID legado
func getFactoryIDForUser(userID int64) (uuid.UUID, error) {
	nxdDB := store.NXDDB()
	if nxdDB == nil {
		return uuid.Nil, fmt.Errorf("NXD DB não disponível")
	}
	var email string
	if err := db.QueryRow("SELECT email FROM users WHERE id = $1", userID).Scan(&email); err != nil {
		return uuid.Nil, fmt.Errorf("usuário não encontrado")
	}
	var factoryID uuid.UUID
	err := nxdDB.QueryRow(`
		SELECT f.id FROM nxd.factories f
		JOIN nxd.users u ON u.id = f.user_id
		WHERE u.email = $1
		LIMIT 1
	`, email).Scan(&factoryID)
	if err != nil {
		// fallback: última factory criada
		err2 := nxdDB.QueryRow(`SELECT id FROM nxd.factories ORDER BY created_at DESC LIMIT 1`).Scan(&factoryID)
		if err2 != nil {
			return uuid.Nil, fmt.Errorf("factory não encontrada")
		}
	}
	return factoryID, nil
}

func GetSectorsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	factoryID, err := getFactoryIDForUser(userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	nxdDB := store.NXDDB()
	rows, err := nxdDB.Query(`
		SELECT id, name, COALESCE(description,''), created_at
		FROM nxd.sectors WHERE factory_id = $1
		ORDER BY name
	`, factoryID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	defer rows.Close()
	type Sector struct {
		ID          string    `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		CreatedAt   time.Time `json:"created_at"`
	}
	var sectors []Sector
	for rows.Next() {
		var s Sector
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.CreatedAt); err == nil {
			sectors = append(sectors, s)
		}
	}
	if sectors == nil {
		sectors = []Sector{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sectors)
}

func CreateSectorHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, "Nome do setor obrigatório", http.StatusBadRequest)
		return
	}
	factoryID, err := getFactoryIDForUser(userID)
	if err != nil {
		http.Error(w, "Factory não encontrada", http.StatusNotFound)
		return
	}
	nxdDB := store.NXDDB()
	var sectorID string
	var createdAt time.Time
	err = nxdDB.QueryRow(`
		INSERT INTO nxd.sectors (factory_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`, factoryID, req.Name, req.Description).Scan(&sectorID, &createdAt)
	if err != nil {
		log.Printf("Erro ao criar setor: %v", err)
		http.Error(w, "Erro ao criar setor", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          sectorID,
		"name":        req.Name,
		"description": req.Description,
		"created_at":  createdAt,
	})
}

func UpdateSectorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func DeleteSectorHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	sectorID := vars["id"]
	factoryID, err := getFactoryIDForUser(userID)
	if err != nil {
		http.Error(w, "Factory não encontrada", http.StatusNotFound)
		return
	}
	nxdDB := store.NXDDB()
	// Desagrupar assets do setor antes de deletar
	nxdDB.Exec(`UPDATE nxd.assets SET group_id = NULL WHERE group_id = $1 AND factory_id = $2`, sectorID, factoryID)
	nxdDB.Exec(`DELETE FROM nxd.sectors WHERE id = $1 AND factory_id = $2`, sectorID, factoryID)
	w.WriteHeader(http.StatusNoContent)
}

// UpdateMachineAssetHandler — PUT /api/machine/asset
// Atualiza o group_id (setor) de um asset no banco NXD
func UpdateMachineAssetHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	var req struct {
		MachineID string `json:"machine_id"`
		SectorID  string `json:"sector_id"` // UUID string ou "" para remover do setor
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.MachineID == "" {
		http.Error(w, "machine_id obrigatório", http.StatusBadRequest)
		return
	}

	factoryID, err := getFactoryIDForUser(userID)
	if err != nil {
		http.Error(w, "Fábrica não encontrada", http.StatusNotFound)
		return
	}

	nxdDB := store.NXDDB()
	if nxdDB == nil {
		http.Error(w, "Banco NXD não disponível", http.StatusInternalServerError)
		return
	}

	// Garante que o asset pertence à factory do usuário autenticado
	var assetCount int
	if err := nxdDB.QueryRow(`SELECT COUNT(*) FROM nxd.assets WHERE id = $1 AND factory_id = $2`,
		req.MachineID, factoryID).Scan(&assetCount); err != nil || assetCount == 0 {
		http.Error(w, "Asset não encontrado ou não autorizado", http.StatusNotFound)
		return
	}

	// Se sector_id for vazio ou "0", remove do setor (NULL)
	if req.SectorID == "" || req.SectorID == "0" {
		_, err = nxdDB.Exec(`UPDATE nxd.assets SET group_id = NULL WHERE id = $1`, req.MachineID)
	} else {
		sectorUUID, parseErr := uuid.Parse(req.SectorID)
		if parseErr != nil {
			http.Error(w, "sector_id inválido", http.StatusBadRequest)
			return
		}
		_, err = nxdDB.Exec(`UPDATE nxd.assets SET group_id = $1 WHERE id = $2`, sectorUUID, req.MachineID)
	}
	if err != nil {
		log.Printf("[UpdateMachineAsset] Erro ao atualizar group_id: %v", err)
		http.Error(w, "Erro ao atualizar setor do ativo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"machine_id": req.MachineID,
		"sector_id":  req.SectorID,
	})
}
func GetDashboardHandler(w http.ResponseWriter, r *http.Request) {}

// GetDashboardDataHandler retorna os dados reais de telemetria para o dashboard (autenticado)
func GetDashboardDataHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	nxdDB := store.NXDDB()
	if nxdDB == nil {
		http.Error(w, "Banco NXD não disponível", http.StatusInternalServerError)
		return
	}

	// Busca a factory NXD pelo userID legado (via email)
	var email string
	if err := db.QueryRow("SELECT email FROM users WHERE id = $1", userID).Scan(&email); err != nil {
		http.Error(w, "Usuário não encontrado", http.StatusNotFound)
		return
	}

	// Busca factory NXD pelo api_key_hash que bate com o que foi gerado no onboarding
	// Alternativa: busca via email → nxd.users → nxd.factories
	var factoryID uuid.UUID
	var factoryName string
	err := nxdDB.QueryRow(`
		SELECT f.id, f.name FROM nxd.factories f
		JOIN nxd.users u ON u.id = f.user_id
		WHERE u.email = $1
		LIMIT 1
	`, email).Scan(&factoryID, &factoryName)
	if err != nil {
		// Tenta buscar factory sem user (onboarding direto via api_key_hash)
		err2 := nxdDB.QueryRow(`
			SELECT id, name FROM nxd.factories
			ORDER BY created_at DESC LIMIT 1
		`).Scan(&factoryID, &factoryName)
		if err2 != nil {
			// Sem factory ainda — retorna estrutura vazia
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"factory_name":   "",
				"assets":         []interface{}{},
				"total_assets":   0,
				"last_update":    nil,
				"online_assets":  0,
			})
			return
		}
	}

	// ── Cache check ────────────────────────────────────────────────────────
	cacheKey := factoryID.String()
	dashboardCacheMu.Lock()
	if entry, ok := dashboardCache[cacheKey]; ok && time.Now().Before(entry.expiresAt) {
		dashboardCacheMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write(entry.payload)
		return
	}
	dashboardCacheMu.Unlock()

	// ── Optimized single query: assets + latest metrics via LATERAL JOIN ──
	// Replaces the old N+2 query pattern (1 GROUP BY + N DISTINCT ON per asset
	// + 1 MAX for factory). Now it's a single round-trip that uses the
	// composite index (asset_id, metric_key, ts DESC).
	type AssetSummary struct {
		ID          string                 `json:"id"`
		DisplayName string                 `json:"display_name"`
		SourceTagID string                 `json:"source_tag_id"`
		GroupID     *string                `json:"group_id"`
		LastSeen    *time.Time             `json:"last_seen"`
		IsOnline    bool                   `json:"is_online"`
		Metrics     map[string]interface{} `json:"metrics"`
	}

	rows, err := nxdDB.Query(`
		SELECT
			a.id,
			a.display_name,
			a.source_tag_id,
			a.group_id::text,
			lat.last_seen,
			lat.metrics_json
		FROM nxd.assets a
		LEFT JOIN LATERAL (
			SELECT
				MAX(tl.ts)                                          AS last_seen,
				jsonb_object_agg(tl.metric_key, tl.metric_value)   AS metrics_json
			FROM (
				SELECT DISTINCT ON (metric_key)
					metric_key, metric_value, ts
				FROM nxd.telemetry_log
				WHERE asset_id = a.id
				ORDER BY metric_key, ts DESC
			) tl
		) lat ON true
		WHERE a.factory_id = $1
		ORDER BY a.display_name
	`, factoryID)
	if err != nil {
		http.Error(w, "Erro ao buscar ativos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var assets []AssetSummary
	onlineCount := 0
	now := time.Now()
	var lastUpdate *time.Time

	for rows.Next() {
		var a AssetSummary
		var groupID sql.NullString
		var lastSeen sql.NullTime
		var metricsJSON []byte
		if err := rows.Scan(&a.ID, &a.DisplayName, &a.SourceTagID, &groupID, &lastSeen, &metricsJSON); err != nil {
			continue
		}
		if groupID.Valid && groupID.String != "" {
			a.GroupID = &groupID.String
		}
		if lastSeen.Valid {
			a.LastSeen = &lastSeen.Time
			a.IsOnline = now.Sub(lastSeen.Time) < 60*time.Second
			// Track factory-wide last update
			if lastUpdate == nil || lastSeen.Time.After(*lastUpdate) {
				t := lastSeen.Time
				lastUpdate = &t
			}
		}
		if a.IsOnline {
			onlineCount++
		}
		if len(metricsJSON) > 0 {
			json.Unmarshal(metricsJSON, &a.Metrics)
		}
		if a.Metrics == nil {
			a.Metrics = map[string]interface{}{}
		}
		assets = append(assets, a)
	}

	if assets == nil {
		assets = []AssetSummary{}
	}

	respMap := map[string]interface{}{
		"factory_name":  factoryName,
		"factory_id":    factoryID,
		"assets":        assets,
		"total_assets":  len(assets),
		"online_assets": onlineCount,
		"last_update":   lastUpdate,
	}

	payload, _ := json.Marshal(respMap)

	// Store in cache
	dashboardCacheMu.Lock()
	dashboardCache[cacheKey] = dashboardCacheEntry{payload: payload, expiresAt: time.Now().Add(dashboardCacheTTL)}
	dashboardCacheMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(payload)
}

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

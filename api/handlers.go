package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"hubsystem/core"
	"hubsystem/internal/nxd/store"
	"hubsystem/services"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

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

// CreateFactoryHandler — POST /api/factory/create (rota pública, legado).
// O fluxo recomendado é: Register → Login → Onboarding (que cria fábrica e gera API key).
func CreateFactoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Use o fluxo: Registrar → Login → Onboarding para criar sua fábrica e obter a API Key.",
		"register": "/api/register",
		"login":    "/api/login",
		"onboarding": "POST /api/onboarding (com JWT)",
	})
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
		// Fábrica legada não encontrada — tenta buscar na NXD diretamente
		if nxdDB := store.NXDDB(); nxdDB != nil {
			var email string
			db.QueryRow("SELECT email FROM users WHERE id = $1", userID).Scan(&email)
			if email != "" {
				nxdDB.QueryRow(`
					SELECT f.name FROM nxd.factories f
					JOIN nxd.users u ON u.id = f.user_id
					WHERE u.email = $1 LIMIT 1
				`, email).Scan(&name)
			}
		}
	}

	// Se nome ainda está vazio, tenta buscar na nxd.factories via email
	if name == "" {
		if nxdDB := store.NXDDB(); nxdDB != nil {
			var email string
			db.QueryRow("SELECT email FROM users WHERE id = $1", userID).Scan(&email)
			if email != "" {
				nxdDB.QueryRow(`
					SELECT f.name FROM nxd.factories f
					JOIN nxd.users u ON u.id = f.user_id
					WHERE u.email = $1 LIMIT 1
				`, email).Scan(&name)
			}
		}
	}

	// Verifica se há api_key na NXD (mais confiável do que o legado)
	hasAPIKey := apiKeyHash != nil && *apiKeyHash != ""
	if !hasAPIKey {
		if nxdDB := store.NXDDB(); nxdDB != nil {
			var email string
			db.QueryRow("SELECT email FROM users WHERE id = $1", userID).Scan(&email)
			if email != "" {
				var nxdAPIKey sql.NullString
				nxdDB.QueryRow(`
					SELECT f.api_key FROM nxd.factories f
					JOIN nxd.users u ON u.id = f.user_id
					WHERE u.email = $1 LIMIT 1
				`, email).Scan(&nxdAPIKey)
				if nxdAPIKey.Valid && nxdAPIKey.String != "" {
					hasAPIKey = true
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":        name,
		"cnpj":        cnpj,
		"address":     address,
		"has_api_key": hasAPIKey,
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

	apiKeyPrefix := ""
	if len(apiKey) >= 16 {
		apiKeyPrefix = apiKey[:16]
	}

	// Atualiza no legado
	db.Exec("UPDATE factories SET api_key_hash = $1 WHERE user_id = $2", string(apiKeyHash), userID)

	// Atualiza na nxd.factories com api_key, hash e prefix para auth funcionar corretamente
	if nxdDB := store.NXDDB(); nxdDB != nil {
		var email string
		db.QueryRow("SELECT email FROM users WHERE id = $1", userID).Scan(&email)
		if email != "" {
			// Busca factory NXD via email e atualiza
			nxdDB.Exec(`
				UPDATE nxd.factories f
				SET api_key_hash = $1, api_key = $2, api_key_prefix = $3, updated_at = NOW()
				FROM nxd.users u
				WHERE u.email = $4 AND f.user_id = u.id
			`, string(apiKeyHash), apiKey, apiKeyPrefix, email)
		}
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
	LogAudit(userID, "create", "sector", sectorID, "", req.Name, ClientIP(r))
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
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	sectorID := vars["id"]
	if sectorID == "" {
		http.Error(w, "ID do setor obrigatório", http.StatusBadRequest)
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
	if nxdDB == nil {
		http.Error(w, "Banco NXD não disponível", http.StatusInternalServerError)
		return
	}
	res, err := nxdDB.Exec(`
		UPDATE nxd.sectors SET name = $1, description = $2
		WHERE id = $3 AND factory_id = $4
	`, req.Name, req.Description, sectorID, factoryID)
	if err != nil {
		log.Printf("[UpdateSector] Erro: %v", err)
		http.Error(w, "Erro ao atualizar setor", http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		http.Error(w, "Setor não encontrado ou não autorizado", http.StatusNotFound)
		return
	}
	LogAudit(userID, "update", "sector", sectorID, "", req.Name, ClientIP(r))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          sectorID,
		"name":        req.Name,
		"description": req.Description,
	})
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
	LogAudit(userID, "delete", "sector", sectorID, "", "", ClientIP(r))
	w.WriteHeader(http.StatusNoContent)
}

// UpdateMachineAssetHandler — PUT /api/machine/asset
// Atualiza display_name, description e/ou group_id (setor) de um asset no banco NXD.
// Aceita tanto "sector_id" quanto "group_id" para compatibilidade com o frontend.
func UpdateMachineAssetHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	var req struct {
		MachineID   string `json:"machine_id"`
		SectorID    string `json:"sector_id"`    // UUID string ou "" para remover do setor
		GroupID     string `json:"group_id"`     // alias para sector_id
		DisplayName string `json:"display_name"` // nome exibido no dashboard
		Description string `json:"description"`  // descrição do ativo
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.MachineID == "" {
		http.Error(w, "machine_id obrigatório", http.StatusBadRequest)
		return
	}
	// Normaliza: group_id tem precedência sobre sector_id se ambos forem enviados
	if req.GroupID != "" && req.SectorID == "" {
		req.SectorID = req.GroupID
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

	// Constrói UPDATE dinâmico apenas para os campos preenchidos
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.DisplayName != "" {
		setClauses = append(setClauses, fmt.Sprintf("display_name = $%d", argIdx))
		args = append(args, req.DisplayName)
		argIdx++
	}
	if req.Description != "" {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, req.Description)
		argIdx++
	}

	// Setor: "" ou "0" = remover (NULL), UUID válido = atribuir
	if req.SectorID == "" || req.SectorID == "0" {
		setClauses = append(setClauses, fmt.Sprintf("group_id = $%d", argIdx))
		args = append(args, nil)
		argIdx++
	} else {
		sectorUUID, parseErr := uuid.Parse(req.SectorID)
		if parseErr != nil {
			http.Error(w, "sector_id inválido", http.StatusBadRequest)
			return
		}
		setClauses = append(setClauses, fmt.Sprintf("group_id = $%d", argIdx))
		args = append(args, sectorUUID)
		argIdx++
	}

	if len(setClauses) == 0 {
		http.Error(w, "Nenhum campo para atualizar", http.StatusBadRequest)
		return
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now())
	argIdx++

	query := fmt.Sprintf("UPDATE nxd.assets SET %s WHERE id = $%d AND factory_id = $%d",
		strings.Join(setClauses, ", "), argIdx, argIdx+1)
	args = append(args, req.MachineID, factoryID)

	if _, err = nxdDB.Exec(query, args...); err != nil {
		log.Printf("[UpdateMachineAsset] Erro ao atualizar asset: %v", err)
		http.Error(w, "Erro ao atualizar ativo", http.StatusInternalServerError)
		return
	}
	LogAudit(userID, "update", "asset", req.MachineID, "", req.DisplayName, ClientIP(r))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"machine_id":   req.MachineID,
		"sector_id":    req.SectorID,
		"display_name": req.DisplayName,
		"description":  req.Description,
	})
}
// GetDashboardHandler — GET /api/dashboard?api_key=XXX
// Dashboard via API key (para o DX / dispositivos externos sem JWT).
// Retorna resumo da fábrica: ativos, status online/offline, últimas métricas.
func GetDashboardHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := r.URL.Query().Get("api_key")
	if apiKey == "" {
		http.Error(w, "api_key obrigatório", http.StatusUnauthorized)
		return
	}

	nxdDB := store.NXDDB()
	if nxdDB == nil {
		http.Error(w, "Banco NXD não disponível", http.StatusInternalServerError)
		return
	}

	factory, err := store.GetFactoryByAPIKey(nxdDB, apiKey)
	if err != nil || factory == nil {
		http.Error(w, "API key inválida", http.StatusUnauthorized)
		return
	}

	factoryUUID, _ := uuid.Parse(factory.ID)

	// Busca assets com últimas métricas
	rows, err := nxdDB.Query(`
		SELECT a.id, a.display_name, a.source_tag_id, a.group_id,
		       MAX(tl.ts) as last_seen,
		       json_object_agg(tl.metric_key, tl.metric_value) FILTER (WHERE tl.metric_key IS NOT NULL) as metrics
		FROM nxd.assets a
		LEFT JOIN LATERAL (
			SELECT DISTINCT ON (metric_key) metric_key, metric_value, ts
			FROM nxd.telemetry_log
			WHERE asset_id = a.id
			ORDER BY metric_key, ts DESC
		) tl ON true
		WHERE a.factory_id = $1
		GROUP BY a.id, a.display_name, a.source_tag_id, a.group_id
		ORDER BY a.display_name
	`, factoryUUID)
	if err != nil {
		log.Printf("[GetDashboardHandler] Erro ao buscar assets: %v", err)
		http.Error(w, "Erro ao buscar dados", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type AssetEntry struct {
		ID          string                 `json:"id"`
		DisplayName string                 `json:"display_name"`
		SourceTagID string                 `json:"source_tag_id"`
		GroupID     *string                `json:"group_id"`
		LastSeen    *time.Time             `json:"last_seen"`
		IsOnline    bool                   `json:"is_online"`
		Metrics     map[string]interface{} `json:"metrics"`
	}

	var assets []AssetEntry
	onlineCount := 0
	now := time.Now()
	var lastUpdate *time.Time

	for rows.Next() {
		var a AssetEntry
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
		assets = []AssetEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"factory_name":  factory.Name,
		"factory_id":    factory.ID,
		"assets":        assets,
		"total_assets":  len(assets),
		"online_assets": onlineCount,
		"last_update":   lastUpdate,
	})
}

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

// DashboardSummaryHandler — GET /api/dashboard/summary?api_key=XXX
// Retorna resumo compacto da fábrica via API key (sem JWT).
// Nota: a rota está registrada fora do authRouter, por isso não usa JWT.
func DashboardSummaryHandler(w http.ResponseWriter, r *http.Request) {
	// Tenta autenticar via API key (rota não-JWT)
	apiKey := r.URL.Query().Get("api_key")
	if apiKey != "" {
		GetDashboardHandler(w, r)
		return
	}
	// Fallback: sem api_key, retorna resumo básico do sistema
	nxdDB := store.NXDDB()
	var totalAssets, onlineAssets int
	if nxdDB != nil {
		nxdDB.QueryRow(`SELECT COUNT(*) FROM nxd.assets`).Scan(&totalAssets)
		nxdDB.QueryRow(`
			SELECT COUNT(DISTINCT asset_id) FROM nxd.telemetry_log
			WHERE ts >= NOW() - INTERVAL '60 seconds'
		`).Scan(&onlineAssets)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_assets":  totalAssets,
		"online_assets": onlineAssets,
		"timestamp":     time.Now().Format(time.RFC3339),
	})
}

// ReportIAHandler — POST /api/report/ia  or  GET /api/ia/analysis
// Gera um relatório completo de análise operacional da fábrica (ou setor)
// usando os dados reais de telemetria do nxd.telemetry_log + Gemini 2.0 Flash.
//
// Query params opcionais:
//   sector_id=<uuid>   — filtra por setor específico
//   period=<minutes>   — janela de dados (padrão: 60 min)
func ReportIAHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	sectorID := r.URL.Query().Get("sector_id")

	// buildTelemetryContext já faz toda a consulta real ao nxd.telemetry_log
	// com PostgreSQL $1,$2 placeholders e suporte a filtro de setor.
	telemetryCtx, err := buildTelemetryContext(userID, sectorID)
	if err != nil {
		log.Printf("[ReportIA] Erro ao montar contexto de telemetria: %v", err)
		telemetryCtx = "Sem dados de telemetria disponíveis no momento."
	}

	// Prompt especializado para relatório completo (diferente do chat interativo)
	reportPrompt := `Gere um RELATÓRIO COMPLETO DE ANÁLISE OPERACIONAL da fábrica com base nos dados de telemetria fornecidos.

O relatório deve conter:
1. **RESUMO EXECUTIVO** — Status geral da operação (1 parágrafo)
2. **ATIVOS MONITORADOS** — Lista de CLPs com status (ONLINE/OFFLINE) e última leitura
3. **MÉTRICAS CRÍTICAS** — Valores mais importantes: temperatura, pressão, velocidade, códigos de falha
4. **ANOMALIAS DETECTADAS** — Qualquer valor fora do padrão (Fault_Code > 0, Health_Score < 0.8, etc.)
5. **RECOMENDAÇÕES** — Ações sugeridas baseadas nos dados
6. **PRÓXIMA REVISÃO** — Quando verificar novamente

Use formatação Markdown. Seja objetivo e técnico. Responda em português brasileiro.`

	ctx := r.Context()
	analysis, err := callGemini(ctx, reportPrompt, telemetryCtx)
	if err != nil {
		log.Printf("[ReportIA] Erro Gemini: %v", err)
		http.Error(w, "Erro ao gerar relatório de IA: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Conta ativos e registros para metadata do relatório
	nxdDB := store.NXDDB()
	var assetCount, telemetryCount int
	var factoryID uuid.UUID
	if nxdDB != nil {
		var ferr error
		factoryID, ferr = getFactoryIDForUser(userID)
		if ferr == nil {
			nxdDB.QueryRow(`SELECT COUNT(*) FROM nxd.assets WHERE factory_id = $1`, factoryID).Scan(&assetCount)
			nxdDB.QueryRow(`SELECT COUNT(*) FROM nxd.telemetry_log WHERE factory_id = $1 AND ts >= NOW() - INTERVAL '1 hour'`, factoryID).Scan(&telemetryCount)
		}
	}
	sources := fmt.Sprintf("Relatório executivo. Setor: %s. Ativos: %d. Registros (1h): %d.", sectorID, assetCount, telemetryCount)
	if sectorID == "" {
		sources = fmt.Sprintf("Relatório executivo. Ativos: %d. Registros de telemetria (1h): %d.", assetCount, telemetryCount)
	}
	if db := GetDB(); db != nil {
		title := "Relatório executivo"
		saveIAReport(db, userID, factoryID.String(), title, analysis, sources)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"generated_at":       time.Now().Format(time.RFC3339),
		"sector_id":          sectorID,
		"assets_count":       assetCount,
		"telemetry_records":  telemetryCount,
		"analysis":           analysis,
		"sources":            sources,
	})
}
// AnalyticsHandler — GET /api/analytics?api_key=XXX
// Retorna métricas agregadas por ativo: min, max, avg, contagem de leituras.
// Suporta parâmetro opcional: period=<minutes> (padrão: 60).
func AnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := r.URL.Query().Get("api_key")
	if apiKey == "" {
		http.Error(w, "api_key obrigatório", http.StatusUnauthorized)
		return
	}

	nxdDB := store.NXDDB()
	if nxdDB == nil {
		http.Error(w, "Banco NXD não disponível", http.StatusInternalServerError)
		return
	}

	factory, err := store.GetFactoryByAPIKey(nxdDB, apiKey)
	if err != nil || factory == nil {
		http.Error(w, "API key inválida", http.StatusUnauthorized)
		return
	}
	factoryUUID, _ := uuid.Parse(factory.ID)

	// Período de análise — padrão 60 min
	periodMinutes := 60
	if p := r.URL.Query().Get("period"); p != "" {
		if v, err := fmt.Sscanf(p, "%d", &periodMinutes); v != 1 || err != nil || periodMinutes < 1 {
			periodMinutes = 60
		}
	}

	rows, err := nxdDB.Query(`
		SELECT
			a.id,
			a.display_name,
			a.source_tag_id,
			tl.metric_key,
			MIN(tl.metric_value) as min_val,
			MAX(tl.metric_value) as max_val,
			AVG(tl.metric_value) as avg_val,
			COUNT(*) as samples,
			MAX(tl.ts) as last_ts
		FROM nxd.telemetry_log tl
		JOIN nxd.assets a ON a.id = tl.asset_id
		WHERE tl.factory_id = $1
		  AND tl.ts >= NOW() - ($2 || ' minutes')::INTERVAL
		GROUP BY a.id, a.display_name, a.source_tag_id, tl.metric_key
		ORDER BY a.display_name, tl.metric_key
	`, factoryUUID, periodMinutes)
	if err != nil {
		log.Printf("[Analytics] Erro ao buscar métricas: %v", err)
		http.Error(w, "Erro ao buscar analytics", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type MetricSummary struct {
		MetricKey string    `json:"metric_key"`
		Min       float64   `json:"min"`
		Max       float64   `json:"max"`
		Avg       float64   `json:"avg"`
		Samples   int       `json:"samples"`
		LastTs    time.Time `json:"last_ts"`
	}
	type AssetAnalytics struct {
		AssetID     string          `json:"asset_id"`
		DisplayName string          `json:"display_name"`
		SourceTagID string          `json:"source_tag_id"`
		Metrics     []MetricSummary `json:"metrics"`
	}

	assetMap := map[string]*AssetAnalytics{}
	assetOrder := []string{}

	for rows.Next() {
		var assetID, displayName, sourceTagID, metricKey string
		var minVal, maxVal, avgVal float64
		var samples int
		var lastTs time.Time
		if err := rows.Scan(&assetID, &displayName, &sourceTagID, &metricKey, &minVal, &maxVal, &avgVal, &samples, &lastTs); err != nil {
			continue
		}
		if _, exists := assetMap[assetID]; !exists {
			assetMap[assetID] = &AssetAnalytics{
				AssetID:     assetID,
				DisplayName: displayName,
				SourceTagID: sourceTagID,
				Metrics:     []MetricSummary{},
			}
			assetOrder = append(assetOrder, assetID)
		}
		assetMap[assetID].Metrics = append(assetMap[assetID].Metrics, MetricSummary{
			MetricKey: metricKey,
			Min:       minVal,
			Max:       maxVal,
			Avg:       avgVal,
			Samples:   samples,
			LastTs:    lastTs,
		})
	}

	result := make([]*AssetAnalytics, 0, len(assetOrder))
	for _, id := range assetOrder {
		result = append(result, assetMap[id])
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"factory_id":     factory.ID,
		"period_minutes": periodMinutes,
		"assets":         result,
		"generated_at":   time.Now().Format(time.RFC3339),
	})
}

// DeleteMachineHandler — DELETE /api/machine/delete
// Remove um asset (máquina) da fábrica do usuário autenticado.
// Requer JWT. Body: {"machine_id":"<uuid>"}
func DeleteMachineHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	var req struct {
		MachineID string `json:"machine_id"`
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

	res, err := nxdDB.Exec(`DELETE FROM nxd.assets WHERE id = $1 AND factory_id = $2`, req.MachineID, factoryID)
	if err != nil {
		log.Printf("[DeleteMachine] Erro: %v", err)
		http.Error(w, "Erro ao deletar ativo", http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		http.Error(w, "Ativo não encontrado ou não autorizado", http.StatusNotFound)
		return
	}
	LogAudit(userID, "delete", "asset", req.MachineID, "", "", ClientIP(r))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"machine_id": req.MachineID,
		"message":    "Ativo removido com sucesso",
	})
}

// HealthStatusHandler — GET /api/connection/status?api_key=XXX
// Retorna status de conexão de cada ativo da fábrica (online/offline).
// Um ativo é considerado ONLINE se enviou dados nos últimos 60 segundos.
func HealthStatusHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := r.URL.Query().Get("api_key")
	if apiKey == "" {
		http.Error(w, "api_key obrigatório", http.StatusUnauthorized)
		return
	}

	nxdDB := store.NXDDB()
	if nxdDB == nil {
		http.Error(w, "Banco NXD não disponível", http.StatusInternalServerError)
		return
	}

	factory, err := store.GetFactoryByAPIKey(nxdDB, apiKey)
	if err != nil || factory == nil {
		http.Error(w, "API key inválida", http.StatusUnauthorized)
		return
	}
	factoryUUID, _ := uuid.Parse(factory.ID)

	rows, err := nxdDB.Query(`
		SELECT
			a.id,
			a.display_name,
			a.source_tag_id,
			MAX(tl.ts) as last_seen
		FROM nxd.assets a
		LEFT JOIN nxd.telemetry_log tl ON tl.asset_id = a.id
		WHERE a.factory_id = $1
		GROUP BY a.id, a.display_name, a.source_tag_id
		ORDER BY a.display_name
	`, factoryUUID)
	if err != nil {
		log.Printf("[HealthStatus] Erro: %v", err)
		http.Error(w, "Erro ao buscar status", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type AssetStatus struct {
		AssetID     string     `json:"asset_id"`
		DisplayName string     `json:"display_name"`
		SourceTagID string     `json:"source_tag_id"`
		IsOnline    bool       `json:"is_online"`
		LastSeen    *time.Time `json:"last_seen"`
		Status      string     `json:"status"`
	}

	var assets []AssetStatus
	onlineCount := 0
	now := time.Now()

	for rows.Next() {
		var a AssetStatus
		var lastSeen sql.NullTime
		if err := rows.Scan(&a.AssetID, &a.DisplayName, &a.SourceTagID, &lastSeen); err != nil {
			continue
		}
		if lastSeen.Valid {
			a.LastSeen = &lastSeen.Time
			a.IsOnline = now.Sub(lastSeen.Time) < 60*time.Second
		}
		if a.IsOnline {
			a.Status = "online"
			onlineCount++
		} else {
			a.Status = "offline"
		}
		assets = append(assets, a)
	}

	if assets == nil {
		assets = []AssetStatus{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"factory_id":    factory.ID,
		"factory_name":  factory.Name,
		"assets":        assets,
		"total_assets":  len(assets),
		"online_assets": onlineCount,
		"offline_assets": len(assets) - onlineCount,
		"timestamp":     time.Now().Format(time.RFC3339),
	})
}

// ConnectionLogsHandler — GET /api/connection/logs?api_key=XXX
// Retorna logs de auditoria/conexão da fábrica (ingest, erros, etc.).
// Suporta parâmetros: limit=<int> (padrão: 50), level=error|all
func ConnectionLogsHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := r.URL.Query().Get("api_key")
	if apiKey == "" {
		http.Error(w, "api_key obrigatório", http.StatusUnauthorized)
		return
	}

	nxdDB := store.NXDDB()
	if nxdDB == nil {
		http.Error(w, "Banco NXD não disponível", http.StatusInternalServerError)
		return
	}

	factory, err := store.GetFactoryByAPIKey(nxdDB, apiKey)
	if err != nil || factory == nil {
		http.Error(w, "API key inválida", http.StatusUnauthorized)
		return
	}
	factoryUUID, _ := uuid.Parse(factory.ID)

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := fmt.Sscanf(l, "%d", &limit); v != 1 || err != nil || limit < 1 || limit > 500 {
			limit = 50
		}
	}

	// Filtro de nível
	levelFilter := r.URL.Query().Get("level")
	statusFilter := ""
	if levelFilter == "error" {
		statusFilter = " AND status = 'error'"
	}

	rows, err := nxdDB.Query(fmt.Sprintf(`
		SELECT ts, action, COALESCE(device_id,''), COALESCE(status,''), COALESCE(message,''), COALESCE(ip_address,'')
		FROM nxd.audit_log
		WHERE factory_id = $1%s
		ORDER BY ts DESC
		LIMIT $2
	`, statusFilter), factoryUUID, limit)
	if err != nil {
		log.Printf("[ConnectionLogs] Erro: %v", err)
		http.Error(w, "Erro ao buscar logs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type LogEntry struct {
		Timestamp time.Time `json:"ts"`
		Action    string    `json:"action"`
		DeviceID  string    `json:"device_id"`
		Status    string    `json:"status"`
		Message   string    `json:"message"`
		IPAddress string    `json:"ip_address"`
	}

	var logs []LogEntry
	for rows.Next() {
		var e LogEntry
		if err := rows.Scan(&e.Timestamp, &e.Action, &e.DeviceID, &e.Status, &e.Message, &e.IPAddress); err != nil {
			continue
		}
		logs = append(logs, e)
	}

	if logs == nil {
		logs = []LogEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"factory_id": factory.ID,
		"logs":       logs,
		"count":      len(logs),
		"limit":      limit,
	})
}

// HealthHandler — GET /api/health
// Retorna status do sistema com conectividade real ao banco.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	nxdOK := false
	apiDBOK := false

	if nxdDB := store.NXDDB(); nxdDB != nil {
		if err := nxdDB.Ping(); err == nil {
			nxdOK = true
		}
	}
	if db != nil {
		if err := db.Ping(); err == nil {
			apiDBOK = true
		}
	}

	overall := "ok"
	statusCode := http.StatusOK
	if !nxdOK {
		overall = "degraded"
		statusCode = http.StatusServiceUnavailable
	}

	buildVersion := os.Getenv("BUILD_VERSION")
	if buildVersion == "" {
		buildVersion = "dev"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    overall,
		"timestamp": time.Now().Format(time.RFC3339),
		"nxd_db":   nxdOK,
		"api_db":   apiDBOK,
		"version":  buildVersion,
	})
}

// SystemParamsHandler — GET /api/system
// Retorna parâmetros de configuração do sistema visíveis ao frontend.
func SystemParamsHandler(w http.ResponseWriter, r *http.Request) {
	nxdDB := store.NXDDB()

	// Conta ativos e leituras recentes
	var assetCount int
	var recentReadings int
	var iaEnabled bool = true

	if nxdDB != nil {
		nxdDB.QueryRow(`SELECT COUNT(*) FROM nxd.assets`).Scan(&assetCount)
		nxdDB.QueryRow(`SELECT COUNT(*) FROM nxd.telemetry_log WHERE ts >= NOW() - INTERVAL '5 minutes'`).Scan(&recentReadings)
	}

	buildVersion := os.Getenv("BUILD_VERSION")
	if buildVersion == "" {
		buildVersion = "dev"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ia_operando_100": iaEnabled,
		"total_assets":    assetCount,
		"recent_readings": recentReadings,
		"version":         buildVersion,
		"features": map[string]bool{
			"import_jobs":       true,
			"ia_chat":           true,
			"ia_report":         true,
			"historical_import": true,
			"totp_2fa":          true,
		},
	})
}

package api

import (
	"encoding/json"
	"hubsystem/internal/nxd/store"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ─── Business config (parâmetros financeiros por setor) ───────────────────

// ListBusinessConfigHandler — GET /api/business-config
func ListBusinessConfigHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	factoryID, err := getFactoryIDForUser(userID)
	if err != nil || factoryID == uuid.Nil {
		http.Error(w, "Fábrica não encontrada", http.StatusNotFound)
		return
	}
	nxdDB := store.NXDDB()
	if nxdDB == nil {
		http.Error(w, "Banco NXD indisponível", http.StatusServiceUnavailable)
		return
	}
	list, err := store.ListBusinessConfigs(nxdDB, factoryID)
	if err != nil {
		log.Printf("[BusinessConfig] List: %v", err)
		http.Error(w, "Erro ao listar configurações", http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []store.BusinessConfigRow{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"configs": list})
}

// UpsertBusinessConfigHandler — POST /api/business-config
// Body: { "sector_id": "uuid|null", "valor_venda_ok": 10.5, "custo_refugo_un": 2, "custo_parada_h": 150 }
func UpsertBusinessConfigHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	factoryID, err := getFactoryIDForUser(userID)
	if err != nil || factoryID == uuid.Nil {
		http.Error(w, "Fábrica não encontrada", http.StatusNotFound)
		return
	}
	nxdDB := store.NXDDB()
	if nxdDB == nil {
		http.Error(w, "Banco NXD indisponível", http.StatusServiceUnavailable)
		return
	}
	var body struct {
		SectorID      *string  `json:"sector_id"`
		ValorVendaOk  float64  `json:"valor_venda_ok"`
		CustoRefugoUn float64  `json:"custo_refugo_un"`
		CustoParadaH  float64  `json:"custo_parada_h"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}
	var sectorID *uuid.UUID
	if body.SectorID != nil && *body.SectorID != "" {
		u, err := uuid.Parse(*body.SectorID)
		if err != nil {
			http.Error(w, "sector_id inválido", http.StatusBadRequest)
			return
		}
		sectorID = &u
	}
	_, err = store.UpsertBusinessConfig(nxdDB, factoryID, sectorID, body.ValorVendaOk, body.CustoRefugoUn, body.CustoParadaH)
	if err != nil {
		log.Printf("[BusinessConfig] Upsert: %v", err)
		http.Error(w, "Erro ao salvar configuração", http.StatusInternalServerError)
		return
	}
	entityID := "default"
	if sectorID != nil {
		entityID = sectorID.String()
	}
	LogAudit(userID, "upsert", "business_config", entityID, "", "ok", ClientIP(r))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ─── Tag mapping (mapeamento de tags por ativo) ─────────────────────────────

// ListTagMappingsHandler — GET /api/tag-mappings
func ListTagMappingsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	factoryID, err := getFactoryIDForUser(userID)
	if err != nil || factoryID == uuid.Nil {
		http.Error(w, "Fábrica não encontrada", http.StatusNotFound)
		return
	}
	nxdDB := store.NXDDB()
	if nxdDB == nil {
		http.Error(w, "Banco NXD indisponível", http.StatusServiceUnavailable)
		return
	}
	list, err := store.ListTagMappingsByFactory(nxdDB, factoryID)
	if err != nil {
		log.Printf("[TagMapping] List: %v", err)
		http.Error(w, "Erro ao listar mapeamentos", http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []store.TagMappingRow{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"mappings": list})
}

// UpsertTagMappingHandler — POST /api/tag-mappings
// Body: { "asset_id": "uuid", "tag_ok": "Total_Pecas", "tag_nok": "Refugo", "tag_status": "Running", "reading_rule": "delta" }
func UpsertTagMappingHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	factoryID, err := getFactoryIDForUser(userID)
	if err != nil || factoryID == uuid.Nil {
		http.Error(w, "Fábrica não encontrada", http.StatusNotFound)
		return
	}
	nxdDB := store.NXDDB()
	if nxdDB == nil {
		http.Error(w, "Banco NXD indisponível", http.StatusServiceUnavailable)
		return
	}
	var body struct {
		AssetID     string `json:"asset_id"`
		TagOK       string `json:"tag_ok"`
		TagNOK      string `json:"tag_nok"`
		TagStatus   string `json:"tag_status"`
		ReadingRule string `json:"reading_rule"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}
	if body.AssetID == "" {
		http.Error(w, "asset_id obrigatório", http.StatusBadRequest)
		return
	}
	assetID, err := uuid.Parse(body.AssetID)
	if err != nil {
		http.Error(w, "asset_id inválido", http.StatusBadRequest)
		return
	}
	_, err = store.UpsertTagMapping(nxdDB, assetID, body.TagOK, body.TagNOK, body.TagStatus, body.ReadingRule)
	if err != nil {
		log.Printf("[TagMapping] Upsert: %v", err)
		http.Error(w, "Erro ao salvar mapeamento", http.StatusInternalServerError)
		return
	}
	LogAudit(userID, "upsert", "tag_mapping", body.AssetID, "", body.TagOK+","+body.TagNOK, ClientIP(r))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ─── Resumo financeiro (agregado por período) ──────────────────────────────

// GetFinancialSummaryHandler — GET /api/financial-summary?period=24h|7d|1h&sector_id=uuid
func GetFinancialSummaryHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	factoryID, err := getFactoryIDForUser(userID)
	if err != nil || factoryID == uuid.Nil {
		http.Error(w, "Fábrica não encontrada", http.StatusNotFound)
		return
	}
	nxdDB := store.NXDDB()
	if nxdDB == nil {
		http.Error(w, "Banco NXD indisponível", http.StatusServiceUnavailable)
		return
	}
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "24h"
	}
	sectorIDStr := r.URL.Query().Get("sector_id")
	var sectorID *uuid.UUID
	if sectorIDStr != "" {
		u, err := uuid.Parse(sectorIDStr)
		if err != nil {
			http.Error(w, "sector_id inválido", http.StatusBadRequest)
			return
		}
		sectorID = &u
	}

	now := time.Now()
	var start time.Time
	switch period {
	case "1h":
		start = now.Add(-1 * time.Hour)
	case "7d":
		start = now.Add(-7 * 24 * time.Hour)
	default:
		start = now.Add(-24 * time.Hour)
	}

	res, breakdown, err := store.ComputeFinancialAggregate(nxdDB, factoryID, sectorID, start, now)
	if err != nil {
		log.Printf("[FinancialSummary] %v", err)
		http.Error(w, "Erro ao calcular resumo financeiro", http.StatusInternalServerError)
		return
	}
	if res == nil {
		res = &store.FinancialAggregateResult{PeriodStart: start, PeriodEnd: now}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"summary":   res,
		"breakdown": breakdown,
		"period":    period,
	})
}

// GetFinancialSummaryRangesHandler — GET /api/financial-summary/ranges
// Retorna hoje, 24h e 7d em uma única chamada (para cards na UI).
func GetFinancialSummaryRangesHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	factoryID, err := getFactoryIDForUser(userID)
	if err != nil || factoryID == uuid.Nil {
		http.Error(w, "Fábrica não encontrada", http.StatusNotFound)
		return
	}
	nxdDB := store.NXDDB()
	if nxdDB == nil {
		http.Error(w, "Banco NXD indisponível", http.StatusServiceUnavailable)
		return
	}
	sectorIDStr := r.URL.Query().Get("sector_id")
	var sectorID *uuid.UUID
	if sectorIDStr != "" {
		u, _ := uuid.Parse(sectorIDStr)
		sectorID = &u
	}
	now := time.Now()
	periods := []struct {
		Key   string
		Start time.Time
	}{
		{"today", time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())},
		{"24h", now.Add(-24 * time.Hour)},
		{"7d", now.Add(-7 * 24 * time.Hour)},
		{"30d", now.Add(-30 * 24 * time.Hour)},
	}
	out := make(map[string]interface{})
	for _, p := range periods {
		res, _, err := store.ComputeFinancialAggregate(nxdDB, factoryID, sectorID, p.Start, now)
		if err != nil {
			out[p.Key] = map[string]string{"error": err.Error()}
			continue
		}
		out[p.Key] = res
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// GetFinancialExecutiveExportHandler — GET /api/financial-summary/export?period=7d|30d&sector_id=uuid
// Retorna resumo para exportação CSV/PDF: período atual, período anterior (comparativo), perdas evitadas, custo parada evitado.
func GetFinancialExecutiveExportHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	factoryID, err := getFactoryIDForUser(userID)
	if err != nil || factoryID == uuid.Nil {
		http.Error(w, "Fábrica não encontrada", http.StatusNotFound)
		return
	}
	nxdDB := store.NXDDB()
	if nxdDB == nil {
		http.Error(w, "Banco NXD indisponível", http.StatusServiceUnavailable)
		return
	}
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "7d"
	}
	sectorIDStr := r.URL.Query().Get("sector_id")
	var sectorID *uuid.UUID
	if sectorIDStr != "" {
		u, _ := uuid.Parse(sectorIDStr)
		sectorID = &u
	}
	now := time.Now()
	var startCurrent, startPrevious time.Time
	switch period {
	case "30d":
		startCurrent = now.Add(-30 * 24 * time.Hour)
		startPrevious = now.Add(-60 * 24 * time.Hour)
	case "24h":
		startCurrent = now.Add(-24 * time.Hour)
		startPrevious = now.Add(-48 * time.Hour)
	default:
		startCurrent = now.Add(-7 * 24 * time.Hour)
		startPrevious = now.Add(-14 * 24 * time.Hour)
	}
	resCurrent, _, err := store.ComputeFinancialAggregate(nxdDB, factoryID, sectorID, startCurrent, now)
	if err != nil {
		log.Printf("[FinancialExport] current: %v", err)
		http.Error(w, "Erro ao calcular resumo", http.StatusInternalServerError)
		return
	}
	resPrevious, _, err := store.ComputeFinancialAggregate(nxdDB, factoryID, sectorID, startPrevious, startCurrent)
	if err != nil {
		resPrevious = nil
	}
	perdasEvitadas := 0.0
	custoParadaEvitado := 0.0
	if resPrevious != nil {
		if resCurrent.PerdaRefugo < resPrevious.PerdaRefugo {
			perdasEvitadas = resPrevious.PerdaRefugo - resCurrent.PerdaRefugo
		}
		if resCurrent.CustoParada < resPrevious.CustoParada {
			custoParadaEvitado = resPrevious.CustoParada - resCurrent.CustoParada
		}
	}
	out := map[string]interface{}{
		"period":                period,
		"period_start":          startCurrent,
		"period_end":            now,
		"faturamento_bruto":      resCurrent.FaturamentoBruto,
		"perda_refugo":          resCurrent.PerdaRefugo,
		"custo_parada":          resCurrent.CustoParada,
		"perdas_evitadas":       perdasEvitadas,
		"custo_parada_evitado":  custoParadaEvitado,
		"sector_id":             sectorIDStr,
		"previous_period":       resPrevious != nil,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

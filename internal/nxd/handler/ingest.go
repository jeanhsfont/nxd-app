package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"hubsystem/internal/nxd/store"
)

// TelemetryIngest handles POST /nxd/telemetry/ingest. Authenticates by gateway_key (hash lookup).
func TelemetryIngest(w http.ResponseWriter, r *http.Request) {
	db := store.NXDDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "NXD não configurado"})
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload store.TelemetryIngestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "payload inválido"})
		return
	}
	if payload.GatewayKey == "" || payload.SourceTagID == "" || len(payload.Metrics) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "gateway_key, source_tag_id e metrics obrigatórios"})
		return
	}
	hash := sha256.Sum256([]byte(payload.GatewayKey))
	hashStr := hex.EncodeToString(hash[:])
	factory, err := store.GetFactoryByGatewayKeyHash(db, hashStr)
	if err != nil || factory == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "gateway não autorizado"})
		return
	}
	correlationID := uuid.New().String()
	ts := payload.Ts
	if ts.IsZero() {
		ts = time.Now()
	}
	assetID, err := store.CreateAsset(db, factory.ID, nil, payload.SourceTagID, "", "", nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "erro ao obter/criar ativo"})
		return
	}
	var rows []store.TelemetryRow
	for _, m := range payload.Metrics {
		if m.MetricKey == "" {
			continue
		}
		status := m.Status
		if status == "" {
			status = "OK"
		}
		rows = append(rows, store.TelemetryRow{
			Ts:          ts,
			MetricKey:   m.MetricKey,
			MetricValue: m.Value,
			Status:      status,
			Raw:         nil,
		})
	}
	if err := store.InsertTelemetryBatch(db, factory.ID, assetID, correlationID, rows); err != nil {
		log.Printf("NXD ingest: insert telemetry: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "erro ao gravar telemetria"})
		return
	}
	for _, m := range payload.Metrics {
		_ = store.UpsertAssetMetricCatalog(db, factory.ID, assetID, m.MetricKey, ts)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "success",
		"correlation_id": correlationID,
		"asset_id":       assetID.String(),
	})
}

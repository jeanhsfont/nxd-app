package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"hubsystem/internal/nxd/middleware"
	"hubsystem/internal/nxd/store"

	"github.com/google/uuid"
)

// Health returns gateway/ingest health for the factory: last ts, status (online if last ts within 5 min).
func Health(w http.ResponseWriter, r *http.Request) {
	db := store.NXDDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "NXD não configurado"})
		return
	}
	factoryIDStr := r.URL.Query().Get("factory_id")
	if factoryIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "factory_id obrigatório"})
		return
	}
	factoryID, err := uuid.Parse(factoryIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "factory_id inválido"})
		return
	}
	// Optional: enforce RBAC so only members can see health
	_, _ = middleware.FactoryIDFromContext(r.Context())
	lastTs, err := store.LastTelemetryTs(db, factoryID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "erro ao buscar health"})
		return
	}
	status := "OFFLINE"
	if time.Since(lastTs) < 5*time.Minute {
		status = "ONLINE"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"factory_id": factoryIDStr,
		"status":     status,
		"last_ts":    lastTs.Format(time.RFC3339),
	})
}

package handler

import (
	"encoding/json"
	"net/http"

	"hubsystem/internal/nxd/middleware"
	"hubsystem/internal/nxd/store"
)

// AuditLog returns recent audit log entries for the factory.
func AuditLog(w http.ResponseWriter, r *http.Request) {
	db := store.NXDDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "NXD não configurado"})
		return
	}
	factoryID, ok := middleware.FactoryIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "factory_id obrigatório"})
		return
	}
	list, err := store.ListAuditLog(db, factoryID, 100)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "erro ao listar auditoria"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"audit": list})
}

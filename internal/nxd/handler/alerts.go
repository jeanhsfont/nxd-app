package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"hubsystem/internal/nxd/middleware"
	"hubsystem/internal/nxd/store"

	"github.com/google/uuid"
)

// ListAlertRules returns alert rules for the factory.
func ListAlertRules(w http.ResponseWriter, r *http.Request) {
	db := store.NXDDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	factoryID, ok := middleware.FactoryIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	list, err := store.ListAlertRules(db, factoryID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"rules": list})
}

// CreateAlertRuleRequest is the body for POST /nxd/alert-rules.
type CreateAlertRuleRequest struct {
	ScopeType    string  `json:"scope_type"`
	ScopeID      string  `json:"scope_id"`
	ConditionType string `json:"condition_type"`
	Threshold    float64 `json:"threshold"`
	Channel      string  `json:"channel"`
}

// CreateAlertRule creates a rule.
func CreateAlertRule(w http.ResponseWriter, r *http.Request) {
	db := store.NXDDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	factoryID, ok := middleware.FactoryIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var req CreateAlertRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.ScopeType == "" || req.ScopeID == "" || req.ConditionType == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "scope_type, scope_id, condition_type obrigatórios"})
		return
	}
	id, err := store.CreateAlertRule(db, factoryID, req.ScopeType, req.ScopeID, req.ConditionType, req.Threshold, req.Channel)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id.String()})
}

// ListAlerts returns alerts for the factory.
func ListAlerts(w http.ResponseWriter, r *http.Request) {
	db := store.NXDDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	factoryID, ok := middleware.FactoryIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	unackOnly := r.URL.Query().Get("unack_only") == "true"
	list, err := store.ListAlerts(db, factoryID, unackOnly)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"alerts": list})
}

// AckAlert acknowledges an alert.
func AckAlert(w http.ResponseWriter, r *http.Request) {
	db := store.NXDDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	factoryID, ok := middleware.FactoryIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	alertID, err := uuid.Parse(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userID := middleware.UserIDFromRequest(r)
	if err := store.AckAlert(db, alertID, factoryID, userID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
